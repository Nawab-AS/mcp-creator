package backend

import (
	"slices"
	"sync/atomic"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// handle-close.go - handles the close event of the main window, allowing for user interaction before closing the app
// Example Usage:
//
//  Golang:
//	func someFunction() {
//		backend.forceClose()
//	}
//
//  Frontend (JS):
//  (Synchronous)
//  closeID = <backend bindings>.Close()
//  // some UI interaction
//  <backend bindings>.Close(closeID)
//
//  (Asynchronous; Events)
//  <backend bindings>.Close()
//  Events.On("attempt-close", (closeID) => {
//      // some UI interaction
//      <backend bindings>.Close(closeID)
//  })

var allowClose bool = false
var closing atomic.Bool = atomic.Bool{}
var allowCloseIDs []int
var mainWindow application.Window

// forceClose is closing from
func (b *Backend) forceClose() {
	allowClose = true
	mainWindow.Close()
}

func (b *Backend) registerCloseHandler() {
	mainWindow = b.app.Window.Current()
	closing.Store(false)

	// handle window close event
	mainWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		if closing.Load() {
			return
		}

		if allowClose {
			closing.Store(true)
			b.ServiceShutdown()
			b.app.Quit()
			return
		}

		e.Cancel()
	})

	// force user interaction to close
	Events.On("close", func(e *application.CustomEvent) {
		closeID, ok := e.Data.(int)
		if !ok {
			return
		}

		if slices.Contains(allowCloseIDs, closeID) {
			b.forceClose()
		}
	})
}
