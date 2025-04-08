package running

import (
	"flag"
	"log"
	"os"
	"path/filepath"
)

var (
	userHomeDir string
	serverMode  = flag.Bool("server", false, "Run in server mode")
)

func Init() error {
	var err error
	userHomeDir, err = os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get user's home directory: %v", err)
		return err
	}

	return nil
}

func UserHomeDir() string {
	return userHomeDir
}

func IsServerMode() bool {
	return *serverMode
}

func ExecutableName() string {
	return filepath.Base(Executable())
}

func Executable() string {
	executable, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
		return ""
	}
	return executable
}
