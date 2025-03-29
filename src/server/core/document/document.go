package document

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"search-indexer/server/core/parser"
)

type Document struct {
	ID           string `json:"-"`
	FullPath     string `json:"full_path"`
	Size         int64  `json:"size"`
	ModifiedTime int64  `json:"modified_time"`
	Hash         string `json:"hash"`

	Content Content `json:"content"`
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
		ID:           fmt.Sprintf("%x", md5.Sum([]byte(fullPath))),
		FullPath:     fullPath,
		Size:         info.Size(),
		ModifiedTime: info.ModTime().UnixNano(),
		Hash:         fmt.Sprintf("%x", md5.Sum(content)),
		Content: Content{
			Words: parser.ParseString(string(content)),
		},
	}, nil
}
