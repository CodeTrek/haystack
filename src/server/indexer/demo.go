package indexer

import (
	"context"
	"log"
	"path/filepath"
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
	baseDir := "D:\\Edge\\src\\chrome"
	log.Println(time.Now().Format("2006-01-02 15:04:05:"), "Indexing:", baseDir)

	files, err := fsutils.ListFiles(baseDir, fsutils.ListFileOptions{
		Filter: utils.NewSimpleFilterExclude([]string{"node_modules/", "dist/", "build/", "out/", "obj/", ".*"}, baseDir),
	})

	// files, err := fsutils.ListFiles(baseDir, fsutils.ListFileOptions{
	// 	Filter: &GitIgnoreFilter{
	// 		ignore: gitutils.NewGitIgnore(baseDir),
	// 	},
	// })

	if err != nil {
		log.Println("Error listing files:", err)
		files = []fsutils.FileInfo{}
	}

	log.Println(time.Now().Format("2006-01-02 15:04:05:"), len(files), "files found.")

	filter := utils.NewSimpleFilter([]string{"*.cc", "*.h", "*.md", "*.js", "*.ts", "*.cpp", "*.txt", "*.mm", "*.java"}, baseDir)
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

	log.Println(time.Now().Format("2006-01-02 15:04:05:"), len(filteredFiles), "files matched.")

	succ := 0
	faied := 0
	last := time.Now()
	for n, file := range filteredFiles {
		select {
		case <-shutdown.Done():
			return
		default:
		}

		_, err := document.Parse(file, baseDir)
		if err != nil {
			faied++
		} else {
			succ++
		}

		if time.Since(last) > 200*time.Millisecond || n == len(filteredFiles)-1 {
			last = time.Now()
			log.Printf("\r%s Parsing %d / %d, succ: %d, failed, %d", time.Now().Format("2006-01-02 15:04:05:"), n+1, len(filteredFiles), succ, faied)
		}
	}

	log.Println("\n", time.Now().Format("2006-01-02 15:04:05:"), len(filteredFiles), "parsed files, succ:", succ, "failed:", faied)
}
