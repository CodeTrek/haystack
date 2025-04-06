package searcher

import (
	"errors"
	"haystack/conf"
	"haystack/server/core/storage"
	"haystack/server/core/workspace"
	"haystack/shared/types"
	"haystack/utils"
	"log"
	"path/filepath"
	"regexp"
	"strings"
)

var rePrefix = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_-]+`)

type SimpleContentSearchEngine struct {
	Workspace *workspace.Workspace
	OrClauses []*SimpleContentSearchEngineOrClause
	Limit     types.SearchLimit
	Filters   QueryFilters
}

type SimpleContentSearchEngineOrClause struct {
	AndTerms []*SimpleContentSearchEngineTerm
}

type SimpleContentSearchEngineTerm struct {
	Pattern string
	Prefix  string
}

func (q *SimpleContentSearchEngine) CollectDocuments() (*storage.SearchResult, error) {
	rs := []*storage.SearchResult{}
	// Collect the documents for each or clause
	for _, orClause := range q.OrClauses {
		r, err := orClause.CollectDocuments(q.Workspace.ID)
		if err != nil {
			continue
		}

		rs = append(rs, r)
	}

	if len(rs) == 0 {
		return &storage.SearchResult{}, nil
	}

	// Merge the results, we use the first result as the base and merge all other results into it
	result := rs[0]
	for _, r := range rs[1:] {
		for docid := range r.DocIds {
			result.DocIds[docid] = struct{}{}
		}
	}

	return result, nil
}

func (q *SimpleContentSearchEngineOrClause) CollectDocuments(workspaceId string) (*storage.SearchResult, error) {
	// Collect the documents for each term
	rs := []*storage.SearchResult{}
	for _, term := range q.AndTerms {
		r := term.CollectDocuments(workspaceId)
		rs = append(rs, &r)
	}

	if len(rs) == 0 {
		return &storage.SearchResult{
			DocIds: make(map[string]struct{}),
		}, nil
	}

	// Merge the results, the documents should match all "AND" terms
	// We use the first result as the base and remove documents that don't match the other results
	result := rs[0]
	for _, r := range rs[1:] {
		for docid := range r.DocIds {
			delete(result.DocIds, docid)
		}
	}

	return result, nil
}

func (q *SimpleContentSearchEngineTerm) CollectDocuments(workspaceId string) storage.SearchResult {
	r := storage.Search(workspaceId, q.Prefix, -1)
	log.Printf("SimpleContentSearchEngineTerm.CollectDocuments: `%s` found %d documents", q.String(), len(r.DocIds))
	return r
}

func NewSimpleContentSearchEngine(workspace *workspace.Workspace, limit *types.SearchLimit,
	filter *types.SearchFilters) *SimpleContentSearchEngine {
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
	var pathFilter = ""
	if filter != nil {
		pathFilter = filepath.Clean(filter.Path)
		if filter.Include != "" {
			includeFilter = utils.NewSimpleFilter(strings.Split(filter.Include, ","), workspace.Path)
		}

		if filter.Exclude != "" {
			excludeFilter = utils.NewSimpleFilter(strings.Split(filter.Exclude, ","), workspace.Path)
		}
	}

	return &SimpleContentSearchEngine{
		Workspace: workspace,
		Limit:     queryLimit,
		Filters: QueryFilters{
			Path:    pathFilter,
			Include: includeFilter,
			Exclude: excludeFilter,
		},
	}
}

func (q *SimpleContentSearchEngine) IsLineMatch(line string) bool {
	return false
}

func (q *SimpleContentSearchEngine) Compile(query string) error {
	query = strings.TrimSpace(query)
	if query == "" {
		return errors.New("query is empty")
	}

	orClauses := []*SimpleContentSearchEngineOrClause{}
	for _, orClause := range strings.Split(query, "|") {
		orClause = strings.TrimSpace(orClause)
		if orClause == "" {
			continue
		}

		andPatterns := []*SimpleContentSearchEngineTerm{}
		for _, andPattern := range strings.Split(orClause, " ") {
			andPattern = strings.TrimSpace(andPattern)
			if andPattern == "" || andPattern == "AND" {
				continue
			}

			prefixes := rePrefix.FindAllString(andPattern, 1)
			if len(prefixes) > 0 {
				andPatterns = append(andPatterns, &SimpleContentSearchEngineTerm{
					Pattern: andPattern,
					Prefix:  strings.ToLower(prefixes[0]),
				})
			}
		}

		if len(andPatterns) == 0 {
			continue
		}

		orClauses = append(orClauses, &SimpleContentSearchEngineOrClause{
			AndTerms: andPatterns,
		})
	}

	if len(orClauses) == 0 {
		return errors.New("query is empty")
	}

	q.OrClauses = orClauses
	return nil
}

func (t *SimpleContentSearchEngineTerm) String() string {
	return t.Pattern
}
