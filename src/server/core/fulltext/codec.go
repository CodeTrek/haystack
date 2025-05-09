package fulltext

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"strings"
)

const (
	DocWordsPrefix  = "dw:"
	DocMetaPrefix   = "dm:"
	DocPathPrefix   = "dp:"
	WorkspacePrefix = "ws:"
	KeywordPrefix   = "kw:"
	MergeIndexKey   = "merge-index"
)

func EncodeWorkspaceKey(workspaceid string) []byte {
	return []byte(fmt.Sprintf("%s%s", WorkspacePrefix, workspaceid))
}

func DecodeWorkspaceKey(key string) string {
	if !strings.HasPrefix(key, WorkspacePrefix) {
		return ""
	}

	key = strings.TrimPrefix(key, WorkspacePrefix)

	return key
}

func EncodeDocumentPathKey(workspaceid string, docid string) []byte {
	return []byte(fmt.Sprintf("%s%s|%s", DocPathPrefix, workspaceid, docid))
}

func DecodeDocumentPathKey(key string) (string, string) {
	if !strings.HasPrefix(key, DocPathPrefix) {
		return "", ""
	}

	key = strings.TrimPrefix(key, DocPathPrefix)

	parts := strings.Split(key, "|")
	if len(parts) != 2 {
		return "", ""
	}

	workspaceid := parts[0]
	docid := parts[1]

	return workspaceid, docid
}

func EncodeDocumentMetaKey(workspaceid string, docid string) []byte {
	return []byte(fmt.Sprintf("%s%s|%s", DocMetaPrefix, workspaceid, docid))
}

func EncodeDocumentMetaValue(doc *Document) ([]byte, error) {
	return json.Marshal(doc)
}

func DecodeDocumentMetaValue(data []byte) (*Document, error) {
	doc := Document{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	return &doc, nil
}

func EncodeDocumentWordsKey(workspaceid string, docid string) []byte {
	return []byte(fmt.Sprintf("%s%s|%s", DocWordsPrefix, workspaceid, docid))
}

func DecodeDocumentWordsKey(key string) (string, string) {
	if !strings.HasPrefix(key, DocWordsPrefix) {
		return "", ""
	}

	key = strings.TrimPrefix(key, DocWordsPrefix)

	parts := strings.Split(key, "|")
	if len(parts) != 2 {
		return "", ""
	}

	workspaceid := parts[0]
	docid := parts[1]

	return workspaceid, docid
}

func EncodeKeywordSearchKey(workspaceid string, query string) []byte {
	return []byte(fmt.Sprintf("%s%s|%s", KeywordPrefix, workspaceid, query))
}

func EncodeKeywordIndexKeyPrefix(workspaceid string, keyword string) []byte {
	return []byte(fmt.Sprintf("%s%s|%s|", KeywordPrefix, workspaceid, keyword))
}

func EncodeKeywordIndexKey(workspaceid string, keyword string, doccount int) []byte {
	return []byte(fmt.Sprintf("%s%d|%d",
		string(EncodeKeywordIndexKeyPrefix(workspaceid, keyword)), doccount, time.Now().UnixMicro()))
}

func DecodeKeywordIndexKey(key string) (string, string, int, string) {
	if !strings.HasPrefix(key, KeywordPrefix) {
		return "", "", 0, ""
	}

	key = strings.TrimPrefix(key, KeywordPrefix)

	parts := strings.Split(key, "|")
	if len(parts) != 4 {
		return "", "", 0, ""
	}

	workspaceid := parts[0]
	keyword := parts[1]
	doccount, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", "", 0, ""
	}
	tick := parts[3]

	return workspaceid, keyword, doccount, tick
}

func EncodeKeywordIndexValue(docids []string) []byte {
	return []byte(strings.Join(docids, "|"))
}

func DecodeKeywordIndexValue(data string) []string {
	return strings.Split(data, "|")
}
