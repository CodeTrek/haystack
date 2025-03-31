package indexer

import (
	"log"
	"search-indexer/running"
	"search-indexer/server/core/storage"
	"search-indexer/server/core/workspace"
	"sync"
	"time"
)

type UpdateDoc struct {
	Workspace *workspace.Workspace
	Document  *storage.Document
}

type Writer struct {
	docs chan *UpdateDoc
}

func NewWriter() *Writer {
	return &Writer{
		docs: make(chan *UpdateDoc, 256),
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
			docs := []*UpdateDoc{doc}
			docs = append(docs, w.getDocs(127)...)

			w.processDocs(docs)
		case <-timer.C:
			storage.FlushPendingWrites(false)
		case <-running.GetShutdown().Done():
			for {
				docs := w.getDocs(128)
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

func (w *Writer) processDocs(docs []*UpdateDoc) {
	m := make(map[string][]*storage.Document)
	for _, doc := range docs {
		m[doc.Workspace.Meta.ID] = append(m[doc.Workspace.Meta.ID], doc.Document)
	}

	for workspaceID, docs := range m {
		storage.SaveDocuments(workspaceID, docs)
	}
}

func (w *Writer) getDocs(limit int) []*UpdateDoc {
	docs := []*UpdateDoc{}
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

func (w *Writer) Add(workspace *workspace.Workspace, doc *storage.Document) {
	w.docs <- &UpdateDoc{
		Workspace: workspace,
		Document:  doc,
	}
}
