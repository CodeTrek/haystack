package indexer

import (
	"fmt"
	"haystack/conf"
	"haystack/server/core/storage"
	"haystack/server/core/workspace"
	"haystack/shared/running"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// ParseFile represents a file to be parsed
type ParseFile struct {
	Workspace *workspace.Workspace
	FilePath  string
	Included  bool
}

// Parser handles concurrent file parsing operations
type Parser struct {
	ch chan ParseFile
}

// NewParser creates a new Parser instance
func NewParser() *Parser {
	return &Parser{}
}

// Start initializes the parser with worker goroutines
func (p *Parser) Start(wg *sync.WaitGroup) {

	p.ch = make(chan ParseFile, 32)

	for i := 0; i < conf.Get().Server.IndexWorkers; i++ {
		wg.Add(1)
		go p.run(i, wg)
	}
}

// run executes the parsing logic in a worker goroutine
func (p *Parser) run(id int, wg *sync.WaitGroup) {
	log.Printf("Parser %d started", id)
	defer wg.Done()
	defer log.Printf("Parser %d stopped", id)

	for {
		select {
		case <-running.GetShutdown().Done():
			return
		case file := <-p.ch:
			p.processFile(file)
		}
	}
}

// processFile handles the parsing of a single file
func (p *Parser) processFile(file ParseFile) error {
	doc, newDoc, err := parse(file)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	// If the document is nil, it means the file has not changed
	if doc == nil {
		return nil
	}

	writer.Add(file.Workspace, doc, newDoc)

	file.Workspace.AddIndexingFiles(1)

	return nil
}

// Add queues a file for parsing
func (p *Parser) Add(workspace *workspace.Workspace, relPath string, included bool) {
	p.ch <- ParseFile{
		Workspace: workspace,
		FilePath:  relPath,
		Included:  included,
	}
}

// parse reads and processes a file, returning a Document
func parse(file ParseFile) (*storage.Document, bool, error) {
	fullPath := filepath.Join(file.Workspace.Path, file.FilePath)
	id := GetDocumentId(fullPath)

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, false, fmt.Errorf("failed to stat file: %w", err)
	}

	fileSizeExceedLimit := info.Size() > conf.Get().Server.MaxFileSize
	if fileSizeExceedLimit {
		log.Printf("File %s (%.2f MiB) is too large to index, skipping", file.FilePath, float64(info.Size())/1024/1024)
	}

	existing, _ := storage.GetDocument(file.Workspace.ID, id, false)
	// If the document exists and the modified time is the same, return nil
	if existing != nil && existing.ModifiedTime == info.ModTime().UnixNano() && existing.Included == file.Included {
		return nil, false, nil
	}

	var hash string
	var words []string
	if !file.Included || fileSizeExceedLimit {
		hash = ""
		words = []string{}
	} else {
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, false, fmt.Errorf("failed to read file: %w", err)
		}

		hash := GetContentHash(content)
		// If the document exists and the hash is the same, return nil
		if existing != nil && existing.Hash == hash {
			return nil, false, nil
		}

		// We only index the content if the file size is below the limit
		words = parseString(string(content))
	}

	return &storage.Document{
		ID:           id,
		RelPath:      file.FilePath,
		FullPath:     fullPath,
		Size:         info.Size(),
		ModifiedTime: info.ModTime().UnixNano(),
		LastSyncTime: time.Now().UnixNano(),
		Included:     file.Included,
		Hash:         hash,
		Words:        words,
		PathWords:    parseString(file.FilePath),
	}, existing == nil, nil
}

var re = regexp.MustCompile(`[a-zA-Z0-9_][a-zA-Z0-9_-]+`)

// parseString extracts unique words from a string
func parseString(str string) []string {
	words := re.FindAllString(str, -1)

	uniqueWords := make(map[string]struct{})
	for _, word := range words {
		if isValidWord(word) {
			uniqueWords[strings.ToLower(word)] = struct{}{}
		}
	}

	result := make([]string, 0, len(uniqueWords))
	for word := range uniqueWords {
		result = append(result, word)
	}

	sort.Strings(result)
	return result
}

// isValidWord checks if a word meets the criteria for inclusion
func isValidWord(word string) bool {
	if len(word) < 3 || len(word) > 80 {
		return false
	}

	for _, r := range word {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
			return false
		}
	}
	return true
}
