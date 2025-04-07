import * as vscode from 'vscode';
import { HaystackProvider } from './search/haystackProvider';
import path from 'path';

export function activate(context: vscode.ExtensionContext) {
    const haystackProvider = new HaystackProvider();

    // Register the command
    let disposableCommand = vscode.commands.registerCommand('haystack.search', async () => {
        const searchTerm = await vscode.window.showInputBox({
            placeHolder: 'Enter search term',
            prompt: 'Search in files'
        });

        if (searchTerm) {
            const results = await haystackProvider.provideTextSearchResults(
                searchTerm,
                { includePattern: { '**/*': true }, excludePattern: { '**/node_modules/**': true } },
                new vscode.CancellationTokenSource().token
            );

            if (results.length > 0) {
                const quickPickItems = results.map(result => ({
                    label: path.basename(result.uri.fsPath),
                    description: result.preview.text,
                    detail: result.uri.fsPath,
                    result: result
                }));

                const selected = await vscode.window.showQuickPick(quickPickItems, {
                    placeHolder: 'Select a result to open'
                });

                if (selected) {
                    const document = await vscode.workspace.openTextDocument(selected.result.uri);
                    const editor = await vscode.window.showTextDocument(document);
                    editor.selection = new vscode.Selection(selected.result.range.start, selected.result.range.end);
                }
            } else {
                vscode.window.showInformationMessage('No results found');
            }
        }
    });
    context.subscriptions.push(disposableCommand);
}

export function deactivate() {}
