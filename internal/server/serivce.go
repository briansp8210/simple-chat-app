package server

import (
	"context"
	"log"

	pb "github.com/briansp8210/simple-chat-app/protobuf"
	_ "github.com/lib/pq"
)

func (s *chatServer) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	log.Printf("Receive Registration from user %s\n", in.Username)

	_, err := s.db.ExecContext(ctx, "INSERT INTO users (username, password_hash) VALUES ($1, $2)", in.Username, in.PasswordHash)
	if err != nil {
		return &pb.RegisterResponse{}, err
	}
	return &pb.RegisterResponse{}, nil
}
