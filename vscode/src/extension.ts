import * as vscode from 'vscode';

const TYPES = [
    'int', 'float', 'bool', 'string', 'object',
    'ints', 'floats', 'bools', 'strings', 'objects', 'null'
];

const CONTRACTS = ['strict', 'flexible', 'dynamic'];

const ACTIONS = ['file', 'env', 'join', 'asString', 'asObject'];

const VALUES = ['true', 'false', 'null'];

export function activate(context: vscode.ExtensionContext) {
    const provider = vscode.languages.registerCompletionItemProvider(
        'tycl',
        {
            provideCompletionItems(document, position) {
                const line = document.lineAt(position).text;
                const textBeforeCursor = line.substring(0, position.character);
                const items: vscode.CompletionItem[] = [];
                
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
        },
        ':', '=', ' ', '\n' 
    );
    context.subscriptions.push(provider);
}

export function deactivate() {}