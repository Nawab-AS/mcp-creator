package backend

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type Models struct{ ctx context.Context }

func (a *Models) Startup(ctx context.Context) { a.ctx = ctx }

// TODO: make this actually download a model
func (a *Models) DownloadModel(modelName string) string {
	fmt.Println("Downloading model:", modelName)

	dir, _ := os.UserConfigDir()
	filePath := filepath.Join(dir, "mcp-creator", "tasks.json")

	fmt.Println("Current OS data path is", filePath)
	return "Downloaded model: " + modelName
}

type Model struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Installed   bool   `json:"installed"`
	Path        string `json:"path,omitempty"`
}

func (a *Models) GetModels() []Model {
	return []Model{
		{
			Name:        "slow",
			Description: "Fast on older hardware and simple datasets.",
			Installed:   true,
			Path:        "path/to/slow",
		},
		{
			Name:        "balanced",
			Description: "Good for modern hardware and balanced performance.",
			Installed:   false,
		},
		{
			Name:        "accurate",
			Description: "Good on powerful hardware and complex datasets.",
			Installed:   false,
		},
		{
			Name:        "fast",
			Description: "Fast on modern hardware and simple datasets.",
			Installed:   true,
			Path:        "path/to/fast",
		},
	}
}
