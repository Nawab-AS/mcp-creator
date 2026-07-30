package backend

import (
	"context"
	"errors"
	"io/fs"
	"os"

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
