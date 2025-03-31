package searcher

import (
	"log"
	"search-indexer/running"
	"search-indexer/server/core/storage"
	"search-indexer/server/core/workspace"
	"strings"
	"sync"
	"time"
)

func Run(wg *sync.WaitGroup) {
	log.Println("Starting searcher...")

	time.Sleep(1 * time.Second)
	if wss := workspace.GetAllWorkspaces(); len(wss) > 0 {
		workspaceId := wss[0]
		Search(workspaceId, []string{"Sync", "SavedTabGroup"})
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer log.Println("Searcher shutdown.")
		running.WaitingForShutdown()
	}()
}

func Search(workspaceId string, query []string) {
	start := time.Now()
	results := []storage.SearchResult{}
	for _, q := range query {
		results = append(results, storage.Search(workspaceId, q))
	}

	docIds := results[0].DocIds
	for _, r := range results[1:] {
		for docid := range docIds {
			if _, ok := r.DocIds[docid]; !ok {
				delete(docIds, docid)
			}
		}
	}

	docPaths := []string{}
	for docid := range docIds {
		doc, err := storage.GetDocument(workspaceId, docid, false)
		if err != nil {
			continue
		}
		docPaths = append(docPaths, doc.FullPath)
	}

	log.Println("Search results for", query, "in", time.Since(start), len(docPaths), "results:\n", strings.Join(docPaths, "\n"))
}
