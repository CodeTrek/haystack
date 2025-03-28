package main

import (
	"fmt"
	"search-indexer/utils"
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

func main() {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05:"), "Starting...")
	var done = make(chan []fsutils.FileInfo)
	go func(done chan []fsutils.FileInfo) {
		baseDir := "D:\\Edge\\src\\chrome"

		files, err := fsutils.ListFiles(baseDir, fsutils.ListFileOptions{
			Filter: utils.NewSimpleFilterExclude([]string{"node_modules/", "dist/", "build/", "out/", "obj/", ".*"}, baseDir),
		})

		// files, err := fsutils.ListFiles(baseDir, fsutils.ListFileOptions{
		// 	Filter: &GitIgnoreFilter{
		// 		ignore: gitutils.NewGitIgnore(baseDir),
		// 	},
		// })

		if err != nil {
			fmt.Println("Error listing files:", err)
			files = []fsutils.FileInfo{}
		}

		done <- files
	}(done)

	files := <-done

	// utils.NewSimpleFilter([]string{"*.cc", "*.h", "*.md", "*.js", "*.ts", "*.cpp", "*.txt", "*.mm", "*.java"})

	fmt.Println(time.Now().Format("2006-01-02 15:04:05:"), len(files), "files found.")
}
