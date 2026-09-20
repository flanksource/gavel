import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { TodoTagListResponse } from '../../types';
import { todoTagsCacheKey, useTodoTagCounts, useTodoTagIndex } from './tagQueries';

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
  // The label picker on the new-todo window offers the last known vocabulary
  // at once, then swaps in the fresh taxonomy — which is cached per workspace.
  it('renders the cached taxonomy while loading and caches the fresh one per workspace', async () => {
    const ui = { name: 'ui', color: 'blue', scope: 'workspace' } as const;
    const flaky = { name: 'flaky', color: 'red', scope: 'global' } as const;
    const cached: TodoTagListResponse = { definitions: [ui], counts: { ui: 3 } };
    const fresh: TodoTagListResponse = { definitions: [ui, flaky], counts: { ui: 4, flaky: 1 } };
    localStorage.setItem(todoTagsCacheKey(dir), JSON.stringify(cached));
    let release!: () => void;
    const gate = new Promise<void>(resolve => { release = resolve; });
    vi.stubGlobal('fetch', vi.fn(async () => {
      await gate;
      return { ok: true, json: async () => fresh } as Response;
    }));

    render(
      <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
        <Probe />
      </QueryClientProvider>,
    );

    expect(screen.getByTestId('tags').textContent).toBe('ui:3');
    expect(screen.getByTestId('tags').dataset.loading).toBe('false');
    release();
    await waitFor(() => expect(screen.getByTestId('tags').textContent).toBe('ui:4,flaky:1'));
    expect(JSON.parse(localStorage.getItem(todoTagsCacheKey(dir)) ?? 'null')).toEqual(fresh);
    expect(localStorage.getItem(todoTagsCacheKey('/work/other'))).toBeNull();
  });
});
