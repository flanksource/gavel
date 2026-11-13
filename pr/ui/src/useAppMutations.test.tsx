import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { PropsWithChildren } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { todoQueryKeys } from './components/todos/todoQueries';
import { queryKeys } from './query';
import type { PRItem, SearchConfig, Snapshot } from './types';
import { useAppMutations } from './useAppMutations';

const pullRequest: PRItem = {
  number: 7,
  title: 'Coordinate application mutations',
  author: 'octocat',
  repo: 'acme/gavel',
  source: 'query-cache',
  target: 'main',
  state: 'OPEN',
  isDraft: false,
  url: 'https://example.com/acme/gavel/pull/7',
  updatedAt: '2026-08-02T09:00:00Z',
};

const initialConfig: SearchConfig = { repos: ['acme/gavel'], org: 'acme' };
const snapshot: Snapshot = {
  prs: [pullRequest],
  fetchedAt: '2026-08-02T09:00:00Z',
  nextFetchIn: 60,
  incremental: false,
  paused: false,
  config: initialConfig,
  unread: { 'acme/gavel#7': true },
};

function createClient() {
  return new QueryClient({ defaultOptions: { mutations: { retry: false }, queries: { retry: false } } });
}

function wrapper(client: QueryClient) {
  return function Provider({ children }: PropsWithChildren) {
    return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
  };
}

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('useAppMutations', () => {
  it.each([200, 503])('tracks refresh progress and settles after HTTP %i', async status => {
    let finishRequest!: (response: Response) => void;
    vi.stubGlobal('fetch', vi.fn(() => new Promise<Response>(resolve => { finishRequest = resolve; })));
    const client = createClient();
    const invalidate = vi.spyOn(client, 'invalidateQueries').mockResolvedValue(undefined);
    const { result } = renderHook(useAppMutations, { wrapper: wrapper(client) });

    expect(result.current.refreshing).toBe(false);
    act(() => result.current.refresh());
    await waitFor(() => expect(result.current.refreshing).toBe(true));
    act(() => finishRequest(new Response(JSON.stringify(status === 200 ? { status: 'refresh requested' } : { error: 'offline' }), { status })));

    await waitFor(() => expect(result.current.refreshing).toBe(false));
    expect(result.current.error).toBe(status === 200 ? '' : 'Refresh pull requests: offline');
    expect(invalidate).toHaveBeenCalledTimes(status === 200 ? 1 : 0);
    expect(fetch).toHaveBeenCalledWith('/api/prs/refresh', { method: 'POST' });
  });

  it('optimistically marks a PR seen and rolls the exact snapshot back on failure', async () => {
    let finishRequest: (response: Response) => void = () => undefined;
    vi.stubGlobal('fetch', vi.fn(() => new Promise<Response>(resolve => { finishRequest = resolve; })));
    const client = createClient();
    client.setQueryData(queryKeys.prSnapshot(), snapshot);
    const { result } = renderHook(useAppMutations, { wrapper: wrapper(client) });

    act(() => result.current.markSeen(pullRequest));
    await waitFor(() => expect(client.getQueryData<Snapshot>(queryKeys.prSnapshot())?.unread).toEqual({}));
    act(() => finishRequest(new Response(JSON.stringify({ error: 'cache unavailable' }), {
      status: 503,
      headers: { 'Content-Type': 'application/json' },
    })));
    await waitFor(() => expect(result.current.error).toBe('Mark pull request seen: cache unavailable'));
    expect(client.getQueryData(queryKeys.prSnapshot())).toEqual(snapshot);
  });

  it('updates server-echoed snapshot fields and invalidates only the PR snapshot on refresh', async () => {
    const requests: Array<{ url: string; init?: RequestInit }> = [];
    vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      requests.push({ url, init });
      const payload = url.endsWith('/pause')
        ? { paused: true }
        : url.endsWith('/bots')
          ? { includeBots: true }
          : url.endsWith('/closed')
            ? { showClosed: true }
            : { status: 'refresh requested' };
      return new Response(JSON.stringify(payload), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      });
    }));
    const client = createClient();
    client.setQueryData(queryKeys.prSnapshot(), snapshot);
    const invalidate = vi.spyOn(client, 'invalidateQueries').mockResolvedValue(undefined);
    const { result } = renderHook(useAppMutations, { wrapper: wrapper(client) });

    act(() => {
      result.current.togglePause();
      result.current.setIncludeBots(true);
      result.current.setShowClosed(true);
      result.current.refresh();
    });

    await waitFor(() => expect(client.getQueryData<Snapshot>(queryKeys.prSnapshot())).toMatchObject({
      paused: true,
      includeBots: true,
      showClosed: true,
    }));
    await waitFor(() => expect(invalidate).toHaveBeenCalledWith({ queryKey: queryKeys.prSnapshot(), exact: true }));
    expect(invalidate).toHaveBeenCalledTimes(1);
    expect(requests.map(request => request.url)).toEqual([
      '/api/prs/pause',
      '/api/prs/bots',
      '/api/prs/closed',
      '/api/prs/refresh',
    ]);
  });

  it.each([
    { hidden: true, method: 'POST' },
    { hidden: false, method: 'DELETE' },
  ])('sends the project dir to the hidden-projects filter ($method) and refreshes everything derived from it', async ({ hidden, method }) => {
    const dir = '/work/my project';
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ name: 'hidden-projects', kind: 'workspace', values: hidden ? [dir] : [] }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    })));
    const client = createClient();
    const invalidate = vi.spyOn(client, 'invalidateQueries').mockResolvedValue(undefined);
    const { result } = renderHook(useAppMutations, { wrapper: wrapper(client) });

    act(() => result.current.setProjectHidden({ dir, hidden }));

    await waitFor(() => expect(invalidate).toHaveBeenCalledWith({ queryKey: queryKeys.projects(), exact: true }));
    expect(fetch).toHaveBeenCalledWith('/api/filters/hidden-projects/values', {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ value: dir }),
    });
    expect(invalidate.mock.calls.map(([filters]) => filters?.queryKey)).toEqual([
      queryKeys.projects(),
      queryKeys.projectGitSummary(),
      queryKeys.filters(),
      todoQueryKeys.workspaceList('hidden-projects'),
      queryKeys.orgs(),
    ]);
  });

  it('sends the org login to the hidden-github-orgs filter', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ name: 'hidden-github-orgs', kind: 'github-org', values: ['acme'] }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    })));
    const client = createClient();
    const invalidate = vi.spyOn(client, 'invalidateQueries').mockResolvedValue(undefined);
    const { result } = renderHook(useAppMutations, { wrapper: wrapper(client) });

    act(() => result.current.setOrgHidden({ login: 'acme', hidden: true }));

    await waitFor(() => expect(invalidate).toHaveBeenCalledWith({ queryKey: queryKeys.orgs() }));
    expect(fetch).toHaveBeenCalledWith('/api/filters/hidden-github-orgs/values', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ value: 'acme' }),
    });
  });

  it('reports a failed hide and leaves the project queries untouched', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ error: 'project "ghost" not found' }), {
      status: 404,
      headers: { 'Content-Type': 'application/json' },
    })));
    const client = createClient();
    const invalidate = vi.spyOn(client, 'invalidateQueries').mockResolvedValue(undefined);
    const { result } = renderHook(useAppMutations, { wrapper: wrapper(client) });

    act(() => result.current.setProjectHidden({ dir: '/work/ghost', hidden: true }));

    await waitFor(() => expect(result.current.error).toBe('Update project visibility: project "ghost" not found'));
    expect(invalidate).not.toHaveBeenCalled();
  });

  it('optimistically saves config, accepts the server response, and rolls back a later failure', async () => {
    const requestedConfig: SearchConfig = { repos: ['acme/clicky'], all: true };
    const savedConfig: SearchConfig = { ...requestedConfig, org: 'server-normalized' };
    let finishRequest: (response: Response) => void = () => undefined;
    vi.stubGlobal('fetch', vi.fn(() => new Promise<Response>(resolve => { finishRequest = resolve; })));
    const client = createClient();
    client.setQueryData(queryKeys.prSnapshot(), snapshot);
    const { result } = renderHook(useAppMutations, { wrapper: wrapper(client) });

    act(() => result.current.saveConfig(requestedConfig));
    await waitFor(() => expect(client.getQueryData<Snapshot>(queryKeys.prSnapshot())?.config).toEqual(requestedConfig));
    act(() => finishRequest(new Response(JSON.stringify(savedConfig), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    })));

    await waitFor(() => expect(client.getQueryData<Snapshot>(queryKeys.prSnapshot())?.config).toEqual(savedConfig));
    expect(fetch).toHaveBeenCalledWith('/api/config', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(requestedConfig),
    });

    act(() => result.current.saveConfig({ repos: [] }));
    await waitFor(() => expect(client.getQueryData<Snapshot>(queryKeys.prSnapshot())?.config).toEqual({ repos: [] }));
    act(() => finishRequest(new Response(JSON.stringify({ error: 'settings unavailable' }), {
      status: 503,
      headers: { 'Content-Type': 'application/json' },
    })));
    await waitFor(() => expect(result.current.error).toBe('Save pull request config: settings unavailable'));
    expect(client.getQueryData<Snapshot>(queryKeys.prSnapshot())?.config).toEqual(savedConfig);
  });
});
