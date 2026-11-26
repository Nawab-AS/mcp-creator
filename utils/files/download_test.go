package files

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadFileCreatesNewTarget(t *testing.T) {
	const content = "downloaded content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(content))
	}))
	defer server.Close()

	targetPath := filepath.Join(t.TempDir(), "new-target.txt")
	if err := (&Files{}).DownloadFile(server.URL, targetPath, nil); err != nil {
		t.Fatalf("DownloadFile() error = %v", err)
	}

	got, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read downloaded target: %v", err)
	}
	if string(got) != content {
		t.Fatalf("downloaded content = %q, want %q", got, content)
	}
}
