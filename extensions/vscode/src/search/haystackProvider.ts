import * as vscode from 'vscode';
import axios from 'axios';
import {
    SearchContentRequest,
    SearchContentResponse,
    SearchContentResult,
    TextSearchOptions,
    TextSearchResult
} from '../types/search';

const HAYSTACK_URL = 'http://127.0.0.1:13134/api/v1/search/content';

export class HaystackProvider {
    private workspaceRoot: string;

    constructor() {
        const workspaceFolders = vscode.workspace.workspaceFolders;
        this.workspaceRoot = workspaceFolders ? workspaceFolders[0].uri.fsPath : '';
    }

    async provideTextSearchResults(
        query: string,
        options: TextSearchOptions,
        token: vscode.CancellationToken
    ): Promise<TextSearchResult[]> {
        const searchRequest: SearchContentRequest = {
            workspace: this.workspaceRoot,
            query: query,
            case_sensitive: false,
            filters: {
                include: options.includePattern ? Object.keys(options.includePattern)[0] : undefined,
                exclude: options.excludePattern ? Object.keys(options.excludePattern)[0] : undefined
            },
            limit: {
                max_results: options.maxResults
            }
        };

        try {
            const response = await axios.post<SearchContentResponse>(HAYSTACK_URL, searchRequest);
            if (response.data.code === 0 && response.data.data?.results) {
                return this.convertResults(response.data.data.results);
            } else {
                vscode.window.showErrorMessage(`Search failed: ${response.data.message}`);
                return [];
            }
        } catch (error) {
            vscode.window.showErrorMessage(`Failed to connect to Haystack server: ${error}`);
            return [];
        }
    }

    private convertResults(results: SearchContentResult[]): TextSearchResult[] {
        const vscodeResults: TextSearchResult[] = [];

        for (const result of results) {
            if (!result.lines) continue;

            for (const lineMatch of result.lines) {
                const line = lineMatch.line;
                const range = new vscode.Range(
                    new vscode.Position(line.line_number - 1, 0),
                    new vscode.Position(line.line_number - 1, line.content.length)
                );

                vscodeResults.push({
                    uri: vscode.Uri.file(result.file),
                    range: range,
                    preview: {
                        text: line.content,
                        matches: [range]
                    }
                });
            }
        }

        return vscodeResults;
    }
}
