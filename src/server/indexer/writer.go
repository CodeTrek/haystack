package indexer

import (
	"log"
	"search-indexer/running"
	"search-indexer/server/core/storage"
	"search-indexer/server/core/workspace"
	"sync"
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
		docs: make(chan *UpdateDoc, 32),
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

	for {
		select {
		case <-running.GetShutdown().Done():
			return
		case doc := <-w.docs:
			storage.SaveDocuments(doc.Workspace.Meta.ID, []*storage.Document{doc.Document})
		}
	}
}

func (w *Writer) Add(workspace *workspace.Workspace, doc *storage.Document) {
	w.docs <- &UpdateDoc{
		Workspace: workspace,
		Document:  doc,
	}
}
