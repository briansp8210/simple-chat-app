package client

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	pb "github.com/briansp8210/simple-chat-app/protobuf"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/crypto/sha3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *chatClient) buildFrontEnd() {
	c.app = tview.NewApplication()
	c.pages = tview.NewPages()

	c.modal = tview.NewModal().
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			c.pages.HidePage("page-hover-modal")
		})
	hoverModal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(c.modal, 0, 1, false).
			AddItem(nil, 0, 1, false), 0, 1, false).
		AddItem(nil, 0, 1, false)

	// Following components are used in login page.

	var loginForm *tview.Form
	loginForm = tview.NewForm().
		SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tcell.ColorGray).
		AddInputField("[white]Username", "", 0, nil, nil).
		AddPasswordField("[white]Password", "", 0, '*', nil).
		AddButton("Register", func() {
			c.registerHandler(loginForm)
		}).
		AddButton("Login", func() {
			c.loginHandler(loginForm)
		}).
		AddButton("Quit", func() {
			c.app.Stop()
		})
	loginForm.SetBorder(true).SetTitle(" Simple Chat App ").SetTitleAlign(tview.AlignCenter)

	loginPageGrid := tview.NewGrid().
		SetRows(0, 12, 0).
		SetColumns(0, 50, 0).
		AddItem(loginForm, 1, 1, 1, 1, 0, 0, true)

	c.pages.AddPage("page-login", loginPageGrid, true, true)

	// Following components are used in main page.

	c.conversationList = tview.NewList()

	c.chatTextView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			c.app.Draw()
		})

	c.msgInputField = tview.NewInputField().
		SetLabel("Message: ").
		SetAcceptanceFunc(tview.InputFieldMaxLength(128)).
		SetLabelColor(tcell.ColorWhite).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				c.sendMessageHandler()
			}
		})

	info := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetHighlightedFunc(func(added, removed, remaining []string) {

			c.pages.SwitchToPage("page-login")
		})
	fmt.Fprintf(info, `F1 ["1"][darkcyan]Logout[white][""] `)
	fmt.Fprintf(info, `F2 ["2"][darkcyan]Add-conversation[white][""]`)

	var addConversationForm *tview.Form
	addConversationForm = tview.NewForm().
		SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tcell.ColorGray).
		AddInputField("[white]User/Group Name", "", 0, nil, nil).
		AddDropDown("Type", []string{"PRIVATE", "GROUP"}, -1, nil).
		AddButton("OK", func() {
			c.addConversationHandler(addConversationForm)
		})
	addConversationForm.SetBorder(true).SetTitle(" Add Conversation ").SetTitleAlign(tview.AlignCenter)
	hoverAddConversationForm := tview.NewFlex().
		AddItem(nil, 0, 2, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(addConversationForm, 10, 1, true).
			AddItem(nil, 0, 1, false), 0, 1, true).
		AddItem(nil, 0, 2, false)

	grid := tview.NewGrid().
		SetRows(0, 1, 1).
		SetColumns(30, 0).
		SetBorders(true).
		AddItem(c.conversationList, 0, 0, 2, 1, 0, 0, true).
		AddItem(c.chatTextView, 0, 1, 1, 1, 0, 0, true).
		AddItem(c.msgInputField, 1, 1, 1, 1, 0, 0, true).
		AddItem(info, 2, 0, 1, 2, 0, 0, true)

	c.pages.AddPage("page-main", grid, true, false)
	c.pages.AddPage("page-add-conversation-form", hoverAddConversationForm, true, false)
	c.pages.AddPage("page-hover-modal", hoverModal, true, false)

	c.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF1:
			c.logoutHandler()
			return nil
		case tcell.KeyF2:
			c.pages.ShowPage("page-add-conversation-form")
			return nil
		}
		return event
	})

	c.app.SetRoot(c.pages, true).SetFocus(c.pages).EnableMouse(true)
}

func (c *chatClient) registerHandler(form *tview.Form) {
	password := sha3.Sum512([]byte(form.GetFormItem(1).(*tview.InputField).GetText()))
	req := &pb.RegisterRequest{
		Username: form.GetFormItem(0).(*tview.InputField).GetText(),
		Password: password[:],
	}

	if _, err := c.client.Register(context.Background(), req); err != nil {
		if st := status.Convert(err); st.Code() == codes.AlreadyExists {
			c.showHoverModal(st.Message())
			return
		} else {
			log.Fatal(err)
		}
	}
	c.showHoverModal("Success")
}

func (c *chatClient) loginHandler(form *tview.Form) {
	username := form.GetFormItem(0).(*tview.InputField).GetText()
	password := sha3.Sum512([]byte(form.GetFormItem(1).(*tview.InputField).GetText()))

	rsp, err := c.client.Login(context.Background(), &pb.LoginRequest{Username: username, Password: password[:]})
	if err != nil {
		if st := status.Convert(err); st.Code() == codes.NotFound || st.Code() == codes.Unauthenticated {
			c.showHoverModal(st.Message())
			return
		} else {
			log.Fatal(err)
		}
	}
	c.currentUser = &userContext{
		id:               rsp.UserId,
		name:             username,
		conversations:    make([]*conversation, 0, len(rsp.Conversations)),
		conversationsMap: make(map[int32]*conversation, len(rsp.Conversations)),
	}
	for _, con := range rsp.Conversations {
		c.addConversation(con)
	}
	go c.startStreamingMessages()
	c.pages.SwitchToPage("page-main")
}

func (c *chatClient) logoutHandler() {
	if _, err := c.client.Logout(context.Background(), &pb.LogoutRequest{UserId: c.currentUser.id}); err != nil {
		log.Fatal(err)
	}
	c.currentUser = nil
	c.conversationList.Clear()
	c.chatTextView.Clear()
	c.msgInputField.SetText("")
	c.pages.SwitchToPage("page-login")
}

func (c *chatClient) conversationSelectedHandler() {
	conversation := c.currentUser.conversations[c.conversationList.GetCurrentItem()]
	c.chatTextView.Clear()

	if conversation.messages == nil {
		// Explicitly make a slice s.t. even when this conversation has no message, `conversation.messages` will not be nil.
		conversation.messages = make([]*pb.Message, 0)
		rsp, err := c.client.GetMessages(context.Background(), &pb.GetMessagesRequest{ConversationId: conversation.Id})
		if err != nil {
			log.Fatal(err)
		}
		for _, msg := range rsp.Messages {
			conversation.messages = append(conversation.messages, msg)
		}
		for id, name := range rsp.MemberIdToName {
			c.userIdToName[id] = name
		}
	}

	for _, msg := range conversation.messages {
		c.showMessage(msg)
	}
}

func (c *chatClient) sendMessageHandler() {
	conversation := c.currentUser.conversations[c.conversationList.GetCurrentItem()]
	contents := strings.TrimSpace(c.msgInputField.GetText())
	c.msgInputField.SetText("")
	if len(contents) == 0 {
		return
	}

	msg := &pb.Message{
		SenderId:        c.currentUser.id,
		ConversationId:  conversation.Id,
		MessageDataType: "TEXT",
		Contents:        contents,
	}
	rsp, err := c.client.SendMessage(context.Background(), &pb.SendMessageRequest{Message: msg})
	if err != nil {
		log.Fatal(err)
	}

	msg.Id = rsp.MessageId
	msg.Ts = rsp.Ts
	conversation.messages = append(conversation.messages, msg)
	c.showMessage(msg)
}

func (c *chatClient) addConversationHandler(form *tview.Form) {
	_, t := form.GetFormItem(1).(*tview.DropDown).GetCurrentOption()
	name := form.GetFormItem(0).(*tview.InputField).GetText()

	req := &pb.AddConversationRequest{Conversation: &pb.Conversation{Type: t}}
	switch t {
	case "PRIVATE":
		req.MemberNames = []string{c.currentUser.name, name}
	case "GROUP":
		req.MemberNames = []string{c.currentUser.name}
		req.Conversation.Name = name
	}

	if rsp, err := c.client.AddConversation(context.Background(), req); err != nil {
		if st := status.Convert(err); st.Code() == codes.NotFound {
			c.showHoverModal(st.Message())
		} else {
			log.Fatal(err)
		}
	} else {
		c.addConversation(rsp.Conversation)
		c.pages.HidePage("page-add-conversation-form")
	}
}

func (c *chatClient) startStreamingMessages() {
	stream, err := c.client.StreamMessages(context.Background(), &pb.StreamMessagesRequest{UserId: c.currentUser.id})
	if err != nil {
		log.Fatal(err)
	}

	for {
		msg, err := stream.Recv()
		switch err {
		case nil:
			if _, exist := c.currentUser.conversationsMap[msg.ConversationId]; !exist {
				rsp, err := c.client.GetConversation(context.Background(), &pb.GetConversationRequest{ConversationId: msg.ConversationId})
				if err != nil {
					log.Fatal(err)
				}
				c.addConversation(rsp.Conversation)
				c.app.Draw()
			}
			c.currentUser.conversationsMap[msg.ConversationId].messages = append(c.currentUser.conversationsMap[msg.ConversationId].messages, msg)
			if c.currentUser.conversations[c.conversationList.GetCurrentItem()].Id == msg.ConversationId {
				c.showMessage(msg)
			}
		case io.EOF:
			break
		default:
			log.Fatal(err)
		}
	}
}

func (c *chatClient) showMessage(msg *pb.Message) {
	switch msg.SenderId {
	case c.currentUser.id:
		fmt.Fprintf(c.chatTextView, "[green]%s[white] ", tview.Escape(fmt.Sprintf("[%s]", c.getUsername(msg.SenderId))))
	default:
		fmt.Fprintf(c.chatTextView, "%s ", tview.Escape(fmt.Sprintf("[%s]", c.getUsername(msg.SenderId))))
	}

	switch msg.MessageDataType {
	case "TEXT":
		fmt.Fprint(c.chatTextView, msg.Contents)
	}

	fmt.Fprintf(c.chatTextView, " [gray](%s)[white]\n", msg.Ts.AsTime().Local().Format("2006/01/02 15:04"))
}

func (c *chatClient) showHoverModal(msg string) {
	c.modal.SetText(msg)
	c.pages.ShowPage("page-hover-modal")
	c.app.SetFocus(c.modal)
}

func (c *chatClient) getConversationName(conversation *pb.Conversation) (name string) {
	switch conversation.Type {
	case "PRIVATE":
		if conversation.MemberIds[0] == c.currentUser.id {
			name = c.getUsername(conversation.MemberIds[1])
		} else {
			name = c.getUsername(conversation.MemberIds[0])
		}
	case "GROUP":
		name = conversation.Name
	}
	return
}

func (c *chatClient) getUsername(id int32) string {
	if _, exist := c.userIdToName[id]; !exist {
		rsp, err := c.client.GetUsernames(context.Background(), &pb.GetUsernamesRequest{UserIds: []int32{id}})
		if err != nil {
			log.Fatal(err)
		}
		c.userIdToName[id] = rsp.IdToUsername[id]
	}
	return c.userIdToName[id]
}

func (c *chatClient) addConversation(pbCon *pb.Conversation) {
	con := &conversation{Conversation: pbCon}
	c.currentUser.conversations = append(c.currentUser.conversations, con)
	c.currentUser.conversationsMap[con.Id] = con
	c.conversationList.AddItem(c.getConversationName(pbCon), "", 0, c.conversationSelectedHandler)
}
