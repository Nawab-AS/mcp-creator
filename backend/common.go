package backend

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mcp-creator/backend/mcpserver"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	// "fmt"
	rt "github.com/wailsapp/wails/v2/pkg/runtime"
)

var FS_MUTEX sync.Mutex

type Common struct {
	ctx         context.Context
	onnxStartup sync.Once
}

func (a *Common) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *Common) StartBackend() {
	a.onnxStartup.Do(func() {
		go func() {
			if err := mcpserver.OnnxStartup(a); err != nil {
				fmt.Printf("Error during ONNX Runtime startup: %v\n", err)
			}
			a.InitProjects()
		}()
	})
}

func (a *Common) CopyToClipboard(text string) {
	rt.ClipboardSetText(a.ctx, text)
}

func (a *Common) EmitEvent(eventName string, data ...interface{}) {
	rt.EventsEmit(a.ctx, eventName, data...)
}

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

func FsExists(path string, isDir bool) (bool, error) {
	return (&Common{}).FsExists(path, isDir)
}
func (m *Common) FsExists(path string, isDir bool) (bool, error) {
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
	return (&Common{}).GetConfigPath()
}
func (m *Common) GetConfigPath() (string, error) {
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
	return (&Common{}).ModelStorageDirectory(modelName)
}
func (m *Common) ModelStorageDirectory(modelName string) (string, error) {
	configDirectory, err := m.GetConfigPath()
	if err != nil {
		return "", fmt.Errorf("get user config directory: %w", err)
	}

	err = os.MkdirAll(filepath.Join(configDirectory, "models"), 0755)
	if err != nil {
		return "", fmt.Errorf("could not create models folder: %w", err)
	}

	return filepath.Join(configDirectory, "models", m.Hash(modelName)), nil
}

func (m *Common) Hash(input string) string {
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash[:8])
}

func ProjectDBPath(projectName string) (string, string, error) {
	return (&Common{}).ProjectDBPath(projectName)
}
func (m *Common) ProjectDBPath(projectName string) (string, string, error) {
	configDirectory, err := m.GetConfigPath()
	if err != nil {
		return "", "", fmt.Errorf("get user config directory: %w", err)
	}

	err = os.MkdirAll(filepath.Join(configDirectory, "projects"), 0755)
	if err != nil {
		return "", "", fmt.Errorf("could not create projects folder: %w", err)
	}

	return filepath.Join(configDirectory, "projects"), fmt.Sprintf("%x.db", m.Hash(projectName)), nil
}

// Download file from source URL
var modelDownloadClient = &http.Client{Timeout: 10 * time.Minute}

type countingWriter struct {
	writer  io.Writer
	written atomic.Int32
}

const progressUpdateInterval = 250 * time.Millisecond

func (w *countingWriter) Write(data []byte) (int, error) {
	n, err := w.writer.Write(data)
	w.written.Add(int32(n))
	return n, err
}

func DownloadFile(sourceURL string, targetPath string, onProgress func(float32)) error {
	return (&Common{}).DownloadFile(sourceURL, targetPath, onProgress)
}
func (m *Common) DownloadFile(sourceURL string, targetPath string, onProgress func(float32)) error {
	request, err := http.NewRequest(http.MethodGet, sourceURL, nil)
	if err != nil {
		return fmt.Errorf("create download request: %w", err)
	}

	response, err := modelDownloadClient.Do(request)
	if err != nil {
		return fmt.Errorf("download %s: %w", sourceURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("download %s: unexpected status %s", sourceURL, response.Status)
	}

	file, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", targetPath, err)
	}

	writer := &countingWriter{writer: file}
	stopProgress := make(chan struct{})
	progressDone := make(chan struct{})
	go func() {
		defer close(progressDone)
		ticker := time.NewTicker(progressUpdateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if response.ContentLength > 0 {
					onProgress(float32(writer.written.Load()) / float32(response.ContentLength))
				}
			case <-stopProgress:
				return
			}
		}
	}()

	_, copyErr := io.Copy(writer, response.Body)
	close(stopProgress)
	<-progressDone
	closeErr := file.Close()
	if copyErr != nil {
		return fmt.Errorf("write %s: %w", targetPath, copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close %s: %w", targetPath, closeErr)
	}

	onProgress(1)
	return nil
}
