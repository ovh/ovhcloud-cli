// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package views

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// View is the interface that all browser views must implement.
// Each view encapsulates its own state, rendering, and key handling.
type View interface {
	// Render returns the view content as a string.
	// width and height are the available dimensions for rendering.
	Render(width, height int) string

	// HandleKey processes keyboard input and returns a command.
	// It may return a ViewTransition to switch to a different view.
	HandleKey(msg tea.KeyMsg) tea.Cmd

	// Update handles async messages (API responses, timers, etc.)
	// Returns a command and optionally a new view if transitioning.
	Update(msg tea.Msg) (tea.Cmd, View)

	// Title returns the view's title for the header bar.
	Title() string

	// HelpText returns contextual help text for the footer.
	HelpText() string
}

// ViewTransition is a message that signals a view change.
type ViewTransition struct {
	NewView View
	Cmd     tea.Cmd // Optional command to run after transition
}

// TransitionTo creates a ViewTransition message.
func TransitionTo(view View, cmd tea.Cmd) tea.Msg {
	return ViewTransition{NewView: view, Cmd: cmd}
}

// Context provides shared state and services to all views.
// This is passed to view constructors and stored in views that need it.
type Context struct {
	// CloudProject is the current OVH cloud project ID.
	CloudProject string
	// CloudProjectName is the display name of the current project.
	CloudProjectName string
	// Width and Height are the terminal dimensions.
	Width  int
	Height int
	// Notification is a temporary message to display.
	Notification       string
	NotificationExpiry int64 // Unix timestamp
}

// SetNotification sets a notification message with expiry.
func (c *Context) SetNotification(msg string, durationSec int) {
	c.Notification = msg
	c.NotificationExpiry = time.Now().Unix() + int64(durationSec)
}

// BaseView provides common functionality for views.
// Embed this in view structs to get default implementations.
type BaseView struct {
	ctx *Context
}

// NewBaseView creates a BaseView with the given context.
func NewBaseView(ctx *Context) BaseView {
	return BaseView{ctx: ctx}
}

// Context returns the shared context.
func (v *BaseView) Context() *Context {
	return v.ctx
}

// Update default implementation - returns nil (no transition, no command).
func (v *BaseView) Update(msg tea.Msg) (tea.Cmd, View) {
	return nil, nil
}
