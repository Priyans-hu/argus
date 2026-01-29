import * as vscode from 'vscode';

let outputChannel: vscode.OutputChannel | undefined;

export function getOutputChannel(): vscode.OutputChannel {
    if (!outputChannel) {
        outputChannel = vscode.window.createOutputChannel('Argus');
    }
    return outputChannel;
}

export function log(message: string): void {
    getOutputChannel().appendLine(`[${new Date().toLocaleTimeString()}] ${message}`);
}

export function show(): void {
    getOutputChannel().show(true);
}

export function dispose(): void {
    outputChannel?.dispose();
    outputChannel = undefined;
}
