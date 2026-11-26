package utils

import (
	"fmt"
	"os"
)

func (u *Utils) Uninstall() error {
	if err := os.RemoveAll(u.Files.GetStoragePath()); err != nil {
		fmt.Printf("Error removing config directory: %v\n", err)
		return err
	}
	return nil
}
