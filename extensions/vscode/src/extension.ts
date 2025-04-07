import * as vscode from 'vscode';
import { SearchViewProvider } from './search/SearchViewProvider';
import { HaystackProvider } from './search/haystackProvider';

export function activate(context: vscode.ExtensionContext) {
    const haystackProvider = new HaystackProvider();
    const searchViewProvider = new SearchViewProvider(context.extensionUri, haystackProvider);

    context.subscriptions.push(
        vscode.window.registerWebviewViewProvider(
            SearchViewProvider.viewType,
            searchViewProvider
        )
    );
}

export function deactivate() {}
