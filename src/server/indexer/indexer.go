package indexer

import (
	"fmt"
	"haystack/conf"
	"haystack/server/core/storage"
	"haystack/server/core/workspace"
	"log"
	"os"
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

func RefreshIndexIfNeeded() {
	workspacePaths := workspace.GetAllPaths()
	for _, w := range workspacePaths {
		SyncIfNeeded(w)
	}
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

	if workspace.LastFullSync.IsZero() ||
		workspace.LastFullSync.Before(time.Now().Add(-time.Hour*24)) {
		if err := scanner.Add(workspace); err != nil {
			return fmt.Errorf("failed to add workspace to scanner queue: %v", err)
		}
	} else {
		log.Printf("Workspace %s is up to date, skipping", workspacePath)
	}
	return nil
}

func AddOrSyncFile(workspace *workspace.Workspace, fullPath string) error {
	docid := GetDocumentId(fullPath)
	doc, err := storage.GetDocument(workspace.ID, docid, false)
	if err != nil {
		return err
	}

	if doc == nil {
		stat, _ := os.Stat(fullPath)
		if !stat.IsDir() && stat.Size() < conf.Get().Server.MaxFileSize {
			// Add new file to the parser queue
			parser.Add(workspace, fullPath)
		}
	} else {
		stat, err := os.Stat(fullPath)
		if err != nil || stat.IsDir() || stat.Size() > conf.Get().Server.MaxFileSize {
			// Remove the file from the index
			RemoveFile(workspace, doc.FullPath)
		} else {
			// Sync existing file to the parser queue
			parser.Add(workspace, fullPath)
		}
	}

	return nil
}

func RemoveFile(workspace *workspace.Workspace, filePath string) error {
	docid := GetDocumentId(filePath)

	storage.DeleteDocument(workspace.ID, docid)
	return nil
}

func RefreshFileIfNeeded(workspaceId string, docs map[string]*storage.Document) []string {
	workspace, err := workspace.Get(workspaceId)
	if err != nil {
		return []string{}
	}

	removedDocs := []string{}
	for _, doc := range docs {
		stat, err := os.Stat(doc.FullPath)

		// If the file becomes a directory or there is an error, remove it
		if err != nil || stat.IsDir() {
			RemoveFile(workspace, doc.FullPath)
			removedDocs = append(removedDocs, doc.ID)
			continue
		}

		// If the file has been modified, add it to the parser queue
		if stat.ModTime().UnixNano() != doc.ModifiedTime {
			parser.Add(workspace, doc.FullPath)
		}
	}

	// Return the list of removed documents
	return removedDocs
}
