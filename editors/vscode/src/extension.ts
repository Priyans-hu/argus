import * as vscode from 'vscode';
import { findBinary } from './binary';
import { createStatusBar, dispose as disposeStatusBar } from './statusBar';
import { log, dispose as disposeOutput } from './output';
import {
    scanCommand,
    watchCommand,
    stopWatchCommand,
    initCommand,
    insightsCommand,
    disposeWatch,
} from './commands';

export function activate(context: vscode.ExtensionContext): void {
    log('Argus extension activated');

    // Check if binary is available
    const binary = findBinary();
    if (!binary) {
        log('Warning: argus binary not found. Install with: go install github.com/Priyans-hu/argus/cmd/argus@latest');
    }

    // Create status bar
    const statusBar = createStatusBar();
    context.subscriptions.push(statusBar);

    // Register commands
    context.subscriptions.push(
        vscode.commands.registerCommand('argus.scan', scanCommand),
        vscode.commands.registerCommand('argus.watch', watchCommand),
        vscode.commands.registerCommand('argus.stopWatch', stopWatchCommand),
        vscode.commands.registerCommand('argus.init', initCommand),
        vscode.commands.registerCommand('argus.insights', insightsCommand),
    );

    // Auto-scan on open if configured
    const config = vscode.workspace.getConfiguration('argus');
    if (config.get<boolean>('autoScanOnOpen', false)) {
        log('Auto-scan enabled, running scan...');
        vscode.commands.executeCommand('argus.scan');
    }
}

export function deactivate(): void {
    disposeWatch();
    disposeStatusBar();
    disposeOutput();
}
