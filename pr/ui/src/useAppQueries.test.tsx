import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, render, screen, waitFor } from '@testing-library/react';
import type { PropsWithChildren } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { PRItem, ProcStatus, Snapshot } from './types';
import { queryKeys } from './query';
import { useAppQueries } from './useAppQueries';

const pullRequest: PRItem = {
  number: 7,
  title: 'Cache bootstrap reads',
  author: 'octocat',
  repo: 'acme/gavel',
  source: 'query-cache',
  target: 'main',
  state: 'OPEN',
  isDraft: false,
  url: 'https://example.com/acme/gavel/pull/7',
  updatedAt: '2026-08-02T09:00:00Z',
};

const snapshot: Snapshot = {
  prs: [pullRequest],
  fetchedAt: '2026-08-02T09:00:00Z',
  nextFetchIn: 60,
  incremental: false,
  paused: false,
  config: { repos: ['acme/gavel'] },
};

const procStatus: Record<string, ProcStatus> = {
  gavel: { hasProcfile: true, running: true },
};

class FakeEventSource {
  static instances: FakeEventSource[] = [];
  readonly listeners = new Map<string, EventListener>();
  onerror: ((event: Event) => void) | null = null;
  readonly close = vi.fn();

  constructor(readonly url: string) {
    FakeEventSource.instances.push(this);
  }

  addEventListener(type: string, listener: EventListener) {
    this.listeners.set(type, listener);
  }

  emit(type: string, data: unknown) {
    this.listeners.get(type)?.(new MessageEvent(type, { data: JSON.stringify(data) }));
  }
}

function Probe({ name, enabled }: { name: string; enabled: boolean }) {
  const state = useAppQueries({ enabled, initialConfig: { repos: [] } });
  return (
    <div data-testid={name} data-error={state.processError}>
      {state.snapshot.prs.length}/{state.projects.length}/{Object.keys(state.procStatus).length}
    </div>
  );
}

function createClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false, refetchOnWindowFocus: false } },
  });
}

function Provider({ client, children }: PropsWithChildren<{ client: QueryClient }>) {
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

afterEach(() => {
  FakeEventSource.instances = [];
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('useAppQueries', () => {
  it('deduplicates bootstrap reads, stores stream frames in the cache, and reuses data after remount', async () => {
    const calls: string[] = [];
    vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      calls.push(url);
      const payload = url === '/api/prs'
        ? snapshot
        : url === '/api/projects'
          ? [{ name: 'gavel', dir: '/work/gavel', repos: ['acme/gavel'] }]
          : procStatus;
      return { ok: true, json: async () => payload } as Response;
    }));
    vi.stubGlobal('EventSource', FakeEventSource);
    const client = createClient();

    const first = render(
      <Provider client={client}>
        <Probe name="first" enabled />
        <Probe name="second" enabled />
      </Provider>,
    );

    await waitFor(() => expect(screen.getByTestId('first').textContent).toBe('1/1/1'));
    expect(calls.filter(url => url === '/api/prs')).toHaveLength(1);
    expect(calls.filter(url => url === '/api/projects')).toHaveLength(1);
    expect(calls.filter(url => url === '/api/proc/status')).toHaveLength(1);

    const streamed = { ...snapshot, prs: [{ ...pullRequest, number: 8 }] };
    act(() => FakeEventSource.instances.find(stream => stream.url === '/api/prs/stream')?.emit('message', streamed));
    await waitFor(() => expect(client.getQueryData(queryKeys.prSnapshot())).toMatchObject({ prs: [{ number: 8 }] }));

    act(() => FakeEventSource.instances.find(stream => stream.url === '/api/proc/status/stream')?.emit('message', 'invalid'));
    await waitFor(() => expect(screen.getByTestId('first').dataset.error).toContain('invalid update'));

    first.unmount();
    render(
      <Provider client={client}>
        <Probe name="remounted" enabled />
      </Provider>,
    );
    await waitFor(() => expect(screen.getByTestId('remounted').textContent).toBe('1/1/1'));
    expect(calls.filter(url => url === '/api/prs')).toHaveLength(1);
    expect(calls.filter(url => url === '/api/projects')).toHaveLength(1);
    expect(calls.filter(url => url === '/api/proc/status')).toHaveLength(1);
  });

  it('aborts in-flight bootstrap reads and closes streams when disabled', async () => {
    const signals: AbortSignal[] = [];
    vi.stubGlobal('fetch', vi.fn((_input: RequestInfo | URL, init?: RequestInit) => {
      if (init?.signal) signals.push(init.signal);
      return new Promise<Response>(() => {});
    }));
    vi.stubGlobal('EventSource', FakeEventSource);
    const client = createClient();

    const view = render(
      <Provider client={client}>
        <Probe name="probe" enabled />
      </Provider>,
    );
    await waitFor(() => expect(signals).toHaveLength(3));

    view.rerender(
      <Provider client={client}>
        <Probe name="probe" enabled={false} />
      </Provider>,
    );

    await waitFor(() => expect(signals.every(signal => signal.aborted)).toBe(true));
    expect(FakeEventSource.instances.every(stream => stream.close.mock.calls.length === 1)).toBe(true);
  });
});
