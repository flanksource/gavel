import { afterEach, describe, expect, it, vi } from 'vitest';
import { fetchJSON, fetchText, mutationJSON, queryKeys } from './query';

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('queryKeys', () => {
  it('isolates each resource identity and request parameter', () => {
    expect(queryKeys.projectStatus('gavel')).not.toEqual(queryKeys.projectStatus('clicky'));
    expect(queryKeys.projectStatus('gavel', false)).not.toEqual(queryKeys.projectStatus('gavel', true));
    expect(queryKeys.projectStatus('gavel', true)).toEqual([...queryKeys.projectStatusScope('gavel'), 'results']);
    expect(queryKeys.projectDiff('gavel', 'one.go', 0)).not.toEqual(queryKeys.projectDiff('gavel', 'two.go', 0));
    expect(queryKeys.projectDiff('gavel', 'one.go', 0)).not.toEqual(queryKeys.projectDiff('gavel', 'one.go', 1));
    expect(queryKeys.processStatus('gavel')).not.toEqual(queryKeys.processStatus('clicky'));
    expect(queryKeys.processLogs('gavel', 'api', 5)).not.toEqual(queryKeys.processLogs('gavel', 'api', 500));
    expect(queryKeys.processLogs('gavel', 'api', 5)).not.toEqual(queryKeys.processLogs('gavel', 'worker', 5));
    expect(queryKeys.prDetail('acme/gavel', 7)).not.toEqual(queryKeys.prDetail('acme/gavel', 8));
    expect(queryKeys.prDetail('acme/gavel', 7)).not.toEqual(queryKeys.prDetail('acme/clicky', 7));
  });
});

describe('query requests', () => {
  it('passes the React Query AbortSignal to JSON GETs', async () => {
    const controller = new AbortController();
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({ state: 'ready' }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    }));
    vi.stubGlobal('fetch', fetchMock);

    await expect(fetchJSON<{ state: string }>({
      url: '/api/example',
      signal: controller.signal,
      context: 'Failed to load example',
    })).resolves.toEqual({ state: 'ready' });
    expect(fetchMock).toHaveBeenCalledWith('/api/example', { signal: controller.signal });
  });

  it('surfaces contextual JSON errors from the backend', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ error: 'database unavailable' }), {
      status: 503,
      headers: { 'Content-Type': 'application/json' },
    })));

    await expect(fetchJSON({
      url: '/api/example',
      signal: new AbortController().signal,
      context: 'Failed to load example',
    })).rejects.toThrow('Failed to load example: database unavailable');
  });

  it('passes the AbortSignal to text GETs and contextualizes HTTP errors', async () => {
    const controller = new AbortController();
    const fetchMock = vi.fn(async () => new Response('permission denied', { status: 403 }));
    vi.stubGlobal('fetch', fetchMock);

    await expect(fetchText({
      url: '/api/example.log',
      signal: controller.signal,
      context: 'Failed to load logs',
    })).rejects.toThrow('Failed to load logs: permission denied');
    expect(fetchMock).toHaveBeenCalledWith('/api/example.log', { signal: controller.signal });
  });

  it('sends JSON mutations and surfaces contextual backend errors', async () => {
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({ paused: true }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    }));
    vi.stubGlobal('fetch', fetchMock);

    await expect(mutationJSON<{ paused: boolean }>({
      url: '/api/prs/pause',
      method: 'POST',
      context: 'Toggle pull request polling',
    })).resolves.toEqual({ paused: true });
    expect(fetchMock).toHaveBeenCalledWith('/api/prs/pause', { method: 'POST' });

    fetchMock.mockResolvedValueOnce(new Response(JSON.stringify({ error: 'poller unavailable' }), {
      status: 503,
      headers: { 'Content-Type': 'application/json' },
    }));
    await expect(mutationJSON({
      url: '/api/prs/pause',
      method: 'POST',
      context: 'Toggle pull request polling',
    })).rejects.toThrow('Toggle pull request polling: poller unavailable');
  });

  it('serializes mutation bodies without weakening request context', async () => {
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({ includeBots: true }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    }));
    vi.stubGlobal('fetch', fetchMock);

    await mutationJSON({
      url: '/api/prs/bots',
      method: 'POST',
      body: { include: true },
      context: 'Update bot pull request visibility',
    });

    expect(fetchMock).toHaveBeenCalledWith('/api/prs/bots', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ include: true }),
    });
  });

  it('accepts empty successful mutations and preserves plain-text backend errors', async () => {
    const fetchMock = vi.fn(async () => new Response(null, { status: 204 }));
    vi.stubGlobal('fetch', fetchMock);

    await expect(mutationJSON<void>({
      url: '/api/projects/acme',
      method: 'DELETE',
      context: 'Delete project',
    })).resolves.toBeUndefined();

    fetchMock.mockResolvedValueOnce(new Response('project is running', { status: 409 }));
    await expect(mutationJSON({
      url: '/api/projects/acme',
      method: 'DELETE',
      context: 'Delete project',
    })).rejects.toThrow('Delete project: project is running');
  });
});
