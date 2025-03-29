package document

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"search-indexer/server/core/parser"
)

type Document struct {
	FullPath     string `json:"full_path"`
	Size         int64  `json:"size"`
	ModifiedTime int64  `json:"modified_time"`
	Hash         string `json:"hash"`

	Content Content `json:"-"`
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
		Size:         info.Size(),
		ModifiedTime: info.ModTime().UnixNano(),
		Hash:         fmt.Sprintf("%x", md5.Sum(content)),
		Content: Content{
			Words: parser.ParseString(string(content)),
		},
	}, nil
}
