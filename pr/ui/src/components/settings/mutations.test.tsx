import type { ReactNode } from 'react';
import { act, renderHook } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { PromptSpecDetail } from '@flanksource/clicky-ui/ai';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { promptDetailQuery } from '../oneShotQueries';
import {
  settingsConfigQuery,
  settingsPromptsQuery,
  settingsRunContextQuery,
  settingsTraceQuery,
} from './queries';
import { useSavePromptMutation, useSaveSettingsMutation } from './mutations';

function testClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
}

function wrapper(client: QueryClient) {
  return function QueryWrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
  };
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('settings mutations', () => {
  it('stores the server settings response and invalidates only derived data for that project layer', async () => {
    const client = testClient();
    const scopeQuery = 'project=widget';
    const otherScopeQuery = 'project=other';
    const saved = { config: { ai: { model: 'server-normalized' } }, path: '/work/widget/.gavel.yaml' };
    const currentDetail = {
      id: 'verify',
      scope: 'project',
      source: 'default',
      raw: 'current',
      spec: {},
      body: 'current',
    } satisfies PromptSpecDetail;
    client.setQueryData(settingsConfigQuery(scopeQuery).queryKey, { config: { ai: { model: 'old' } } });
    client.setQueryData(settingsConfigQuery(otherScopeQuery).queryKey, { config: { ai: { model: 'other' } } });
    client.setQueryData(settingsTraceQuery(scopeQuery).queryKey, { merged: {} });
    client.setQueryData(settingsTraceQuery(otherScopeQuery).queryKey, { merged: {} });
    client.setQueryData(promptDetailQuery('verify', scopeQuery).queryKey, currentDetail);
    client.setQueryData(promptDetailQuery('verify', otherScopeQuery).queryKey, currentDetail);
    client.setQueryData(settingsPromptsQuery().queryKey, {});
    client.setQueryData(settingsRunContextQuery().queryKey, { modes: [], runtimes: [], models: [], efforts: [], tools: [] });
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => saved }));
    const { result } = renderHook(() => useSaveSettingsMutation(scopeQuery), { wrapper: wrapper(client) });

    await act(() => result.current.mutateAsync({ ai: { model: 'requested' } }));

    expect(fetch).toHaveBeenCalledWith('/api/settings/gavel?project=widget', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ai: { model: 'requested' } }),
    });
    expect(client.getQueryData(settingsConfigQuery(scopeQuery).queryKey)).toEqual(saved);
    expect(client.getQueryState(settingsConfigQuery(otherScopeQuery).queryKey)?.isInvalidated).toBe(false);
    expect(client.getQueryState(settingsTraceQuery(scopeQuery).queryKey)?.isInvalidated).toBe(true);
    expect(client.getQueryState(settingsTraceQuery(otherScopeQuery).queryKey)?.isInvalidated).toBe(false);
    expect(client.getQueryState(promptDetailQuery('verify', scopeQuery).queryKey)?.isInvalidated).toBe(true);
    expect(client.getQueryState(promptDetailQuery('verify', otherScopeQuery).queryKey)?.isInvalidated).toBe(false);
    expect(client.getQueryState(settingsPromptsQuery().queryKey)?.isInvalidated).toBe(false);
    expect(client.getQueryState(settingsRunContextQuery().queryKey)?.isInvalidated).toBe(false);
  });

  it('surfaces a contextual settings save error without changing cached settings', async () => {
    const client = testClient();
    const scopeQuery = 'scope=global';
    const current = { config: { ai: { model: 'current' } }, path: '/home/.gavel.yaml' };
    client.setQueryData(settingsConfigQuery(scopeQuery).queryKey, current);
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: false,
      status: 422,
      text: async () => 'invalid model policy',
    }));
    const { result } = renderHook(() => useSaveSettingsMutation(scopeQuery), { wrapper: wrapper(client) });

    await expect(act(() => result.current.mutateAsync({ ai: { model: 'bad' } })))
      .rejects.toThrow('Failed to save global settings: invalid model policy');
    expect(client.getQueryData(settingsConfigQuery(scopeQuery).queryKey)).toEqual(current);
  });

  it('updates one prompt detail and invalidates the config and trace backed by the same project file', async () => {
    const client = testClient();
    const scopeQuery = 'project=widget';
    const otherScopeQuery = 'project=other';
    const saved = {
      id: 'verify',
      scope: 'project',
      source: 'inline',
      raw: '---\nmodel: new\n---\nbody',
      spec: { model: 'new' },
      body: 'body',
    } satisfies PromptSpecDetail;
    client.setQueryData(promptDetailQuery('verify', scopeQuery).queryKey, { ...saved, raw: 'old' });
    client.setQueryData(promptDetailQuery('commit', scopeQuery).queryKey, { ...saved, id: 'commit' });
    client.setQueryData(settingsConfigQuery(scopeQuery).queryKey, { config: {} });
    client.setQueryData(settingsConfigQuery(otherScopeQuery).queryKey, { config: {} });
    client.setQueryData(settingsTraceQuery(scopeQuery).queryKey, { merged: {} });
    client.setQueryData(settingsTraceQuery(otherScopeQuery).queryKey, { merged: {} });
    client.setQueryData(settingsPromptsQuery().queryKey, {});
    client.setQueryData(settingsRunContextQuery().queryKey, { modes: [], runtimes: [], models: [], efforts: [], tools: [] });
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => saved }));
    const { result } = renderHook(() => useSavePromptMutation('verify', scopeQuery), { wrapper: wrapper(client) });

    await act(() => result.current.mutateAsync({ source: 'inline', raw: saved.raw, baseRaw: 'old' }));

    expect(client.getQueryData(promptDetailQuery('verify', scopeQuery).queryKey)).toEqual(saved);
    expect(client.getQueryState(promptDetailQuery('commit', scopeQuery).queryKey)?.isInvalidated).toBe(false);
    expect(client.getQueryState(settingsConfigQuery(scopeQuery).queryKey)?.isInvalidated).toBe(true);
    expect(client.getQueryState(settingsConfigQuery(otherScopeQuery).queryKey)?.isInvalidated).toBe(false);
    expect(client.getQueryState(settingsTraceQuery(scopeQuery).queryKey)?.isInvalidated).toBe(true);
    expect(client.getQueryState(settingsTraceQuery(otherScopeQuery).queryKey)?.isInvalidated).toBe(false);
    expect(client.getQueryState(settingsPromptsQuery().queryKey)?.isInvalidated).toBe(false);
    expect(client.getQueryState(settingsRunContextQuery().queryKey)?.isInvalidated).toBe(false);
  });

  it('surfaces a contextual prompt save error without replacing its cached detail', async () => {
    const client = testClient();
    const scopeQuery = 'scope=global';
    const current = {
      id: 'verify',
      scope: 'global',
      source: 'default',
      raw: 'current',
      spec: {},
      body: 'current',
    } satisfies PromptSpecDetail;
    client.setQueryData(promptDetailQuery('verify', scopeQuery).queryKey, current);
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: false,
      status: 400,
      text: async () => '{"error":"frontmatter is invalid"}',
    }));
    const { result } = renderHook(() => useSavePromptMutation('verify', scopeQuery), { wrapper: wrapper(client) });

    await expect(act(() => result.current.mutateAsync({ source: 'inline', raw: 'broken', baseRaw: 'current' })))
      .rejects.toThrow('Failed to save prompt verify: frontmatter is invalid');
    expect(client.getQueryData(promptDetailQuery('verify', scopeQuery).queryKey)).toEqual(current);
  });
});
