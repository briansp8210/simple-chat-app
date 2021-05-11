package server

import (
	"database/sql"
	"log"
	"net"
	"strconv"
	"time"

	pb "github.com/briansp8210/simple-chat-app/protobuf"
	"github.com/gomodule/redigo/redis"
	"google.golang.org/grpc"
)

type chatServer struct {
	pb.UnimplementedChatServer

	serverId   string
	grpcServer *grpc.Server
	db         *sql.DB
	redisPool  *redis.Pool
	pubSubConn *redis.PubSubConn

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

	redisPool := &redis.Pool{
		MaxIdle:     8,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "redis:6379")
		},
	}

	redisConn := redisPool.Get()
	serverId, err := redis.Int(redisConn.Do("INCR", "serverId"))
	if err != nil {
		log.Fatal(err)
	}

	server := &chatServer{
		serverId:   strconv.Itoa(serverId),
		grpcServer: grpcServer,
		db:         db,
		redisPool:  redisPool,
		pubSubConn: &redis.PubSubConn{Conn: redisConn},
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

func (s *chatServer) CreateUserContext(id int32, msgChan chan *pb.Message, termChan chan interface{}) {
	s.users[id] = &userContext{
		msgChan:  msgChan,
		termChan: termChan,
	}
}

func (s *chatServer) DeleteUserContext(id int32) {
	delete(s.users, id)
}
