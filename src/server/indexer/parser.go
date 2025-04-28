package indexer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/codetrek/haystack/conf"
	"github.com/codetrek/haystack/server/core/fulltext"
	"github.com/codetrek/haystack/server/core/workspace"
)

// ParseFile represents a file to be parsed
type ParseFile struct {
	Workspace   *workspace.Workspace
	RelFilePath string
}

// Parser handles concurrent file parsing operations
type Parser struct {
	ch   chan ParseFile
	stop chan struct{}
	done chan struct{}
}

// NewParser creates a new Parser instance
func NewParser() *Parser {
	return &Parser{
		ch:   make(chan ParseFile, 32),
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
}

// Start initializes the parser with worker goroutines
func (p *Parser) Start(wg *sync.WaitGroup) {
	for i := 0; i < conf.Get().Server.IndexWorkers; i++ {
		wg.Add(1)
		go p.run(i, wg)
	}
}

func (p *Parser) Stop() {
	close(p.stop)
	for range conf.Get().Server.IndexWorkers {
		<-p.done
	}
	close(p.done)
	defer log.Printf("Parser stopped")
}

// run executes the parsing logic in a worker goroutine
func (p *Parser) run(id int, wg *sync.WaitGroup) {
	log.Printf("Parser %d started", id)
	defer wg.Done()

	for {
		select {
		case <-p.stop:
			p.done <- struct{}{}
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
func (p *Parser) Add(workspace *workspace.Workspace, relPath string) {
	p.ch <- ParseFile{
		Workspace:   workspace,
		RelFilePath: relPath,
	}
}

// parse reads and processes a file, returning a Document
func parse(file ParseFile) (*fulltext.Document, bool, error) {
	fullPath := filepath.Join(file.Workspace.Path, file.RelFilePath)
	id := GetDocumentId(fullPath)

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, false, fmt.Errorf("failed to stat file: %w", err)
	}

	fileSizeExceedLimit := info.Size() > conf.Get().Server.MaxFileSize
	if fileSizeExceedLimit {
		log.Printf("File `%s` (%.2f MiB) is too large to index, skipping", file.RelFilePath, float64(info.Size())/1024/1024)
	}

	existing, _ := fulltext.GetDocument(file.Workspace.ID, id, false)
	// If the document exists and the modified time is the same, return nil
	if existing != nil &&
		existing.ModifiedTime == info.ModTime().UnixNano() {
		return nil, false, nil
	}

	var hash string
	var words []string
	if fileSizeExceedLimit {
		hash = ""
		words = []string{}
	} else {
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, false, fmt.Errorf("failed to read file: %w", err)
		}
		if !IsLikelyText(content) {
			log.Printf("File `%s` is not a text file, skipping", file.RelFilePath)
			return nil, false, nil
		}

		hash := GetContentHash(content)
		// If the document exists and the hash is the same, return nil
		if existing != nil && existing.Hash == hash {
			return nil, false, nil
		}

		// We only index the content if the file size is below the limit
		words = parseString(string(content))
	}

	return &fulltext.Document{
		ID:           id,
		RelPath:      file.RelFilePath,
		Size:         info.Size(),
		ModifiedTime: info.ModTime().UnixNano(),
		LastSyncTime: time.Now().UnixNano(),
		Hash:         hash,
		Words:        words,
		PathWords:    parseString(file.RelFilePath),
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
