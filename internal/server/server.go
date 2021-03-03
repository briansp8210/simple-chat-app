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

	serverId   string
	grpcServer *grpc.Server
	db         *sql.DB
	redisConn  redis.Conn
	pubSubConn redis.PubSubConn

	users map[int32]*userContext
}

type userContext struct {
	msgChan  chan *pb.Message
	termChan chan interface{}
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

	serverId, err := redis.Int(redisConn.Do("INCR", "serverId"))
	if err != nil {
		log.Fatal(err)
	}

	c, err := redis.Dial("tcp", "redis:6379")
	if err != nil {
		log.Fatal(err)
	}
	pubSubConn := redis.PubSubConn{Conn: c}

	server := &chatServer{
		serverId:   strconv.Itoa(serverId),
		grpcServer: grpcServer,
		db:         db,
		redisConn:  redisConn,
		pubSubConn: pubSubConn,
		users:      make(map[int32]*userContext),
	}
	pb.RegisterChatServer(grpcServer, server)
	return server
}

func (s *chatServer) Serve(lis *net.Listener) {
	go s.pubHandler()
	if err := s.grpcServer.Serve(*lis); err != nil {
		log.Fatal(err)
	}
}

func (s *chatServer) AddUserContext(id int32, msgChan chan *pb.Message, termChan chan interface{}) {
	s.users[id] = &userContext{
		msgChan:  msgChan,
		termChan: termChan,
	}
}
