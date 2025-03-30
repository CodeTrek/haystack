package indexer

import (
	"fmt"
	"log"
	"search-indexer/server/core/workspace"
	"sync"
	"time"
)

var scanner Scanner
var parser Parser
var writer Writer

func Run(wg *sync.WaitGroup) {
	fmt.Println("Starting indexer...")

	scanner.start(wg)
	parser.start(wg)
	writer.start(wg)

	fmt.Println("Indexer started.")
}

func SyncIfNeeded(path string) {
	workspace, err := workspace.GetOrCreate(path)
	if err != nil {
		log.Fatalf("Failed to get or create workspace: %v", err)
	}

	if workspace.Meta.LastFullSync.IsZero() ||
		workspace.Meta.LastFullSync.Before(time.Now().Add(-time.Hour*24)) {
		scanner.add(workspace)
	}
}
