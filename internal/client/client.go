package client

import (
	"fmt"
	"log"

	pb "github.com/briansp8210/simple-chat-app/protobuf"
	"github.com/rivo/tview"
	"google.golang.org/grpc"
)

type chatClient struct {
	client pb.ChatClient

	app              *tview.Application
	pages            *tview.Pages
	modal            *tview.Modal
	conversationList *tview.List
	chatTextView     *tview.TextView
	msgInputField    *tview.InputField

	currentUser  *userContext
	userIdToName map[int32]string
}

type userContext struct {
	id               int32
	name             string
	conversations    []*conversation
	conversationsMap map[int32]*conversation
}

type conversation struct {
	*pb.Conversation
	messages []*pb.Message
	listIdx  int
}

func NewChatClient(host string, port int) *chatClient {
	client := &chatClient{}
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	client.client = pb.NewChatClient(conn)
	client.buildFrontEnd()
	client.userIdToName = make(map[int32]string)
	return client
}

func (c *chatClient) Run() {
	if err := c.app.Run(); err != nil {
		panic(err)
	}
}
