import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { queryKeys } from '../query';
import type { ActivitySnapshot, CacheStatus } from '../types';
import { ActivityView } from './ActivityView';

const firstSnapshot: ActivitySnapshot = {
  entries: [],
  stats: { total: 13, cacheHits: 5, errors: 0, totalBytes: 2048, totalNs: 13_000_000, byKind: {} },
};

const cacheStatus: CacheStatus = {
  enabled: true,
  driver: 'postgres',
  dsnSource: 'config',
  dsnMasked: 'postgres://***',
  retentionSec: 3600,
  counts: { pull_requests: 5 },
};

class FakeEventSource {
  static instances: FakeEventSource[] = [];
  readonly listener = vi.fn();
  readonly close = vi.fn();
  onerror: ((event: Event) => void) | null = null;

  constructor(readonly url: string) {
    FakeEventSource.instances.push(this);
  }

  addEventListener(type: string, listener: EventListener) {
    if (type === 'message') this.listener.mockImplementation(listener);
  }

  emit(payload: unknown) {
    this.listener(new MessageEvent('message', { data: JSON.stringify(payload) }));
  }
}

function createClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false, refetchOnWindowFocus: false } } });
}

function renderActivity(client: QueryClient) {
  return render(
    <QueryClientProvider client={client}>
      <ActivityView />
    </QueryClientProvider>,
  );
}

function setVisibility(value: DocumentVisibilityState) {
  Object.defineProperty(document, 'visibilityState', { configurable: true, value });
  document.dispatchEvent(new Event('visibilitychange'));
}

afterEach(() => {
  Object.defineProperty(document, 'visibilityState', { configurable: true, value: 'visible' });
  FakeEventSource.instances = [];
  vi.useRealTimers();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('ActivityView queries', () => {
  it('owns bootstrap and stream snapshots in React Query and reuses them after remount', async () => {
    const calls: string[] = [];
    vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      calls.push(url);
      return {
        ok: true,
        json: async () => url === '/api/activity' ? firstSnapshot : cacheStatus,
      } as Response;
    }));
    vi.stubGlobal('EventSource', FakeEventSource);
    const client = createClient();

    const first = renderActivity(client);
    await screen.findByText('13', { selector: '.text-2xl' });
    expect(client.getQueryData(queryKeys.activityCache())).toEqual(cacheStatus);
    expect(calls.filter(url => url === '/api/activity')).toHaveLength(1);
    expect(calls.filter(url => url === '/api/activity/cache')).toHaveLength(1);

    const streamed: ActivitySnapshot = {
      entries: [],
      stats: { ...firstSnapshot.stats, total: 21 },
    };
    act(() => FakeEventSource.instances[0]?.emit(streamed));
    await screen.findByText('21', { selector: '.text-2xl' });
    expect(client.getQueryData(queryKeys.activity())).toEqual(streamed);

    act(() => FakeEventSource.instances[0]?.emit({ entries: 'invalid', stats: streamed.stats }));
    expect((await screen.findByRole('alert')).textContent).toContain('invalid update');
    expect(client.getQueryData(queryKeys.activity())).toEqual(streamed);

    first.unmount();
    renderActivity(client);
    await screen.findByText('21', { selector: '.text-2xl' });
    expect(calls.filter(url => url === '/api/activity')).toHaveLength(1);
    expect(calls.filter(url => url === '/api/activity/cache')).toHaveLength(1);
  });

  it('does not read or stream activity while the document is hidden', async () => {
    setVisibility('hidden');
    vi.stubGlobal('fetch', vi.fn());
    vi.stubGlobal('EventSource', FakeEventSource);

    renderActivity(createClient());

    await waitFor(() => expect(screen.getByText('HTTP Activity')).toBeTruthy());
    expect(fetch).not.toHaveBeenCalled();
    expect(FakeEventSource.instances).toHaveLength(0);
  });

  it('resets the activity cache through a mutation and surfaces contextual failures', async () => {
    let resetFails = true;
    vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url === '/api/activity') {
        return new Response(JSON.stringify(firstSnapshot), { status: 200, headers: { 'Content-Type': 'application/json' } });
      }
      if (url === '/api/activity/cache') {
        return new Response(JSON.stringify(cacheStatus), { status: 200, headers: { 'Content-Type': 'application/json' } });
      }
      if (resetFails) {
        return new Response(JSON.stringify({ error: 'recorder unavailable' }), { status: 503, headers: { 'Content-Type': 'application/json' } });
      }
      return new Response(JSON.stringify({ status: 'reset' }), { status: 200, headers: { 'Content-Type': 'application/json' } });
    }));
    vi.stubGlobal('EventSource', FakeEventSource);
    const client = createClient();
    renderActivity(client);
    await screen.findByText('13', { selector: '.text-2xl' });

    fireEvent.click(screen.getByRole('button', { name: 'Reset' }));
    expect((await screen.findByRole('alert')).textContent).toContain('Reset HTTP activity: recorder unavailable');
    expect(client.getQueryData(queryKeys.activity())).toEqual(firstSnapshot);

    resetFails = false;
    fireEvent.click(screen.getByRole('button', { name: 'Reset' }));
    await screen.findByText('0', { selector: '.text-2xl' });
    expect(client.getQueryData(queryKeys.activity())).toEqual({
      entries: [],
      stats: { total: 0, cacheHits: 0, errors: 0, totalBytes: 0, totalNs: 0, byKind: {} },
    });
  });
});
