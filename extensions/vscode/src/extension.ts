import * as vscode from 'vscode';
import { SearchViewProvider } from './search/SearchViewProvider';
import { HaystackProvider } from './search/haystackProvider';

// Constants
const isDev = () => process.env.VSCODE_DEBUG_MODE === 'true' || process.env.IS_DEV === 'true' || __dirname.includes('.vscode');
const getStateKey = (key: string) => `${isDev() ? 'dev.' : ''}${key}`;

// Global state
let haystackProvider: HaystackProvider | undefined;
let searchViewProvider: SearchViewProvider | undefined;

/**
 * This function is called when your extension is activated
 */
export async function activate(context: vscode.ExtensionContext) {
    console.log(`haystack is running in ${isDev() ? 'dev' : 'prod'} mode`);

    // Simple activation logging without version checks
    console.log('Haystack extension activated');

    haystackProvider = new HaystackProvider();
    searchViewProvider = new SearchViewProvider(context.extensionUri, haystackProvider);

    // Register search view
    context.subscriptions.push(
        vscode.window.registerWebviewViewProvider(
            SearchViewProvider.viewType,
            searchViewProvider,
            {
                webviewOptions: {
                    retainContextWhenHidden: true
                }
            }
        )
    );

    // Register command to search selected text
    context.subscriptions.push(
        vscode.commands.registerCommand('haystack.searchSelectedText', async () => {
            // Get active editor
            const editor = vscode.window.activeTextEditor;
            if (!editor) {
                vscode.window.showInformationMessage('No active editor found');
                return;
            }

            // Get selected text
            const selection = editor.selection;
            if (selection.isEmpty) {
                vscode.window.showInformationMessage('No text selected');
                return;
            }

            const selectedText = editor.document.getText(selection);
            if (!selectedText) {
                vscode.window.showInformationMessage('Selected text is empty');
                return;
            }

            // Show search view if it's not visible
            await vscode.commands.executeCommand('haystackSearch.focus');

            // Perform search with selected text
            if (searchViewProvider) {
                searchViewProvider.searchText(selectedText);
            }
        })
    );

    // Register command to sync the workspace index
    context.subscriptions.push(
        vscode.commands.registerCommand('haystack.syncWorkspace', async () => {
            try {
                vscode.window.withProgress({
                    location: vscode.ProgressLocation.Notification,
                    title: "Refreshing Haystack index...",
                    cancellable: false
                }, async (progress) => {
                    progress.report({ increment: 0 });

                    try {
                        // Call the sync API
                        await haystackProvider?.syncWorkspace();

                        progress.report({ increment: 100 });
                        vscode.window.showInformationMessage('Haystack index refreshed successfully.');
                    } catch (error) {
                        vscode.window.showErrorMessage(`Failed to refresh Haystack index: ${error}`);
                    }
                });
            } catch (error) {
                vscode.window.showErrorMessage(`Error during Haystack index refresh: ${error}`);
            }
        })
    );

    // Delay workspace creation
    setTimeout(async () => {
        try {
            await haystackProvider?.createWorkspace();
        } catch (error) {
            // Silent fail
        }
    }, 1000);

    // Listen for workspace folder changes
    context.subscriptions.push(
        vscode.workspace.onDidChangeWorkspaceFolders(async (event) => {
            if (event.added.length > 0) {
                try {
                    await haystackProvider?.createWorkspace();
                } catch (error) {
                    // Silent fail
                }
            }
        })
    );

    // Add provider to subscriptions for proper cleanup
    context.subscriptions.push({
        dispose: () => {
            if (haystackProvider) {
                haystackProvider.dispose();
                haystackProvider = undefined;
            }
            if (searchViewProvider) {
                searchViewProvider.dispose();
                searchViewProvider = undefined;
            }
        }
    });
}

export function deactivate() {
    if (haystackProvider) {
        haystackProvider.dispose();
        haystackProvider = undefined;
    }
    if (searchViewProvider) {
        searchViewProvider.dispose();
        searchViewProvider = undefined;
    }
}
