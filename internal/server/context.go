package server

import (
	"database/sql"
	"log"
	"net"
	"strconv"

	pb "github.com/briansp8210/simple-chat-app/protobuf"
	"github.com/gomodule/redigo/redis"
	"google.golang.org/grpc"
)

type chatServer struct {
	pb.UnimplementedChatServer

	instanceID string
	grpcServer *grpc.Server
	db         *sql.DB
	redisConn  redis.Conn
}

func NewChatServer() *chatServer {
	grpcServer := grpc.NewServer()
	db, err := sql.Open("postgres", "user=postgres host=db sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	redisConn, err := redis.Dial("tcp", "redis:6379")
	if err != nil {
		log.Fatal(err)
	}
	id, err := redis.Int(redisConn.Do("INCR", "chatServerInstanceID"))
	if err != nil {
		log.Fatal(err)
	}

	server := &chatServer{
		instanceID: strconv.Itoa(id),
		grpcServer: grpcServer,
		db:         db,
		redisConn:  redisConn,
	}
	pb.RegisterChatServer(grpcServer, server)
	return server
}

func (s *chatServer) Serve(lis *net.Listener) {
	if err := s.grpcServer.Serve(*lis); err != nil {
		log.Fatal(err)
	}
}
