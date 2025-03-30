package storage

import (
	"encoding/json"
	"fmt"
	"search-indexer/server/core/document"
	"strings"
)

const (
	DocPrefix       = "doc:"
	WorkspacePrefix = "ws:"
	KeywordPrefix   = "kw:"
)

func KEncodeDocument(workspaceid string, docid string) []byte {
	return []byte(fmt.Sprintf("%s%s,%s%s", WorkspacePrefix, workspaceid, DocPrefix, docid))
}

func KDecodeDocument(key string) (string, string) {
	parts := strings.Split(key, ",")
	if len(parts) != 2 {
		return "", ""
	}

	if !strings.HasPrefix(parts[0], WorkspacePrefix) || !strings.HasPrefix(parts[1], DocPrefix) {
		return "", ""
	}

	workspaceid := strings.TrimPrefix(parts[0], WorkspacePrefix)
	docid := strings.TrimPrefix(parts[1], DocPrefix)
	return workspaceid, docid
}

func KEncodeKeyword(workspaceid string, keyword string, docid string) []byte {
	return []byte(fmt.Sprintf("%s%s,%s%s,%s%s", WorkspacePrefix, workspaceid, KeywordPrefix, keyword, DocPrefix, docid))
}

func KDecodeKeyword(key string) (string, string, string) {
	parts := strings.Split(key, ",")
	if len(parts) != 3 {
		return "", "", ""
	}

	if !strings.HasPrefix(parts[0], WorkspacePrefix) ||
		!strings.HasPrefix(parts[1], KeywordPrefix) ||
		!strings.HasPrefix(parts[2], DocPrefix) {
		return "", "", ""
	}

	workspaceid := strings.TrimPrefix(parts[0], WorkspacePrefix)
	keyword := strings.TrimPrefix(parts[1], KeywordPrefix)
	docid := strings.TrimPrefix(parts[2], DocPrefix)
	return workspaceid, keyword, docid
}

func VEncodeDocument(doc *document.Document) ([]byte, error) {
	return json.Marshal(doc)
}

func VDecodeDocument(data []byte) (*document.Document, error) {
	doc := document.Document{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	return &doc, nil
}
