package client

import (
	"encoding/json"
	"flag"
	"fmt"
	"haystack/conf"
	"haystack/shared/running"
	"haystack/shared/types"
	"strings"
)

func handleSearch(args []string) {
	// Create a new FlagSet for the search command
	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)

	// Define flags for search command
	maxResults := searchCmd.Int("limit", conf.Get().Client.DefaultLimit.MaxResults, "Maximum number of results")
	maxResultsPerFile := searchCmd.Int("limit-per-file", conf.Get().Client.DefaultLimit.MaxResultsPerFile, "Maximum number of results per file")
	path := searchCmd.String("path", "", "Path to search in")
	include := searchCmd.String("include", "", "File patterns to include")
	exclude := searchCmd.String("exclude", "", "File patterns to exclude")
	workspace := searchCmd.String("workspace", conf.Get().Client.DefaultWorkspace, "Workspace path to search in")
	caseSensitive := searchCmd.Bool("case-sensitive", false, "Enable case-sensitive search")

	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Println("Usage: " + running.ExecutableName() + " search <query> [options]")
		fmt.Println("Options:")
		searchCmd.PrintDefaults()
		return
	}

	// Parse the remaining arguments
	searchCmd.Parse(args)

	// Get the search query (all non-flag arguments)
	query := strings.Join(searchCmd.Args(), " ")

	if query == "" {
		fmt.Println("Error: Search query cannot be empty")
		fmt.Println("Usage: " + running.ExecutableName() + " search [options] <query>")
		fmt.Println("Options:")
		searchCmd.PrintDefaults()
		return
	}

	// Prepare the search request
	searchReq := types.SearchContentRequest{
		Workspace:     *workspace,
		Query:         query,
		CaseSensitive: *caseSensitive,
		Limit: &types.SearchLimit{
			MaxResults:        *maxResults,
			MaxResultsPerFile: *maxResultsPerFile,
		},
	}

	// Add filters if specified
	if *path != "" || *include != "" || *exclude != "" {
		searchReq.Filters = &types.SearchFilters{
			Path:    *path,
			Include: *include,
			Exclude: *exclude,
		}
	}

	// Execute the search
	fmt.Printf("Searching for: %s (limit: %d, limit-per-file: %d)\n", query, *maxResults, *maxResultsPerFile)
	results, err := sendSearchRequest(searchReq)
	if err != nil {
		fmt.Printf("Error searching: %v\n", err)
		return
	}

	// Display results
	displaySearchResults(results)
}

func sendSearchRequest(req types.SearchContentRequest) (*types.SearchContentResults, error) {
	// Marshal request to JSON
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	result, err := serverRequest("/search/content", reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	// Parse response
	var searchResp types.SearchContentResults
	if err := json.Unmarshal(*result.Body.Data, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &searchResp, nil
}

func displaySearchResults(resp *types.SearchContentResults) {
	if len(resp.Results) == 0 {
		fmt.Println("No results found.")
		return
	}

	totalHits := 0
	for _, result := range resp.Results {
		totalHits += len(result.Lines)
	}

	fmt.Printf("Found %d results in %d files:\n", totalHits, len(resp.Results))
	fmt.Println("----------------------------------------")

	for _, result := range resp.Results {
		fmt.Printf("File: %s\n", result.File)

		for _, match := range result.Lines {
			// // Show context before the match
			// for _, beforeLine := range match.Before {
			// 	fmt.Printf("  %4d: %s\n", beforeLine.LineNumber, beforeLine.Content)
			// }

			// Show the matching line
			fmt.Printf("→ %4d: %s\n", match.Line.LineNumber, match.Line.Content)

			// // Show context after the match
			// for _, afterLine := range match.After {
			// 	fmt.Printf("  %4d: %s\n", afterLine.LineNumber, afterLine.Content)
			// }
			// fmt.Println()
		}

		if result.Truncate {
			fmt.Println("  (Results truncated...)")
		}
		fmt.Println("----------------------------------------")
	}

	if resp.Truncate {
		fmt.Println("(Search results were truncated. Try narrowing your search.)")
	}
}
