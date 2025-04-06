package server

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"haystack/server/core/workspace"
	"haystack/server/searcher"
	"haystack/shared/types"
	"haystack/utils"
)

// handleSearchContent handles the search content endpoint
// It will search the content of the server
func handleSearchContent(w http.ResponseWriter, r *http.Request) {
	var request types.SearchContentRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if request.Workspace == "" {
		json.NewEncoder(w).Encode(types.SearchContentResponse{
			Code:    1,
			Message: "Workspace is required",
		})
		return
	}

	// Normalize the workspace path
	// If the path is not absolute, return an error
	workspacePath := utils.NormalizePath(request.Workspace)
	if !filepath.IsAbs(workspacePath) {
		json.NewEncoder(w).Encode(types.SearchContentResponse{
			Code:    1,
			Message: "Workspace is not absolute",
		})
		return
	}

	// Get the workspace by path
	// If the workspace is not found, return an error
	workspace, err := workspace.GetByPath(workspacePath)
	if err != nil {
		json.NewEncoder(w).Encode(types.SearchContentResponse{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	// If the query is empty, return an error
	if request.Query == "" {
		json.NewEncoder(w).Encode(types.SearchContentResponse{
			Code:    1,
			Message: "Query is required",
		})
		return
	}

	start := time.Now()
	// Search the content of the workspace
	results, truncate := searcher.SearchContent(workspace, request.Query, request.Filters, request.Limit, request.CaseSensitive)
	defer func() {
		totalHits := 0
		for _, result := range results {
			totalHits += len(result.Lines)
		}
		req, _ := json.Marshal(request)
		log.Printf("Process /api/v1/search/content `%s`: took %s, found %d results in %d files, truncate: %t",
			string(req), time.Since(start), totalHits, len(results), truncate)
	}()

	json.NewEncoder(w).Encode(types.SearchContentResponse{
		Code:    0,
		Message: "Ok",
		Data: struct {
			Results  []types.SearchContentResult `json:"results,omitempty"`
			Truncate bool                        `json:"truncate,omitempty"`
		}{
			Results:  results,
			Truncate: truncate,
		},
	})
}
