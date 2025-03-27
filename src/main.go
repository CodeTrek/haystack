package main

import (
	"fmt"
	"search-indexer/helper/git"
)

type GitIgnoreFilter struct {
	ignore *git.GitIgnore
}

func (f *GitIgnoreFilter) Match(path string, isDir bool) bool {
	return !f.ignore.IsIgnored(path, isDir)
}

func main() {
	var done = make(chan []git.FileInfo)
	go func(done chan []git.FileInfo) {
		baseDir := "D:\\Edge\\src\\chrome"

		files, err := git.ListFiles(baseDir, git.ListFileOptions{
			Filter: &GitIgnoreFilter{
				ignore: git.NewGitIgnore(baseDir),
			},
		})

		if err != nil {
			fmt.Println("Error listing files:", err)
			files = []git.FileInfo{}
		}

		done <- files
	}(done)

	files := <-done

	fmt.Println(len(files), "files found.")
}
