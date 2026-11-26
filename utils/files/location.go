package files

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

// TODO: fix docs

// location.go - File location utility functions
// Example usage:
//
//	func someFunction() {
//		storagePath := files.GetStoragePath()
//		fmt.Println("Storage path:", storagePath)
//	}
//
//	func someFunction() {
//		subDirPath := files.GetStoragePath("my_subdir", "another_subdir")
//		fmt.Println("Subdirectory path:", subDirPath)
//	}

// GetStorageSubDir returns the path to a subdirectory within the storage directory, creating it if it doesn't exist.
func (f *Files) GetStoragePath(path ...string) string {
	// error ignored because main.go already checks (or panics)
	configDir, _ := os.UserConfigDir()

	path = append([]string{configDir, "mcp-creator"}, path...)
	joinedPath := filepath.Join(path...)

	err := os.MkdirAll(joinedPath, f.GetPermissions("dir", false))
	if err != nil {
		panic(fmt.Sprintf("Failed to create storage directory %s: %v", joinedPath, err))
	}

	return joinedPath
}

func (f *Files) GetFileStoragePath(path ...string) string {
	if len(path) == 0 {
		return f.GetStoragePath()
	}

	parentDir := f.GetStoragePath(path[:len(path)-1]...)
	return filepath.Join(parentDir, path[len(path)-1])
}

func Hash(input string) string {
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash[:8])
}

func (f *Files) ProjectDBPath(projectName string) string {
	configDirectory := f.GetStoragePath("projects")

	return filepath.Join(configDirectory, fmt.Sprintf("%s.db", Hash(projectName)))
}

func (f *Files) ModelStorageDirectory(modelName string) string {
	configDirectory := f.GetStoragePath("projects")

	return filepath.Join(configDirectory, "models", Hash(modelName))
}
