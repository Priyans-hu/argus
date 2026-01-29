import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';
import { execFile } from 'child_process';
import { log } from './output';

const COMMON_PATHS = [
    path.join(process.env.HOME || '', 'go', 'bin', 'argus'),
    '/usr/local/bin/argus',
    '/usr/bin/argus',
];

export function findBinary(): string | undefined {
    // 1. Check user settings
    const configPath = vscode.workspace.getConfiguration('argus').get<string>('binaryPath');
    if (configPath && fs.existsSync(configPath)) {
        log(`Found argus at configured path: ${configPath}`);
        return configPath;
    }

    // 2. Try `which argus` (Unix) or `where argus` (Windows)
    const whichCmd = process.platform === 'win32' ? 'where' : 'which';
    try {
        const result = require('child_process').execFileSync(whichCmd, ['argus'], {
            encoding: 'utf-8',
            timeout: 3000,
        });
        const found = result.trim().split('\n')[0];
        if (found && fs.existsSync(found)) {
            log(`Found argus via PATH: ${found}`);
            return found;
        }
    } catch {
        // Not found in PATH
    }

    // 3. Check common installation paths
    for (const p of COMMON_PATHS) {
        if (fs.existsSync(p)) {
            log(`Found argus at common path: ${p}`);
            return p;
        }
    }

    log('argus binary not found');
    return undefined;
}

export function runArgus(
    args: string[],
    cwd: string,
    onStdout?: (data: string) => void,
    onStderr?: (data: string) => void,
): Promise<{ code: number; stdout: string; stderr: string }> {
    const binary = findBinary();
    if (!binary) {
        return Promise.reject(new Error(
            'argus binary not found. Install it with: go install github.com/Priyans-hu/argus/cmd/argus@latest'
        ));
    }

    return new Promise((resolve, reject) => {
        const proc = execFile(binary, args, { cwd, maxBuffer: 10 * 1024 * 1024 }, (error, stdout, stderr) => {
            if (error && error.killed) {
                reject(new Error('Process was killed'));
                return;
            }
            resolve({
                code: typeof error?.code === 'number' ? error.code : (error ? 1 : 0),
                stdout: stdout || '',
                stderr: stderr || '',
            });
        });

        if (onStdout) {
            proc.stdout?.on('data', (data: Buffer) => onStdout(data.toString()));
        }
        if (onStderr) {
            proc.stderr?.on('data', (data: Buffer) => onStderr(data.toString()));
        }
    });
}

export function spawnArgus(
    args: string[],
    cwd: string,
    onData: (data: string) => void,
    onExit: (code: number | null) => void,
): { kill: () => void } {
    const binary = findBinary();
    if (!binary) {
        onData('Error: argus binary not found\n');
        onExit(1);
        return { kill: () => {} };
    }

    const { spawn } = require('child_process');
    const proc = spawn(binary, args, { cwd });

    proc.stdout?.on('data', (data: Buffer) => onData(data.toString()));
    proc.stderr?.on('data', (data: Buffer) => onData(data.toString()));
    proc.on('exit', (code: number | null) => onExit(code));

    return {
        kill: () => {
            proc.kill('SIGTERM');
        },
    };
}
