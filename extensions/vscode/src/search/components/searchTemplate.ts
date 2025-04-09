export function getSearchTemplate(webview: any) {
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
                    display: flex;
                    flex-direction: column;
                    height: 100vh;
                }
                .search-container {
                    padding: 10px;
                    flex: 1;
                    display: flex;
                    flex-direction: column;
                    overflow: hidden;
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
                    flex: 1;
                    overflow-y: auto;
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

                function displayResults(message) {
                    const container = document.getElementById('searchResults');
                    if (!container) return;

                    container.innerHTML = '';

                    const results = message.results;

                    if (!results || results.length === 0) {
                        const emptyMessage = document.createElement('div');
                        emptyMessage.className = 'search-summary';
                        emptyMessage.textContent = 'No results found';
                        container.appendChild(emptyMessage);
                        return;
                    }

                    const totalMatches = results.reduce((sum, result) => sum + (result.lines?.length || 0), 0);
                    const summary = document.createElement('div');
                    summary.className = 'search-summary';

                    // Check if results are truncated based on server response
                    if (message.truncated) {
                        summary.textContent = \`\${totalMatches} results in \${results.length} files (Results truncated, showing partial matches)\`;
                    } else {
                        summary.textContent = \`\${totalMatches} results in \${results.length} files\`;
                    }
                    container.appendChild(summary);

                    results.forEach(result => {
                        const fileGroup = document.createElement('div');
                        fileGroup.className = 'file-group';

                        const fileHeader = document.createElement('div');
                        fileHeader.className = 'file-header';
                        fileHeader.innerHTML = \`
                            <span class="file-path">\${result.file}\${result.truncate ? ' (truncated)' : ''}</span>
                            <span class="match-count">\${result.lines?.length || 0}</span>
                        \`;
                        fileGroup.appendChild(fileHeader);

                        if (result.lines) {
                            result.lines.forEach(match => {
                                const matchDiv = document.createElement('div');
                                matchDiv.className = 'result-item';

                                const lineNumberSpan = document.createElement('span');
                                lineNumberSpan.className = 'line-number';
                                lineNumberSpan.textContent = match.line.line_number.toString();

                                const lineContentSpan = document.createElement('span');
                                lineContentSpan.className = 'line-content';

                                let content = match.line.content;
                                let highlightedContent = '';

                                const matchPositions = match.line.match || [];
                                if (matchPositions.length >= 2) {
                                    const start = matchPositions[0];
                                    const end = matchPositions[1];

                                    const beforeMatch = content.substring(0, start);
                                    const truncatedBefore = beforeMatch.length > 24
                                        ? '...' + beforeMatch.substring(beforeMatch.length - 24)
                                        : beforeMatch;

                                    highlightedContent += truncatedBefore;
                                    highlightedContent += \`<span class="match-highlight">\${content.substring(start, end)}</span>\`;

                                    const afterMatch = content.substring(end);
                                    const truncatedAfter = afterMatch.length > 128
                                        ? afterMatch.substring(0, 128) + '...'
                                        : afterMatch;
                                    highlightedContent += truncatedAfter;

                                    matchDiv.dataset.start = start;
                                    matchDiv.dataset.end = end;
                                } else {
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
                        displayResults(message);
                    }
                });

                document.getElementById('searchInput').addEventListener('keyup', (event) => {
                    if (event.key === 'Enter') {
                        performSearch();
                    }
                });

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

                document.getElementById('caseSensitiveBtn').addEventListener('click', () => {
                    document.getElementById('caseSensitiveBtn').classList.toggle('active');
                    if (document.getElementById('searchInput').value.trim()) {
                        performSearch();
                    }
                });

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
