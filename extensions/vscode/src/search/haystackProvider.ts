import * as vscode from 'vscode';
import axios from 'axios';
import {
    SearchContentRequest,
    SearchContentResponse,
    SearchContentResult
} from '../types/search';

const HAYSTACK_URL = 'http://127.0.0.1:13134/api/v1';
const WORKSPACE_CREATE_URL = `${HAYSTACK_URL}/workspace/create`;

export class HaystackProvider {
    private workspaceRoot: string;

    constructor() {
        const workspaceFolders = vscode.workspace.workspaceFolders;
        this.workspaceRoot = workspaceFolders ? workspaceFolders[0].uri.fsPath : '';
    }

    async createWorkspace(): Promise<void> {
        if (!this.workspaceRoot) {
            throw new Error('No workspace folder is opened');
        }

        try {
            const response = await axios.post(WORKSPACE_CREATE_URL, {
                workspace: this.workspaceRoot
            });

            if (response.data.code !== 0) {
                throw new Error(response.data.message || 'Failed to create workspace');
            }
        } catch (error) {
            throw new Error(`Failed to create workspace: ${error}`);
        }
    }

    async search(query: string, options: {
        caseSensitive?: boolean;
        include?: string;
        exclude?: string;
        maxResults?: number
    }): Promise<SearchContentResult[]> {
        const searchRequest: SearchContentRequest = {
            workspace: this.workspaceRoot,
            query: query,
            case_sensitive: options.caseSensitive,
            filters: {
                include: options.include,
                exclude: options.exclude
            },
            limit: {
                max_results: options.maxResults
            }
        };

        try {
            const response = await axios.post<SearchContentResponse>(`${HAYSTACK_URL}/search/content`, searchRequest);
            if (response.data.code === 0 && response.data.data?.results) {
                return response.data.data.results;
            } else {
                throw new Error(response.data.message || 'Search failed');
            }
        } catch (error) {
            throw new Error(`Failed to connect to Haystack server: ${error}`);
        }
    }
}
