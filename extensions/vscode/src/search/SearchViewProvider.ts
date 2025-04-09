import * as vscode from 'vscode';
import { HaystackProvider } from './haystackProvider';
import { SearchContentResult } from '../types/search';

export class SearchViewProvider implements vscode.WebviewViewProvider {
    public static readonly viewType = 'haystackSearch';

    // Store the webview view reference
    private _view?: vscode.WebviewView;

    constructor(
        private readonly _extensionUri: vscode.Uri,
        private readonly _haystackProvider: HaystackProvider
    ) {}

    public resolveWebviewView(
        webviewView: vscode.WebviewView,
        context: vscode.WebviewViewResolveContext,
        _token: vscode.CancellationToken,
    ) {
        // Store the webview reference
        this._view = webviewView;

        webviewView.webview.options = {
            enableScripts: true,
            localResourceRoots: [this._extensionUri]
        };

        // Set HTML content only if it hasn't been set before
        // This is crucial - we don't want to reset the HTML when the view becomes visible again
        if (!webviewView.webview.html) {
            webviewView.webview.html = this._getHtmlForWebview(webviewView.webview);
        }

        // Handle messages from the webview
        webviewView.webview.onDidReceiveMessage(async (data) => {
            switch (data.type) {
                case 'search':
                    await this._handleSearch(webviewView.webview, data.query, data.options);
                    break;
                case 'openFile':
                    await this._handleOpenFile(data.file, data.line, data.start, data.end);
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
                maxResults: options.maxResults || 200
            });

            webview.postMessage({
                type: 'searchResults',
                results: results || []
            });
        } catch (error) {
            // Don't show error message for empty results
            console.log(`Search error: ${error}`);

            // Send empty results instead of showing an error
            webview.postMessage({
                type: 'searchResults',
                results: []
            });
        }
    }

    private async _handleOpenFile(file: string, line: number, start?: number, end?: number) {
        try {
            // Resolve the full path relative to workspace root
            const workspaceRoot = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
            if (!workspaceRoot) {
                throw new Error('No workspace folder found');
            }

            // Join the workspace path with the file path, handling both absolute and relative paths
            const fullPath = file.startsWith(workspaceRoot) ? file : vscode.Uri.joinPath(vscode.Uri.file(workspaceRoot), file).fsPath;

            const uri = vscode.Uri.file(fullPath);
            const document = await vscode.workspace.openTextDocument(uri);
            const editor = await vscode.window.showTextDocument(document);

            // Create a position at the target line
            const position = new vscode.Position(line - 1, 0);

            // First reveal the line
            editor.revealRange(new vscode.Range(position, position), vscode.TextEditorRevealType.InCenter);

            // If start and end positions are provided, create a selection for the match
            if (start !== undefined && end !== undefined) {
                const startPos = new vscode.Position(line - 1, start);
                const endPos = new vscode.Position(line - 1, end);
                editor.selection = new vscode.Selection(startPos, endPos);
            } else {
                // Fall back to selecting the beginning of the line
                editor.selection = new vscode.Selection(position, position);
            }
        } catch (error) {
            vscode.window.showErrorMessage(`Failed to open file: ${error}`);
        }
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
                        position: relative;
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
                    .search-option-button {
                        position: absolute;
                        right: 24px;
                        top: 50%;
                        transform: translateY(-50%);
                        cursor: pointer;
                        color: var(--vscode-descriptionForeground);
                        background: none;
                        border: none;
                        width: 20px;
                        height: 18px;
                        display: flex;
                        align-items: center;
                        justify-content: center;
                        font-size: 11px;
                        padding: 0;
                        border-radius: 3px;
                    }
                    .search-option-button:hover {
                        background: var(--vscode-button-hoverBackground);
                        opacity: 0.8;
                    }
                    .search-option-button.active {
                        background: var(--vscode-button-background);
                        color: var(--vscode-button-foreground);
                    }
                    .search-option-button.active:hover {
                        background: var(--vscode-button-background);
                        opacity: 0.9;
                    }
                    .search-options-toggle {
                        position: absolute;
                        right: 6px;
                        top: 50%;
                        transform: translateY(-50%);
                        cursor: pointer;
                        color: var(--vscode-descriptionForeground);
                        background: none;
                        border: none;
                        width: 18px;
                        height: 18px;
                        display: flex;
                        align-items: center;
                        justify-content: center;
                        font-size: 12px;
                        padding: 0;
                        border-radius: 3px;
                    }
                    .search-options-toggle:hover {
                        background: var(--vscode-button-hoverBackground);
                        opacity: 0.8;
                    }
                    .search-options-toggle.active {
                        background: var(--vscode-button-background);
                        color: var(--vscode-button-foreground);
                    }
                    .search-options-toggle.active:hover {
                        background: var(--vscode-button-background);
                        opacity: 0.9;
                    }
                    .search-options {
                        margin-bottom: 8px;
                        display: none;
                    }
                    .search-options.visible {
                        display: block;
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
                    .search-summary {
                        font-size: 12px;
                        color: var(--vscode-descriptionForeground);
                        margin-bottom: 8px;
                        padding: 4px 8px;
                    }
                    .file-group {
                        margin-bottom: 12px;
                        background: var(--vscode-list-inactiveSelectionBackground);
                        border-radius: 4px;
                        overflow: hidden;
                    }
                    .file-header {
                        padding: 6px 8px;
                        background: var(--vscode-list-activeSelectionBackground);
                        display: flex;
                        justify-content: space-between;
                        align-items: center;
                        cursor: pointer;
                    }
                    .file-header:hover {
                        background: var(--vscode-list-activeSelectionBackground);
                    }
                    .file-path {
                        font-size: 12px;
                        color: var(--vscode-foreground);
                        white-space: nowrap;
                        overflow: hidden;
                        text-overflow: ellipsis;
                        direction: rtl;
                        text-align: left;
                        flex: 1;
                        margin-right: 8px;
                    }
                    .match-count {
                        font-size: 11px;
                        color: var(--vscode-descriptionForeground);
                        padding: 2px 6px;
                        border-radius: 10px;
                        background: var(--vscode-badge-background);
                        flex-shrink: 0;
                    }
                    .result-item {
                        padding: 4px 8px 4px 24px;
                        cursor: pointer;
                        display: flex;
                        align-items: center;
                        gap: 8px;
                    }
                    .result-item:hover {
                        background: var(--vscode-list-hoverBackground);
                    }
                    .line-number {
                        font-size: 12px;
                        color: var(--vscode-descriptionForeground);
                        min-width: 13px;
                        text-align: right;
                        flex-shrink: 0;
                    }
                    .line-content {
                        font-family: 'Consolas', 'Courier New', monospace;
                        font-size: 12px;
                        white-space: nowrap;
                        overflow: hidden;
                        text-overflow: ellipsis;
                        flex: 1;
                    }
                    .match-highlight {
                        background-color: var(--vscode-editor-findMatchHighlightBackground);
                        color: var(--vscode-editor-findMatchHighlightForeground);
                        padding: 0 1px;
                        border-radius: 2px;
                    }
                </style>
            </head>
            <body>
                <div class="search-container">
                    <div class="search-input-container">
                        <input type="text" class="search-input" placeholder="Search in files..." id="searchInput">
                        <button class="search-option-button" id="caseSensitiveBtn" title="Case sensitive">Aa</button>
                        <button class="search-options-toggle" id="optionsToggle">⋮</button>
                    </div>
                    <div class="search-options" id="searchOptions">
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

                    function performSearch() {
                        const query = document.getElementById('searchInput').value;
                        if (!query) return;

                        const options = {
                            caseSensitive: document.getElementById('caseSensitiveBtn').classList.contains('active'),
                            include: document.getElementById('includeFiles').value,
                            exclude: document.getElementById('excludeFiles').value
                        };

                        vscode.postMessage({
                            type: 'search',
                            query: query,
                            options: options
                        });
                    }

                    function displayResults(results) {
                        const container = document.getElementById('searchResults');
                        if (!container) return;

                        container.innerHTML = '';

                        // Check if results is empty or undefined
                        if (!results || results.length === 0) {
                            const emptyMessage = document.createElement('div');
                            emptyMessage.className = 'search-summary';
                            emptyMessage.textContent = 'No results found';
                            container.appendChild(emptyMessage);
                            return;
                        }

                        // Add search summary
                        const totalMatches = results.reduce((sum, result) => sum + (result.lines?.length || 0), 0);
                        const summary = document.createElement('div');
                        summary.className = 'search-summary';
                        summary.textContent = \`\${totalMatches} results in \${results.length} files\`;
                        container.appendChild(summary);

                        // Group results by file
                        results.forEach(result => {
                            const fileGroup = document.createElement('div');
                            fileGroup.className = 'file-group';

                            // File header with path and match count
                            const fileHeader = document.createElement('div');
                            fileHeader.className = 'file-header';
                            fileHeader.innerHTML = \`
                                <span class="file-path">\${result.file}</span>
                                <span class="match-count">\${result.lines?.length || 0}</span>
                            \`;
                            fileGroup.appendChild(fileHeader);

                            // Add matches for this file
                            if (result.lines) {
                                result.lines.forEach(match => {
                                    const matchDiv = document.createElement('div');
                                    matchDiv.className = 'result-item';

                                    // Create line number span
                                    const lineNumberSpan = document.createElement('span');
                                    lineNumberSpan.className = 'line-number';
                                    lineNumberSpan.textContent = match.line.line_number.toString();

                                    // Create line content span with highlighted matches
                                    const lineContentSpan = document.createElement('span');
                                    lineContentSpan.className = 'line-content';

                                    let content = match.line.content;
                                    let highlightedContent = '';

                                    // Handle match as number[] (start and end positions)
                                    const matchPositions = match.line.match || [];
                                    if (matchPositions.length >= 2) {
                                        const start = matchPositions[0];
                                        const end = matchPositions[1];

                                        // Truncate text before match if it's too long
                                        const beforeMatch = content.substring(0, start);
                                        const truncatedBefore = beforeMatch.length > 24
                                            ? '...' + beforeMatch.substring(beforeMatch.length - 24)
                                            : beforeMatch;

                                        // Add truncated text before match
                                        highlightedContent += truncatedBefore;

                                        // Add highlighted match
                                        highlightedContent += \`<span class="match-highlight">\${content.substring(start, end)}</span>\`;

                                        // Only show a bit of text after the match
                                        const afterMatch = content.substring(end);
                                        const truncatedAfter = afterMatch.length > 128
                                            ? afterMatch.substring(0, 128) + '...'
                                            : afterMatch;
                                        highlightedContent += truncatedAfter;

                                        // Store match information directly on the element as a data attribute
                                        matchDiv.dataset.start = start;
                                        matchDiv.dataset.end = end;
                                    } else {
                                        // If no match positions, show truncated line
                                        highlightedContent = content.length > 160
                                            ? content.substring(0, 160) + '...'
                                            : content;
                                    }

                                    lineContentSpan.innerHTML = highlightedContent;

                                    matchDiv.appendChild(lineNumberSpan);
                                    matchDiv.appendChild(lineContentSpan);

                                    matchDiv.addEventListener('click', () => {
                                        vscode.postMessage({
                                            type: 'openFile',
                                            file: result.file,
                                            line: match.line.line_number,
                                            start: matchDiv.dataset.start ? parseInt(matchDiv.dataset.start) : undefined,
                                            end: matchDiv.dataset.end ? parseInt(matchDiv.dataset.end) : undefined
                                        });
                                    });
                                    fileGroup.appendChild(matchDiv);
                                });
                            }

                            container.appendChild(fileGroup);
                        });
                    }

                    window.addEventListener('message', event => {
                        const message = event.data;
                        if (message.type === 'searchResults') {
                            displayResults(message.results);
                        }
                    });

                    document.getElementById('searchInput').addEventListener('keyup', (event) => {
                        if (event.key === 'Enter') {
                            performSearch();
                        }
                    });

                    // Add event listeners for Enter key in include/exclude fields
                    document.getElementById('includeFiles').addEventListener('keyup', (event) => {
                        if (event.key === 'Enter') {
                            performSearch();
                        }
                    });

                    document.getElementById('excludeFiles').addEventListener('keyup', (event) => {
                        if (event.key === 'Enter') {
                            performSearch();
                        }
                    });

                    // Case sensitive button toggle
                    document.getElementById('caseSensitiveBtn').addEventListener('click', () => {
                        document.getElementById('caseSensitiveBtn').classList.toggle('active');
                        // Re-run search if there's already a query
                        if (document.getElementById('searchInput').value.trim()) {
                            performSearch();
                        }
                    });

                    // Toggle search options visibility
                    document.getElementById('optionsToggle').addEventListener('click', () => {
                        const options = document.getElementById('searchOptions');
                        const toggle = document.getElementById('optionsToggle');
                        options.classList.toggle('visible');
                        toggle.classList.toggle('active');
                    });
                </script>
            </body>
            </html>
        `;
    }
}
