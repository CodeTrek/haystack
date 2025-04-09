import * as vscode from 'vscode';
import { HaystackProvider } from './haystackProvider';
import { getSearchTemplate } from './components/searchTemplate';
import { SearchHandlers } from './components/searchHandlers';
import { SearchContentResult, SearchMessage } from '../types/search';

export class SearchViewProvider implements vscode.WebviewViewProvider {
    public static readonly viewType = 'haystackSearch';

    // Store the webview view reference
    private _view?: vscode.WebviewView;
    private readonly _searchHandlers: SearchHandlers;
    private _statusBarItem: vscode.StatusBarItem;
    private _statusUpdateInterval: NodeJS.Timeout | null = null;

    constructor(
        private readonly _extensionUri: vscode.Uri,
        private readonly _haystackProvider: HaystackProvider
    ) {
        this._searchHandlers = new SearchHandlers(_haystackProvider);
        this._statusBarItem = vscode.window.createStatusBarItem(
            vscode.StatusBarAlignment.Left,
            100
        );
        // Start status updates immediately when the provider is created
        this.startStatusUpdates();
    }

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
            webviewView.webview.html = getSearchTemplate(webviewView.webview);
        }

        // Handle messages from the webview
        webviewView.webview.onDidReceiveMessage(async (data: SearchMessage) => {
            switch (data.type) {
                case 'search':
                    await this._searchHandlers.handleSearch(webviewView.webview, data.query!, data.options!);
                    break;
                case 'openFile':
                    await this._searchHandlers.handleOpenFile(data.file!, data.line!, data.start, data.end);
                    break;
            }
        });

        // Clean up when the view is disposed
        webviewView.onDidDispose(() => {
            this.stopStatusUpdates();
        });
    }

    private startStatusUpdates() {
        // Clear existing interval if any
        if (this._statusUpdateInterval) {
            clearInterval(this._statusUpdateInterval);
        }

        // Set up new interval (3 seconds)
        this._statusUpdateInterval = setInterval(async () => {
            try {
                const status = await this._haystackProvider.getWorkspaceStatus();
                this.updateStatusBar(status);
            } catch (error) {
                console.error(`Failed to update workspace status: ${error}`);
            }
        }, 3000);
    }

    private updateStatusBar(status: any) {
        if (status.error) {
            this._statusBarItem.text = `$(error) Haystack: (Error)`;
            this._statusBarItem.tooltip = `Error: ${status.error}`;
            this._statusBarItem.show();
        } else if (status.indexing) {
            this._statusBarItem.text = `$(sync~spin) Haystack (indexing)`;
            this._statusBarItem.tooltip = `Indexing workspace:\n• Indexed: ${status.indexedFiles} files\n• Total: ${status.totalFiles} files\nYou can search now, but the results may not be accurate`;
            this._statusBarItem.show();
        } else {
            this._statusBarItem.text = `$(check) Haystack (Ready)`;
            this._statusBarItem.tooltip = `Haystack search is ready\n• Total indexed files: ${status.totalFiles}\n• Status: Ready for search`;
            this._statusBarItem.show();
        }
    }

    private stopStatusUpdates() {
        if (this._statusUpdateInterval) {
            clearInterval(this._statusUpdateInterval);
            this._statusUpdateInterval = null;
        }
        this._statusBarItem.hide();
    }

    dispose() {
        this.stopStatusUpdates();
        this._statusBarItem.dispose();
    }
}
