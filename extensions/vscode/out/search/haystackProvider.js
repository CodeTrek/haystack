"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.HaystackProvider = void 0;
const vscode = __importStar(require("vscode"));
const axios_1 = __importDefault(require("axios"));
const HAYSTACK_URL = 'http://127.0.0.1:13134/api/v1/search/content';
class HaystackProvider {
    constructor() {
        const workspaceFolders = vscode.workspace.workspaceFolders;
        this.workspaceRoot = workspaceFolders ? workspaceFolders[0].uri.fsPath : '';
    }
    provideTextSearchResults(query, options, token) {
        return __awaiter(this, void 0, void 0, function* () {
            var _a;
            const searchRequest = {
                workspace: this.workspaceRoot,
                query: query,
                case_sensitive: false,
                filters: {
                    include: options.includePattern ? Object.keys(options.includePattern)[0] : undefined,
                    exclude: options.excludePattern ? Object.keys(options.excludePattern)[0] : undefined
                },
                limit: {
                    max_results: options.maxResults
                }
            };
            try {
                const response = yield axios_1.default.post(HAYSTACK_URL, searchRequest);
                if (response.data.code === 0 && ((_a = response.data.data) === null || _a === void 0 ? void 0 : _a.results)) {
                    return this.convertResults(response.data.data.results);
                }
                else {
                    vscode.window.showErrorMessage(`Search failed: ${response.data.message}`);
                    return [];
                }
            }
            catch (error) {
                vscode.window.showErrorMessage(`Failed to connect to Haystack server: ${error}`);
                return [];
            }
        });
    }
    convertResults(results) {
        const vscodeResults = [];
        for (const result of results) {
            if (!result.lines)
                continue;
            for (const lineMatch of result.lines) {
                const line = lineMatch.line;
                const range = new vscode.Range(new vscode.Position(line.line_number - 1, 0), new vscode.Position(line.line_number - 1, line.content.length));
                vscodeResults.push({
                    uri: vscode.Uri.file(result.file),
                    range: range,
                    preview: {
                        text: line.content,
                        matches: [range]
                    }
                });
            }
        }
        return vscodeResults;
    }
}
exports.HaystackProvider = HaystackProvider;
