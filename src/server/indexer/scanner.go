package indexer

import (
	"container/list"
	"log"
	"search-indexer/running"
	"search-indexer/server/core/workspace"
	"search-indexer/utils"
	fsutils "search-indexer/utils/fs"
	gitutils "search-indexer/utils/git"
	"sync"
	"time"
)

type Scanner struct {
	current *workspace.Workspace
	queue   *list.List

	mutex sync.Mutex
}

func (s *Scanner) start(wg *sync.WaitGroup) {
	wg.Add(1)
	s.queue = list.New()
	go s.run(wg)
}

func (s *Scanner) run(wg *sync.WaitGroup) {
	defer wg.Done()

	for !running.IsShuttingDown() {
		workspace := s.tryPopJob()
		if workspace == nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		s.setCurrent(workspace)
		s.scan(workspace)
		s.setCurrent(nil)
	}

	log.Println("Scanner stopped")
}

func (s *Scanner) scan(current *workspace.Workspace) {
	baseDir := current.Meta.Path
	filters := current.GetFilters()

	var filter fsutils.ListFileFilter
	if filters.Exclude.UseGitIgnore {
		filter = &GitIgnoreFilter{
			ignore: gitutils.NewGitIgnore(baseDir),
		}
	} else {
		filter = utils.NewSimpleFilterExclude(filters.Exclude.Customized, baseDir)
	}

	fsutils.ListFiles(baseDir, fsutils.ListFileOptions{Filter: filter}, func(fileInfo fsutils.FileInfo) bool {
		log.Println(fileInfo.Path)
		return !running.IsShuttingDown()
	})
}

func (s *Scanner) tryPopJob() *workspace.Workspace {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.queue.Len() > 0 {
		job := s.queue.Remove(s.queue.Front())
		return job.(*workspace.Workspace)
	}
	return nil
}

func (s *Scanner) setCurrent(workspace *workspace.Workspace) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.current = workspace
}

func (s *Scanner) add(workspace *workspace.Workspace) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.queue.PushBack(workspace)
}
