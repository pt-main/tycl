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
Object.defineProperty(exports, "__esModule", { value: true });
exports.activate = activate;
exports.deactivate = deactivate;
const vscode = __importStar(require("vscode"));
const TYPES = [
    'int', 'float', 'bool', 'string', 'object',
    'ints', 'floats', 'bools', 'strings', 'objects', 'null'
];
const CONTRACTS = ['strict', 'flexible', 'dynamic'];
const ACTIONS = ['file', 'env', 'join', 'asString', 'asObject'];
const VALUES = ['true', 'false', 'null'];
function isInComment(document, position) {
    const text = document.getText(new vscode.Range(new vscode.Position(0, 0), position));
    let inComment = false;
    let i = 0;
    while (i < text.length - 1) {
        if (!inComment && text[i] === '/' && text[i + 1] === '*') {
            inComment = true;
            i += 2;
        }
        else if (inComment && text[i] === '*' && text[i + 1] === '/') {
            inComment = false;
            i += 2;
        }
        else {
            i++;
        }
    }
    if (!inComment) {
        i = 0;
        while (i < text.length - 1) {
            if (!inComment && text[i] === '/' && text[i + 1] === '/') {
                inComment = true;
                i += 2;
            }
            else if (inComment && text[i] === '\n') {
                inComment = false;
                i++;
            }
            else {
                i++;
            }
        }
    }
    return inComment;
}
function activate(context) {
    const provider = vscode.languages.registerCompletionItemProvider('tycl', {
        provideCompletionItems(document, position) {
            if (isInComment(document, position)) {
                return [];
            }
            const line = document.lineAt(position).text;
            const textBeforeCursor = line.substring(0, position.character);
            const items = [];
            const trimmed = textBeforeCursor.trimEnd();
            if (trimmed.endsWith(':')) {
                for (const t of TYPES) {
                    const item = new vscode.CompletionItem(t, vscode.CompletionItemKind.TypeParameter);
                    item.detail = 'TYCL type';
                    items.push(item);
                }
                return items;
            }
            if (trimmed.endsWith('=')) {
                for (const v of VALUES) {
                    const item = new vscode.CompletionItem(v, vscode.CompletionItemKind.Value);
                    item.detail = 'TYCL value';
                    items.push(item);
                }
                for (const a of ACTIONS) {
                    const item = new vscode.CompletionItem(`${a}()`, vscode.CompletionItemKind.Function);
                    item.detail = 'TYCL action';
                    item.insertText = new vscode.SnippetString(`${a}($0)`);
                    items.push(item);
                }
                for (const c of CONTRACTS) {
                    const item = new vscode.CompletionItem(c, vscode.CompletionItemKind.Keyword);
                    item.detail = 'TYCL contract';
                    items.push(item);
                }
                return items;
            }
            const lastOpenBrace = trimmed.lastIndexOf('{');
            if (lastOpenBrace !== -1) {
                const afterBrace = trimmed.substring(lastOpenBrace + 1).trim();
                if (afterBrace === '' || position.character <= lastOpenBrace + 2) {
                    for (const c of CONTRACTS) {
                        const item = new vscode.CompletionItem(c, vscode.CompletionItemKind.Keyword);
                        item.detail = 'TYCL contract';
                        items.push(item);
                    }
                    return items;
                }
            }
            for (const t of TYPES) {
                const item = new vscode.CompletionItem(t, vscode.CompletionItemKind.TypeParameter);
                item.detail = 'TYCL type';
                items.push(item);
            }
            for (const c of CONTRACTS) {
                const item = new vscode.CompletionItem(c, vscode.CompletionItemKind.Keyword);
                item.detail = 'TYCL contract';
                items.push(item);
            }
            for (const a of ACTIONS) {
                const item = new vscode.CompletionItem(`${a}()`, vscode.CompletionItemKind.Function);
                item.detail = 'TYCL action';
                item.insertText = new vscode.SnippetString(`${a}($0)`);
                items.push(item);
            }
            return items;
        }
    }, ':', '=', ' ', '\n');
    context.subscriptions.push(provider);
}
function deactivate() { }
//# sourceMappingURL=extension.js.map