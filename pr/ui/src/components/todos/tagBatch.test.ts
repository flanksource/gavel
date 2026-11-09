import { QueryClient } from '@tanstack/react-query';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { readTagBatch } from './tagBatch';

afterEach(() => vi.unstubAllGlobals());

describe('tag request batching', () => {
  it('combines simultaneous workspaces and preserves their individual counts', async () => {
    const fetch = vi.fn(async (_url: RequestInfo | URL, options?: RequestInit) => {
      const { dirs } = JSON.parse(String(options?.body)) as { dirs: string[] };
      return Response.json({ results: dirs.map((dir, i) => ({ dir, labels: { definitions: [], counts: { bug: i + 1 } } })), palette: [] });
    });
    vi.stubGlobal('fetch', fetch);
    const client = new QueryClient();
    const signal = new AbortController().signal;
    const result = await Promise.all(['/work/first', '/work/second'].map(dir => readTagBatch(client, dir, signal)));
    expect(fetch).toHaveBeenCalledTimes(1);
    expect(result.map(index => index.counts)).toEqual([{ bug: 1 }, { bug: 2 }]);
  });

  it('rejects a mismatched response rather than assigning another workspace taxonomy', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => Response.json({ results: [{ dir: '/other', labels: { definitions: [] } }] })));
    await expect(readTagBatch(new QueryClient(), '/work/first', new AbortController().signal)).rejects.toThrow('Invalid tag batch response');
  });

  it('does not dispatch a cancelled batch', async () => {
    const fetch = vi.fn();
    vi.stubGlobal('fetch', fetch);
    const controller = new AbortController();
    const read = readTagBatch(new QueryClient(), '/work/first', controller.signal);
    controller.abort();
    await expect(read).rejects.toThrow();
    expect(fetch).not.toHaveBeenCalled();
  });
});
