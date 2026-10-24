import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import type { PropsWithChildren } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { Project } from './types';
import { useProjectCatalog } from './useProjectCatalog';

vi.mock('./eventHub', () => ({
  openEventStream: () => ({ addEventListener: () => undefined, removeEventListener: () => undefined, close: () => undefined, onerror: null }),
}));

const run = { runId: 'run-1', kind: 'test', passed: 1, failed: 0, skipped: 0, warned: 0, total: 1, lintViolations: 0, lintLinters: 0 };
const latestRuns = {
  projects: [
    { name: 'gavel', dir: '/work/gavel', runs: [run] },
    { name: 'history-only', dir: '/work/history-only', runs: [{ ...run, repo: 'acme/history-only' }] },
    { name: 'archive', dir: '/work/archive', runs: [{ ...run, repo: 'acme/archive' }] },
  ],
};

const configured: Project[] = [{ name: 'gavel', dir: '/work/gavel', repos: ['acme/gavel'] }];

function wrapper({ children }: PropsWithChildren) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('useProjectCatalog hidden projects', () => {
  function stubRuns() {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => latestRuns }));
  }

  it('lists history-only projects whose directory is not hidden', async () => {
    stubRuns();
    const { result } = renderHook(
      () => useProjectCatalog({ configured, hiddenDirs: [], selectedName: '', enabled: true }),
      { wrapper },
    );

    await waitFor(() => expect(result.current.projects.map(project => project.name)).toEqual(['gavel', 'history-only', 'archive']));
  });

  it('drops a history-only project whose directory belongs to a hidden project', async () => {
    stubRuns();
    const { result } = renderHook(
      () => useProjectCatalog({ configured, hiddenDirs: ['/work/archive'], selectedName: 'archive', enabled: true }),
      { wrapper },
    );

    await waitFor(() => expect(result.current.projects.map(project => project.name)).toEqual(['gavel', 'history-only']));
    expect(result.current.selected).toBeNull();
  });
});
