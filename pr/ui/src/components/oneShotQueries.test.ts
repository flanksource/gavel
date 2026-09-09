import { QueryClient } from '@tanstack/react-query';
import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  jobLogsQuery,
  orgsQuery,
  prCommitDiffQuery,
  prFileDiffQuery,
  projectActionSchemaQuery,
  promptDetailQuery,
} from './oneShotQueries';

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('one-shot query options', () => {
  it('includes every request parameter in stable resource keys', () => {
    expect(orgsQuery().queryKey).toEqual(['orgs', { includeIgnored: true }]);
    expect(projectActionSchemaQuery('acme/widget', 'lint').queryKey)
      .not.toEqual(projectActionSchemaQuery('acme/widget', 'test').queryKey);
    expect(jobLogsQuery('acme/widget', 10, 20, 100).queryKey)
      .not.toEqual(jobLogsQuery('acme/widget', 10, 20, 200).queryKey);
    expect(promptDetailQuery('verify', 'project=widget').queryKey)
      .not.toEqual(promptDetailQuery('verify', 'scope=global').queryKey);
    expect(prCommitDiffQuery('acme/widget', 'abc123').queryKey)
      .not.toEqual(prCommitDiffQuery('acme/widget', 'def456').queryKey);
    expect(prFileDiffQuery('acme/widget', 7, 'a.go').queryKey)
      .not.toEqual(prFileDiffQuery('acme/widget', 7, 'b.go').queryKey);
  });

  it('reuses fresh cached data for the same key', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => [{ login: 'acme' }],
    });
    vi.stubGlobal('fetch', fetchMock);
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });

    await client.fetchQuery(orgsQuery());
    await client.fetchQuery(orgsQuery());

    expect(fetchMock).toHaveBeenCalledOnce();
  });

  it('forwards React Query cancellation to fetch', async () => {
    let requestSignal: AbortSignal | undefined;
    vi.stubGlobal('fetch', vi.fn((_url: string, init?: RequestInit) => {
      requestSignal = init?.signal as AbortSignal;
      return new Promise((_resolve, reject) => {
        requestSignal?.addEventListener('abort', () => reject(new DOMException('Aborted', 'AbortError')));
      });
    }));
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const pending = client.fetchQuery(jobLogsQuery('acme/widget', 10, 20));

    await client.cancelQueries({ queryKey: jobLogsQuery('acme/widget', 10, 20).queryKey });

    expect(requestSignal?.aborted).toBe(true);
    await expect(pending).rejects.toBeTruthy();
  });

  it('adds resource context to backend failures', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: false,
      status: 502,
      text: async () => 'upstream unavailable',
    }));
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });

    await expect(client.fetchQuery(projectActionSchemaQuery('acme/widget', 'lint')))
      .rejects.toThrow('Failed to load lint options for acme/widget: upstream unavailable');
  });
});
