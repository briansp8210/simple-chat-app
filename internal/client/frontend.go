package client

import (
	"context"
	"fmt"
	"log"

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
	c.pages.AddPage("page-hover-modal", hoverModal, true, false)

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
		SetFieldBackgroundColor(tcell.ColorBlack)

	info := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetHighlightedFunc(func(added, removed, remaining []string) {

			c.pages.SwitchToPage("page-login")
		})
	fmt.Fprintf(info, `F1 ["1"][darkcyan]Logout[white][""] `)

	grid := tview.NewGrid().
		SetRows(0, 1, 1).
		SetColumns(30, 0).
		SetBorders(true).
		AddItem(c.conversationList, 0, 0, 2, 1, 0, 0, true).
		AddItem(c.chatTextView, 0, 1, 1, 1, 0, 0, true).
		AddItem(c.msgInputField, 1, 1, 1, 1, 0, 0, true).
		AddItem(info, 2, 0, 1, 2, 0, 0, true)

	c.pages.AddPage("page-main", grid, true, false)

	c.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF1:
			c.logoutHandler()
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
	password := sha3.Sum512([]byte(form.GetFormItem(1).(*tview.InputField).GetText()))
	req := &pb.LoginRequest{
		Username: form.GetFormItem(0).(*tview.InputField).GetText(),
		Password: password[:],
	}

	rsp, err := c.client.Login(context.Background(), req)
	if err != nil {
		if st := status.Convert(err); st.Code() == codes.NotFound || st.Code() == codes.Unauthenticated {
			c.showHoverModal(st.Message())
			return
		} else {
			log.Fatal(err)
		}
	}
	c.currentUser = &userContext{id: rsp.UserId}
	for _, convo := range rsp.Conversations {
		c.currentUser.conversations = append(c.currentUser.conversations, &conversation{Conversation: convo})
		c.conversationList.AddItem(convo.Name, "", 0, c.conversationSelectedHandler)
	}
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
		switch msg.SenderId {
		case c.currentUser.id:
			fmt.Fprintf(c.chatTextView, "[green]%s[white] ", tview.Escape(fmt.Sprintf("[%s]", c.userIdToName[msg.SenderId])))
		default:
			fmt.Fprintf(c.chatTextView, "%s ", tview.Escape(fmt.Sprintf("[%s]", c.userIdToName[msg.SenderId])))
		}

		switch msg.MessageDataType {
		case "TEXT":
			fmt.Fprint(c.chatTextView, msg.Contents)
		}

		fmt.Fprintf(c.chatTextView, " [gray](%s)[white]\n", msg.Ts.AsTime().Local().Format("2006/01/02 15:04"))
	}
}

func (c *chatClient) showHoverModal(msg string) {
	c.modal.SetText(msg)
	c.pages.ShowPage("page-hover-modal")
	c.app.SetFocus(c.modal)
}
