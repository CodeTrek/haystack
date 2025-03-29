package storage

import (
	"search-indexer/server/core/document"
)

// As the Document already breakdown into keywords, we can use the document full-path as the document id
// and store the document id and its keywords in the storage, below is the process:
//   - Create a reading snapshot of the storage to allow concurrent read operations
//   - Document full-path is converted to a md5 hash value, and used as the document id
//   - A new entry is created in the storage, with the document id as the key and the json-encoded document.Document
//     as the value: "doc:<hash_value>" -> <Document>
//   - For each keyword in the document, a new entry is created in the storage, with the keyword as the key and the document id
//     as the value: "kw:<keyword>" -> <document_id>

func Save(docs []*document.Document) error {

	return nil
}
