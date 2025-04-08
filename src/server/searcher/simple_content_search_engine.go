package searcher

import (
	"errors"
	"haystack/server/core/storage"
	"haystack/server/core/workspace"
	"log"
	"regexp"
	"strings"
)

var rePrefix = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_-]+`)

type SimpleContentSearchEngine struct {
	Workspace *workspace.Workspace
	OrClauses []*SimpleContentSearchEngineAndClause
}

type SimpleContentSearchEngineAndClause struct {
	AndTerms []*SimpleContentSearchEngineTerm
}

type SimpleContentSearchEngineTerm struct {
	Pattern string
	Prefix  string
	Regex   *regexp.Regexp
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

	if len(q.OrClauses) > 1 {
		log.Printf("Merged Documents: ==>`%s` found %d documents", q.String(), len(result.DocIds))
	}

	return result, nil
}

func (q *SimpleContentSearchEngineAndClause) CollectDocuments(workspaceId string) (*storage.SearchResult, error) {
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
		for docid := range result.DocIds {
			if _, ok := r.DocIds[docid]; !ok {
				delete(result.DocIds, docid)
			}
		}
	}

	if len(q.AndTerms) > 1 {
		log.Printf("Merged Documents: =>`%s` found %d documents", q.String(), len(result.DocIds))
	}

	return result, nil
}

func (q *SimpleContentSearchEngineTerm) CollectDocuments(workspaceId string) storage.SearchResult {
	r := storage.Search(workspaceId, q.Prefix, -1)
	log.Printf("CollectDocuments: |--`%s` found %d documents", q.String(), len(r.DocIds))
	return r
}

func NewSimpleContentSearchEngine(workspace *workspace.Workspace) *SimpleContentSearchEngine {
	return &SimpleContentSearchEngine{
		Workspace: workspace,
	}
}

func (q *SimpleContentSearchEngine) IsLineMatch(line string) [][]int {
	for _, orClause := range q.OrClauses {
		matches := orClause.IsLineMatch(line)
		if len(matches) > 0 {
			return matches
		}
	}

	return [][]int{}
}

func (q *SimpleContentSearchEngineAndClause) IsLineMatch(line string) [][]int {
	if len(q.AndTerms) == 0 {
		return [][]int{}
	}

	// We should match keywords in orders
	matches := [][]int{}
	for _, term := range q.AndTerms {
		match := term.IsLineMatch(line)
		if len(match) == 0 {
			return [][]int{}
		}
		matches = append(matches, match)
		line = line[match[1]:]
	}

	return matches
}

func (q *SimpleContentSearchEngineTerm) IsLineMatch(line string) []int {
	match := q.Regex.FindAllStringIndex(line, 1)
	if len(match) > 0 {
		return []int{match[0][0], match[0][1]}
	}

	return []int{}
}

func (q *SimpleContentSearchEngine) Compile(query string, caseSensitive bool) error {
	query = strings.TrimSpace(query)
	if query == "" {
		return errors.New("query is empty")
	}

	orClauses := []*SimpleContentSearchEngineAndClause{}
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
				regPattern := andPattern
				regPattern = strings.ReplaceAll(regPattern, ".", "\\.")
				regPattern = strings.ReplaceAll(regPattern, "*", ".{0,32}")
				regPattern = strings.ReplaceAll(regPattern, "?", ".?")
				regPattern = strings.ReplaceAll(regPattern, "[", "\\[")
				regPattern = strings.ReplaceAll(regPattern, "]", "\\]")
				regPattern = strings.ReplaceAll(regPattern, "^", "\\^")
				regPattern = strings.ReplaceAll(regPattern, "$", "\\$")
				regPattern = strings.ReplaceAll(regPattern, ":", "\\:")
				casePattern := ""
				if !caseSensitive {
					casePattern = "(?i)"
				}
				reg, err := regexp.Compile(casePattern + "[^a-zA-Z0-9]" + regPattern)
				if err != nil {
					return err
				}

				andPatterns = append(andPatterns, &SimpleContentSearchEngineTerm{
					Pattern: andPattern,
					Prefix:  strings.ToLower(prefixes[0]),
					Regex:   reg,
				})
			}
		}

		if len(andPatterns) == 0 {
			continue
		}

		orClauses = append(orClauses, &SimpleContentSearchEngineAndClause{
			AndTerms: andPatterns,
		})
	}

	if len(orClauses) == 0 {
		return errors.New("query is empty")
	}

	q.OrClauses = orClauses
	return nil
}

func (q *SimpleContentSearchEngine) String() string {
	orClauses := []string{}
	for _, orClause := range q.OrClauses {
		orClauses = append(orClauses, orClause.String())
	}

	return strings.Join(orClauses, " | ")
}

func (t *SimpleContentSearchEngineAndClause) String() string {
	terms := []string{}
	for _, term := range t.AndTerms {
		terms = append(terms, term.String())
	}

	return strings.Join(terms, " AND ")
}

func (t *SimpleContentSearchEngineTerm) String() string {
	return t.Pattern
}
