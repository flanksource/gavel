import { afterEach, describe, expect, it, vi } from 'vitest';
import { copyText } from './clipboard';

function stubClipboard(writeText: unknown) {
  Object.defineProperty(navigator, 'clipboard', {
    configurable: true,
    value: writeText === undefined ? undefined : { writeText },
  });
}

function stubExecCommand(execCommand: unknown) {
  Object.defineProperty(document, 'execCommand', { configurable: true, value: execCommand });
}

describe('copyText', () => {
  afterEach(() => {
    Reflect.deleteProperty(navigator, 'clipboard');
    Reflect.deleteProperty(document, 'execCommand');
    vi.restoreAllMocks();
  });

  it('writes through the async Clipboard API when available', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    stubClipboard(writeText);
    const execCommand = vi.fn().mockReturnValue(true);
    stubExecCommand(execCommand);

    await copyText('hello world');

    expect(writeText).toHaveBeenCalledWith('hello world');
    expect(execCommand).not.toHaveBeenCalled();
  });

  it('falls back to execCommand when the Clipboard API rejects (menu-bar webview)', async () => {
    const writeText = vi.fn().mockRejectedValue(new Error('clipboard permission denied'));
    stubClipboard(writeText);
    const execCommand = vi.fn().mockReturnValue(true);
    stubExecCommand(execCommand);

    await copyText('fallback payload');

    expect(writeText).toHaveBeenCalledWith('fallback payload');
    expect(execCommand).toHaveBeenCalledWith('copy');
  });

  it('falls back to execCommand when navigator.clipboard is absent', async () => {
    stubClipboard(undefined);
    const execCommand = vi.fn().mockReturnValue(true);
    stubExecCommand(execCommand);

    await copyText('no clipboard api');

    expect(execCommand).toHaveBeenCalledWith('copy');
  });

  it('rejects when both the Clipboard API and execCommand fail', async () => {
    stubClipboard(vi.fn().mockRejectedValue(new Error('denied')));
    stubExecCommand(vi.fn().mockReturnValue(false));

    await expect(copyText('unwritable')).rejects.toThrow('Browser rejected the copy request');
  });
});
