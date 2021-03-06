package client

import (
	"context"
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
	c.currentUser.id = rsp.UserId
	c.currentUser.conversations = rsp.Conversations
	c.pages.SwitchToPage("page-main")
}

func (c *chatClient) showHoverModal(msg string) {
	c.modal.SetText(msg)
	c.pages.ShowPage("page-hover-modal")
	c.app.SetFocus(c.modal)
}
