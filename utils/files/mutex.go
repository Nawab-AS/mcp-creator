package files

import "sync"

// fs_mutex.go - File mutex system without needing a variable for each file
//
// Example usage (preferred way):
//
//		func someFunction() {
//			mutex := FileMutex("test.txt").Lock()
//			defer mutex.Unlock()
//			// Do something with the locked file
//		}
//
// Example usage (manual unlocking, only use if you REALLY need to):
//
//		func someFunction() {
//			FileMutex("test.txt").Lock()
//			// Do something with the locked file
//			FileMutex("test.txt").Unlock()
//		}

// FileMutex is a simple wrapper around sync.Mutex to allow .Lock() to return the
// mutex itself for deferred unlocking or .Unlock() for manual unlocking.
type FileLock struct {
	mutex *sync.Mutex
}

func (fl *FileLock) Lock() *FileLock {
	fl.mutex.Lock()
	return fl
}

func (fl *FileLock) Unlock() {
	fl.mutex.Unlock()
}

var mutexMap = make(map[string]*FileLock)
var lockingMutex sync.Mutex // Mutex to protect access to the mutexMap

// FileMutex returns a mutex for the given file path.
func (f *Files) Mutex(filePath string) *FileLock {
	lockingMutex.Lock()
	defer lockingMutex.Unlock()

	mutex, exists := mutexMap[filePath]
	if !exists {
		mutex = &FileLock{mutex: &sync.Mutex{}}
		mutexMap[filePath] = mutex
	}

	return mutex
}
