package server

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"search-indexer/server/core/workspace"
	"search-indexer/server/searcher"
	"search-indexer/shared/requests"
	"search-indexer/utils"
)

// handleSearchContent handles the search content endpoint
// It will search the content of the server
func handleSearchContent(w http.ResponseWriter, r *http.Request) {
	var request requests.SearchContentRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if request.Workspace == "" {
		json.NewEncoder(w).Encode(requests.SearchContentResponse{
			Code:    1,
			Message: "Workspace is required",
		})
		return
	}

	workspacePath := utils.NormalizePath(request.Workspace)
	if !filepath.IsAbs(workspacePath) {
		json.NewEncoder(w).Encode(requests.SearchContentResponse{
			Code:    1,
			Message: "Workspace path must be absolute",
		})
		return
	}

	workspace, err := workspace.GetByPath(workspacePath)
	if err != nil {
		json.NewEncoder(w).Encode(requests.SearchContentResponse{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	if request.Query == "" {
		json.NewEncoder(w).Encode(requests.SearchContentResponse{
			Code:    1,
			Message: "Query is required",
		})
		return
	}

	queries := strings.Split(request.Query, " ")
	results := searcher.SearchContent(workspace.Meta.ID, queries)

	json.NewEncoder(w).Encode(requests.SearchContentResponse{
		Code:    0,
		Message: "Ok",
		Data: struct {
			Results []requests.SearchContentResult `json:"results,omitempty"`
		}{
			Results: results,
		},
	})
}
