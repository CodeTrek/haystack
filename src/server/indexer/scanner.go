package indexer

import (
	"container/list"
	"fmt"
	"log"
	"search-indexer/running"
	"search-indexer/server/conf"
	"search-indexer/server/core/workspace"
	"search-indexer/utils"
	fsutils "search-indexer/utils/fs"
	gitutils "search-indexer/utils/git"
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
			log.Printf("Error scanning workspace %s: %v", workspace.Meta.Path, err)
		} else {
			workspace.UpdateLastFullSync()
			workspace.Save()
		}
		s.setCurrent(nil)
	}
}

// processWorkspace processes a single workspace by scanning its files and applying filters.
func (s *Scanner) processWorkspace(w *workspace.Workspace) error {
	log.Printf("Start processing workspace %s", w.Meta.Path)
	start := time.Now()
	fileCount := 0
	interrupted := false
	defer func() {
		log.Printf("Finished processing workspace %s, cost %s, %d files, interrupted: %t", w.Meta.Path, time.Since(start), fileCount, interrupted)
	}()

	baseDir := w.Meta.Path
	filters := w.GetFilters()

	var exclude fsutils.ListFileFilter
	if filters.Exclude.UseGitIgnore {
		exclude = &GitIgnoreFilter{
			ignore: gitutils.NewGitIgnore(baseDir),
		}
	} else {
		exclude = utils.NewSimpleFilterExclude(filters.Exclude.Customized, baseDir)
	}

	include := utils.NewSimpleFilter(filters.Include, baseDir)
	startTime := time.Now()
	lastTime := time.Now()
	err := fsutils.ListFiles(baseDir, fsutils.ListFileOptions{Filter: exclude}, func(fileInfo fsutils.FileInfo) bool {
		if include.Match(fileInfo.Path, false) {
			if fileInfo.Size <= conf.Get().MaxFileSize {
				parser.Add(w, fileInfo.Path)
				fileCount++
			} else {
				log.Printf("File %s (%f MiB) is too large to index, skipping", fileInfo.Path, float64(fileInfo.Size)/1024/1024)
			}
		}

		if time.Since(lastTime) > 1000*time.Millisecond {
			log.Printf("Scanning %s, %d files found, elapsed %s", w.Meta.Path, fileCount, time.Since(startTime))
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

	s.mu.Lock()
	defer s.mu.Unlock()
	s.queue.PushBack(w)
	return nil
}
