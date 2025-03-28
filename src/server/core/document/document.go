package document

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"search-indexer/server/core/parser"
)

type Document struct {
	FullPath     string
	RelPath      string
	Size         int64
	ModifiedTime int64  // Last modified time in nanoseconds
	Hash         string // MD5 hash of the file content

	Content Content
}

func Parse(relPath string, baseDir string) (*Document, error) {
	fullPath := filepath.Join(baseDir, relPath)

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	return &Document{
		FullPath:     fullPath,
		RelPath:      relPath,
		Size:         info.Size(),
		ModifiedTime: info.ModTime().UnixNano(),
		Hash:         fmt.Sprintf("%x", md5.Sum(content)),
		Content: Content{
			Words: parser.ParseString(string(content)),
		},
	}, nil
}
