import * as vscode from 'vscode';
import { SearchViewProvider } from './search/SearchViewProvider';
import { HaystackProvider } from './search/haystackProvider';

let haystackProvider: HaystackProvider | undefined;

export async function activate(context: vscode.ExtensionContext) {
    haystackProvider = new HaystackProvider();
    const searchViewProvider = new SearchViewProvider(context.extensionUri, haystackProvider);

    // Create workspace when extension is activated
    try {
        await haystackProvider.createWorkspace();
    } catch (error) {
//        vscode.window.showErrorMessage(`Failed to create workspace: ${error}`);
    }

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

    // Listen for workspace folder changes
    context.subscriptions.push(
        vscode.workspace.onDidChangeWorkspaceFolders(async (event) => {
            if (event.added.length > 0) {
                try {
                    await haystackProvider?.createWorkspace();
                } catch (error) {
//                    vscode.window.showErrorMessage(`Failed to create workspace: ${error}`);
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
        }
    });
}

export function deactivate() {
    if (haystackProvider) {
        haystackProvider.dispose();
        haystackProvider = undefined;
    }
}
