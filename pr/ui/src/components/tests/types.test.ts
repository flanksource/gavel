import { QueryClient } from '@tanstack/react-query';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { runSnapshotQuery } from './types';

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('runSnapshotQuery', () => {
  it('isolates project, directory, and run identities', () => {
    expect(runSnapshotQuery({ project: 'gavel', runId: 'run-1' }).queryKey)
      .not.toEqual(runSnapshotQuery({ project: 'gavel', runId: 'run-2' }).queryKey);
    expect(runSnapshotQuery({ project: 'gavel', runId: 'run-1' }).queryKey)
      .not.toEqual(runSnapshotQuery({ dir: '', runId: 'run-1' }).queryKey);
  });

  it('passes cancellation through and reports the run identity on failure', async () => {
    let requestSignal: AbortSignal | undefined;
    vi.stubGlobal('fetch', vi.fn((_url: string, init?: RequestInit) => {
      requestSignal = init?.signal as AbortSignal;
      return Promise.resolve({
        ok: false,
        status: 404,
        json: async () => ({ error: 'snapshot missing' }),
      });
    }));
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });

    await expect(client.fetchQuery(runSnapshotQuery({ project: 'gavel', runId: 'run-404' })))
      .rejects.toThrow('Failed to load run run-404 for project gavel: snapshot missing');
    expect(requestSignal).toBeInstanceOf(AbortSignal);
  });
});
