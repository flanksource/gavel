import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, render, screen, waitFor } from '@testing-library/react';
import type { PropsWithChildren } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { queryKeys } from './query';
import type { PRItem } from './types';
import { usePRDetailStream } from './usePRDetailStream';

const firstPR: PRItem = {
  number: 7,
  title: 'Cache detail frames',
  author: 'octocat',
  repo: 'acme/gavel',
  source: 'detail-cache',
  target: 'main',
  state: 'OPEN',
  isDraft: false,
  url: 'https://example.com/acme/gavel/pull/7',
  updatedAt: '2026-08-02T09:00:00Z',
};
const secondPR = { ...firstPR, number: 8, title: 'Isolate detail frames' };

class FakeEventSource {
  static instances: FakeEventSource[] = [];
  readonly listeners = new Map<string, EventListener>();
  readonly close = vi.fn();
  onerror: ((event: Event) => void) | null = null;

  constructor(readonly url: string) {
    FakeEventSource.instances.push(this);
  }

  addEventListener(type: string, listener: EventListener) {
    this.listeners.set(type, listener);
  }

  emit(type: string, data: unknown) {
    this.listeners.get(type)?.(new MessageEvent(type, { data: JSON.stringify(data) }));
  }

  fail() {
    this.onerror?.(new Event('error'));
  }
}

function Probe({ selected }: { selected: PRItem | null }) {
  const stream = usePRDetailStream(selected);
  return (
    <>
      <div data-testid="detail" data-loading={stream.loading}>
        {stream.detail?.pr?.title || stream.detail?.error || 'empty'}
      </div>
      <button onClick={stream.refresh}>Refresh detail</button>
    </>
  );
}

function Provider({ client, children }: PropsWithChildren<{ client: QueryClient }>) {
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

afterEach(() => {
  FakeEventSource.instances = [];
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('usePRDetailStream', () => {
  it('stores progressive frames under the parameterized PR key and closes replaced and terminal streams', async () => {
    vi.stubGlobal('EventSource', FakeEventSource);
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const view = render(
      <Provider client={client}>
        <Probe selected={firstPR} />
      </Provider>,
    );
    await waitFor(() => expect(FakeEventSource.instances).toHaveLength(1));
    const firstStream = FakeEventSource.instances[0];

    act(() => firstStream.emit('pr', { pr: { title: 'First detail' }, comments: [{ id: 1 }] }));
    act(() => firstStream.emit('runs', { runs: { '42': { id: 42 } } }));
    expect(await screen.findByText('First detail')).toBeTruthy();
    expect(client.getQueryData(queryKeys.prDetail(firstPR.repo, firstPR.number))).toMatchObject({
      pr: { title: 'First detail' },
      comments: [{ id: 1 }],
      runs: { '42': { id: 42 } },
    });

    view.rerender(
      <Provider client={client}>
        <Probe selected={secondPR} />
      </Provider>,
    );
    await waitFor(() => expect(FakeEventSource.instances).toHaveLength(2));
    expect(firstStream.close).toHaveBeenCalledTimes(1);
    expect(client.getQueryData(queryKeys.prDetail(secondPR.repo, secondPR.number))).toBeUndefined();

    const secondStream = FakeEventSource.instances[1];
    act(() => secondStream.emit('done', null));
    expect(secondStream.close).toHaveBeenCalledTimes(1);
    expect(screen.getByTestId('detail').dataset.loading).toBe('false');

    act(() => screen.getByRole('button', { name: 'Refresh detail' }).click());
    await waitFor(() => expect(FakeEventSource.instances).toHaveLength(3));
  });

  it('accepts a PR frame with no comments and stores an empty list', async () => {
    vi.stubGlobal('EventSource', FakeEventSource);
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <Provider client={client}>
        <Probe selected={firstPR} />
      </Provider>,
    );
    await waitFor(() => expect(FakeEventSource.instances).toHaveLength(1));

    act(() => FakeEventSource.instances[0].emit('pr', { pr: { title: 'No comments yet' }, comments: null }));

    expect(await screen.findByText('No comments yet')).toBeTruthy();
    expect(client.getQueryData(queryKeys.prDetail(firstPR.repo, firstPR.number))).toMatchObject({
      pr: { title: 'No comments yet' },
      comments: [],
    });
    expect(screen.getByTestId('detail').dataset.loading).toBe('false');
  });

  it('rejects a PR frame whose comments field is not a list', async () => {
    vi.stubGlobal('EventSource', FakeEventSource);
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <Provider client={client}>
        <Probe selected={firstPR} />
      </Provider>,
    );
    await waitFor(() => expect(FakeEventSource.instances).toHaveLength(1));

    act(() => FakeEventSource.instances[0].emit('pr', { pr: { title: 'Broken' }, comments: 'nope' }));

    expect(await screen.findByText('Pull request detail stream received an invalid PR update.')).toBeTruthy();
  });

  it('surfaces a transport failure in the selected cache entry and closes the stream', async () => {
    vi.stubGlobal('EventSource', FakeEventSource);
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <Provider client={client}>
        <Probe selected={firstPR} />
      </Provider>,
    );
    await waitFor(() => expect(FakeEventSource.instances).toHaveLength(1));

    act(() => FakeEventSource.instances[0].fail());

    expect(await screen.findByText('Connection lost')).toBeTruthy();
    expect(client.getQueryData(queryKeys.prDetail(firstPR.repo, firstPR.number))).toEqual({ error: 'Connection lost' });
    expect(FakeEventSource.instances[0].close).toHaveBeenCalledTimes(1);
  });
});
