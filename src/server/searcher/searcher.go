package searcher

import (
	"log"
	"search-indexer/server/core/storage"
	"search-indexer/server/indexer"
	"search-indexer/shared/requests"
	"search-indexer/shared/running"
	"sync"
)

func Run(wg *sync.WaitGroup) {
	log.Println("Starting searcher...")

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer log.Println("Searcher shutdown.")
		running.WaitingForShutdown()
	}()
}

func SearchContent(workspaceId string, query []string) []requests.SearchContentResult {
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

	docs := map[string]*storage.Document{}
	docPaths := []string{}
	for docid := range docIds {
		doc, err := storage.GetDocument(workspaceId, docid, false)
		if err != nil {
			continue
		}
		docs[docid] = doc
		docPaths = append(docPaths, doc.FullPath)
	}

	indexer.RefreshFileIfNeeded(workspaceId, docs)

	// TODO: Add lines to the results
	finalResults := []requests.SearchContentResult{}
	for _, doc := range docs {
		finalResults = append(finalResults, requests.SearchContentResult{
			File: doc.FullPath,
		})
	}

	return finalResults
}
