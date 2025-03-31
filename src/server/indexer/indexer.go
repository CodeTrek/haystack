package indexer

import (
	"fmt"
	"log"
	"os"
	"search-indexer/server/core/storage"
	"search-indexer/server/core/workspace"
	"sync"
	"time"
)

var (
	scanner = NewScanner()
	parser  = NewParser()
	writer  = NewWriter()
)

// Run starts the indexer components in separate goroutines.
func Run(wg *sync.WaitGroup) {
	log.Println("Starting indexer...")

	scanner.Start(wg)
	parser.Start(wg)
	writer.Start(wg)

	log.Println("Indexer started.")
}

// SyncIfNeeded checks if a workspace needs to be synced and adds it to the scanner queue if necessary.
// A workspace needs to be synced if:
// 1. It has never been synced (LastFullSync is zero)
// 2. It was last synced more than 24 hours ago
func SyncIfNeeded(workspacePath string) error {
	workspace, err := workspace.GetOrCreate(workspacePath)
	if err != nil {
		return fmt.Errorf("failed to get or create workspace: %v", err)
	}

	if workspace.Meta.LastFullSync.IsZero() ||
		workspace.Meta.LastFullSync.Before(time.Now().Add(-time.Hour*24)) {
		if err := scanner.Add(workspace); err != nil {
			return fmt.Errorf("failed to add workspace to scanner queue: %v", err)
		}
	} else {
		log.Printf("Workspace %s is up to date, skipping", workspacePath)
	}
	return nil
}

func SyncFile(workspace *workspace.Workspace, filePath string) error {
	parser.Add(workspace, filePath)
	return nil
}

func RemoveFile(workspace *workspace.Workspace, filePath string) error {
	// TODO: Implement
	return nil
}

func RefreshFileIfNeeded(workspaceId string, docs map[string]*storage.Document) error {
	workspace, err := workspace.Get(workspaceId)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %v", err)
	}

	for _, doc := range docs {
		stat, err := os.Stat(doc.FullPath)
		if stat.IsDir() || err != nil {
			RemoveFile(workspace, doc.FullPath)
			continue
		}

		parser.Add(workspace, doc.FullPath)
	}
	return nil
}
