package indexer

import (
	"log"
	"search-indexer/running"
	"search-indexer/server/core/storage"
	"search-indexer/server/core/workspace"
	"sync"
	"time"
)

type WriteDoc struct {
	Workspace *workspace.Workspace
	Document  *storage.Document
	CreateNew bool
}

type Writer struct {
	docs chan *WriteDoc
}

func NewWriter() *Writer {
	return &Writer{
		docs: make(chan *WriteDoc, 64),
	}
}

func (w *Writer) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go w.run(wg)
}

func (w *Writer) run(wg *sync.WaitGroup) {
	log.Println("Writer started")
	defer wg.Done()
	defer log.Println("Writer stopped")
	timer := time.NewTicker(1000 * time.Millisecond)
	defer timer.Stop()

	for {
		select {
		case doc := <-w.docs:
			docs := []*WriteDoc{doc}
			docs = append(docs, w.getPendingWrites(63)...)

			w.processDocs(docs)
		case <-running.GetShutdown().Done():
			for {
				docs := w.getPendingWrites(64)
				if len(docs) == 0 {
					break
				}
				w.processDocs(docs)

				// Sleep to wait for remaining docs to be added to the channel
				time.Sleep(10 * time.Millisecond)
			}

			return
		}
	}
}

func (w *Writer) processDocs(docs []*WriteDoc) {
	newDocs := make(map[string][]*storage.Document)
	existingDocs := make(map[string][]*storage.Document)
	for _, doc := range docs {
		if doc.CreateNew {
			newDocs[doc.Workspace.Meta.ID] = append(newDocs[doc.Workspace.Meta.ID], doc.Document)
		} else {
			existingDocs[doc.Workspace.Meta.ID] = append(existingDocs[doc.Workspace.Meta.ID], doc.Document)
		}
	}

	for workspaceID, docs := range newDocs {
		storage.SaveNewDocuments(workspaceID, docs)
	}

	for workspaceID, docs := range existingDocs {
		storage.UpdateDocuments(workspaceID, docs)
	}
}

func (w *Writer) getPendingWrites(limit int) []*WriteDoc {
	docs := []*WriteDoc{}
	for {
		select {
		case doc := <-w.docs:
			docs = append(docs, doc)
			if len(docs) >= limit {
				return docs
			}
		default:
			return docs
		}
	}
}

func (w *Writer) Add(workspace *workspace.Workspace, doc *storage.Document, createNew bool) {
	w.docs <- &WriteDoc{
		Workspace: workspace,
		Document:  doc,
		CreateNew: createNew,
	}
}
