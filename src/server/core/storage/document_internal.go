package storage

import (
	"fmt"
	"haystack/server/core/storage/pebble"
	"haystack/utils"
	"log"
	"sort"
	"time"
)

var (
	pendingWrites = map[string]*WorkspacePendingWrite{}
)

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

type flushPendingWritesTask struct {
	closing bool
}

func (t *flushPendingWritesTask) Run() {
	// Flush pending writes to the database
	flushPendingWrites(t.closing)
}

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
	if closing {
		log.Println("Final flushing pending writes...")
		defer func() {
			log.Println("Final fulshed pending writes")
		}()
	}

	batch := db.Batch()
	count := 0

	wordsCount := 0
	docsCount := 0
	for _, wp := range pendingWrites {
		for kw, relatedDocs := range wp.Keywords {
			// Skip the keyword if it has been updated in the last 2 seconds
			// and has less than 50 documents
			if !closing && len(relatedDocs.DocIds) < 50 && time.Since(relatedDocs.UpdatedAt) < 2*time.Second {
				continue
			}

			wordsCount++
			docsCount += len(relatedDocs.DocIds)

			writeKeywordIndex(batch, wp.WorkspaceID, kw, relatedDocs.DocIds)
			delete(wp.Keywords, kw)

			// delete empty workspace
			if len(wp.Keywords) == 0 && len(wp.PathWords) == 0 {
				delete(pendingWrites, wp.WorkspaceID)
			}

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

// updateKeywordIndexCached updates the keyword index in write cached
// It will add the document to the keyword index cache to merge with other documents and flush later
func updateKeywordIndexCached(workspaceid string, docid string, keywords []string) {
	cache := getPendingWrite(workspaceid)
	for _, kw := range keywords {
		// Add to write cache to merge with other documents and flush later
		cache.Keywords[kw] = RelatedDocs{
			DocIds:    append(cache.Keywords[kw].DocIds, docid),
			UpdatedAt: time.Now(),
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
	batch.Put(EncodeDocumentPathKey(workspaceid, doc.ID), []byte(doc.RelPath))
}

type saveNewDocumentsTask struct {
	WorkspaceID string
	Docs        []*Document
	done        chan error
}

func (t *saveNewDocumentsTask) Wait() error {
	defer close(t.done)
	return <-t.done
}

func (t *saveNewDocumentsTask) Run() {
	if db.IsClosed() {
		log.Println("Database is closed, skip saving new documents")
		t.done <- nil
		return
	}

	batch := db.Batch()

	for _, doc := range t.Docs {
		saveDocument(batch, t.WorkspaceID, doc)
		updateKeywordIndexCached(t.WorkspaceID, doc.ID, doc.Words)
		// TODO: update path words index
	}

	err := batch.Commit()
	if err != nil {
		log.Println("Failed to save new documents:", err)
	}

	t.done <- err
}

type updateDocumentsTask struct {
	WorkspaceID string
	Docs        []*Document
	done        chan error
}

func (t *updateDocumentsTask) Wait() error {
	defer close(t.done)
	return <-t.done
}

func (t *updateDocumentsTask) Run() {
	if db.IsClosed() {
		log.Println("Database is closed, skip updating documents")
		t.done <- fmt.Errorf("database is closed")
		return
	}

	batch := db.Batch()

	for _, updatedDoc := range t.Docs {
		// Convert the updated document words to a map for faster lookup
		updatedWordsMap := map[string]struct{}{}
		for _, kw := range updatedDoc.Words {
			updatedWordsMap[kw] = struct{}{}
		}

		// Get the current document words from the database
		currentWords, err := GetDocumentWords(t.WorkspaceID, updatedDoc.ID)
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
			removeDocumentFromKeywordsIndex(batch, t.WorkspaceID, kw, updatedDoc.ID)
		}

		// Add new words to the keywords index
		if len(newWords) > 0 {
			updateKeywordIndexCached(t.WorkspaceID, updatedDoc.ID, newWords)
		}

		// Save the updated document
		saveDocument(batch, t.WorkspaceID, updatedDoc)
	}

	err := batch.Commit()
	if err != nil {
		log.Println("Failed to update documents:", err)
	}
	t.done <- err
}

type deleteDocumentTask struct {
	WorkspaceID string
	DocID       string
	done        chan error
}

func (t *deleteDocumentTask) Wait() error {
	defer close(t.done)
	return <-t.done
}

func (t *deleteDocumentTask) Run() {
	if db.IsClosed() {
		log.Println("Database is closed, skip deleting document")
		t.done <- fmt.Errorf("database is closed")
		return
	}

	doc, err := GetDocument(t.WorkspaceID, t.DocID, true)
	if err != nil {
		t.done <- err
		log.Println("Failed to get document:", err)
		return
	}

	if doc == nil {
		t.done <- nil
		return
	}

	defer log.Println("Document '", doc.FullPath, "' deleted from workspace '", t.WorkspaceID, "'")

	batch := db.Batch()

	// delete the document from the keywords
	for _, kw := range doc.Words {
		removeDocumentFromKeywordsIndex(batch, t.WorkspaceID, kw, t.DocID)
	}

	// delete the document meta and words
	batch.Delete(EncodeDocumentMetaKey(t.WorkspaceID, t.DocID))
	batch.Delete(EncodeDocumentWordsKey(t.WorkspaceID, t.DocID))
	batch.Delete(EncodeDocumentPathKey(t.WorkspaceID, t.DocID))

	err = batch.Commit()
	if err != nil {
		log.Println("Failed to delete document:", err)
	}

	t.done <- err
}
