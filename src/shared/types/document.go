package types

type DocumentUpdateRequest struct {
	Workspace string `json:"workspace"`
	Path      string `json:"path"`
}

type DocumentDeleteRequest struct {
	Workspace string `json:"workspace"`
	Path      string `json:"path"`
}
