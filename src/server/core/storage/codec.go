package storage

import (
	"encoding/json"
	"fmt"

	"strings"
)

const (
	DocPrefix       = "doc:"
	WorkspacePrefix = "ws:"
	KeywordPrefix   = "kw:"
)

func KEncodeWorkspace(workspaceid string) []byte {
	return []byte(fmt.Sprintf("%s%s", WorkspacePrefix, workspaceid))
}

func KDecodeWorkspace(key string) string {
	if !strings.HasPrefix(key, WorkspacePrefix) {
		return ""
	}

	key = strings.TrimPrefix(key, WorkspacePrefix)

	return key
}

func KEncodeDocument(workspaceid string, docid string) []byte {
	return []byte(fmt.Sprintf("%s%s|%s", DocPrefix, workspaceid, docid))
}

func KDecodeDocument(key string) (string, string) {
	if !strings.HasPrefix(key, DocPrefix) {
		return "", ""
	}

	key = strings.TrimPrefix(key, DocPrefix)

	parts := strings.Split(key, "|")
	if len(parts) != 2 {
		return "", ""
	}

	workspaceid := parts[0]
	docid := parts[1]

	return workspaceid, docid
}

func KEncodeKeyword(workspaceid string, keyword string, docid string) []byte {
	return []byte(fmt.Sprintf("%s%s|%s|%s", KeywordPrefix, workspaceid, keyword, docid))
}

func KDecodeKeyword(key string) (string, string, string) {
	if !strings.HasPrefix(key, KeywordPrefix) {
		return "", "", ""
	}

	key = strings.TrimPrefix(key, KeywordPrefix)

	parts := strings.Split(key, "|")
	if len(parts) != 3 {
		return "", "", ""
	}

	workspaceid := parts[0]
	keyword := parts[1]
	docid := parts[2]

	return workspaceid, keyword, docid
}

func VEncodeDocument(doc *Document) ([]byte, error) {
	return json.Marshal(doc)
}

func VDecodeDocument(data []byte) (*Document, error) {
	doc := Document{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	return &doc, nil
}
