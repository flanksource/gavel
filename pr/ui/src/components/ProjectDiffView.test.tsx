import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ProjectDiffView } from './ProjectDiffView';

vi.mock('@flanksource/clicky-ui/data', () => ({
  GitDiffPanel: ({ loading, payload, error }: { loading: boolean; payload: { patch?: string } | null; error: string }) => (
    <div>{loading ? 'loading' : error || payload?.patch || 'empty'}</div>
  ),
}));

vi.mock('@flanksource/clicky-ui/icons', () => ({ UiDiff: () => <span /> }));

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

function createClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false, gcTime: Infinity } } });
}

function Provider({ client, children }: { client: QueryClient; children: ReactNode }) {
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

describe('ProjectDiffView queries', () => {
  it('deduplicates concurrent consumers and reuses the cached diff after remount', async () => {
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({ patch: 'diff:one.go' }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    }));
    vi.stubGlobal('fetch', fetchMock);
    const client = createClient();

    const first = render(
      <Provider client={client}>
        <ProjectDiffView projectName="gavel" path="one.go" />
        <ProjectDiffView projectName="gavel" path="one.go" />
      </Provider>,
    );

    expect(await screen.findAllByText('diff:one.go')).toHaveLength(2);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    first.unmount();

    render(
      <Provider client={client}>
        <ProjectDiffView projectName="gavel" path="one.go" />
      </Provider>,
    );

    expect(await screen.findByText('diff:one.go')).toBeTruthy();
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it('cancels the prior request when the diff identity changes', async () => {
    const signals: AbortSignal[] = [];
    vi.stubGlobal('fetch', vi.fn((_input: RequestInfo | URL, init?: RequestInit) => {
      signals.push(init?.signal as AbortSignal);
      return new Promise<Response>(() => undefined);
    }));
    const client = createClient();

    const view = render(
      <Provider client={client}>
        <ProjectDiffView projectName="gavel" path="one.go" />
      </Provider>,
    );
    await waitFor(() => expect(signals).toHaveLength(1));

    view.rerender(
      <Provider client={client}>
        <ProjectDiffView projectName="gavel" path="two.go" />
      </Provider>,
    );

    await waitFor(() => expect(signals).toHaveLength(2));
    expect(signals[0].aborted).toBe(true);
  });
});
