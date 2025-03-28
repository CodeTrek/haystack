package parser

import (
	"regexp"
	"sort"
)

// define custom tokens keys
const (
	TSep = iota + 1
)

func Init() {
}

func ParseString(str string) []string {
	re := regexp.MustCompile(`[a-zA-Z0-9]+`)
	words := re.FindAllString(str, -1)

	uniqueWords := make(map[string]struct{})
	for _, word := range words {
		if isValidWord(word) {
			uniqueWords[word] = struct{}{}
		}
	}

	result := make([]string, 0, len(uniqueWords))
	for word := range uniqueWords {
		result = append(result, word)
	}

	sort.Strings(result)
	return result
}
