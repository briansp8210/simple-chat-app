package server

import (
	"bytes"
	"context"
	"database/sql"
	"log"

	pb "github.com/briansp8210/simple-chat-app/protobuf"
	"github.com/gomodule/redigo/redis"
	_ "github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *chatServer) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.Empty, error) {
	log.Printf("Receive Registration from user %s\n", in.Username)

	_, err := s.db.ExecContext(ctx, "INSERT INTO users (username, password_hash) VALUES ($1, $2)", in.Username, in.PasswordHash)
	if err != nil {
		return &pb.Empty{}, err
	}
	return &pb.Empty{}, nil
}

func (s *chatServer) Login(ctx context.Context, in *pb.LoginRequest) (*pb.Empty, error) {
	log.Printf("Receive Login from user %s\n", in.Username)

	row := s.db.QueryRow("SELECT password_hash FROM users WHERE username = $1", in.Username)
	var correctHash []byte
	if err := row.Scan(&correctHash); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "User %s not found", in.Username)
		}
		log.Fatal(err)
	}

	if !bytes.Equal(in.PasswordHash, correctHash) {
		return nil, status.Error(codes.Unauthenticated, "Invalid password")
	}

	if _, err := s.redisConn.Do("HSET", "userLocation", in.Username, s.instanceID); err != nil {
		log.Fatal(err)
	}

	return &pb.Empty{}, nil
}

func (s *chatServer) Logout(ctx context.Context, in *pb.LogoutRequest) (*pb.Empty, error) {
	log.Printf("Receive Logout from user %s\n", in.Username)

	count, err := redis.Int(s.redisConn.Do("HDEL", "userLocation", in.Username))
	if err != nil {
		log.Fatal(err)
	}
	if count != 1 {
		return nil, status.Errorf(codes.NotFound, "User %s not found", in.Username)
	}

	return &pb.Empty{}, nil
}
