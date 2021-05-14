package server

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	pb "github.com/briansp8210/simple-chat-app/protobuf"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/gomodule/redigo/redis"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func (s *chatServer) Register(ctx context.Context, in *pb.RegisterRequest) (*empty.Empty, error) {
	log.Printf("Receive Registration from user %s\n", in.Username)

	if _, err := s.db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", in.Username, in.Password); err != nil {
		switch pqErr := err.(*pq.Error); pqErr.Code.Name() {
		case "unique_violation":
			return nil, status.Errorf(codes.AlreadyExists, "Username %s is already used", in.Username)
		case "check_violation":
			return nil, status.Errorf(codes.InvalidArgument, "Username should be alphanumeric")
		default:
			log.Fatal(err)
		}
	}
	return &empty.Empty{}, nil
}

func (s *chatServer) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Printf("Receive Login from user %s\n", in.Username)

	var uid int32
	var correctPassword []byte
	if err := s.db.QueryRow("SELECT id, password FROM users WHERE username = $1", in.Username).Scan(&uid, &correctPassword); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "User %s not found", in.Username)
		}
		log.Fatal(err)
	}

	if !bytes.Equal(in.Password, correctPassword) {
		return nil, status.Error(codes.Unauthenticated, "Invalid password")
	}

	redisConn := s.redisPool.Get()
	defer redisConn.Close()
	if _, err := redisConn.Do("HSET", fmt.Sprintf("user:%d", uid), "serverId", s.serverId); err != nil {
		log.Fatal(err)
	}

	query := `SELECT conversations.id, conversations.name, conversations.type FROM participants
		INNER JOIN conversations ON participants.conversation_id = conversations.id
		WHERE participants.user_id = $1`
	rows, err := s.db.Query(query, uid)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	conversations := make([]*pb.Conversation, 0)
	for rows.Next() {
		c := &pb.Conversation{}
		if err := rows.Scan(&c.Id, &c.Name, &c.Type); err != nil {
			log.Fatal(err)
		}
		conversations = append(conversations, c)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return &pb.LoginResponse{UserId: uid, Conversations: conversations}, nil
}

func (s *chatServer) Logout(ctx context.Context, in *pb.LogoutRequest) (*empty.Empty, error) {
	log.Printf("Logout\n")

	redisConn := s.redisPool.Get()
	defer redisConn.Close()
	_, err := redis.Int(redisConn.Do("DEL", fmt.Sprintf("user:%d", in.UserId)))
	if err != nil {
		log.Fatal(err)
	}

	return &empty.Empty{}, nil
}

func (s *chatServer) AddConversation(ctx context.Context, in *pb.AddConversationRequest) (*pb.AddConversationResponse, error) {
	log.Printf("AddConversation\n")

	memberIds := make([]int32, 0, len(in.MemberNames))
	for _, memberName := range in.MemberNames {
		uid, err := s.getUserId(memberName)
		if err != nil {
			return nil, err
		}
		memberIds = append(memberIds, uid)
	}

	if err := s.db.QueryRow("INSERT INTO conversations (name, type) VALUES ($1, $2) RETURNING id", in.Conversation.Name, in.Conversation.Type).Scan(&in.Conversation.Id); err != nil {
		switch pqErr := err.(*pq.Error); pqErr.Code.Name() {
		case "unique_violation":
			return nil, status.Errorf(codes.AlreadyExists, "Conversation already exists")
		default:
			log.Fatal(err)
		}
	}

	for _, memberId := range memberIds {
		if _, err := s.db.Exec("INSERT INTO participants (user_id, conversation_id) VALUES ($1, $2)", memberId, in.Conversation.Id); err != nil {
			log.Fatal(err)
		}
	}

	redisConn := s.redisPool.Get()
	defer redisConn.Close()
	if _, err := redisConn.Do("SADD", redis.Args{}.Add(fmt.Sprintf("conversationMembers:%d", in.Conversation.Id)).AddFlat(memberIds)...); err != nil {
		log.Fatal(err)
	}

	return &pb.AddConversationResponse{Conversation: in.Conversation}, nil
}

func (s *chatServer) GetConversation(ctx context.Context, in *pb.GetConversationRequest) (*pb.GetConversationResponse, error) {
	log.Printf("GetConversation\n")

	conversation := &pb.Conversation{}
	row := s.db.QueryRow("SELECT * FROM conversations WHERE id = $1", in.ConversationId)
	if err := row.Scan(&conversation.Id, &conversation.Name, &conversation.Type); err != nil {
		log.Fatal(err)
	}

	return &pb.GetConversationResponse{Conversation: conversation}, nil
}

func (s *chatServer) JoinGroup(ctx context.Context, in *pb.JoinGroupRequest) (*pb.JoinGroupResponse, error) {
	log.Printf("JoinGroup\n")

	group := &pb.Conversation{}
	if err := s.db.QueryRow("SELECT * FROM conversations WHERE name = $1", in.GroupName).Scan(&group.Id, &group.Name, &group.Type); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "Gorup %s not found", in.GroupName)
		}
		log.Fatal(err)
	}

	if _, err := s.db.Exec("INSERT INTO participants (user_id, conversation_id) VALUES ($1, $2)", in.UserId, group.Id); err != nil {
		switch pqErr := err.(*pq.Error); pqErr.Code.Name() {
		case "unique_violation":
			return nil, status.Errorf(codes.AlreadyExists, "You have already joined %s", in.GroupName)
		default:
			log.Fatal(err)
		}
	}

	redisConn := s.redisPool.Get()
	defer redisConn.Close()
	if _, err := redisConn.Do("SADD", redis.Args{}.Add(fmt.Sprintf("conversationMembers:%d", group.Id)).AddFlat(in.UserId)...); err != nil {
		log.Fatal(err)
	}

	return &pb.JoinGroupResponse{Group: group}, nil
}

func (s *chatServer) GetMessages(ctx context.Context, in *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	log.Printf("GetMessages\n")

	rows, err := s.db.Query("SELECT * FROM messages WHERE conversation_id = $1 ORDER BY ts ASC", in.ConversationId)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	rsp := &pb.GetMessagesResponse{
		Messages:       make([]*pb.Message, 0),
		MemberIdToName: make(map[int32]string),
	}
	for rows.Next() {
		m := &pb.Message{}
		var t time.Time
		if err := rows.Scan(&m.Id, &m.SenderId, &m.ConversationId, &t, &m.MessageDataType, &m.Contents); err != nil {
			log.Fatal(err)
		}
		if m.Ts, err = ptypes.TimestampProto(t); err != nil {
			log.Fatal(err)
		}
		rsp.Messages = append(rsp.Messages, m)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	for _, id := range s.getConversationMemberIds(in.ConversationId) {
		rsp.MemberIdToName[id] = s.getUsername(id)
	}

	return rsp, nil
}

func (s *chatServer) GetUsernames(ctx context.Context, in *pb.GetUsernamesRequest) (*pb.GetUsernamesResponse, error) {
	log.Printf("GetUsernames\n")

	idToUsername := make(map[int32]string)
	for _, uid := range in.UserIds {
		idToUsername[uid] = s.getUsername(uid)
	}

	return &pb.GetUsernamesResponse{IdToUsername: idToUsername}, nil
}

func (s *chatServer) StreamMessages(in *pb.StreamMessagesRequest, stream pb.Chat_StreamMessagesServer) error {
	log.Printf("StreamMessages\n")

	msgChan := make(chan *pb.Message)
	termChan := make(chan interface{})
	s.CreateUserContext(in.UserId, msgChan, termChan)

	for {
		select {
		case msg := <-msgChan:
			if err := stream.Send(msg); err != nil {
				log.Fatal(err)
			}
		case <-termChan:
			s.DeleteUserContext(in.UserId)
			return nil
		}
	}
}

func (s *chatServer) SendMessage(ctx context.Context, in *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	log.Printf("SendMessage\n")

	var id int32
	var t time.Time
	if err := s.db.QueryRow("INSERT INTO messages (sender_id, conversation_id, data_type, contents) VALUES ($1, $2, $3, $4) RETURNING id, ts", in.Message.SenderId, in.Message.ConversationId, in.Message.MessageDataType, in.Message.Contents).Scan(&id, &t); err != nil {
		log.Fatal(err)
	}
	ts, err := ptypes.TimestampProto(t)
	if err != nil {
		log.Fatal(err)
	}
	in.Message.Id = id
	in.Message.Ts = ts

	redisConn := s.redisPool.Get()
	defer redisConn.Close()
	serverIdToReceiverIds := make(map[string][]int32)
	for _, memberId := range s.getConversationMemberIds(in.Message.ConversationId) {
		if memberId != in.Message.SenderId {
			serverId, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("user:%d", memberId), "serverId"))
			switch err {
			case redis.ErrNil: // This member is offline, ignore him
			case nil:
				serverIdToReceiverIds[serverId] = append(serverIdToReceiverIds[serverId], memberId)
			default:
				log.Fatal(err)
			}
		}
	}
	for serverId, receiverIds := range serverIdToReceiverIds {
		data, err := proto.Marshal(&pb.MessageWithReceivers{Message: in.Message, ReceiverIds: receiverIds})
		if err != nil {
			log.Fatal(err)
		}
		if _, err := redisConn.Do("PUBLISH", serverId, data); err != nil {
			log.Fatal(err)
		}
	}

	return &pb.SendMessageResponse{MessageId: id, Ts: ts}, nil
}

func (s *chatServer) pubHandler() {
	if err := s.pubSubConn.Subscribe(s.serverId); err != nil {
		log.Fatal(err)
	}

	for {
		switch n := s.pubSubConn.Receive().(type) {
		case error:
			log.Fatal(n.(error))
		case redis.Message:
			var msg pb.MessageWithReceivers
			if err := proto.Unmarshal(n.Data, &msg); err != nil {
				log.Fatal(err)
			}
			for _, userId := range msg.ReceiverIds {
				s.users[userId].msgChan <- msg.Message
			}
		}
	}
}

func (s *chatServer) getUserId(name string) (id int32, err error) {
	redisConn := s.redisPool.Get()
	defer redisConn.Close()
	idStr, err := redis.String(redisConn.Do("GET", fmt.Sprintf("userNameToId:%s", name)))
	switch err {
	case nil:
		idInt, err := strconv.Atoi(idStr)
		if err != nil {
			log.Fatal(err)
		}
		return int32(idInt), nil
	case redis.ErrNil:
		if err := s.db.QueryRow("SELECT id FROM users WHERE username = $1", name).Scan(&id); err != nil {
			if err == sql.ErrNoRows {
				return -1, status.Errorf(codes.NotFound, "User %s not found", name)
			}
			log.Fatal(err)
		}
		if _, err := redisConn.Do("SET", fmt.Sprintf("userNameToId:%s", name), id); err != nil {
			log.Fatal(err)
		}
		return id, nil
	default:
		log.Fatal(err)
	}
	return // Never reach here
}

func (s *chatServer) getUsername(id int32) (name string) {
	redisConn := s.redisPool.Get()
	defer redisConn.Close()
	name, err := redis.String(redisConn.Do("GET", fmt.Sprintf("userIdToName:%d", id)))
	switch err {
	case nil:
	case redis.ErrNil:
		if err := s.db.QueryRow("SELECT username FROM users WHERE id = $1", id).Scan(&name); err != nil {
			log.Fatal(err)
		}
		if _, err := redisConn.Do("SET", fmt.Sprintf("userIdToName:%d", id), name); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal(err)
	}
	return
}

func (s *chatServer) getConversationMemberIds(conversationId int32) (int32MemberIds []int32) {
	redisConn := s.redisPool.Get()
	defer redisConn.Close()
	memberIds, err := redis.Ints(redisConn.Do("SMEMBERS", fmt.Sprintf("conversationMembers:%d", conversationId)))
	if err != nil {
		log.Fatal(err)
	}

	if len(memberIds) == 0 {
		query := `SELECT participants.user_id FROM conversations
			INNER JOIN participants ON conversations.id = participants.conversation_id
			WHERE conversations.id = $1`
		rows, err := s.db.Query(query, conversationId)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			var memberId int32
			if err := rows.Scan(&memberId); err != nil {
				log.Fatal(err)
			}
			int32MemberIds = append(int32MemberIds, memberId)
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}
		if _, err := redisConn.Do("SADD", redis.Args{}.Add(fmt.Sprintf("conversationMembers:%d", conversationId)).AddFlat(int32MemberIds)...); err != nil {
			log.Fatal(err)
		}
	} else {
		for _, id := range memberIds {
			int32MemberIds = append(int32MemberIds, int32(id))
		}
	}
	return
}
