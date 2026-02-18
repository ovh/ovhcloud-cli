// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package views

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// LoadingView displays a loading indicator with an optional message.
type LoadingView struct {
	BaseView
	message    string
	showSplash bool
}

// NewLoadingView creates a new loading view.
func NewLoadingView(ctx *Context, message string, showSplash bool) *LoadingView {
	return &LoadingView{
		BaseView:   NewBaseView(ctx),
		message:    message,
		showSplash: showSplash,
	}
}

// ASCII OVHcloud logo for loading screen
const ovhcloudASCIILogo = `
   ____  __      __ _    _        _                    _ 
  / __ \ \ \    / /| |  | |      | |                  | |
 | |  | | \ \  / / | |__| |  ___ | |  ___   _   _   __| |
 | |  | |  \ \/ /  |  __  | / __|| | / _ \ | | | | / _` + "`" + ` |
 | |__| |   \  /   | |  | || (__ | || (_) || |_| || (_| |
  \____/     \/    |_|  |_| \___||_| \___/  \__,_| \__,_|
`

func (v *LoadingView) Render(width, height int) string {
	var content strings.Builder

	if v.showSplash {
		content.WriteString(StyleLogo.Render(ovhcloudASCIILogo))
		content.WriteString("\n\n")
		content.WriteString(StyleNotificationSuccess.Render("        ⏳ " + v.message))
		content.WriteString("\n")
	} else {
		content.WriteString(StyleLoading.Render("⏳ " + v.message))
	}

	return content.String()
}

func (v *LoadingView) HandleKey(msg tea.KeyMsg) tea.Cmd {
	// Loading view doesn't handle keys
	return nil
}

func (v *LoadingView) Title() string {
	return " Loading "
}

func (v *LoadingView) HelpText() string {
	return "Loading... Please wait"
}

// ErrorView displays an error message.
type ErrorView struct {
	BaseView
	errorMsg string
}

// NewErrorView creates a new error view.
func NewErrorView(ctx *Context, errorMsg string) *ErrorView {
	return &ErrorView{
		BaseView: NewBaseView(ctx),
		errorMsg: errorMsg,
	}
}

func (v *ErrorView) Render(width, height int) string {
	return StyleError.Render("❌ Error: " + v.errorMsg)
}

func (v *ErrorView) HandleKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "enter":
		// Signal to go back to previous view
		return func() tea.Msg {
			return GoBackMsg{}
		}
	}
	return nil
}

func (v *ErrorView) Title() string {
	return " Error "
}

func (v *ErrorView) HelpText() string {
	return "Esc/Enter: Go back"
}

// GoBackMsg signals the controller to go back to the previous view.
type GoBackMsg struct{}

// EmptyView displays an empty state with a creation prompt.
type EmptyView struct {
	BaseView
	productName string
	canCreate   bool
}

// NewEmptyView creates a new empty view.
func NewEmptyView(ctx *Context, productName string, canCreate bool) *EmptyView {
	return &EmptyView{
		BaseView:    NewBaseView(ctx),
		productName: productName,
		canCreate:   canCreate,
	}
}

func (v *EmptyView) Render(width, height int) string {
	var content strings.Builder

	content.WriteString(StyleSubheader.Render("No " + v.productName + " found\n\n"))

	if v.canCreate {
		content.WriteString(StyleHelp.Render("Press 'c' to create a new " + v.productName))
	}

	return content.String()
}

func (v *EmptyView) HandleKey(msg tea.KeyMsg) tea.Cmd {
	// 'c' for create will be handled by the controller
	return nil
}

func (v *EmptyView) Title() string {
	return " " + v.productName + " "
}

func (v *EmptyView) HelpText() string {
	if v.canCreate {
		return "c: Create • p: Projects • q: Quit"
	}
	return "p: Projects • q: Quit"
}
