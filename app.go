package main

import (
	"context"
	"mcp-creator/backend"
)

// App struct
type App struct {
	ctx context.Context
	*backend.Projects
	*backend.Models
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		Projects: &backend.Projects{},
		Models:   &backend.Models{},
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	a.Projects.Startup(ctx)
	a.Models.Startup(ctx)
}
