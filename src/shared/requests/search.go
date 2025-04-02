package requests

type SearchContentLimit struct {
	Files        int `json:"files,omitempty"`
	LinesPerFile int `json:"lines_per_file,omitempty"`
}

type SearchContentRequest struct {
	Workspace string `json:"workspace,omitempty"`
	Query     string `json:"query,omitempty"`
	Filters   struct {
		Path string `json:"path,omitempty"`
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
