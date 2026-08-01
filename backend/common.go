package backend

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	// "fmt"
	rt "github.com/wailsapp/wails/v2/pkg/runtime"
)

type Common struct{ ctx context.Context }

func (a *Common) Startup(ctx context.Context) { a.ctx = ctx }

type Response struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

func (a *Common) SelectDirDialog(title string) string {
	dir, err := rt.OpenDirectoryDialog(a.ctx, rt.OpenDialogOptions{
		Title: title,
	})
	if err != nil {
		return ""
	}

	return dir
}

func fsExists(path string, isDir bool) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		// Path exists. verify dir/file
		if isDir {
			return info.IsDir(), nil
		}
		return !info.IsDir(), nil
	}

	// non-existant
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}

	return false, err // Another error occurred (e.g., permission denied)
}

func GetConfigPath() (string, error) {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not get user config dir: %w", err)
	}

	appConfigDir := filepath.Join(baseDir, "mcp-creator")
	err = os.MkdirAll(appConfigDir, 0755) // if not already existing
	if err != nil {
		return "", fmt.Errorf("could not create config folder: %w", err)
	}

	return appConfigDir, nil
}

func ModelStorageDirectory(modelName string) (string, error) {
	configDirectory, err := GetConfigPath()
	if err != nil {
		return "", fmt.Errorf("get user config directory: %w", err)
	}

	err = os.MkdirAll(filepath.Join(configDirectory, "models"), 0755)
	if err != nil {
		return "", fmt.Errorf("could not create models folder: %w", err)
	}

	nameHash := sha256.Sum256([]byte(modelName))
	return filepath.Join(configDirectory, "models", fmt.Sprintf("%x", nameHash[:8])), nil
}
