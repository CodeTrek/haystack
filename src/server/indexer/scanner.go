package indexer

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"search-indexer/running"
	"search-indexer/server/core/workspace"
	"search-indexer/utils"
	fsutils "search-indexer/utils/fs"
	gitutils "search-indexer/utils/git"
	"sync"
	"time"
)

// Scanner represents a file system scanner that processes workspaces in a queue.
// It is responsible for scanning files in workspaces and applying appropriate filters.
type Scanner struct {
	current *workspace.Workspace
	queue   *list.List
	mu      sync.RWMutex
}

// NewScanner creates a new Scanner instance.
func NewScanner() *Scanner {
	return &Scanner{
		queue: list.New(),
	}
}

// Start begins the scanning process in a goroutine.
// It will continue running until the application is shutting down.
func (s *Scanner) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go s.run(wg)
}

// run is the main scanning loop that processes workspaces from the queue.
func (s *Scanner) run(wg *sync.WaitGroup) {
	defer wg.Done()

	ctx := context.Background()
	for !running.IsShuttingDown() {
		workspace := s.tryPopJob()
		if workspace == nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(500 * time.Millisecond):
				continue
			}
		}

		s.setCurrent(workspace)
		if err := s.scan(workspace); err != nil {
			log.Printf("Error scanning workspace %s: %v", workspace.Meta.Path, err)
		}
		s.setCurrent(nil)
	}

	log.Println("Scanner stopped")
}

// scan processes a single workspace by scanning its files and applying filters.
func (s *Scanner) scan(w *workspace.Workspace) error {
	baseDir := w.Meta.Path
	filters := w.GetFilters()

	var filter fsutils.ListFileFilter
	if filters.Exclude.UseGitIgnore {
		filter = &GitIgnoreFilter{
			ignore: gitutils.NewGitIgnore(baseDir),
		}
	} else {
		filter = utils.NewSimpleFilterExclude(filters.Exclude.Customized, baseDir)
	}

	return fsutils.ListFiles(baseDir, fsutils.ListFileOptions{Filter: filter}, func(fileInfo fsutils.FileInfo) bool {
		log.Println(fileInfo.Path)
		return !running.IsShuttingDown()
	})
}

// tryPopJob attempts to remove and return the next workspace from the queue.
// Returns nil if the queue is empty.
func (s *Scanner) tryPopJob() *workspace.Workspace {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.queue.Len() > 0 {
		job := s.queue.Remove(s.queue.Front())
		return job.(*workspace.Workspace)
	}
	return nil
}

// setCurrent updates the current workspace being processed.
func (s *Scanner) setCurrent(w *workspace.Workspace) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.current = w
}

// Add adds a workspace to the scanning queue.
func (s *Scanner) Add(w *workspace.Workspace) error {
	if w == nil {
		return fmt.Errorf("cannot add nil workspace to scanner queue")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.queue.PushBack(w)
	return nil
}
