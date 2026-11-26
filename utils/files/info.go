package files

import (
	"os"
)

// info.go - File information utility functions
// Example usage:
//
//	func someFunction() {
//		isWritable := files.IsFileWritable("/absolute/path/to/file.txt")
//		if isWritable {
//			fmt.Println("File is writable")
//		} else {
//			fmt.Println("File is not writable")
//		}
//	}
//
// files.IsDirWritable is used in the exact same way, but checks for directories instead of files.
//
// 	func someFunction() {
//		exists := files.FileExists("/absolute/path/to/file.txt")
//		if exists {
//			fmt.Println("File exists")
//		} else {
//			fmt.Println("File does not exist")
//		}
//	}

// IsDirWritable checks if targetPath exists, is a directory, and allows creating/deleting files.
func (f *Files) IsDirWritable(targetPath string) bool {
	info, err := os.Stat(targetPath)
	if err != nil || !info.IsDir() {
		return false
	}

	tmpFile, err := os.CreateTemp(targetPath, ".write_test_*")
	if err != nil {
		return false
	}

	closeErr := tmpFile.Close()
	removeErr := os.Remove(tmpFile.Name())

	return closeErr == nil && removeErr == nil
}

// IsFileWritable checks if targetPath exists, is a regular file, and can be opened for writing.
func (f *Files) IsFileWritable(targetPath string) bool {
	info, err := os.Stat(targetPath)
	if err != nil || info.IsDir() {
		return false
	}

	file, err := os.OpenFile(targetPath, os.O_WRONLY, 0)
	if err != nil {
		return false
	}

	return file.Close() == nil
}

// FileExists checks if the given path exists and is a file.
func (f *Files) Exists(path string) (bool, error) {
	// check if path is a file
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}
