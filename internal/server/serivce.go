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
		// Ref.: https://www.postgresql.org/docs/9.2/errcodes-appendix.html
		switch pqErr := err.(*pq.Error); pqErr.Code {
		case "23505":
			return nil, status.Errorf(codes.AlreadyExists, "Username %s is already used", in.Username)
		case "23514":
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

	rows, err := s.db.Query("SELECT * FROM conversations WHERE $1 = ANY (member_ids)", uid)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	conversations := make([]*pb.Conversation, 0)
	for rows.Next() {
		c := &pb.Conversation{}
		if err := rows.Scan(&c.Id, &c.Name, &c.Type, pq.Array(&c.MemberIds)); err != nil {
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

	for _, member := range in.MemberNames {
		if uid, err := s.getUserId(member); err != nil {
			return nil, err
		} else {
			in.Conversation.MemberIds = append(in.Conversation.MemberIds, uid)
		}
	}

	if err := s.db.QueryRow("INSERT INTO conversations (name, type, member_ids) VALUES ($1, $2, $3) RETURNING id", in.Conversation.Name, "PRIVATE", pq.Array(in.Conversation.MemberIds)).Scan(&in.Conversation.Id); err != nil {
		if pqErr := err.(*pq.Error); pqErr.Code == "23505" {
			return nil, status.Errorf(codes.AlreadyExists, "Conversation already exists")
		} else {
			log.Fatal(err)
		}
	}

	redisConn := s.redisPool.Get()
	defer redisConn.Close()
	if _, err := redisConn.Do("SADD", redis.Args{}.Add(fmt.Sprintf("conversationMembers:%d", in.Conversation.Id)).AddFlat(in.Conversation.MemberIds)...); err != nil {
		log.Fatal(err)
	}

	return &pb.AddConversationResponse{Conversation: in.Conversation}, nil
}

func (s *chatServer) GetConversation(ctx context.Context, in *pb.GetConversationRequest) (*pb.GetConversationResponse, error) {
	log.Printf("GetConversation\n")

	conversation := &pb.Conversation{}
	row := s.db.QueryRow("SELECT * FROM conversations WHERE id = $1", in.ConversationId)
	if err := row.Scan(&conversation.Id, &conversation.Name, &conversation.Type, pq.Array(&conversation.MemberIds)); err != nil {
		log.Fatal(err)
	}

	return &pb.GetConversationResponse{Conversation: conversation}, nil
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
		if err := s.db.QueryRow("SELECT member_ids FROM conversations WHERE id = $1", conversationId).Scan(pq.Array(&int32MemberIds)); err != nil {
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
