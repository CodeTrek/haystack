package requests

type SearchContentLimit struct {
	MaxLines        int `json:"max_lines,omitempty"`
	MaxFiles        int `json:"max_files,omitempty"`
	MaxLinesPerFile int `json:"max_lines_per_file,omitempty"`
}

// SearchContentRequest is the request for searching the content of a workspace
// @param Workspace: is the path to the workspace
// @param Query: is the query to search for, refer to the search query syntax in the server/server/search.md
// @param Filters: is the filters to apply to the search
// @param Limit: is the limit to apply to the search
// @param Filters.Path: is the path to the workspace
// @param Filters.Include: is the include to apply to the search
// @param Filters.Exclude: is the exclude to apply to the search
// @param Limit.MaxLines: is the max lines to apply to the search
// @param Limit.MaxFiles: is the max files to apply to the search
// @param Limit.MaxLinesPerFile: is the max lines per file to apply to the search
type SearchContentRequest struct {
	Workspace string `json:"workspace,omitempty"`
	Query     string `json:"query,omitempty"`
	Filters   struct {
		Path    string `json:"path,omitempty"`
		Include string `json:"include,omitempty"`
		Exclude string `json:"exclude,omitempty"`
	} `json:"filters,omitempty"`
	Limit SearchContentLimit `json:"limit,omitempty"`
}

type SearchContentLine struct {
	LineNumber int    `json:"line_number"`
	Content    string `json:"content"`
}

type SearchContentResult struct {
	File  string              `json:"file"`
	Lines []SearchContentLine `json:"lines,omitempty"`
}

type SearchContentResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Results []SearchContentResult `json:"results,omitempty"`
	} `json:"data,omitempty"`
}
