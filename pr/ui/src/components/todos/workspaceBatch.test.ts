import { QueryClient } from '@tanstack/react-query';
import { afterEach, expect, it, vi } from 'vitest';
import { readWorkspaceBatch } from './workspaceBatch';

afterEach(() => vi.unstubAllGlobals());

it('batches concurrent workspace reads while isolating failures', async () => {
  const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => Response.json({ results: [
    { dir: '/first', items: [{ ref: 'one' }], counts: { total: 1 } },
    { dir: '/second', error: { message: 'Workspace unavailable' } },
  ] }));
  vi.stubGlobal('fetch', fetchMock);
  const client = new QueryClient();
  const first = readWorkspaceBatch(client, '/first', new AbortController().signal);
  const second = readWorkspaceBatch(client, '/second', new AbortController().signal);
  const results = await Promise.allSettled([first, second]);
  expect(fetchMock).toHaveBeenCalledTimes(1);
  expect(JSON.parse(fetchMock.mock.calls[0][1]!.body as string)).toEqual({ dirs: ['/first', '/second'] });
  expect(results[0]).toMatchObject({ status: 'fulfilled', value: { dir: '/first', items: [{ ref: 'one' }] } });
  expect(results[1]).toMatchObject({ status: 'rejected', reason: new Error('Workspace unavailable') });
});

it('rejects a reordered response and does not read cancelled workspaces', async () => {
  const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => Response.json({ results: [{ dir: '/other', items: [], counts: {} }] }));
  vi.stubGlobal('fetch', fetchMock);
  const client = new QueryClient();
  const cancelled = new AbortController(); cancelled.abort();
  const outcomes = await Promise.allSettled([
    readWorkspaceBatch(client, '/cancelled', cancelled.signal),
    readWorkspaceBatch(client, '/first', new AbortController().signal),
  ]);
  expect(JSON.parse(fetchMock.mock.calls[0][1]!.body as string)).toEqual({ dirs: ['/first'] });
  expect(outcomes.map(outcome => outcome.status)).toEqual(['rejected', 'rejected']);
  expect(outcomes[1]).toMatchObject({ reason: new Error('Invalid TODO batch response') });
});
