package document

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"search-indexer/server/core/parser"
	"search-indexer/server/core/storage"
)

type Document struct {
	Doc storage.Document
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
		Doc: storage.Document{
			ID:           fmt.Sprintf("%x", md5.Sum([]byte(fullPath))),
			FullPath:     fullPath,
			Size:         info.Size(),
			ModifiedTime: info.ModTime().UnixNano(),
			Hash:         fmt.Sprintf("%x", md5.Sum(content)),
			Words:        parser.ParseString(string(content)),
		},
	}, nil
}
