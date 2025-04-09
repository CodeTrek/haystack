package indexer

import (
	"container/list"
	"fmt"
	"haystack/server/core/workspace"
	"haystack/shared/running"
	"haystack/utils"
	fsutils "haystack/utils/fs"
	gitutils "haystack/utils/git"
	"log"
	"sync"
	"time"
)

type GitIgnoreFilter struct {
	ignore *gitutils.GitIgnore
}

func (f *GitIgnoreFilter) Match(path string, isDir bool) bool {
	return !f.ignore.IsIgnored(path, isDir)
}

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
	defer log.Println("Scanner stopped")

	for !running.IsShuttingDown() {
		workspace := s.tryPopJob()
		if workspace == nil {
			select {
			case <-running.GetShutdown().Done():
				return
			case <-time.After(500 * time.Millisecond):
				continue
			}
		}

		s.setCurrent(workspace)
		if err := s.processWorkspace(workspace); err != nil {
			log.Printf("Error scanning workspace %s: %v", workspace.Path, err)
		} else {
			workspace.UpdateLastFullSync()
			workspace.Save()
		}
		s.setCurrent(nil)
	}
}

// processWorkspace processes a single workspace by scanning its files and applying filters.
func (s *Scanner) processWorkspace(w *workspace.Workspace) error {
	log.Printf("Start processing workspace %s", w.Path)
	start := time.Now()
	fileCount := 0
	interrupted := false
	defer func() {
		log.Printf("Finished processing workspace %s, cost %s, %d files, interrupted: %t", w.Path, time.Since(start), fileCount, interrupted)
	}()

	baseDir := w.Path
	filters := w.GetFilters()

	var exclude fsutils.ListFileFilter
	if filters.Exclude.UseGitIgnore {
		exclude = &GitIgnoreFilter{
			ignore: gitutils.NewGitIgnore(baseDir, true),
		}
	} else {
		exclude = utils.NewSimpleFilterExclude(filters.Exclude.Customized, baseDir)
	}

	include := utils.NewSimpleFilter(filters.Include, baseDir)
	startTime := time.Now()
	lastTime := time.Now()
	err := fsutils.ListFiles(baseDir, fsutils.ListFileOptions{Filter: exclude}, func(fileInfo fsutils.FileInfo) bool {
		if include.Match(fileInfo.Path, false) {
			parser.Add(w, fileInfo.Path)
			fileCount++

			w.Mutex.Lock()
			w.IndexingStatus.TotalFiles++
			w.Mutex.Unlock()
		}

		if time.Since(lastTime) > 1000*time.Millisecond {
			log.Printf("Scanning %s, %d files found, elapsed %s", w.Path, fileCount, time.Since(startTime))
			lastTime = time.Now()
		}

		interrupted = running.IsShuttingDown()
		return !interrupted
	})

	if err != nil {
		return err
	}
	if interrupted {
		return fmt.Errorf("interrupted")
	}
	return nil
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

	w.Mutex.Lock()
	defer w.Mutex.Unlock()
	if w.IndexingStatus != nil {
		return fmt.Errorf("workspace is indexing")
	}

	now := time.Now()
	w.IndexingStatus = &workspace.IndexingStatus{
		StartedAt:    &now,
		TotalFiles:   0,
		IndexedFiles: 0,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.queue.PushBack(w)
	return nil
}
