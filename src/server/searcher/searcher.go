package searcher

import (
	"haystack/server/core/storage"
	"haystack/server/indexer"
	"haystack/shared/requests"
	"haystack/shared/running"
	"log"
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

// SearchContent searches the content of the workspace
// query is a list of words to search for
// returns a list of results
func SearchContent(workspaceId string, query string) []requests.SearchContentResult {
	results := []storage.SearchResult{}
	results = append(results, storage.Search(workspaceId, query))

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
