package searcher

import (
	"bufio"
	"haystack/server/core/storage"
	"haystack/server/core/workspace"
	"haystack/server/indexer"
	"haystack/shared/running"
	"haystack/shared/types"
	"haystack/utils"
	"log"
	"os"
	"path/filepath"
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

	engine := NewSimpleContentSearchEngine(workspace, limit, filters)
	err := engine.Compile(query)
	if err != nil {
		log.Println("Failed to compile query:", err)
		return []types.SearchContentResult{}
	}

	// Collect the all related documents
	results, err := engine.CollectDocuments()
	if err != nil {
		return []types.SearchContentResult{}
	}

	docIds := results.DocIds
	docs := map[string]*storage.Document{}
	for docid := range docIds {
		doc, err := storage.GetDocument(workspace.ID, docid, false)
		if err != nil {
			continue
		}

		if doc != nil {
			docs[docid] = doc
		}
	}

	removedDocs := indexer.RefreshFileIfNeeded(workspace.ID, docs)
	for _, docid := range removedDocs {
		delete(docs, docid)
	}

	// TODO: Add lines to the results
	finalResults := []types.SearchContentResult{}
	for _, doc := range docs {
		relPath, err := filepath.Rel(workspace.Path, doc.FullPath)
		if err != nil {
			continue
		}

		file, err := os.Open(doc.FullPath)
		if err != nil {
			log.Println("Failed to open file:", doc.FullPath, ", error:", err)
			continue
		}
		scanner := bufio.NewScanner(file)

		fileMatch := types.SearchContentResult{
			File:  filepath.Clean(relPath),
			Lines: []types.LineMatch{},
		}

		lineNumber := 1
		for scanner.Scan() {
			line := scanner.Text()
			if engine.IsLineMatch(line) {
				fileMatch.Lines = append(fileMatch.Lines, types.LineMatch{
					Line: types.SearchContentLine{
						LineNumber: lineNumber,
						Content:    line,
					},
				})
			}
			lineNumber++
		}

		file.Close()
		if len(fileMatch.Lines) > 0 {
			finalResults = append(finalResults, fileMatch)
		}
	}

	return finalResults
}
