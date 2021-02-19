package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"golang.org/x/crypto/sha3"
	"google.golang.org/grpc"

	pb "github.com/briansp8210/simple-chat-app/protobuf"
)

var (
	host = flag.String("host", "127.0.0.1", "Binding IP address")
	port = flag.Int("port", 61234, "Binding port number")
)

func main() {
	flag.Parse()

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", *host, *port), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	c := pb.NewChatClient(conn)

	hash := sha3.Sum512([]byte("password1"))
	_, err = c.Register(context.Background(), &pb.RegisterRequest{
		Username:     "user1",
		PasswordHash: hash[:],
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.Login(context.Background(), &pb.LoginRequest{
		Username:     "user1",
		PasswordHash: hash[:],
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("user1 login successfully")

	_, err = c.Logout(context.Background(), &pb.LogoutRequest{
		Username: "user1",
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("user1 logout successfully")
}
