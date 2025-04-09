package storage

import (
	"fmt"
	"haystack/server/core/storage/pebble"
	"haystack/utils"
	"log"
	"sort"
	"sync"
	"time"
)

type Document struct {
	ID           string `json:"-"`
	FullPath     string `json:"full_path"`
	Size         int64  `json:"size"`
	Hash         string `json:"hash"`
	ModifiedTime int64  `json:"modified_time"`
	LastSyncTime int64  `json:"last_sync_time"`

	Words     []string `json:"-"` // words in the document content
	PathWords []string `json:"-"` // words in the document relative-path
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
	Keywords  map[string]RelatedDocs
	PathWords map[string]RelatedDocs
}

var pendingWrites = map[string]*WorkspacePendingWrite{}
var pendingWritesMutex sync.Mutex

func getPendingWrite(workspaceid string) *WorkspacePendingWrite {
	wp := pendingWrites[workspaceid]
	if wp == nil {
		wp = &WorkspacePendingWrite{
			WorkspaceID: workspaceid,
			Keywords:    make(map[string]RelatedDocs),
			PathWords:   make(map[string]RelatedDocs),
		}
		pendingWrites[workspaceid] = wp
	}

	return wp
}

// flushPendingWrites flushes the pending writes to the database
func flushPendingWrites(closing bool) {
	pendingWritesMutex.Lock()
	defer pendingWritesMutex.Unlock()

	batch := db.Batch()
	count := 0

	wordsCount := 0
	docsCount := 0
	for _, wp := range pendingWrites {
		for kw, relatedDocs := range wp.Keywords {
			// Skip the keyword if it has been updated in the last 1 seconds
			// and has less than 50 documents
			if !closing && len(relatedDocs.DocIds) < 50 && time.Since(relatedDocs.UpdatedAt) < 1*time.Second {
				continue
			}

			wordsCount++
			docsCount += len(relatedDocs.DocIds)

			writeKeywordIndex(batch, wp.WorkspaceID, kw, relatedDocs.DocIds)
			delete(wp.Keywords, kw)

			count++

			// Flush the batch if it has more than 1024 writes
			if count >= 1024 {
				if err := batch.Commit(); err != nil {
					log.Println("Failed to flush pending writes:", err)
				}
				batch = db.Batch()
				count = 0
			}
		}
	}

	if count > 0 {
		if err := batch.Commit(); err != nil {
			log.Println("Failed to flush pending writes:", err)
		}
	}
}

// writeKeywordIndex writes a keyword to the database
func writeKeywordIndex(batch *pebble.Batch, workspaceid string, kw string, docids []string) {
	sort.Strings(docids)
	content := EncodeKeywordIndexValue(docids)
	hash := utils.Md5Hash(content)
	batch.Put(EncodeKeywordIndexKey(workspaceid, kw, len(docids), hash), content)
}

// removeDocumentFromKeywordsIndex removes a document from the keywords index
// It will remove the document from the keywords index and rewrite the keyword with new docids
func removeDocumentFromKeywordsIndex(batch *pebble.Batch, workspaceid string, kw string, removingDocid string) {
	db.Scan(EncodeKeywordSearchKey(workspaceid, kw), func(key, value []byte) bool {
		newDocids := []string{}

		// Get the current docids from the keywords index
		// and remove the removingDocid from the docids
		docids := DecodeKeywordIndexValue(value)
		changed := false
		for _, id := range docids {
			if id != removingDocid {
				newDocids = append(newDocids, id)
			} else {
				changed = true
			}
		}

		if changed {
			// rewrite the keyword with new docids
			// and delete the old keyword
			batch.Delete(key)
			if len(newDocids) > 0 {
				writeKeywordIndex(batch, workspaceid, kw, newDocids)
			}
		}

		return true
	})
}

// saveDocument saves a document to the database
func saveDocument(batch *pebble.Batch, workspaceid string, doc *Document) {
	doc.LastSyncTime = time.Now().UnixNano()
	meta, err := EncodeDocumentMetaValue(doc)
	if err != nil {
		return
	}

	// Save the document meta and words
	batch.Put(EncodeDocumentMetaKey(workspaceid, doc.ID), meta)
	batch.Put(EncodeDocumentWordsKey(workspaceid, doc.ID), EncodeKeywordIndexValue(doc.Words))
}

// updateKeywordIndexCached updates the keyword index in write cached
// It will add the document to the keyword index cache to merge with other documents and flush later
func updateKeywordIndexCached(workspaceid string, docid string, keywords []string) {
	pendingWritesMutex.Lock()
	defer pendingWritesMutex.Unlock()

	cache := getPendingWrite(workspaceid)
	for _, kw := range keywords {
		// Add to write cache to merge with other documents and flush later
		cache.Keywords[kw] = RelatedDocs{
			DocIds:    append(cache.Keywords[kw].DocIds, docid),
			UpdatedAt: time.Now(),
		}
	}
}

// GetDocument returns a document from the database
// It returns nil if the document does not exist
func GetDocument(workspaceid string, docid string, includeWords bool) (*Document, error) {
	data, err := db.Get(EncodeDocumentMetaKey(workspaceid, docid))
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	doc, err := DecodeDocumentMetaValue(data)
	if err != nil {
		return nil, err
	}

	if includeWords {
		words, err := GetDocumentWords(workspaceid, docid)
		if err != nil {
			return nil, err
		}

		doc.Words = words
	}

	return doc, nil
}

// GetDocumentWords returns the words of a document
// It returns an empty array if the document does not exist
func GetDocumentWords(workspaceid string, docid string) ([]string, error) {
	words, err := db.Get(EncodeDocumentWordsKey(workspaceid, docid))
	if err != nil {
		return nil, err
	}

	if words == nil {
		return []string{}, nil
	}

	return DecodeKeywordIndexValue(words), nil
}

// SaveNewDocuments saves new documents to the database
// It also updates the pending writes cache to merge with other documents and flush later
func SaveNewDocuments(workspaceid string, docs []*Document) error {
	if db.IsClosed() {
		return fmt.Errorf("database is closed")
	}

	batch := db.Batch()

	for _, doc := range docs {
		saveDocument(batch, workspaceid, doc)
		updateKeywordIndexCached(workspaceid, doc.ID, doc.Words)
		// TODO: update path words index
	}

	if err := batch.Commit(); err != nil {
		return err
	}

	return nil
}

// UpdateDocuments updates the words of a document
// It also updates the pending writes cache to merge with other documents and flush later
func UpdateDocuments(workspaceid string, updatedDocs []*Document) error {
	if db.IsClosed() {
		return fmt.Errorf("database is closed")
	}

	batch := db.Batch()

	for _, updatedDoc := range updatedDocs {
		// Convert the updated document words to a map for faster lookup
		updatedWordsMap := map[string]struct{}{}
		for _, kw := range updatedDoc.Words {
			updatedWordsMap[kw] = struct{}{}
		}

		// Get the current document words from the database
		currentWords, err := GetDocumentWords(workspaceid, updatedDoc.ID)
		if err != nil {
			continue
		}

		// Convert the current document words to a map for faster lookup
		currentWordsMap := map[string]struct{}{}
		for _, kw := range currentWords {
			currentWordsMap[kw] = struct{}{}
		}

		removedWords := []string{}
		newWords := []string{}

		// Find the words that are added to the current document
		for _, kw := range updatedDoc.Words {
			if _, ok := currentWordsMap[kw]; !ok {
				newWords = append(newWords, kw)
			}
		}

		// Find the words that are removed from the current document
		for _, kw := range currentWords {
			if _, ok := updatedWordsMap[kw]; !ok {
				removedWords = append(removedWords, kw)
			}
		}

		// Remove removed words from the keywords index
		for _, kw := range removedWords {
			removeDocumentFromKeywordsIndex(batch, workspaceid, kw, updatedDoc.ID)
		}

		// Add new words to the keywords index
		if len(newWords) > 0 {
			updateKeywordIndexCached(workspaceid, updatedDoc.ID, newWords)
		}

		// Save the updated document
		saveDocument(batch, workspaceid, updatedDoc)
	}
	return batch.Commit()
}

// DeleteDocument deletes a document from the database
// It will delete the document from the keywords index and the document meta
func DeleteDocument(workspaceid string, docid string) error {
	if db.IsClosed() {
		return fmt.Errorf("database is closed")
	}

	doc, err := GetDocument(workspaceid, docid, true)
	if doc == nil {
		return err
	}

	defer log.Println("Document '", doc.FullPath, "' deleted from workspace '", workspaceid, "'")

	batch := db.Batch()

	// delete the document from the keywords
	for _, kw := range doc.Words {
		removeDocumentFromKeywordsIndex(batch, workspaceid, kw, docid)
	}

	// delete the document meta and words
	batch.Delete(EncodeDocumentMetaKey(workspaceid, docid))
	batch.Delete(EncodeDocumentWordsKey(workspaceid, docid))

	return batch.Commit()
}
