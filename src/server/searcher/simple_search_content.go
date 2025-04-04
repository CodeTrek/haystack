package searcher

import (
	"errors"
	"haystack/conf"
	"haystack/shared/types"
	"haystack/utils"
	"path/filepath"
	"regexp"
	"strings"
)

var rePrefix = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_-]+`)

type SimpleSearchContent struct {
	OrClauses []*SimpleSearchContentOrClause
	Limit     types.SearchLimit
	Filters   QueryFilters
}

type SimpleSearchContentOrClause struct {
	AndTerms []*SimpleSearchContentTerm
}

type SimpleSearchContentTerm struct {
	Pattern string
	Prefix  string
}

func NewSimpleQuery(baseDir string, limit *types.SearchLimit, filter *types.SearchFilters) *SimpleSearchContent {
	queryLimit := conf.Get().Server.SearchLimit

	if limit != nil {
		if limit.MaxResults > 0 && limit.MaxResults < queryLimit.MaxResults {
			queryLimit.MaxResults = limit.MaxResults
		}

		if limit.MaxResultsPerFile > 0 && limit.MaxResultsPerFile < queryLimit.MaxResultsPerFile {
			queryLimit.MaxResultsPerFile = limit.MaxResultsPerFile
		}
	}

	var includeFilter *utils.SimpleFilter
	var excludeFilter *utils.SimpleFilter
	if filter != nil {
		if filter.Include != "" {
			includeFilter = utils.NewSimpleFilter(strings.Split(filter.Include, ","), baseDir)
		}

		if filter.Exclude != "" {
			excludeFilter = utils.NewSimpleFilter(strings.Split(filter.Exclude, ","), baseDir)
		}
	}

	return &SimpleSearchContent{
		Limit: queryLimit,
		Filters: QueryFilters{
			Path:    filepath.Clean(filter.Path),
			Include: includeFilter,
			Exclude: excludeFilter,
		},
	}
}

func (q *SimpleSearchContent) Compile(query string) error {
	query = strings.TrimSpace(query)
	if query == "" {
		return errors.New("query is empty")
	}

	orClauses := []*SimpleSearchContentOrClause{}
	for _, orClause := range strings.Split(query, "|") {
		orClause = strings.TrimSpace(orClause)
		if orClause == "" {
			continue
		}

		andPatterns := []*SimpleSearchContentTerm{}
		for _, andPattern := range strings.Split(orClause, " ") {
			andPattern = strings.TrimSpace(andPattern)
			if andPattern == "" || andPattern == "AND" {
				continue
			}

			prefixes := rePrefix.FindAllString(andPattern, 1)
			if len(prefixes) > 0 {
				andPatterns = append(andPatterns, &SimpleSearchContentTerm{
					Pattern: andPattern,
					Prefix:  prefixes[0],
				})
			}
		}

		if len(andPatterns) == 0 {
			continue
		}

		orClauses = append(orClauses, &SimpleSearchContentOrClause{
			AndTerms: andPatterns,
		})
	}

	if len(orClauses) == 0 {
		return errors.New("query is empty")
	}

	q.OrClauses = orClauses
	return nil
}

func (q *SimpleSearchContent) Run() ([]types.SearchContentResult, error) {
	return nil, nil
}

func (q *SimpleSearchContent) CollectFiles() ([]string, error) {
	return nil, nil
}
