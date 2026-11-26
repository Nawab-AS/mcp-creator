package files

import (
	"encoding/json"
	"os"
)

// write-struct.go - File writing utility functions
// Example usage:
//
//	func someFunction() {
//		data := MyStruct{Field1: "value1", Field2: 42}
//		err := files.WriteStructToFile(data, "/absolute/path/to/file.json")
//		if err != nil {
//			fmt.Println("Error writing struct to file:", err)
//		} else {
//			fmt.Println("Struct written to file successfully")
//		}
//	}

func (f *Files) WriteStructToFile(data any, filePath string) error {
	mutex := f.Mutex(filePath).Lock()
	defer mutex.Unlock()

	file, err := os.Create(f.GetFileStoragePath(filePath))
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(data)
	if err != nil {
		return err
	}
	return nil
}
