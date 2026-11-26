package main

import (
	"embed"
	"fmt"
	"mcp-creator/backend"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if err := createStorageDir(); err != nil {
		panic(fmt.Errorf("failed to create storage directory: %w", err))
	}

	// Create an instance of the app structure
	app := application.New(application.Options{
		Name: "mcp-creator",
		Services: []application.Service{
			application.NewService(&backend.Backend{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:         "MCP-Creator",
		Width:         800,
		Height:        500,
		DisableResize: true,
	})

	err := app.Run()
	if err != nil {
		panic(err)
	}
}

func createStorageDir() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("could not get user config dir: %w", err)
	}

	appConfigDir := filepath.Join(configDir, "mcp-creator")
	err = os.MkdirAll(appConfigDir, 0755) // if not already existing
	if err != nil {
		return fmt.Errorf("could not create config folder: %w", err)
	}

	return nil
}
