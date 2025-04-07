package client

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"haystack/shared/types"
)

const (
	apiBaseURL = "http://127.0.0.1:13134/api/v1"
)

func Run() {
	// Check if there are enough command line arguments
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	// Get the command (first argument)
	command := os.Args[1]

	switch command {
	case "search":
		handleSearch(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage: haystack.exe <command> [arguments]")
	fmt.Println("Commands:")
	fmt.Println("  search <query>    Search for documents matching the query")
}

func handleSearch(args []string) {
	// Create a new FlagSet for the search command
	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)

	// Define flags for search command
	maxResults := searchCmd.Int("limit", 100, "Maximum number of results")
	path := searchCmd.String("path", "", "Path to search in")
	include := searchCmd.String("include", "", "File patterns to include")
	exclude := searchCmd.String("exclude", "", "File patterns to exclude")
	workspace := searchCmd.String("workspace", "D:\\Edge\\src", "Workspace path to search in")
	caseSensitive := searchCmd.Bool("case-sensitive", false, "Enable case-sensitive search")

	// Parse the remaining arguments
	searchCmd.Parse(args)

	// Get the search query (all non-flag arguments)
	query := strings.Join(searchCmd.Args(), " ")

	if query == "" {
		fmt.Println("Error: Search query cannot be empty")
		fmt.Println("Usage: haystack.exe search [options] <query>")
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
			MaxResults: *maxResults,
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
	fmt.Printf("Searching for: %s (limit: %d)\n", query, *maxResults)
	results, err := sendSearchRequest(searchReq)
	if err != nil {
		fmt.Printf("Error searching: %v\n", err)
		return
	}

	// Display results
	displaySearchResults(results)
}

func sendSearchRequest(req types.SearchContentRequest) (*types.SearchContentResponse, error) {
	// Marshal request to JSON
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Send request
	resp, err := client.Post(
		apiBaseURL+"/search/content",
		"application/json",
		bytes.NewBuffer(reqData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to API: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Check if status code indicates error
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var searchResp types.SearchContentResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &searchResp, nil
}

func displaySearchResults(resp *types.SearchContentResponse) {
	if resp.Code != 0 {
		fmt.Printf("Error from server: %s\n", resp.Message)
		return
	}

	if len(resp.Data.Results) == 0 {
		fmt.Println("No results found.")
		return
	}

	totalHits := 0
	for _, result := range resp.Data.Results {
		totalHits += len(result.Lines)
	}

	fmt.Printf("Found %d results in %d files:\n", totalHits, len(resp.Data.Results))
	fmt.Println("----------------------------------------")

	for _, result := range resp.Data.Results {
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

	if resp.Data.Truncate {
		fmt.Println("(Search results were truncated. Try narrowing your search.)")
	}
}
