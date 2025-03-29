package indexer

import (
	"context"
	"log"
	"path/filepath"
	"search-indexer/server/conf"
	"search-indexer/server/core/document"
	"search-indexer/utils"
	fsutils "search-indexer/utils/fs"
	gitutils "search-indexer/utils/git"
	"time"
)

type GitIgnoreFilter struct {
	ignore *gitutils.GitIgnore
}

func (f *GitIgnoreFilter) Match(path string, isDir bool) bool {
	return !f.ignore.IsIgnored(path, isDir)
}

func demo(shutdown context.Context) {
	conf := conf.Get()
	baseDir := conf.Workspaces[0].Path
	log.Println("Indexing:", baseDir)

	var filter fsutils.ListFileFilter
	if conf.Workspaces[0].Exclude.UseGitIgnore {
		log.Println("Using gitignore filter")
		filter = &GitIgnoreFilter{
			ignore: gitutils.NewGitIgnore(baseDir),
		}
	} else {
		log.Println("Using customized filter")
		filter = utils.NewSimpleFilterExclude(conf.Workspaces[0].Exclude.Customized, baseDir)
	}

	files, err := fsutils.ListFiles(baseDir, fsutils.ListFileOptions{
		Filter: utils.NewSimpleFilterExclude(conf.Workspaces[0].Exclude.Customized, baseDir),
	})

	if err != nil {
		log.Println("Error listing files:", err)
		files = []fsutils.FileInfo{}
	}

	log.Println(len(files), "files found.")

	filter = utils.NewSimpleFilter(conf.Workspaces[0].Files, baseDir)
	filteredFiles := []string{}
	for _, file := range files {
		select {
		case <-shutdown.Done():
			return
		default:
		}

		baseName := filepath.Base(file.Path)
		if filter.Match(baseName, false) {
			filteredFiles = append(filteredFiles, file.Path)
		}
	}

	log.Println(len(filteredFiles), "files matched.")

	succ := 0
	faied := 0
	last := time.Now()
	wordCount := 0
	for n, file := range filteredFiles {
		select {
		case <-shutdown.Done():
			return
		default:
		}

		doc, err := document.Parse(file, baseDir)
		if err != nil {
			faied++
		} else {
			succ++
			wordCount += len(doc.Content.Words)
		}

		if time.Since(last) > 200*time.Millisecond || n == len(filteredFiles)-1 {
			last = time.Now()
			log.Printf("Parsing progress %d / %d, succ: %d, failed, %d, wordCount: %d", n+1, len(filteredFiles), succ, faied, wordCount)
		}
	}

	log.Println(len(filteredFiles), "parsed files, succ:", succ, "failed:", faied, "wordCount:", wordCount)

}
