package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"search-indexer/running"
	"search-indexer/server/core/document"
	"search-indexer/server/core/storage/leveldb"
	"sync"
	"time"
)

// As the Document already breakdown into keywords, we can use the document full-path as the document id
// and store the document id and its keywords in the storage, below is the process:
//   - Create a reading snapshot of the storage to allow concurrent read operations
//   - Document full-path is converted to a md5 hash value, and used as the document id
//   - A new entry is created in the storage:
//       key: "ws:<workspace_id>,doc:<document_id>"
//       value: <Document>
//   - For each keyword in the document, a new entry is created in the storage:
//       key: "ws:<workspace_id>,kw:<keyword>,doc:<document_id>"
//       value: <weight>

var db *leveldb.DB

const StorageVersion = "1.0"

func Init(wg *sync.WaitGroup) error {
	dataPath := filepath.Join(running.RootPath(), "data")
	dbPath := filepath.Join(dataPath, "leveldb")
	versionPath := filepath.Join(dataPath, "version")

	os.MkdirAll(dataPath, 0755)
	os.WriteFile(versionPath, []byte(StorageVersion), 0644)

	var err error
	db, err = leveldb.OpenDB(dbPath)
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		running.WaitingForShutdown()
		log.Println("Shutting down storage...")

		for {
			if err := db.Close(); err != nil {
				log.Printf("failed to close storage: %v", err)
			}

			if db.IsClosed() {
				break
			}

			log.Println("Waiting for storage to be closed...")
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			<-ctx.Done()
		}
	}()

	return nil
}

func Save(docs []*document.Document, workspaceid string) error {
	if db.IsClosed() {
		return fmt.Errorf("database is closed")
	}

	batch := db.Batch()

	for _, doc := range docs {
		v, err := VEncodeDocument(doc)
		if err != nil {
			return err
		}

		batch.Put(KEncodeDocument(workspaceid, doc.ID), v)
		for _, kw := range doc.Content.Words {
			batch.Put(KEncodeKeyword(workspaceid, kw, doc.ID), []byte("1"))
		}
	}

	if err := batch.Write(); err != nil {
		return err
	}

	if err := db.TakeSnapshot(); err != nil {
		log.Printf("failed to take snapshot: %v", err)
	}

	return nil
}
