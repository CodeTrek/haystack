import * as vscode from 'vscode';
import axios from 'axios';
import {
    SearchContentRequest,
    SearchContentResponse,
    SearchContentResult
} from '../types/search';

// Get the appropriate host based on the environment
function getHaystackHost(): string {
    // If we're in a remote environment, use the remote host
    if (vscode.env.remoteName) {
        // In remote environment, we can safely use localhost
        return 'localhost';
    }

    // In local environment, prefer 127.0.0.1 for better compatibility
    return '127.0.0.1';
}

const HAYSTACK_PORT = '13134';
const HAYSTACK_URL = `http://${getHaystackHost()}:${HAYSTACK_PORT}/api/v1`;
const WORKSPACE_CREATE_URL = `${HAYSTACK_URL}/workspace/create`;
const WORKSPACE_GET_URL = `${HAYSTACK_URL}/workspace/get`;
const DOCUMENT_UPDATE_URL = `${HAYSTACK_URL}/document/update`;
const DOCUMENT_DELETE_URL = `${HAYSTACK_URL}/document/delete`;

export class HaystackProvider {
    private workspaceRoot: string;
    private updateTimeouts: Map<string, NodeJS.Timeout>;
    private periodicTaskInterval: NodeJS.Timeout | null;
    private statusUpdateInterval: NodeJS.Timeout | null;

    constructor() {
        const workspaceFolders = vscode.workspace.workspaceFolders;
        this.workspaceRoot = workspaceFolders ? workspaceFolders[0].uri.fsPath : '';
        this.updateTimeouts = new Map();
        this.periodicTaskInterval = null;
        this.statusUpdateInterval = null;

        // Start periodic workspace creation task
        this.startPeriodicTask();

        // Listen for file save events
        vscode.workspace.onDidSaveTextDocument(async (document) => {
            if (document.uri.scheme === 'file') {
                try {
                    // Convert absolute path to relative path
                    const relativePath = vscode.workspace.asRelativePath(document.uri);

                    // Clear existing timeout if any
                    const existingTimeout = this.updateTimeouts.get(relativePath);
                    if (existingTimeout) {
                        clearTimeout(existingTimeout);
                    }

                    // Set new timeout
                    const timeout = setTimeout(async () => {
                        await this.updateDocument(relativePath);
                        this.updateTimeouts.delete(relativePath);
                    }, 500);

                    this.updateTimeouts.set(relativePath, timeout);
                } catch (error) {
                    console.error(`Failed to update document: ${error}`);
                }
            }
        });

        // Listen for file delete events
        vscode.workspace.onDidDeleteFiles(async (event) => {
            for (const uri of event.files) {
                if (uri.scheme === 'file') {
                    try {
                        const relativePath = vscode.workspace.asRelativePath(uri);
                        await this.deleteDocument(relativePath);
                    } catch (error) {
                        console.error(`Failed to delete document: ${error}`);
                    }
                }
            }
        });

        // Listen for file restore events
        vscode.workspace.onDidCreateFiles(async (event) => {
            for (const uri of event.files) {
                if (uri.scheme === 'file') {
                    try {
                        const relativePath = vscode.workspace.asRelativePath(uri);
                        await this.updateDocument(relativePath);
                    } catch (error) {
                        console.error(`Failed to update restored document: ${error}`);
                    }
                }
            }
        });
    }

    private startPeriodicTask() {
        // Clear existing interval if any
        if (this.periodicTaskInterval) {
            clearInterval(this.periodicTaskInterval);
        }

        // Set up new interval (24 hours in milliseconds)
        const TWENTY_FOUR_HOURS = 24 * 60 * 60 * 1000;
        this.periodicTaskInterval = setInterval(async () => {
            try {
                await this.createWorkspace();
            } catch (error) {
                console.error(`Failed to create workspace in periodic task: ${error}`);
            }
        }, TWENTY_FOUR_HOURS);
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

    async updateDocument(filePath: string): Promise<void> {
        if (!this.workspaceRoot) {
            throw new Error('No workspace folder is opened');
        }

        try {
            const response = await axios.post(DOCUMENT_UPDATE_URL, {
                workspace: this.workspaceRoot,
                path: filePath
            });

            if (response.data.code !== 0) {
                throw new Error(response.data.message || 'Failed to update document');
            }
        } catch (error) {
            throw new Error(`Failed to update document: ${error}`);
        }
    }

    async deleteDocument(filePath: string): Promise<void> {
        if (!this.workspaceRoot) {
            throw new Error('No workspace folder is opened');
        }

        try {
            const response = await axios.post(DOCUMENT_DELETE_URL, {
                workspace: this.workspaceRoot,
                path: filePath
            });

            if (response.data.code !== 0) {
                throw new Error(response.data.message || 'Failed to delete document');
            }
        } catch (error) {
            throw new Error(`Failed to delete document: ${error}`);
        }
    }

    async search(query: string, options: {
        caseSensitive?: boolean;
        include?: string;
        exclude?: string;
        maxResults?: number;
        maxResultsPerFile?: number;
    }): Promise<{ results: SearchContentResult[]; truncated: boolean }> {
        const searchRequest: SearchContentRequest = {
            workspace: this.workspaceRoot,
            query: query,
            case_sensitive: options.caseSensitive,
            filters: {
                include: options.include,
                exclude: options.exclude
            },
            limit: {
                max_results: options.maxResults,
                max_results_per_file: options.maxResultsPerFile
            }
        };

        try {
            const response = await axios.post<SearchContentResponse>(`${HAYSTACK_URL}/search/content`, searchRequest);
            if (response.data.code === 0) {
                return {
                    results: response.data.data?.results || [],
                    truncated: response.data.data?.truncate || false
                };
            } else {
                console.log(`Search returned no results: ${response.data.message || 'Unknown reason'}`);
                return { results: [], truncated: false };
            }
        } catch (error) {
            throw new Error(`Failed to connect to Haystack server: ${error}`);
        }
    }

    async getWorkspaceStatus(): Promise<{ indexing: boolean; totalFiles: number; indexedFiles: number; error?: string }> {
        if (!this.workspaceRoot) {
            return {
                indexing: false,
                totalFiles: 0,
                indexedFiles: 0,
                error: 'No workspace folder is opened'
            };
        }

        try {
            const response = await axios.post(WORKSPACE_GET_URL, {
                workspace: this.workspaceRoot
            });

            if (response.data.code !== 0) {
                return {
                    indexing: false,
                    totalFiles: 0,
                    indexedFiles: 0,
                    error: response.data.message || 'Failed to get workspace status'
                };
            }

            // Data might be undefined due to omitempty
            if (!response.data.data) {
                return {
                    indexing: false,
                    totalFiles: 0,
                    indexedFiles: 0,
                    error: 'Workspace not found'
                };
            }

            return {
                indexing: response.data.data.indexing,
                totalFiles: response.data.data.total_files,
                indexedFiles: response.data.data.total_files
            };
        } catch (error) {
            return {
                indexing: false,
                totalFiles: 0,
                indexedFiles: 0,
                error: `Failed to get workspace status: ${error}`
            };
        }
    }

    startStatusUpdates(callback: (status: { indexing: boolean; totalFiles: number; indexedFiles: number; error?: string }) => void) {
        // Clear existing interval if any
        if (this.statusUpdateInterval) {
            clearInterval(this.statusUpdateInterval);
        }

        // Set up new interval (3 seconds)
        this.statusUpdateInterval = setInterval(async () => {
            try {
                const status = await this.getWorkspaceStatus();
                callback(status);
            } catch (error) {
                console.error(`Failed to update workspace status: ${error}`);
            }
        }, 3000);
    }

    stopStatusUpdates() {
        if (this.statusUpdateInterval) {
            clearInterval(this.statusUpdateInterval);
            this.statusUpdateInterval = null;
        }
    }

    dispose() {
        // Clear all timeouts
        for (const timeout of this.updateTimeouts.values()) {
            clearTimeout(timeout);
        }
        this.updateTimeouts.clear();

        // Clear periodic task
        if (this.periodicTaskInterval) {
            clearInterval(this.periodicTaskInterval);
            this.periodicTaskInterval = null;
        }

        // Clear status update interval
        if (this.statusUpdateInterval) {
            clearInterval(this.statusUpdateInterval);
            this.statusUpdateInterval = null;
        }
    }
}
