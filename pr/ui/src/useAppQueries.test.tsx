import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, render, screen, waitFor } from '@testing-library/react';
import type { PropsWithChildren } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { PRItem, ProcStatus, Snapshot } from './types';
import { queryKeys } from './query';
import { useProcStatus } from './procStatusQuery';
import { PROJECTS_CACHE_KEY, useAppQueries } from './useAppQueries';

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

function Probe({ name, enabled, onRender }: { name: string; enabled: boolean; onRender?: () => void }) {
  const state = useAppQueries({ enabled, initialConfig: { repos: [] } });
  onRender?.();
  return (
    <div data-testid={name} data-error={state.processError} data-project-error={state.projectError}>
      {state.snapshot.prs.length}/{state.projects.length}
    </div>
  );
}

function ProcProbe({ name }: { name: string }) {
  const procStatus = useProcStatus();
  return <div data-testid={name}>{Object.keys(procStatus).sort().join(',')}</div>;
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
  localStorage.clear();
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
        <ProcProbe name="procs" />
      </Provider>,
    );

    await waitFor(() => expect(screen.getByTestId('first').textContent).toBe('1/1'));
    await waitFor(() => expect(screen.getByTestId('procs').textContent).toBe('gavel'));
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
        <ProcProbe name="remounted-procs" />
      </Provider>,
    );
    await waitFor(() => expect(screen.getByTestId('remounted').textContent).toBe('1/1'));
    expect(screen.getByTestId('remounted-procs').textContent).toBe('gavel');
    expect(calls.filter(url => url === '/api/prs')).toHaveLength(1);
    expect(calls.filter(url => url === '/api/projects')).toHaveLength(1);
    expect(calls.filter(url => url === '/api/proc/status')).toHaveLength(1);
  });

  // The new-todo window and sidebar render from the last known catalog while
  // /api/projects is in flight, and a successful load refreshes that cache.
  it('renders cached projects before the request resolves and caches the fresh list', async () => {
    const cached = [{ name: 'cached', dir: '/work/cached', repos: [] }];
    const fresh = [{ name: 'alpha', dir: '/work/alpha', repos: [] }, { name: 'beta', dir: '/work/beta', repos: [] }];
    localStorage.setItem(PROJECTS_CACHE_KEY, JSON.stringify(cached));
    let releaseProjects!: () => void;
    const projectsGate = new Promise<void>(resolve => { releaseProjects = resolve; });
    vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url === '/api/projects') {
        await projectsGate;
        return { ok: true, json: async () => fresh } as Response;
      }
      return { ok: true, json: async () => (url === '/api/prs' ? snapshot : procStatus) } as Response;
    }));
    vi.stubGlobal('EventSource', FakeEventSource);

    render(
      <Provider client={createClient()}>
        <Probe name="cached" enabled />
      </Provider>,
    );

    await waitFor(() => expect(screen.getByTestId('cached').textContent).toBe('1/1'));
    releaseProjects();
    await waitFor(() => expect(screen.getByTestId('cached').textContent).toBe('1/2'));
    expect(JSON.parse(localStorage.getItem(PROJECTS_CACHE_KEY) ?? 'null')).toEqual(fresh);
  });

  // The dashboard's Project type declares repos as a plain array and the projects
  // sidebar indexes it directly, so a null list has to be rejected here rather
  // than crashing a render deep in the tree.
  it('rejects a projects payload whose repos list is not an array', async () => {
    vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      const payload = url === '/api/prs'
        ? snapshot
        : url === '/api/projects'
          ? [{ name: 'time-kiosk', dir: '/work/time-kiosk', repos: null }]
          : procStatus;
      return { ok: true, json: async () => payload } as Response;
    }));
    vi.stubGlobal('EventSource', FakeEventSource);

    render(
      <Provider client={createClient()}>
        <Probe name="probe" enabled />
      </Provider>,
    );

    await waitFor(() => expect(screen.getByTestId('probe').dataset.projectError).toContain('invalid project'));
    expect(screen.getByTestId('probe').textContent).toBe('1/0');
  });

  // Process-status frames carry live CPU/memory samples, so a changed frame
  // lands every few seconds. They must re-render only the components showing
  // process state, never the app root that owns the streams.
  it('delivers process-status frames to useProcStatus without re-rendering the app root', async () => {
    vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      const payload = url === '/api/prs'
        ? snapshot
        : url === '/api/projects'
          ? [{ name: 'gavel', dir: '/work/gavel', repos: ['acme/gavel'] }]
          : procStatus;
      return { ok: true, json: async () => payload } as Response;
    }));
    vi.stubGlobal('EventSource', FakeEventSource);
    const rootRender = vi.fn();

    render(
      <Provider client={createClient()}>
        <Probe name="root" enabled onRender={rootRender} />
        <ProcProbe name="procs" />
      </Provider>,
    );
    await waitFor(() => expect(screen.getByTestId('root').textContent).toBe('1/1'));
    await waitFor(() => expect(screen.getByTestId('procs').textContent).toBe('gavel'));
    const settledRenders = rootRender.mock.calls.length;

    const frames = [
      { gavel: { hasProcfile: true, running: true }, widgets: { hasProcfile: true, running: false } },
      { gavel: { hasProcfile: true, running: false }, widgets: { hasProcfile: true, running: true } },
    ];
    const stream = FakeEventSource.instances.find(s => s.url === '/api/proc/status/stream');
    for (const frame of frames) act(() => stream?.emit('message', frame));

    await waitFor(() => expect(screen.getByTestId('procs').textContent).toBe('gavel,widgets'));
    expect(rootRender).toHaveBeenCalledTimes(settledRenders);
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
