package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetFileStoragePathCreatesParentAndReturnsFilePath(t *testing.T) {
	tempFile, err := os.CreateTemp(t.TempDir(), "target-")
	if err != nil {
		t.Fatal(err)
	}
	filename := filepath.Base(tempFile.Name())
	if err := tempFile.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(tempFile.Name()); err != nil {
		t.Fatal(err)
	}

	files := &Files{}
	singleFilePath := files.GetFileStoragePath(filename)
	if want := filepath.Join(files.GetStoragePath(), filename); singleFilePath != want {
		t.Fatalf("GetFileStoragePath(filename) = %q, want %q", singleFilePath, want)
	}
	if _, err := os.Stat(singleFilePath); !os.IsNotExist(err) {
		t.Fatalf("expected single-file path not to exist yet, stat error = %v", err)
	}

	parentName := "test-" + filename
	t.Cleanup(func() {
		if err := os.RemoveAll(filepath.Join(files.GetStoragePath(), parentName)); err != nil {
			t.Errorf("remove test storage directory: %v", err)
		}
	})

	got := files.GetFileStoragePath(parentName, filename)
	want := filepath.Join(files.GetStoragePath(parentName), filename)
	if got != want {
		t.Fatalf("GetFileStoragePath() = %q, want %q", got, want)
	}
	if _, err := os.Stat(got); !os.IsNotExist(err) {
		t.Fatalf("expected file path not to exist yet, stat error = %v", err)
	}
}

func TestProjectDBPath(t *testing.T) {
	files := &Files{}

	got := files.ProjectDBPath("test-project")
	want := filepath.Join(files.GetStoragePath("projects"), Hash("test-project")+".db")
	if got != want {
		t.Fatalf("ProjectDBPath() = %q, want %q", got, want)
	}
}
