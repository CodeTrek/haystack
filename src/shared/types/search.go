package types

type SearchLimit struct {
	MaxResults        int `yaml:"max_results" json:"max_results,omitempty"`
	MaxResultsPerFile int `yaml:"max_results_per_file" json:"max_results_per_file,omitempty"`
}

type SearchFilters struct {
	Path    string `json:"path,omitempty"`
	Include string `json:"include,omitempty"`
	Exclude string `json:"exclude,omitempty"`
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
	Workspace     string         `json:"workspace,omitempty"`
	Query         string         `json:"query,omitempty"`
	CaseSensitive bool           `json:"case_sensitive,omitempty"`
	Filters       *SearchFilters `json:"filters,omitempty"`
	Limit         *SearchLimit   `json:"limit,omitempty"`
}

type LineMatch struct {
	Before []SearchContentLine `json:"before,omitempty"`
	Line   SearchContentLine   `json:"line"`
	After  []SearchContentLine `json:"after,omitempty"`
}

type SearchContentLine struct {
	LineNumber int    `json:"line_number"`
	Content    string `json:"content"`
}

type SearchContentResult struct {
	File     string      `json:"file"`
	Lines    []LineMatch `json:"lines,omitempty"`
	Truncate bool        `json:"truncate,omitempty"`
}

type SearchContentResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Results  []SearchContentResult `json:"results,omitempty"`
		Truncate bool                  `json:"truncate,omitempty"`
	} `json:"data,omitempty"`
}
