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
	p.ch = make(chan ParseFile, conf.Get().IndexWorkers)

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
			if err := p.processFile(id, file); err != nil {
				log.Printf("Parser %d error processing file %s: %v", id, file.FilePath, err)
			} else {
				log.Printf("Parser %d processed file %s", id, file.FilePath)
			}
		}
	}
}

// processFile handles the parsing of a single file
func (p *Parser) processFile(id int, file ParseFile) error {
	baseDir := file.Workspace.Meta.Path
	doc, err := parse(file.FilePath, baseDir)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	// TODO: Store the parsed document
	_ = doc
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
func parse(relPath string, baseDir string) (*storage.Document, error) {
	fullPath := filepath.Join(baseDir, relPath)

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return &storage.Document{
		ID:           fmt.Sprintf("%x", md5.Sum([]byte(fullPath))),
		FullPath:     fullPath,
		Size:         info.Size(),
		ModifiedTime: info.ModTime().UnixNano(),
		Hash:         fmt.Sprintf("%x", md5.Sum(content)),
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
