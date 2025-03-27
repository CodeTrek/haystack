package main

import (
	"fmt"
	"search-indexer/utils/fs"
	"search-indexer/utils/git"
	"time"
)

type GitIgnoreFilter struct {
	ignore *gitutils.GitIgnore
}

func (f *GitIgnoreFilter) Match(path string, isDir bool) bool {
	return !f.ignore.IsIgnored(path, isDir)
}

type SimpleFilter struct {
	ignore *gitutils.GitIgnoreRules
}

func (f *SimpleFilter) Match(path string, isDir bool) bool {
	return !f.ignore.IsIgnored(path, isDir)
}

func main() {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05:"), "Starting...")
	var done = make(chan []fsutils.FileInfo)
	go func(done chan []fsutils.FileInfo) {
		baseDir := "C:\\Edge\\src\\chrome"

		/*
			ignoreRules, _ := gitutils.NewGitIgnoreRulesFromString(`
				.git
				.gitignore
				.cache
			`, baseDir)

			files, err := fsutils.ListFiles(baseDir, fsutils.ListFileOptions{
				Filter: &SimpleFilter{
					ignore: ignoreRules,
				},
			})
		*/

		files, err := fsutils.ListFiles(baseDir, fsutils.ListFileOptions{
			Filter: &GitIgnoreFilter{
				ignore: gitutils.NewGitIgnore(baseDir),
			},
		})
		if err != nil {
			fmt.Println("Error listing files:", err)
			files = []fsutils.FileInfo{}
		}

		done <- files
	}(done)

	files := <-done

	fmt.Println(time.Now().Format("2006-01-02 15:04:05:"), len(files), "files found.")
}
