package server

import (
	"database/sql"
	"log"
	"net"

	pb "github.com/briansp8210/simple-chat-app/protobuf"
	"google.golang.org/grpc"
)

type chatServer struct {
	pb.UnimplementedChatServer

	grpcServer *grpc.Server
	db         *sql.DB
}

func NewChatServer() *chatServer {
	grpcServer := grpc.NewServer()
	db, err := sql.Open("postgres", "user=postgres host=db sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	server := &chatServer{grpcServer: grpcServer, db: db}
	pb.RegisterChatServer(grpcServer, server)
	return server
}

func (s *chatServer) Serve(lis *net.Listener) {
	if err := s.grpcServer.Serve(*lis); err != nil {
		log.Fatal(err)
	}
}
