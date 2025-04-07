import * as assert from 'assert';
import * as vscode from 'vscode';
import { HaystackProvider } from '../../search/haystackProvider';

export function runTests() {
    suite('Haystack Search Tests', () => {
        let haystackProvider: HaystackProvider;

        setup(() => {
            haystackProvider = new HaystackProvider();
        });

        test('should return search results for a valid query', async () => {
            const query = 'test';
            const token = new vscode.CancellationTokenSource().token;
            const results = await haystackProvider.provideTextSearchResults(query, {
                includeDeclaration: true,
                maxResults: 10
            }, token);

            assert.ok(results.length > 0, 'Expected search results to be returned');
        });

        test('should return no results for an invalid query', async () => {
            const query = 'nonexistentquery';
            const token = new vscode.CancellationTokenSource().token;
            const results = await haystackProvider.provideTextSearchResults(query, {
                includeDeclaration: true,
                maxResults: 10
            }, token);

            assert.strictEqual(results.length, 0, 'Expected no search results to be returned');
        });

        // Additional tests can be added here
    });
}
