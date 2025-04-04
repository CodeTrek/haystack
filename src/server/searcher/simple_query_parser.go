package searcher

import (
	"errors"
	"regexp"
	"strings"
)

type SimpleQuery struct {
	OrClauses []*SimpleOrClause
}

type SimpleOrClause struct {
	AndPatterns []*SimplePattern
}

type SimplePattern struct {
	Pattern string
	Prefix  string
}

var rePrefix = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_-]+`)

func ParseQuerySimple(query string) (*SimpleQuery, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("query is empty")
	}

	orClauses := []*SimpleOrClause{}
	for _, orClause := range strings.Split(query, "|") {
		orClause = strings.TrimSpace(orClause)
		if orClause == "" {
			continue
		}

		andPatterns := []*SimplePattern{}
		for _, andPattern := range strings.Split(orClause, " ") {
			andPattern = strings.TrimSpace(andPattern)
			if andPattern == "" || andPattern == "AND" {
				continue
			}

			prefixes := rePrefix.FindAllString(andPattern, 1)
			if len(prefixes) > 0 {
				andPatterns = append(andPatterns, &SimplePattern{
					Pattern: andPattern,
					Prefix:  prefixes[0],
				})
			}
		}

		if len(andPatterns) == 0 {
			continue
		}

		orClauses = append(orClauses, &SimpleOrClause{
			AndPatterns: andPatterns,
		})
	}

	if len(orClauses) == 0 {
		return nil, errors.New("query is empty")
	}

	return &SimpleQuery{
		OrClauses: orClauses,
	}, nil
}
