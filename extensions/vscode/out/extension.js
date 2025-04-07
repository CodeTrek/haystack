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
exports.activate = activate;
exports.deactivate = deactivate;
const vscode = __importStar(require("vscode"));
const haystackProvider_1 = require("./search/haystackProvider");
const path_1 = __importDefault(require("path"));
function activate(context) {
    const haystackProvider = new haystackProvider_1.HaystackProvider();
    // Register the command
    let disposableCommand = vscode.commands.registerCommand('haystack.search', () => __awaiter(this, void 0, void 0, function* () {
        const searchTerm = yield vscode.window.showInputBox({
            placeHolder: 'Enter search term',
            prompt: 'Search in files'
        });
        if (searchTerm) {
            const results = yield haystackProvider.provideTextSearchResults(searchTerm, { includePattern: { '**/*': true }, excludePattern: { '**/node_modules/**': true } }, new vscode.CancellationTokenSource().token);
            if (results.length > 0) {
                const quickPickItems = results.map(result => ({
                    label: path_1.default.basename(result.uri.fsPath),
                    description: result.preview.text,
                    detail: result.uri.fsPath,
                    result: result
                }));
                const selected = yield vscode.window.showQuickPick(quickPickItems, {
                    placeHolder: 'Select a result to open'
                });
                if (selected) {
                    const document = yield vscode.workspace.openTextDocument(selected.result.uri);
                    const editor = yield vscode.window.showTextDocument(document);
                    editor.selection = new vscode.Selection(selected.result.range.start, selected.result.range.end);
                }
            }
            else {
                vscode.window.showInformationMessage('No results found');
            }
        }
    }));
    context.subscriptions.push(disposableCommand);
}
function deactivate() { }
