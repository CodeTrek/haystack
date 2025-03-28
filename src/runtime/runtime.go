package runtime

import (
	"flag"
	"log"
	"os"
	"path/filepath"
)

var (
	rootPath   string
	serverMode = flag.Bool("server", false, "Run in server mode")
)

func Init() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get user's home directory: %v", err)
		return err
	}

	rootPath = filepath.Join(homeDir, ".search-indexer")
	if err := os.Mkdir(rootPath, 0755); err != nil {
		if !os.IsExist(err) {
			log.Fatalf("Failed to create data directory: %v", err)
			return err
		}
	}

	return nil
}

func RootPath() string {
	return rootPath
}

func IsServerMode() bool {
	return *serverMode
}

func DefaultListenPort() int {
	return 13134
}

var serverConf *string

func ServerConf() string {
	if serverConf == nil {
		search := []string{
			"./server.yaml",
			"./server.sample.yaml",
			filepath.Join(rootPath, "server.yaml"),
		}

		for _, path := range search {
			serverConf = &path
			if _, err := os.Stat(path); err == nil {
				break
			}
		}
	}

	return *serverConf
}

func ClientConf() string {
	return filepath.Join(rootPath, "client.yaml")
}
