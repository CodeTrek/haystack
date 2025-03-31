package storage

import (
	"crypto/md5"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

type Document struct {
	ID           string `json:"-"`
	FullPath     string `json:"full_path"`
	Size         int64  `json:"size"`
	ModifiedTime int64  `json:"modified_time"`
	Hash         string `json:"hash"`

	Words []string `json:"words"`
}

// As the Document already breakdown into keywords, we can use the document full-path as the document id
// and store the document id and its keywords in the storage, below is the process:
//   - Create a reading snapshot of the storage to allow concurrent read operations
//   - Document full-path is converted to a md5 hash value, and used as the document id
//   - A new entry is created in the storage:
//       key: "doc:<workspace_id>|<document_id>"
//       value: <Document>
//   - For each keyword in the document, a new entry is created in the storage:
//       key: "kw:<workspace_id>|<keyword>|<document_count>|<document_hash>"
//       value: <document_ids>

type RelatedDocs struct {
	DocIds    []string
	UpdatedAt time.Time
}

type WorkspacePendingWrite struct {
	WorkspaceID string

	// Map of keyword to document ids
	Keywords map[string]RelatedDocs
}

var pendingWrites = map[string]*WorkspacePendingWrite{}
var pendingWritesMutex sync.Mutex

func getPendingWrite(workspaceid string) *WorkspacePendingWrite {
	wp := pendingWrites[workspaceid]
	if wp == nil {
		wp = &WorkspacePendingWrite{
			WorkspaceID: workspaceid,
			Keywords:    make(map[string]RelatedDocs),
		}
		pendingWrites[workspaceid] = wp
	}

	return wp
}

func FlushPendingWrites(final bool) {
	pendingWritesMutex.Lock()
	defer pendingWritesMutex.Unlock()

	batch := db.Batch()
	count := 0

	wordsCount := 0
	docsCount := 0
	for _, wp := range pendingWrites {
		for kw, relatedDocs := range wp.Keywords {
			if !final && len(relatedDocs.DocIds) < 10 && time.Since(relatedDocs.UpdatedAt) < 10*time.Second {
				continue
			}

			wordsCount++
			docsCount += len(relatedDocs.DocIds)

			sort.Strings(relatedDocs.DocIds)
			content := strings.Join(relatedDocs.DocIds, "|")
			hash := fmt.Sprintf("%x", md5.Sum([]byte(content)))
			batch.Put(KEncodeKeyword(wp.WorkspaceID, kw, len(relatedDocs.DocIds), hash), []byte(content))
			delete(wp.Keywords, kw)

			count++

			if count >= 1024 {
				if err := batch.Write(true); err != nil {
					log.Println("Failed to flush pending writes:", err)
				}
				batch = db.Batch()
				count = 0
			}
		}
	}

	if count > 0 {
		if err := batch.Write(true); err != nil {
			log.Println("Failed to flush pending writes:", err)
		}
	}

	if wordsCount > 0 {
		log.Printf("Flushed %d words with %d docs", wordsCount, docsCount)
	}
}

func SaveDocuments(workspaceid string, docs []*Document) error {
	if db.IsClosed() {
		return fmt.Errorf("database is closed")
	}

	batch := db.Batch()

	pendingWritesMutex.Lock()
	defer pendingWritesMutex.Unlock()
	cache := getPendingWrite(workspaceid)

	for _, doc := range docs {
		v, err := VEncodeDocument(doc)
		if err != nil {
			return err
		}

		batch.Put(KEncodeDocument(workspaceid, doc.ID), v)

		for _, kw := range doc.Words {
			// Add to write cache to merge with other documents and flush later
			cache.Keywords[kw] = RelatedDocs{
				DocIds:    append(cache.Keywords[kw].DocIds, doc.ID),
				UpdatedAt: time.Now(),
			}
		}
	}

	if err := batch.Write(true); err != nil {
		return err
	}

	return nil
}
