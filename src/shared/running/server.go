package running

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
)

func CheckAndLockServer(lockFile string) (func(), error) {
	// Ensure the directory exists
	dir := filepath.Dir(lockFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create lock directory: %v", err)
	}

	// Create a new file lock
	fileLock := flock.New(lockFile)

	// Try to acquire an exclusive lock
	locked, err := fileLock.TryLock()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %v", err)
	}
	if !locked {
		return nil, fmt.Errorf("another instance is already running")
	}

	// Return cleanup function
	cleanup := func() {
		fileLock.Unlock()
		os.Remove(lockFile)
	}

	return cleanup, nil
}
