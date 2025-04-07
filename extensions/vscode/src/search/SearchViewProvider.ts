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
                    .search-input {
                        width: 100%;
                        padding: 5px;
                        margin-bottom: 10px;
                        background: var(--vscode-input-background);
                        color: var(--vscode-input-foreground);
                        border: 1px solid var(--vscode-input-border);
                    }
                    .search-options {
                        margin-bottom: 10px;
                    }
                    .search-option {
                        margin-bottom: 5px;
                    }
                    .search-results {
                        margin-top: 10px;
                    }
                    .result-item {
                        padding: 5px;
                        cursor: pointer;
                        border-bottom: 1px solid var(--vscode-list-inactiveSelectionBackground);
                    }
                    .result-item:hover {
                        background: var(--vscode-list-hoverBackground);
                    }
                    .result-path {
                        font-size: 0.9em;
                        color: var(--vscode-descriptionForeground);
                    }
                    .result-preview {
                        font-family: var(--vscode-editor-font-family);
                        margin-top: 3px;
                    }
                </style>
            </head>
            <body>
                <div class="search-container">
                    <input type="text" class="search-input" placeholder="Search in files..." id="searchInput">
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
                    let searchTimeout;

                    document.getElementById('searchInput').addEventListener('input', (e) => {
                        clearTimeout(searchTimeout);
                        searchTimeout = setTimeout(() => performSearch(), 300);
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
