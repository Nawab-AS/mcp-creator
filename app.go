package main

import (
	"context"
	"mcp-creator/backend"
)

// App struct
type App struct {
	ctx context.Context
	*backend.Models
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		Models: &backend.Models{},
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	a.Models.Startup(ctx)
}
