import * as vscode from 'vscode';

export function getSearchTemplate(webview: any, extensionUri: vscode.Uri) {
    // Get paths to resource files
    const scriptUri = webview.asWebviewUri(vscode.Uri.joinPath(extensionUri, 'resources', 'search.js'));
    const styleUri = webview.asWebviewUri(vscode.Uri.joinPath(extensionUri, 'resources', 'search.css'));

    // Set CSP (Content Security Policy)
    const csp = `
        default-src 'none';
        style-src ${webview.cspSource};
        script-src ${webview.cspSource};
        font-src ${webview.cspSource};
    `;

    return `
        <!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <meta http-equiv="Content-Security-Policy" content="${csp}">
            <title>Haystack Search</title>
            <link href="${styleUri}" rel="stylesheet" type="text/css">
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
            <script src="${scriptUri}" type="text/javascript"></script>
        </body>
        </html>
    `;
}

function getNonce() {
    let text = '';
    const possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    for (let i = 0; i < 32; i++) {
        text += possible.charAt(Math.floor(Math.random() * possible.length));
    }
    return text;
}
