package client

import (
	"fmt"
	"log"

	pb "github.com/briansp8210/simple-chat-app/protobuf"
	"github.com/rivo/tview"
	"google.golang.org/grpc"
)

type chatClient struct {
	client *pb.ChatClient
	app    *tview.Application
}

func NewChatClient(host string, port int) *chatClient {
	client := &chatClient{}
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	pbClient := pb.NewChatClient(conn)
	client.client = &pbClient
	client.buildFrontEnd()
	return client
}

func (c *chatClient) Run() {
	if err := c.app.Run(); err != nil {
		panic(err)
	}
}
