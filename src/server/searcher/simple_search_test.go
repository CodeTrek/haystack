package searcher

import (
	"testing"
)

func TestParseQuerySimple(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		want    *SimpleSearchContent
		wantErr bool
	}{
		{
			name:    "empty query",
			query:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			query:   "   ",
			wantErr: true,
		},
		{
			name:  "single pattern",
			query: "test",
			want: &SimpleSearchContent{
				OrClauses: []*SimpleSearchContentOrClause{
					{
						AndTerms: []*SimpleSearchContentTerm{
							{
								Pattern: "test",
								Prefix:  "test",
							},
						},
					},
				},
			},
		},
		{
			name:  "multiple AND patterns",
			query: "test1 test2",
			want: &SimpleSearchContent{
				OrClauses: []*SimpleSearchContentOrClause{
					{
						AndTerms: []*SimpleSearchContentTerm{
							{
								Pattern: "test1",
								Prefix:  "test1",
							},
							{
								Pattern: "test2",
								Prefix:  "test2",
							},
						},
					},
				},
			},
		},
		{
			name:  "OR clauses",
			query: "test1 test2 | test3",
			want: &SimpleSearchContent{
				OrClauses: []*SimpleSearchContentOrClause{
					{
						AndTerms: []*SimpleSearchContentTerm{
							{
								Pattern: "test1",
								Prefix:  "test1",
							},
							{
								Pattern: "test2",
								Prefix:  "test2",
							},
						},
					},
					{
						AndTerms: []*SimpleSearchContentTerm{
							{
								Pattern: "test3",
								Prefix:  "test3",
							},
						},
					},
				},
			},
		},
		{
			name:  "pattern with prefix",
			query: "prefix:value",
			want: &SimpleSearchContent{
				OrClauses: []*SimpleSearchContentOrClause{
					{
						AndTerms: []*SimpleSearchContentTerm{
							{
								Pattern: "prefix:value",
								Prefix:  "prefix",
							},
						},
					},
				},
			},
		},
		{
			name:  "complex query with prefixes and OR",
			query: "field1:value1 field2:value2 | field3:value3",
			want: &SimpleSearchContent{
				OrClauses: []*SimpleSearchContentOrClause{
					{
						AndTerms: []*SimpleSearchContentTerm{
							{
								Pattern: "field1:value1",
								Prefix:  "field1",
							},
							{
								Pattern: "field2:value2",
								Prefix:  "field2",
							},
						},
					},
					{
						AndTerms: []*SimpleSearchContentTerm{
							{
								Pattern: "field3:value3",
								Prefix:  "field3",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &SimpleSearchContent{}
			err := got.Compile(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseQuerySimple() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if len(got.OrClauses) != len(tt.want.OrClauses) {
				t.Errorf("ParseQuerySimple() got %d OR clauses, want %d", len(got.OrClauses), len(tt.want.OrClauses))
				return
			}

			for i, orClause := range got.OrClauses {
				wantOrClause := tt.want.OrClauses[i]
				if len(orClause.AndTerms) != len(wantOrClause.AndTerms) {
					t.Errorf("OR clause %d: got %d AND patterns, want %d", i, len(orClause.AndTerms), len(wantOrClause.AndTerms))
					continue
				}

				for j, pattern := range orClause.AndTerms {
					wantPattern := wantOrClause.AndTerms[j]
					if pattern.Pattern != wantPattern.Pattern {
						t.Errorf("pattern %d in OR clause %d: got pattern %q, want %q", j, i, pattern.Pattern, wantPattern.Pattern)
					}
					if pattern.Prefix != wantPattern.Prefix {
						t.Errorf("pattern %d in OR clause %d: got prefix %q, want %q", j, i, pattern.Prefix, wantPattern.Prefix)
					}
				}
			}
		})
	}
}
