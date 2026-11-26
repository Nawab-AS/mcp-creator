package backend

import (
	"context"
	"fmt"
	"mcp-creator/backend/mcpserver"
	"mcp-creator/utils"
	"os/exec"
	"runtime"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

var Utils *utils.Utils
var Events *application.EventManager

type Backend struct {
	app     *application.App // wails app instance
	servers map[string]*mcpserver.Server
}

func (b *Backend) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
	Events = b.app.Event
	b.registerCloseHandler()

	// Start onnxruntime
	go func() {
		if err := mcpserver.OnnxStartup(Utils); err != nil {
			fmt.Printf("Error during ONNX Runtime startup: %v\n", err)
			b.ShowInfoDialog(err.Error())
			return
		}
		b.InitProjects()
	}()

	return nil
}

func (b *Backend) ServiceShutdown() {
	for name, server := range b.servers {
		if err := server.Close(); err != nil {
			fmt.Printf("Error stopping project %s: %v\n", name, err)
		}
	}
}

// used from frontend (it doesn't have authority to set the clipboard directly)
func (b *Backend) CopyToClipboard(text string) {
	b.app.Clipboard.SetText(text)
}

// used from ./mcp-server (it does not have access to the wails context)
func (b *Backend) ShowInfoDialog(message string) {
	b.app.Dialog.Info().
		SetTitle("MCP Creator").
		SetMessage(message).
		Show()
}

func (b *Backend) Uninstall() {
	// Stop all servers
	for name, server := range b.servers {
		if err := server.Close(); err != nil {
			fmt.Printf("Error stopping project %s: %v\n", name, err)
		}
	}

	b.servers = map[string]*mcpserver.Server{} // Clear the servers map (stops any duplicate closing)
	projects = nil                             // TODO: review this

	err := Utils.Uninstall()
	if err != nil {
		b.ShowInfoDialog("Error during uninstall: " + err.Error())
	}

	b.forceClose()
}

func (b *Backend) OpenStorageDir() {
	configPath := Utils.Files.GetStoragePath()

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", configPath)
	case "windows":
		// Explorer requires backslashes for local paths on Windows
		configPath = strings.ReplaceAll(configPath, "/", "\\")
		cmd = exec.Command("explorer", configPath)
	case "linux":
		cmd = exec.Command("xdg-open", configPath)
	default:
		cmd = exec.Command("xdg-open", configPath)
	}

	cmd.Start()
}

func (b *Backend) SelectDirDialog(title string) string {
	dir, err := b.app.Dialog.OpenFile().
		SetTitle(title).
		CanChooseDirectories(true).
		CanChooseFiles(false).
		PromptForSingleSelection()
	if err != nil {
		return ""
	}

	return dir
}
