package storage

import (
	"encoding/json"
	"fmt"
	"search-indexer/server/core/document"
	"strings"
)

func EncodeKey(workspaceid string, docid string) []byte {
	return []byte(fmt.Sprintf("ws:%s,doc:%s", workspaceid, docid))
}

func DecodeKey(key string) (string, string) {
	return strings.Split(key, ",")[0], strings.Split(key, ",")[1]
}

func EncodeDocument(doc *document.Document) ([]byte, error) {
	return json.Marshal(doc)
}

func DecodeDocument(data []byte) (*document.Document, error) {
	doc := document.Document{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	return &doc, nil
}
