import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { TodoTagListResponse } from '../../types';
import { todoTagKeys, useTodoTagCounts, useTodoTagIndex } from './tagQueries';

const dir = '/work/acme';

function Probe() {
  const index = useTodoTagIndex(dir);
  const counts = useTodoTagCounts(dir);
  return (
    <div data-testid="tags" data-loading={String(index.loading)}>
      {index.defs.map(def => `${def.name}:${counts[def.name] ?? 0}`).join(',')}
    </div>
  );
}

afterEach(() => {
  localStorage.clear();
  vi.unstubAllGlobals();
});

describe('tag definition cache', () => {
  it('reuses the account cache and refreshes taxonomy on explicit invalidation', async () => {
    const ui = { name: 'ui', color: 'blue', scope: 'workspace' } as const;
    const flaky = { name: 'flaky', color: 'red', scope: 'global' } as const;
    const cached: TodoTagListResponse = { definitions: [ui], counts: { ui: 3 } };
    const fresh: TodoTagListResponse = { definitions: [ui, flaky], counts: { ui: 4, flaky: 1 } };
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    client.setQueryData(todoTagKeys.list(dir), cached);
    let release!: () => void;
    const gate = new Promise<void>(resolve => { release = resolve; });
    vi.stubGlobal('fetch', vi.fn(async () => {
      await gate;
      return { ok: true, json: async () => ({ results: [{ dir, labels: fresh }] }) } as Response;
    }));

    render(
      <QueryClientProvider client={client}>
        <Probe />
      </QueryClientProvider>,
    );

    expect(screen.getByTestId('tags').textContent).toBe('ui:3');
    expect(screen.getByTestId('tags').dataset.loading).toBe('false');
    await act(async () => { void client.invalidateQueries({ queryKey: todoTagKeys.list(dir) }); release(); });
    await waitFor(() => expect(screen.getByTestId('tags').textContent).toBe('ui:4,flaky:1'));
    expect(client.getQueryData(todoTagKeys.list(dir))).toEqual(fresh);
    expect(client.getQueryData(todoTagKeys.list('/work/other'))).toBeUndefined();
    expect(localStorage.length).toBe(0);
  });
});
