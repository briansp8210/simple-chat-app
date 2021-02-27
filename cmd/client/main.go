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

	hash1 := sha3.Sum512([]byte("password1"))
	_, err = c.Register(context.Background(), &pb.RegisterRequest{
		Username: "user1",
		Password: hash1[:],
	})
	if err != nil {
		log.Fatal(err)
	}
	hash2 := sha3.Sum512([]byte("password2"))
	_, err = c.Register(context.Background(), &pb.RegisterRequest{
		Username: "user2",
		Password: hash2[:],
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.Login(context.Background(), &pb.LoginRequest{
		Username: "user1",
		Password: hash1[:],
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("user1 login successfully")

	rsp, err := c.AddConversation(context.Background(), &pb.AddConversationRequest{
		MemberNames: []string{"user1", "user2"},
		Conversation: &pb.Conversation{
			Name: "user2",
			Type: "PRIVATE",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.GetMessages(context.Background(), &pb.GetMessagesRequest{
		ConversationId: rsp.ConversationId,
	})

	_, err = c.Logout(context.Background(), &pb.LogoutRequest{
		Username: "user1",
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("user1 logout successfully")
}
