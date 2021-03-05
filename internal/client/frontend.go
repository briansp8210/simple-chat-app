package client

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (c *chatClient) buildFrontEnd() {
	c.app = tview.NewApplication()
	pages := tview.NewPages()

	// Following components are used in login page.

	loginForm := tview.NewForm().
		SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tcell.ColorGray).
		AddInputField("[white]Username", "", 0, nil, nil).
		AddPasswordField("[white]Password", "", 0, '*', nil).
		AddButton("Register", nil).
		AddButton("Login", nil).
		AddButton("Quit", func() {
			c.app.Stop()
		})
	loginForm.SetBorder(true).SetTitle(" Simple Chat App ").SetTitleAlign(tview.AlignCenter)

	loginPageGrid := tview.NewGrid().
		SetRows(0, 12, 0).
		SetColumns(0, 50, 0).
		AddItem(loginForm, 1, 1, 1, 1, 0, 0, true)

	pages.AddPage("page-login", loginPageGrid, true, true)

	c.app.SetRoot(pages, true).SetFocus(pages).EnableMouse(true)
}
