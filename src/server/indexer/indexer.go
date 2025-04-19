package indexer

import (
	"fmt"
	"haystack/server/core/storage"
	"haystack/server/core/workspace"
	"haystack/shared/types"
	"log"
	"os"
	"path/filepath"
	"sync"
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

func CreateWorkspace(workspacePath string, useGlobalFilter bool, filters *types.Filters) (*workspace.Workspace, error) {
	w, err := workspace.Create(workspacePath)
	if err != nil {
		return nil, err
	}

	w.UseGlobalFilters = useGlobalFilter
	w.Filters = filters
	w.Save()

	Sync(w)
	return w, nil
}

// SyncIfNeeded checks if a workspace needs to be synced and adds it to the scanner queue if necessary.
// A workspace needs to be synced if:
// 1. It has never been successfully synced (LastFullSync is zero)
func SyncIfNeeded(workspacePath string) error {
	workspace, _ := workspace.GetByPath(workspacePath)
	if workspace == nil {
		return fmt.Errorf("workspace not found")
	}

	if workspace.LastFullSync.IsZero() {
		return Sync(workspace)
	} else {
		log.Printf("Workspace %s is up to date, skipping", workspacePath)
	}
	return nil
}

func Sync(workspace *workspace.Workspace) error {
	return scanner.Add(workspace)
}

func AddOrSyncFile(workspace *workspace.Workspace, relPath string) error {
	fullPath := filepath.Join(workspace.Path, relPath)
	docid := GetDocumentId(fullPath)
	doc, err := storage.GetDocument(workspace.ID, docid, false)
	if err != nil {
		return err
	}

	if doc == nil {
		stat, err := os.Stat(fullPath)
		if err != nil || stat.IsDir() {
			return err
		}

		// Add new file to the parser queue
		parser.Add(workspace, relPath, true)
	} else {
		stat, err := os.Stat(fullPath)
		if err != nil || stat.IsDir() {
			// Remove the file from the index
			RemoveFile(workspace, relPath)
		} else {
			// Sync existing file to the parser queue
			parser.Add(workspace, relPath, true)
		}
	}

	return nil
}

func RemoveFile(workspace *workspace.Workspace, relPath string) error {
	fullPath := filepath.Join(workspace.Path, relPath)

	docid := GetDocumentId(fullPath)
	if err := storage.DeleteDocument(workspace.ID, docid); err != nil {
		return err
	}

	workspace.AddTotalFiles(-1)
	workspace.Save()
	return nil
}

func RefreshFilesIfNeeded(workspaceId string, docs map[string]*storage.Document) []string {
	workspace, err := workspace.Get(workspaceId)
	if err != nil {
		return []string{}
	}

	removedDocs := []string{}
	for _, doc := range docs {
		removed, _, err := RefreshFileIfNeeded(workspace, doc)
		if err != nil {
			continue
		}

		if removed {
			removedDocs = append(removedDocs, doc.ID)
		}
	}

	// Return the list of removed documents
	return removedDocs
}

func RefreshFileIfNeeded(workspace *workspace.Workspace, doc *storage.Document) (removed bool, relPath string, err error) {
	relPath, err = filepath.Rel(workspace.Path, doc.FullPath)
	if err != nil {
		return false, "", err
	}

	stat, err := os.Stat(doc.FullPath)
	// If the file becomes a directory or there is an error, remove it
	if err != nil || stat.IsDir() {
		RemoveFile(workspace, relPath)
		return true, relPath, nil
	}

	// If the file has been modified, add it to the parser queue
	if stat.ModTime().UnixNano() != doc.ModifiedTime {
		parser.Add(workspace, relPath, true)
	}

	return false, relPath, nil
}
