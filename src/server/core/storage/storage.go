package storage

import (
	"search-indexer/server/core/document"
)

// As the Document already breakdown into keywords, we can use the document full-path as the document id
// and store the document id and its keywords in the storage, below is the process:
//   - Create a reading snapshot of the storage to allow concurrent read operations
//   - Document full-path is converted to a md5 hash value, and used as the document id
//   - A new entry is created in the storage:
//       key: "doc:<document_id>"
//       value: <Document>
//   - For each keyword in the document, a new entry is created in the storage:
//       key: "kw:<keyword>,doc:<document_id>"
//       value: <weight>

func Init() error {
	return nil
}

func Save(docs []*document.Document, workspaceid string) error {
	return nil
}
