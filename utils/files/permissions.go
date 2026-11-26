package files

import "os"

// permissions.go - File permission utility functions
// Example usage:
//
//	func someFunction() {
//		mode := "dir" or "file" or "script"
//		permissions := files.GetPermissions("dir", true)
//		fmt.Printf("Permissions for sensitive directory: %o\n", permissions)
//	}

func (f *Files) GetPermissions(filetype string, sensitive bool) os.FileMode {
	resolveSensitive := func(sensitiveValue int, nonSensitiveValue int) os.FileMode {
		if sensitive {
			return os.FileMode(sensitiveValue)
		}
		return os.FileMode(nonSensitiveValue)
	}

	switch filetype {
	case "dir":
		// 0700 = only owner can read/write/traverse
		// 0755 = owner can read/write/traverse, others can read/traverse
		return resolveSensitive(0700, 0755)

	case "file":
		// 0600 = only owner can read/write
		// 0644 = owner can read/write, others can read
		return resolveSensitive(0600, 0644)
	case "script":
		// 0700 = only owner can read/write/execute
		// 0755 = owner can read/write/execute, others can read/execute
		return resolveSensitive(0700, 0755)
	}

	// fallback
	// 0755 = owner can read/write/traverse/execute, others can read/traverse/execute
	return 0755
}
