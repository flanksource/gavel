import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { RunSnapshot } from './types';
import { TestRunDetail } from './TestRunDetail';

const runResultsCalls = vi.hoisted(() => ({ props: [] as { snapshot: RunSnapshot; runKey: string }[] }));
const snapshotFetch = vi.hoisted(() => ({ fn: vi.fn() }));

vi.mock('./TestRunResults', () => ({
  TestRunResults: (props: { snapshot: RunSnapshot; runKey: string }) => {
    runResultsCalls.props.push(props);
    return <div data-testid="test-run-results">{props.runKey}</div>;
  },
}));

const snapshot = {
  metadata: { kind: 'test' },
  status: { running: false },
  tests: [],
} as unknown as RunSnapshot;

function createClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false, refetchOnWindowFocus: false, gcTime: Infinity } },
  });
}

function snapshotResponse(): Response {
  return new Response(JSON.stringify(snapshot), {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  });
}

function renderDetail(client = createClient(), project = 'gavel', runId = 'run-1') {
  return render(
    <QueryClientProvider client={client}>
      <TestRunDetail project={project} projectDir="/workspace" runId={runId} />
    </QueryClientProvider>,
  );
}

describe('TestRunDetail', () => {
  beforeEach(() => {
    runResultsCalls.props.length = 0;
    snapshotFetch.fn = vi.fn(async () => snapshotResponse());
    vi.stubGlobal('fetch', snapshotFetch.fn);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('loads the project run through the shared snapshot query', async () => {
    renderDetail();

    await waitFor(() => expect(screen.getByTestId('test-run-results')).toBeTruthy());
    expect(snapshotFetch.fn).toHaveBeenCalledWith(
      '/api/tests/run?runId=run-1&project=gavel',
      expect.objectContaining({ signal: expect.anything() }),
    );
    expect(runResultsCalls.props.at(-1)?.snapshot).toEqual(snapshot);
    expect(runResultsCalls.props.at(-1)?.runKey).toBe('run-1');
  });

  it('reuses the fresh snapshot when the same run is reopened', async () => {
    const client = createClient();
    const first = renderDetail(client);
    await waitFor(() => expect(screen.getByTestId('test-run-results')).toBeTruthy());
    first.unmount();

    renderDetail(client);
    await waitFor(() => expect(screen.getByTestId('test-run-results')).toBeTruthy());

    expect(snapshotFetch.fn).toHaveBeenCalledTimes(1);
  });

  it('does not fetch until the project run identity is complete', async () => {
    renderDetail(createClient(), 'gavel', '');

    await act(async () => undefined);
    expect(snapshotFetch.fn).not.toHaveBeenCalled();
  });

  it('cancels its snapshot request when the run detail unmounts', async () => {
    const request: { signal?: AbortSignal } = {};
    snapshotFetch.fn = vi.fn((_input: RequestInfo | URL, init?: RequestInit) => new Promise<Response>((_resolve, reject) => {
      const signal = init?.signal;
      if (!signal) throw new Error('snapshot request did not receive an AbortSignal');
      request.signal = signal;
      signal.addEventListener('abort', () => reject(signal.reason), { once: true });
    }));
    vi.stubGlobal('fetch', snapshotFetch.fn);
    const view = renderDetail();
    await waitFor(() => expect(snapshotFetch.fn).toHaveBeenCalledTimes(1));

    view.unmount();

    expect(request.signal?.aborted).toBe(true);
  });
});
