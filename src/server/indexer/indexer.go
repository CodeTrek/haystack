package indexer

import (
	"fmt"
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
	fmt.Println("Starting indexer...")

	scanner.Start(wg)
	parser.Start(wg)
	writer.Start(wg)

	fmt.Println("Indexer started.")
}

// SyncIfNeeded checks if a workspace needs to be synced and adds it to the scanner queue if necessary.
// A workspace needs to be synced if:
// 1. It has never been synced (LastFullSync is zero)
// 2. It was last synced more than 24 hours ago
func SyncIfNeeded(path string) error {
	workspace, err := workspace.GetOrCreate(path)
	if err != nil {
		return fmt.Errorf("failed to get or create workspace: %v", err)
	}

	if workspace.Meta.LastFullSync.IsZero() ||
		workspace.Meta.LastFullSync.Before(time.Now().Add(-time.Hour*24)) {
		if err := scanner.Add(workspace); err != nil {
			return fmt.Errorf("failed to add workspace to scanner queue: %v", err)
		}
	}
	return nil
}
