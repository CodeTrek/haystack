package server

import (
	"context"
	"fmt"
	"haystack/conf"
	"haystack/server/core/workspace"
	"haystack/server/searcher"
	"haystack/shared/running"
	"haystack/shared/types"
	"haystack/utils"
	"log"
	"net/http"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ToolName string

const (
	HaystackSearch ToolName = "HaystackSearch"
)

// mcpInit initializes and sets up the Model Context Protocol (MCP) server
func mcpInit() {
	// Create a new MCP server instance
	mcpServer := server.NewMCPServer(
		"Haystack",
		running.Version(),
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithLogging(),
	)

	// Register MCP tools (framework only, implementations will be added later)
	registerMCPTools(mcpServer)

	sse := server.NewSSEServer(mcpServer,
		server.WithBaseURL(fmt.Sprintf("http://localhost:%d", conf.Get().Global.Port)),
		server.WithBasePath("/mcp"))

	http.HandleFunc("/mcp/", func(w http.ResponseWriter, r *http.Request) {
		sse.ServeHTTP(w, r)
		log.Printf("MCP request: %s %s", r.Method, r.URL.Path)
	})
	log.Println("MCP server initialized at /mcp endpoint")
}

// registerMCPTools registers all the MCP tools with the server
func registerMCPTools(mcpServer *server.MCPServer) {
	// Register search tool
	config := conf.Get()

	mcpServer.AddTool(mcp.NewTool(string(HaystackSearch),
		mcp.WithDescription("Search for code in the Haystack index. The search engine supports prefix matching and "+
			"logical operators to help you find exactly what you're looking for in your codebase."),
		mcp.WithString("query",
			mcp.Description("The search query. Supports the following syntax features:\n"+
				"- Basic terms: single words like 'function' or exact phrases in quotes like \"hello world\"\n"+
				"- Prefix matching: 'func*' matches 'function', 'functional', etc. (wildcard only at end of term)\n"+
				"- Logical operators: 'AND' (or space) for conjunction, '|' for OR operator\n"+
				"- Examples: 'error AND handle', 'create | update', 'init*'"),
			mcp.Required(),
		),
		mcp.WithString("workspace",
			mcp.Description("The workspace to search in, normally it's the absolute path to the project directory, "+
				"e.g. /home/user/projects/project1. Please always passing current workspace path."),
			mcp.Required(),
		),
		mcp.WithString("path",
			mcp.Description("The path to search in, related to workspace, e.g. src/core"),
		),
		mcp.WithString("filter", mcp.Description("Filter the search results by file path. The filter supports "+
			"glob patterns, separated by comma(','), e.g. 'src/**/*.go,*.cc' to search only in Go files in the src directory "+
			"or *.cc files in all directory.")),
		mcp.WithString("exclude", mcp.Description("Exclude files from the search. The exclude filter supports glob "+
			"patterns, separated by comma, e.g. 'test/**/*.go' to exclude all Go test files.")),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of results to return. The search will stop once this limit is reached, "+
				"which can improve performance for large codebases.\n"+
				fmt.Sprintf("Currently, the default limit is %d, and the maximum limit is %d.\n",
					config.Client.DefaultLimit.MaxResults, config.Server.Search.Limit.MaxResults))),
	), handleSearch)

	log.Println("MCP tools registered")
}

// Tool handler function stubs - implementations will be added later

// searchHandler handles search requests from MCP
func handleSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	query, ok1 := arguments["query"].(string)
	workspacePath, ok2 := arguments["workspace"].(string)
	limit, _ := arguments["limit"].(int)
	path, _ := arguments["path"].(string)
	filter, _ := arguments["filter"].(string)
	exclude, _ := arguments["exclude"].(string)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("invalid arguments")
	}

	workspacePath = utils.NormalizePath(workspacePath)
	if !filepath.IsAbs(workspacePath) {
		return nil, fmt.Errorf("workspace is not absolute")
	}

	workspace, err := workspace.GetByPath(workspacePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %v", err)
	}

	if path != "" {
		path = utils.NormalizePath(path)
		if filepath.IsAbs(path) {
			return nil, fmt.Errorf("path could not be absolute")
		}
	}

	req := types.SearchContentRequest{
		Query:     query,
		Workspace: workspacePath,
		Limit: &types.SearchLimit{
			MaxResults:        limit,
			MaxResultsPerFile: conf.Get().Server.Search.Limit.MaxResultsPerFile,
		},
		Filters: &types.SearchFilters{
			Path:    path,
			Include: filter,
			Exclude: exclude,
		},
	}

	results, truncate := searcher.SearchContent(workspace, &req)
	resultCount := 0
	for _, result := range results {
		resultCount += len(result.Lines)
	}

	var toTruncated = func(truncated bool) string {
		if truncated {
			return " (truncated)"
		}
		return ""
	}

	tr := &mcp.CallToolResult{}
	tr.Content = append(tr.Content, mcp.TextContent{
		Type: "text",
		Text: fmt.Sprintf("Found %d results in %d files%s", resultCount, len(results), toTruncated(truncate)),
	})

	if len(results) == 0 {
		tr.Content = append(tr.Content, mcp.TextContent{
			Type: "text",
			Text: "No results found.",
		})
		return tr, nil
	}

	for _, result := range results {
		tr.Content = append(tr.Content, mcp.TextContent{
			Type: "text",
			Text: fmt.Sprintf("File: %s, %d result%s", result.File, len(result.Lines), toTruncated(result.Truncate)),
		})
		for _, line := range result.Lines {
			tr.Content = append(tr.Content, mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Line %d: %s", line.Line.LineNumber, line.Line.Content),
			})
		}
		tr.Content = append(tr.Content, mcp.TextContent{
			Type: "text",
			Text: "",
		})
	}

	return tr, nil
}
