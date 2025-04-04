package searcher

import (
	"haystack/server/core/storage"
	"haystack/server/core/workspace"
	"haystack/server/indexer"
	"haystack/shared/running"
	"haystack/shared/types"
	"haystack/utils"
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

type QueryFilters struct {
	Path    string
	Include *utils.SimpleFilter
	Exclude *utils.SimpleFilter
}

// SearchContent searches the content of the workspace
// query is a list of words to search for
// returns a list of results
func SearchContent(workspace *workspace.Workspace, query string,
	filters *types.SearchFilters,
	limit *types.SearchLimit) []types.SearchContentResult {

	results := []storage.SearchResult{}
	results = append(results, storage.Search(workspace.ID, query))

	docIds := results[0].DocIds
	for _, r := range results[1:] {
		for docid := range docIds {
			if _, ok := r.DocIds[docid]; !ok {
				delete(docIds, docid)
			}
		}
	}

	docs := map[string]*storage.Document{}
	for docid := range docIds {
		doc, err := storage.GetDocument(workspace.ID, docid, false)
		if err != nil {
			continue
		}
		docs[docid] = doc
	}

	indexer.RefreshFileIfNeeded(workspace.ID, docs)

	// TODO: Add lines to the results
	finalResults := []types.SearchContentResult{}
	for _, doc := range docs {
		finalResults = append(finalResults, types.SearchContentResult{
			File: doc.FullPath,
		})
	}

	return finalResults
}
