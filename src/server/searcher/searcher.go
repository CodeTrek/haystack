package searcher

import (
	"bufio"
	"haystack/conf"
	"haystack/server/core/storage"
	"haystack/server/core/workspace"
	"haystack/server/indexer"
	"haystack/shared/running"
	"haystack/shared/types"
	"haystack/utils"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
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
func SearchContent(workspace *workspace.Workspace, req *types.SearchContentRequest) ([]types.SearchContentResult, bool) {
	startTime := time.Now()
	var isTimeout = func() bool {
		return time.Since(startTime) > 10*time.Second
	}

	limit := conf.Get().Server.Search.Limit
	if req.Limit != nil {
		if req.Limit.MaxResults > 0 && req.Limit.MaxResults < limit.MaxResults {
			limit.MaxResults = req.Limit.MaxResults
		}

		if req.Limit.MaxResultsPerFile > 0 && req.Limit.MaxResultsPerFile < limit.MaxResultsPerFile {
			limit.MaxResultsPerFile = req.Limit.MaxResultsPerFile
		}
	}

	var includeFilter *utils.SimpleFilter
	var excludeFilter *utils.SimpleFilter
	var pathFilter = ""
	if req.Filters != nil {
		pathFilter = strings.ToLower(
			filepath.FromSlash(filepath.Clean(filepath.Join(workspace.Path, req.Filters.Path)) + "/"))

		if req.Filters.Include != "" {
			includeFilter = utils.NewSimpleFilter(strings.Split(req.Filters.Include, ","), workspace.Path)
		}

		if req.Filters.Exclude != "" {
			excludeFilter = utils.NewSimpleFilter(strings.Split(req.Filters.Exclude, ","), workspace.Path)
		}
	}

	// Check if the file should be included in the search
	var wantFile = func(doc *storage.Document) bool {
		if len(pathFilter) > 0 && !strings.HasPrefix(strings.ToLower(doc.FullPath), pathFilter) {
			return false
		}

		// Excluded by filter
		if excludeFilter != nil && excludeFilter.Match(doc.FullPath, false) {
			return false
		}

		// Not included by include filter
		if includeFilter != nil && !includeFilter.Match(doc.FullPath, false) {
			return false
		}

		return true
	}

	// Compile the query
	engine := NewSimpleContentSearchEngine(workspace)
	err := engine.Compile(req.Query, req.CaseSensitive)
	if err != nil {
		log.Println("Failed to compile query:", err)
		return []types.SearchContentResult{}, false
	}

	finalResults := []types.SearchContentResult{}
	totalHits := 0

	// Match the content of the file line by line
	var matchFileContent = func(relPath string, doc *storage.Document) (types.SearchContentResult, error) {
		fileMatch := types.SearchContentResult{
			File:  filepath.Clean(relPath),
			Lines: []types.LineMatch{},
		}

		// Read file and match line by line
		file, err := os.Open(doc.FullPath)
		if err != nil {
			log.Println("Failed to open file:", doc.FullPath, ", error:", err)
			return fileMatch, err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		lineNumber := 1
		fileHits := 0
		for scanner.Scan() {
			line := scanner.Text()
			matches := engine.IsLineMatch(line)
			if len(matches) > 0 {
				for _, match := range matches {
					fileMatch.Lines = append(fileMatch.Lines, types.LineMatch{
						Line: types.SearchContentLine{
							LineNumber: lineNumber,
							Content:    line,
							Match:      match,
						},
					})

					totalHits++
					fileHits++
					if fileHits >= limit.MaxResultsPerFile {
						fileMatch.Truncate = true
						break
					}
				}
				if fileHits >= limit.MaxResultsPerFile || totalHits >= limit.MaxResults {
					break
				}
			}
			lineNumber++
		}

		return fileMatch, nil
	}

	// Collect the all related documents
	results, err := engine.CollectDocuments()
	if err != nil {
		return []types.SearchContentResult{}, false
	}

	for docid := range results.DocIds {
		if isTimeout() {
			break
		}

		doc, err := storage.GetDocument(workspace.ID, docid, false)
		if err != nil || doc == nil {
			continue
		}

		// Check if the file should be included in the search
		if !wantFile(doc) {
			continue
		}

		// File has been removed, skip it
		removed, relPath, err := indexer.RefreshFileIfNeeded(workspace, doc)
		if err != nil || removed {
			continue
		}

		fileMatch, err := matchFileContent(relPath, doc)
		if err != nil {
			continue
		}

		if len(fileMatch.Lines) > 0 {
			finalResults = append(finalResults, fileMatch)
		}

		if totalHits >= limit.MaxResults {
			break
		}
	}

	return finalResults, totalHits >= limit.MaxResults
}
