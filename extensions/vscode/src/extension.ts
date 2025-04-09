import * as vscode from 'vscode';
import { SearchViewProvider } from './search/SearchViewProvider';
import { HaystackProvider } from './search/haystackProvider';

// Constants
const isDev = () => process.env.VSCODE_DEBUG_MODE === 'true' || process.env.IS_DEV === 'true' || __dirname.includes('.vscode');
const getStateKey = (key: string) => `${isDev() ? 'dev.' : ''}${key}`;

// Global state
let haystackProvider: HaystackProvider | undefined;

/**
 * This function is called when your extension is activated
 */
export async function activate(context: vscode.ExtensionContext) {
    console.log(`haystack is running in ${isDev() ? 'dev' : 'prod'} mode`);

    // Simple activation logging without version checks
    console.log('Haystack extension activated');

    haystackProvider = new HaystackProvider();
    const searchViewProvider = new SearchViewProvider(context.extensionUri, haystackProvider);

    // Create workspace when extension is activated
    try {
        await haystackProvider.createWorkspace();
    } catch (error) {
        // Silent fail
    }

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
        }
    });
}

export function deactivate() {
    if (haystackProvider) {
        haystackProvider.dispose();
        haystackProvider = undefined;
    }
}
