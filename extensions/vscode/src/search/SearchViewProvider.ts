import * as vscode from 'vscode';
import { HaystackProvider } from './haystackProvider';
import { SearchContentResult } from '../types/search';

export class SearchViewProvider implements vscode.WebviewViewProvider {
    public static readonly viewType = 'haystackSearch';

    constructor(
        private readonly _extensionUri: vscode.Uri,
        private readonly _haystackProvider: HaystackProvider
    ) {}

    public resolveWebviewView(
        webviewView: vscode.WebviewView,
        context: vscode.WebviewViewResolveContext,
        _token: vscode.CancellationToken,
    ) {
        webviewView.webview.options = {
            enableScripts: true,
            localResourceRoots: [this._extensionUri]
        };

        webviewView.webview.html = this._getHtmlForWebview(webviewView.webview);

        // Handle messages from the webview
        webviewView.webview.onDidReceiveMessage(async (data) => {
            switch (data.type) {
                case 'search':
                    await this._handleSearch(webviewView.webview, data.query, data.options);
                    break;
                case 'openFile':
                    await this._handleOpenFile(data.file, data.line);
                    break;
            }
        });
    }

    private async _handleSearch(webview: vscode.Webview, query: string, options: {
        caseSensitive: boolean;
        include: string;
        exclude: string;
        maxResults?: number;
    }) {
        try {
            const results = await this._haystackProvider.search(query, {
                caseSensitive: options.caseSensitive,
                include: options.include,
                exclude: options.exclude,
                maxResults: options.maxResults || 100
            });

            webview.postMessage({
                type: 'searchResults',
                results: results
            });
        } catch (error) {
            vscode.window.showErrorMessage(`Search failed: ${error}`);
        }
    }

    private async _handleOpenFile(file: string, line: number) {
        const uri = vscode.Uri.file(file);
        const document = await vscode.workspace.openTextDocument(uri);
        const editor = await vscode.window.showTextDocument(document);
        const position = new vscode.Position(line - 1, 0);
        editor.revealRange(new vscode.Range(position, position), vscode.TextEditorRevealType.InCenter);
    }

    private _getHtmlForWebview(webview: vscode.Webview) {
        return `
            <!DOCTYPE html>
            <html lang="en">
            <head>
                <meta charset="UTF-8">
                <meta name="viewport" content="width=device-width, initial-scale=1.0">
                <title>Haystack Search</title>
                <style>
                    body {
                        padding: 0;
                        margin: 0;
                        font-family: var(--vscode-font-family);
                        color: var(--vscode-foreground);
                    }
                    .search-container {
                        padding: 10px;
                    }
                    .search-input-container {
                        display: flex;
                        gap: 6px;
                        margin-bottom: 8px;
                    }
                    .search-input {
                        flex: 1;
                        padding: 4px 6px;
                        background: var(--vscode-input-background);
                        color: var(--vscode-input-foreground);
                        border: 1px solid var(--vscode-input-border);
                        border-radius: 2px;
                        font-size: 13px;
                        line-height: 18px;
                    }
                    .search-input:focus {
                        outline: 1px solid var(--vscode-focusBorder);
                        outline-offset: -1px;
                    }
                    .search-button {
                        padding: 4px 8px;
                        background: var(--vscode-button-background);
                        color: var(--vscode-button-foreground);
                        border: none;
                        border-radius: 2px;
                        cursor: pointer;
                        font-size: 13px;
                        line-height: 18px;
                    }
                    .search-button:hover {
                        background: var(--vscode-button-hoverBackground);
                    }
                    .search-options {
                        margin-bottom: 8px;
                    }
                    .search-option {
                        margin-bottom: 6px;
                        display: flex;
                        align-items: center;
                        gap: 6px;
                    }
                    .search-option input[type="checkbox"] {
                        margin: 0;
                    }
                    .search-option label {
                        font-size: 13px;
                        user-select: none;
                    }
                    .search-results {
                        margin-top: 8px;
                    }
                    .result-item {
                        padding: 6px 8px;
                        cursor: pointer;
                        border-radius: 2px;
                    }
                    .result-item:hover {
                        background: var(--vscode-list-hoverBackground);
                    }
                    .result-path {
                        font-size: 12px;
                        color: var(--vscode-descriptionForeground);
                        margin-bottom: 2px;
                    }
                    .result-preview {
                        font-family: var(--vscode-editor-font-family);
                        font-size: 13px;
                        margin-top: 2px;
                        white-space: pre-wrap;
                        overflow-wrap: break-word;
                    }
                </style>
            </head>
            <body>
                <div class="search-container">
                    <div class="search-input-container">
                        <input type="text" class="search-input" placeholder="Search in files..." id="searchInput">
                        <button class="search-button" id="searchButton">Search</button>
                    </div>
                    <div class="search-options">
                        <div class="search-option">
                            <input type="checkbox" id="caseSensitive">
                            <label for="caseSensitive">Case sensitive</label>
                        </div>
                        <div class="search-option">
                            <input type="text" class="search-input" id="includeFiles" placeholder="Files to include (e.g. *.ts)">
                        </div>
                        <div class="search-option">
                            <input type="text" class="search-input" id="excludeFiles" placeholder="Files to exclude">
                        </div>
                    </div>
                    <div class="search-results" id="searchResults"></div>
                </div>
                <script>
                    const vscode = acquireVsCodeApi();

                    document.getElementById('searchButton').addEventListener('click', () => {
                        performSearch();
                    });

                    document.getElementById('searchInput').addEventListener('keyup', (event) => {
                        if (event.key === 'Enter') {
                            performSearch();
                        }
                    });

                    function performSearch() {
                        const query = document.getElementById('searchInput').value;
                        if (!query) return;

                        const options = {
                            caseSensitive: document.getElementById('caseSensitive').checked,
                            include: document.getElementById('includeFiles').value,
                            exclude: document.getElementById('excludeFiles').value
                        };

                        vscode.postMessage({
                            type: 'search',
                            query: query,
                            options: options
                        });
                    }

                    window.addEventListener('message', event => {
                        const message = event.data;
                        switch (message.type) {
                            case 'searchResults':
                                displayResults(message.results);
                                break;
                        }
                    });

                    function displayResults(results) {
                        const container = document.getElementById('searchResults');
                        if (!container) return;

                        container.innerHTML = '';

                        results.forEach(result => {
                            const div = document.createElement('div');
                            div.className = 'result-item';
                            div.innerHTML = \`
                                <div class="result-path">\${result.file}</div>
                            \`;
                            div.addEventListener('click', () => {
                                vscode.postMessage({
                                    type: 'openFile',
                                    file: result.file,
                                    line: result.line
                                });
                            });
                            container.appendChild(div);
                        });
                    }
                </script>
            </body>
            </html>
        `;
    }
}
