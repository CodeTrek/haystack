package indexer

import (
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"search-indexer/running"
	"search-indexer/server/conf"
	"search-indexer/server/core/storage"
	"search-indexer/server/core/workspace"
	"sort"
	"sync"
	"time"
)

// ParseFile represents a file to be parsed
type ParseFile struct {
	Workspace *workspace.Workspace
	FilePath  string
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

	for i := 0; i < conf.Get().IndexWorkers; i++ {
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
	doc, err := parse(file)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	// If the document is nil, it means the file has not changed
	if doc == nil {
		return nil
	}

	writer.Add(file.Workspace, doc)
	return nil
}

// Add queues a file for parsing
func (p *Parser) Add(workspace *workspace.Workspace, filePath string) {
	p.ch <- ParseFile{
		Workspace: workspace,
		FilePath:  filePath,
	}
}

// parse reads and processes a file, returning a Document
func parse(file ParseFile) (*storage.Document, error) {
	fullPath := filepath.Join(file.Workspace.Meta.Path, file.FilePath)

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	id := fmt.Sprintf("%x", md5.Sum([]byte(fullPath)))

	existing, _ := storage.GetDocument(file.Workspace.Meta.ID, id, false)

	// If the document exists and the modified time is the same, return nil
	if existing != nil && existing.ModifiedTime == info.ModTime().UnixNano() {
		return nil, nil
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	hash := fmt.Sprintf("%x", md5.Sum(content))

	// If the document exists and the hash is the same, return nil
	if existing != nil && existing.Hash == hash {
		return nil, nil
	}

	return &storage.Document{
		ID:           id,
		FullPath:     fullPath,
		Size:         info.Size(),
		ModifiedTime: info.ModTime().UnixNano(),
		LastSyncTime: time.Now().UnixNano(),
		Hash:         hash,
		Words:        parseString(string(content)),
	}, nil
}

// parseString extracts unique words from a string
func parseString(str string) []string {
	re := regexp.MustCompile(`[a-zA-Z0-9]+`)
	words := re.FindAllString(str, -1)

	uniqueWords := make(map[string]struct{})
	for _, word := range words {
		if isValidWord(word) {
			uniqueWords[word] = struct{}{}
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
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}
