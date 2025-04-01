package storage

type KeywordResult struct {
	Keyword string              `json:"-"`
	DocIds  map[string]struct{} `json:"-"`
}

type SearchResult struct {
	Keywords map[string]*KeywordResult `json:"keywords"`
	DocIds   map[string]struct{}       `json:"docIds"`
}

func Search(workspaceid string, query string) SearchResult {
	results := SearchResult{
		Keywords: make(map[string]*KeywordResult),
		DocIds:   make(map[string]struct{}),
	}

	db.Scan(EncodeKeywordSearchKey(workspaceid, query), func(key, value []byte) bool {
		_, keyword, _, _ := DecodeKeywordIndexKey(string(key))
		docids := DecodeKeywordIndexValue(value)
		if len(docids) > 0 {
			kr, ok := results.Keywords[keyword]
			if !ok {
				kr = &KeywordResult{
					Keyword: keyword,
					DocIds:  make(map[string]struct{}),
				}
				results.Keywords[keyword] = kr
			}

			for _, docid := range docids {
				kr.DocIds[docid] = struct{}{}
				results.DocIds[docid] = struct{}{}
			}
		}
		return true
	})
	return results
}
