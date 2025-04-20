package storage

import (
	"fmt"
	"log"
	"time"
)

const MaxKeywordIndexSize = 1000

var (
	pendingWrites      = map[string]*WorkspacePendingWrite{}
	lastFlushWriteTime = time.Now()

	pendingDeletes      = map[string]*WorkspacePendingWrite{}
	lastFlushDeleteTime = time.Now()
)

type RelatedDocs struct {
	DocIds    []string
	UpdatedAt time.Time
}

type WorkspacePendingWrite struct {
	WorkspaceID string

	// Map of keyword to document ids
	Keywords map[string]RelatedDocs
}

type flushPendingWritesTask struct {
	closing bool
}

func (t *flushPendingWritesTask) Run() {
	// Flush pending writes to the database
	flushPendingWrites(t.closing)
	flushPendingDeletes(t.closing)
}

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

// flushPendingWrites flushes the pending writes to the database
func flushPendingWrites(closing bool) {
	if !closing && time.Since(lastFlushWriteTime) < 1*time.Second {
		return
	}
	lastFlushWriteTime = time.Now()

	if closing {
		log.Println("Final flushing pending writes...")
		defer func() {
			log.Println("Final fulshed pending writes")
		}()
	}

	batch := NewBatchWrite(db)

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

			writeKeywordIndex(batch, wp.WorkspaceID, kw, relatedDocs.DocIds, nil)
			delete(wp.Keywords, kw)

			// delete empty workspace
			if len(wp.Keywords) == 0 {
				delete(pendingWrites, wp.WorkspaceID)
			}
		}
	}

	batch.Commit()
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

// getPendingDelete returns the pending delete cache for the workspace
// It will create a new cache if it does not exist
func getPendingDelete(workspaceid string) *WorkspacePendingWrite {
	wp := pendingDeletes[workspaceid]
	if wp == nil {
		wp = &WorkspacePendingWrite{
			WorkspaceID: workspaceid,
			Keywords:    make(map[string]RelatedDocs),
		}
		pendingDeletes[workspaceid] = wp
	}

	return wp
}

func flushPendingDeletes(closing bool) {
	if !closing && time.Since(lastFlushDeleteTime) < 1*time.Second {
		return
	}
	lastFlushDeleteTime = time.Now()

	if closing {
		log.Println("Final flushing pending deletes...")
		defer func() {
			log.Println("Final fulshed pending deletes")
		}()
	}

	batch := NewBatchWrite(db)

	for _, wp := range pendingDeletes {
		for kw, relatedDocs := range wp.Keywords {
			// Skip the keyword if it has been updated in the last 2 seconds
			// and has less than 50 documents
			if !closing && len(relatedDocs.DocIds) < 50 && time.Since(relatedDocs.UpdatedAt) < 5*time.Second {
				continue
			}

			removeDocumentsFromKeywordIndex(batch, wp.WorkspaceID, kw, relatedDocs.DocIds)
			delete(wp.Keywords, kw)

			// delete empty workspace
			if len(wp.Keywords) == 0 {
				delete(pendingDeletes, wp.WorkspaceID)
			}
		}
	}

	batch.Commit()
}

func removeKeywordsFromDocumentCached(workspaceid string, docid string, keywords []string) {
	w := getPendingDelete(workspaceid)
	for _, kw := range keywords {
		// Add to delete cache to merge with other documents and flush later
		w.Keywords[kw] = RelatedDocs{
			DocIds:    append(w.Keywords[kw].DocIds, docid),
			UpdatedAt: time.Now(),
		}
	}
}

// writeKeywordIndex writes a keyword to the database
func writeKeywordIndex(batch BatchWrite, workspaceid string, kw string, docids []string, key []byte) {
	content := EncodeKeywordIndexValue(docids)
	if len(key) == 0 {
		key = EncodeKeywordIndexKey(workspaceid, kw, len(docids))
	}
	batch.Put(key, content)
}

// removeDocumentsFromKeywordIndex removes a document from the keywords index
// It will remove the document from the keywords index and rewrite the keyword with new docids
func removeDocumentsFromKeywordIndex(batch BatchWrite, workspaceid string, kw string, removingDocids []string) {
	if len(kw) == 0 {
		log.Println("Warning: removing document from keywords index, but keyword is empty")
		return
	}

	removings := map[string]struct{}{}
	for _, id := range removingDocids {
		if id != "" {
			removings[id] = struct{}{}
		}
	}

	if len(removings) == 0 {
		log.Println("Warning: removing document from keywords index, but docid is empty")
		return
	}

	keys := []string{}
	docids := map[string]struct{}{}
	db.Scan(EncodeKeywordIndexKeyPrefix(workspaceid, kw), func(key, value []byte) bool {
		changed := false
		tmpids := []string{}

		ids := DecodeKeywordIndexValue(string(value))
		for _, id := range ids {
			if _, ok := removings[id]; ok {
				// remove the document from the keyword index
				changed = true
				continue
			}
			if id != "" {
				tmpids = append(tmpids, id)
			}
		}

		if changed || len(tmpids) < MaxKeywordIndexSize/2 {
			keys = append(keys, string(key))
			for _, id := range tmpids {
				docids[id] = struct{}{}
			}
		}
		return true
	})

	count := 0
	for len(docids) > 0 {
		docs := []string{}
		for id := range docids {
			if len(docs) >= MaxKeywordIndexSize {
				break
			}
			docs = append(docs, id)
			delete(docids, id)
		}

		var key string
		if len(keys) > 0 {
			key = keys[0]
			keys = keys[1:]
		}

		// writeKeywordIndexCached(workspaceid, kw, docs)
		writeKeywordIndex(batch, workspaceid, kw, docs, []byte(key))
		count++
	}

	for _, key := range keys {
		batch.Delete([]byte(key))
	}
}

// saveDocument saves a document to the database
func saveDocument(batch BatchWrite, workspaceid string, doc *Document) {
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

	batch := NewBatchWrite(db)

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

	batch := NewBatchWrite(db)

	for _, updatedDoc := range t.Docs {
		// Convert the updated document words to a map for faster lookup
		updatedWordsMap := map[string]struct{}{}
		for _, kw := range updatedDoc.Words {
			if kw != "" {
				updatedWordsMap[kw] = struct{}{}
			}
		}

		// Get the current document words from the database
		currentWords, err := GetDocumentWords(t.WorkspaceID, updatedDoc.ID)
		if err != nil {
			continue
		}

		// Convert the current document words to a map for faster lookup
		currentWordsMap := map[string]struct{}{}
		for _, kw := range currentWords {
			if kw != "" {
				currentWordsMap[kw] = struct{}{}
			}
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

		removeKeywordsFromDocumentCached(t.WorkspaceID, updatedDoc.ID, removedWords)
		/*
			// Remove removed words from the keywords index
			for _, kw := range removedWords {
				removeDocumentFromKeywordsIndex(batch, t.WorkspaceID, kw, updatedDoc.ID)
			}
		*/

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
	DocId       string
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

	batch := NewBatchWrite(db)

	doc, err := GetDocument(t.WorkspaceID, t.DocId, true)
	if err != nil {
		t.done <- err
		log.Println("Failed to get document:", err)
		return
	}

	if doc == nil {
		t.done <- fmt.Errorf("document not found")
		return
	}

	defer log.Println("Document '", doc.RelPath, "' deleted from workspace '", t.WorkspaceID, "'")

	removeKeywordsFromDocumentCached(t.WorkspaceID, t.DocId, doc.Words)
	/*
		// delete the document from the keywords
		for _, kw := range doc.Words {
			removeDocumentFromKeywordsIndex(batch, t.WorkspaceID, kw, t.DocId)
		}
	*/

	// delete the document meta and words
	batch.Delete(EncodeDocumentMetaKey(t.WorkspaceID, t.DocId))
	batch.Delete(EncodeDocumentWordsKey(t.WorkspaceID, t.DocId))
	batch.Delete(EncodeDocumentPathKey(t.WorkspaceID, t.DocId))

	err = batch.Commit()
	if err != nil {
		log.Println("Failed to delete document:", err)
	}

	t.done <- err
}
