import * as vscode from 'vscode';
import { runArgus, spawnArgus } from './binary';
import { setState } from './statusBar';
import { log, show } from './output';

let watchProcess: { kill: () => void } | undefined;

function getWorkspacePath(uri?: vscode.Uri): string | undefined {
    if (uri) {
        return uri.fsPath;
    }
    const folders = vscode.workspace.workspaceFolders;
    if (folders && folders.length > 0) {
        return folders[0].uri.fsPath;
    }
    return undefined;
}

function buildScanArgs(): string[] {
    const config = vscode.workspace.getConfiguration('argus');
    const args = ['scan'];

    const format = config.get<string>('defaultFormat', 'claude');
    args.push('--format', format);

    if (config.get<boolean>('parallelMode', true)) {
        args.push('--parallel');
    }

    if (config.get<boolean>('compactMode', false)) {
        args.push('--compact');
    }

    if (config.get<boolean>('mergeMode', true)) {
        args.push('--merge');
    }

    if (config.get<boolean>('enableAI', false)) {
        args.push('--ai');
    }

    return args;
}

export async function scanCommand(uri?: vscode.Uri): Promise<void> {
    const cwd = getWorkspacePath(uri);
    if (!cwd) {
        vscode.window.showErrorMessage('No workspace folder open.');
        return;
    }

    setState('scanning');
    show();
    log(`Scanning ${cwd}...`);

    try {
        const args = buildScanArgs();
        args.push(cwd);

        const result = await runArgus(
            args,
            cwd,
            (data) => log(data.trimEnd()),
            (data) => log(data.trimEnd()),
        );

        if (result.code === 0) {
            setState('done');
            log('Scan complete.');
            vscode.window.showInformationMessage('Argus: Scan complete!');
        } else {
            setState('error');
            log(`Scan failed with exit code ${result.code}`);
            if (result.stderr) {
                log(result.stderr);
            }
            vscode.window.showErrorMessage(`Argus scan failed. Check output for details.`);
        }
    } catch (err) {
        setState('error');
        const message = err instanceof Error ? err.message : String(err);
        log(`Error: ${message}`);
        vscode.window.showErrorMessage(`Argus: ${message}`);
    }
}

export async function watchCommand(): Promise<void> {
    if (watchProcess) {
        vscode.window.showWarningMessage('Argus is already watching. Stop it first.');
        return;
    }

    const cwd = getWorkspacePath();
    if (!cwd) {
        vscode.window.showErrorMessage('No workspace folder open.');
        return;
    }

    setState('watching');
    show();
    log(`Starting watch mode for ${cwd}...`);

    watchProcess = spawnArgus(
        ['watch', cwd],
        cwd,
        (data) => log(data.trimEnd()),
        (code) => {
            watchProcess = undefined;
            if (code === 0 || code === null) {
                setState('idle');
                log('Watch mode stopped.');
            } else {
                setState('error');
                log(`Watch process exited with code ${code}`);
            }
        },
    );
}

export function stopWatchCommand(): void {
    if (!watchProcess) {
        vscode.window.showInformationMessage('Argus is not currently watching.');
        return;
    }

    log('Stopping watch mode...');
    watchProcess.kill();
    watchProcess = undefined;
    setState('idle');
}

export async function initCommand(): Promise<void> {
    const cwd = getWorkspacePath();
    if (!cwd) {
        vscode.window.showErrorMessage('No workspace folder open.');
        return;
    }

    log(`Initializing argus config in ${cwd}...`);

    try {
        const result = await runArgus(['init', cwd], cwd);

        if (result.code === 0) {
            log('Config initialized.');
            vscode.window.showInformationMessage('Argus: Config initialized!');

            // Open the config file
            const configUri = vscode.Uri.joinPath(vscode.Uri.file(cwd), '.argus.yaml');
            const doc = await vscode.workspace.openTextDocument(configUri);
            await vscode.window.showTextDocument(doc);
        } else {
            log(`Init failed: ${result.stderr}`);
            vscode.window.showErrorMessage(`Argus init failed: ${result.stderr}`);
        }
    } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        log(`Error: ${message}`);
        vscode.window.showErrorMessage(`Argus: ${message}`);
    }
}

export async function insightsCommand(): Promise<void> {
    const cwd = getWorkspacePath();
    if (!cwd) {
        vscode.window.showErrorMessage('No workspace folder open.');
        return;
    }

    log(`Fetching usage insights for ${cwd}...`);
    show();

    try {
        const result = await runArgus(
            ['insights', '--format', 'text', cwd],
            cwd,
        );

        if (result.code === 0 && result.stdout) {
            log(result.stdout);
        } else if (result.stderr) {
            log(result.stderr);
        } else {
            log('No usage insights available.');
        }
    } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        log(`Error: ${message}`);
        vscode.window.showErrorMessage(`Argus: ${message}`);
    }
}

export function disposeWatch(): void {
    if (watchProcess) {
        watchProcess.kill();
        watchProcess = undefined;
    }
}
