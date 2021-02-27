package server

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"

	pb "github.com/briansp8210/simple-chat-app/protobuf"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/gomodule/redigo/redis"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *chatServer) Register(ctx context.Context, in *pb.RegisterRequest) (*empty.Empty, error) {
	log.Printf("Receive Registration from user %s\n", in.Username)

	_, err := s.db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", in.Username, in.Password)
	if err != nil {
		return &empty.Empty{}, err
	}
	return &empty.Empty{}, nil
}

func (s *chatServer) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Printf("Receive Login from user %s\n", in.Username)

	var id int32
	var correctPassword []byte
	err := s.db.QueryRow("SELECT id, password FROM users WHERE username = $1", in.Username).Scan(
		&id, &correctPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "User %s not found", in.Username)
		}
		log.Fatal(err)
	}

	if !bytes.Equal(in.Password, correctPassword) {
		return nil, status.Error(codes.Unauthenticated, "Invalid password")
	}

	if _, err := s.redisConn.Do("HSET", fmt.Sprintf("user:%s", in.Username), "id", id, "location", s.instanceID); err != nil {
		log.Fatal(err)
	}

	rows, err := s.db.Query("SELECT * FROM conversations WHERE $1 = ANY (member_ids)", id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	conversations := make([]*pb.Conversation, 0)
	for rows.Next() {
		c := &pb.Conversation{}
		if err := rows.Scan(&c.Id, &c.Name, &c.Type, &c.MemberIds, &c.MessageIds); err != nil {
			log.Fatal(err)
		}
		conversations = append(conversations, c)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return &pb.LoginResponse{UserId: id, Conversations: conversations}, nil
}

func (s *chatServer) Logout(ctx context.Context, in *pb.LogoutRequest) (*empty.Empty, error) {
	log.Printf("Receive Logout from user %s\n", in.Username)

	count, err := redis.Int(s.redisConn.Do("DEL", fmt.Sprintf("user:%s", in.Username)))
	if err != nil {
		log.Fatal(err)
	}
	if count != 1 {
		return nil, status.Errorf(codes.NotFound, "User %s not found", in.Username)
	}

	return &empty.Empty{}, nil
}

func (s *chatServer) AddConversation(ctx context.Context, in *pb.AddConversationRequest) (*pb.AddConversationResponse, error) {
	log.Printf("AddConversation\n")

	for _, member := range in.MemberNames {
		in.Conversation.MemberIds = append(in.Conversation.MemberIds, s.getUserId(member))
	}

	var id int32
	err := s.db.QueryRow("INSERT INTO conversations (name, type, member_ids) VALUES ($1, $2, $3) RETURNING id",
		in.Conversation.Name, "PRIVATE", pq.Array(in.Conversation.MemberIds)).Scan(&id)
	if err != nil {
		log.Fatal(err)
	}

	return &pb.AddConversationResponse{ConversationId: id}, nil
}

func (s *chatServer) GetMessages(ctx context.Context, in *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	log.Printf("GetMessages\n")

	rows, err := s.db.Query("SELECT * FROM messages WHERE conversation_id = $1 ORDER BY ts ASC", in.ConversationId)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	messages := make([]*pb.Message, 0)
	for rows.Next() {
		m := &pb.Message{}
		if err := rows.Scan(&m.Id, &m.SenderId, &m.ConversationId, &m.Ts, &m.MessageDataType, &m.Contents); err != nil {
			log.Fatal(err)
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return &pb.GetMessagesResponse{Messages: messages}, nil
}

func (s *chatServer) getUserId(name string) int32 {
	idStr, err := redis.String(s.redisConn.Do("HGET", fmt.Sprintf("user:%s", name), "id"))
	if err != nil && err != redis.ErrNil {
		log.Fatal(err)
	}
	if err != redis.ErrNil {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Fatal(err)
		}
		return int32(id)
	}

	var id int32
	if err := s.db.QueryRow("SELECT id FROM users WHERE username = $1", name).Scan(&id); err != nil {
		log.Fatal(err)
	}
	if _, err := s.redisConn.Do("HSET", fmt.Sprintf("user:%s", name), "id", id); err != nil {
		log.Fatal(err)
	}
	return id
}
