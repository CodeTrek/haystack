package indexer

import "search-indexer/utils"

func GetDocumentId(fullPath string) string {
	return utils.Md5HashString(fullPath)
}

func GetContentHash(content []byte) string {
	return utils.Md5Hash(content)
}
