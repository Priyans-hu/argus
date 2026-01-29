import * as vscode from 'vscode';

type StatusState = 'idle' | 'scanning' | 'watching' | 'done' | 'error';

let statusBarItem: vscode.StatusBarItem | undefined;

const STATUS_CONFIG: Record<StatusState, { icon: string; text: string; tooltip: string }> = {
    idle: { icon: '$(eye)', text: 'Argus', tooltip: 'Argus: Click to scan' },
    scanning: { icon: '$(sync~spin)', text: 'Argus: Scanning...', tooltip: 'Argus: Scanning codebase' },
    watching: { icon: '$(eye)', text: 'Argus: Watching', tooltip: 'Argus: Watching for changes (click to stop)' },
    done: { icon: '$(check)', text: 'Argus: Done', tooltip: 'Argus: Scan complete' },
    error: { icon: '$(error)', text: 'Argus: Error', tooltip: 'Argus: An error occurred' },
};

export function createStatusBar(): vscode.StatusBarItem {
    statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 100);
    statusBarItem.command = 'argus.scan';
    setState('idle');
    statusBarItem.show();
    return statusBarItem;
}

export function setState(state: StatusState): void {
    if (!statusBarItem) {
        return;
    }
    const config = STATUS_CONFIG[state];
    statusBarItem.text = `${config.icon} ${config.text}`;
    statusBarItem.tooltip = config.tooltip;

    if (state === 'watching') {
        statusBarItem.command = 'argus.stopWatch';
    } else {
        statusBarItem.command = 'argus.scan';
    }

    // Auto-reset from done/error to idle after 5s
    if (state === 'done' || state === 'error') {
        setTimeout(() => setState('idle'), 5000);
    }
}

export function dispose(): void {
    statusBarItem?.dispose();
    statusBarItem = undefined;
}
