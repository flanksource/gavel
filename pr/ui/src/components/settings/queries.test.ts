import { QueryClient } from '@tanstack/react-query';
import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  settingsConfigQuery,
  settingsPromptsQuery,
  settingsRunContextQuery,
  settingsSchemaQuery,
  settingsTraceQuery,
} from './queries';

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('settings queries', () => {
  it('loads each workspace run context with an isolated cache key and directory', async () => {
    const fetchMock = vi.fn(async (url: string) => ({ ok: true, json: async () => ({ workspace: new URL(url, 'http://localhost').searchParams.get('dir') }) }));
    vi.stubGlobal('fetch', fetchMock);
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const workspaces = ['/workspace/alpha', '/workspace/beta'];
    const contexts = await Promise.all(workspaces.map(dir => client.fetchQuery(settingsRunContextQuery(dir))));
    expect(contexts).toEqual(workspaces.map(workspace => ({ workspace })));
    expect(fetchMock.mock.calls.map(([url]) => url)).toEqual(workspaces.map(dir => `/api/todos/run/context?dir=${encodeURIComponent(dir)}`));
  });

  it('isolates settings layers and trace scopes in their keys', () => {
    expect(settingsConfigQuery('scope=global').queryKey)
      .not.toEqual(settingsConfigQuery('project=widget').queryKey);
    expect(settingsTraceQuery('scope=global').queryKey)
      .not.toEqual(settingsTraceQuery('project=widget').queryKey);
    expect(settingsSchemaQuery().queryKey).toEqual(['settings', 'schema']);
    expect(settingsPromptsQuery().queryKey).toEqual(['settings', 'prompts']);
    expect(settingsRunContextQuery().queryKey).toEqual(['todos', 'run-context']);
  });

  it('deduplicates concurrent schema reads and supplies an AbortSignal', async () => {
    let requestSignal: AbortSignal | undefined;
    const fetchMock = vi.fn((_url: string, init?: RequestInit) => {
      requestSignal = init?.signal as AbortSignal;
      return Promise.resolve({ ok: true, json: async () => ({ type: 'object', properties: {} }) });
    });
    vi.stubGlobal('fetch', fetchMock);
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });

    const [first, second] = await Promise.all([
      client.fetchQuery(settingsSchemaQuery()),
      client.fetchQuery(settingsSchemaQuery()),
    ]);

    expect(first).toEqual(second);
    expect(fetchMock).toHaveBeenCalledOnce();
    expect(requestSignal).toBeInstanceOf(AbortSignal);
  });

  it('surfaces the settings resource in errors', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
      text: async () => 'broken yaml',
    }));
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });

    await expect(client.fetchQuery(settingsConfigQuery('project=widget')))
      .rejects.toThrow('Failed to load project settings: broken yaml');
  });
});
