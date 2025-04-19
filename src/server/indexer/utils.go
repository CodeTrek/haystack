package indexer

import (
	"haystack/utils"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

func GetDocumentId(fullPath string) string {
	return utils.Md5HashString(fullPath)
}

func GetContentHash(content []byte) string {
	return utils.Md5Hash(content)
}

// isTextMIME checks if the MIME type is a text type or a known text-like type
// such as JSON or XML. It returns true if the MIME type is likely to be text.
// This function is a simplified version and may not cover all cases.
// It is used to determine if the content is likely to be text based on its MIME type.
func isTextMIME(mtype string) bool {
	if strings.HasPrefix(mtype, "text/") {
		return true
	}

	if strings.HasSuffix(mtype, "+json") ||
		strings.HasSuffix(mtype, "+xml") ||
		mtype == "application/json" ||
		mtype == "application/xml" ||
		mtype == "application/javascript" {
		return true
	}
	return false
}

// isProbablyText checks if the data is likely to be text based on a heuristic.
// It counts the number of printable characters and checks if they are
// above a certain threshold. This is a simple heuristic and may not be
// perfect, but it should work for most text files.
func isProbablyText(data []byte) bool {
	var printable int
	for _, b := range data {
		if (b >= 32 && b <= 126) || b == '\n' || b == '\r' || b == '\t' || (b >= 128) {
			printable++
		}
	}

	// Check if the ratio of printable characters is greater than 95%
	// This is a heuristic and may not be perfect
	// but should work for most text files.
	return float64(printable)/float64(len(data)) > 0.95
}

// IsLikelyText checks if the data is likely to be text based on its MIME type
// and a heuristic for binary data. It returns true if the data is likely text.
func IsLikelyText(data []byte) bool {
	minetype := mimetype.Detect(data)
	if isTextMIME(minetype.String()) {
		return true
	}

	return isProbablyText(data)
}
