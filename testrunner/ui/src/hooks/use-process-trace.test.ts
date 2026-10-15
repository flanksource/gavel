import { afterEach, describe, expect, it, vi } from 'vitest';
import { act, renderHook, waitFor } from '@testing-library/react';
import { PROCESS_TRACE_ENDED, useProcessTrace } from './use-process-trace';
import type { ProcessTraceNode, TracesStatus } from '../types';

const root: ProcessTraceNode = {
  key: '4242@1', pid: 4242, ppid: 1, name: 'gavel', command: 'gavel test', status: 'sleeping', isRoot: true,
  startedAt: '2026-01-01T00:00:00Z', lastSeenAt: '2026-01-01T00:00:09Z',
  cpuPercent: 0, rssBytes: 0, peakCpuPercent: 0, peakRssBytes: 0,
};

const treeExport = '/query/run-1/api/v1/traces/process_tree/export?stream=tree-1&format=ndjson';

const ended: TracesStatus = {
  runs: { 'run-1': { mount: '/query/run-1' } },
  process_run: 'run-1',
  process: { tree: { kind: 'process_tree', stream: 'tree-1' }, samples: { kind: 'process_samples', stream: 'samples-1' } },
};
const running: TracesStatus = { ...ended, process_session: 'session-1' };

function ndjson(...records: unknown[]): Response {
  return new Response(records.map(record => JSON.stringify(record)).join('\n') + '\n');
}

describe('useProcessTrace', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('reads the tree once through the query mount for a finished run, and nothing while the store is not served', async () => {
    const fetchMock = vi.fn(async (url: string) => {
      if (url === treeExport) return ndjson(root);
      throw new Error(`unexpected ${url}`);
    });
    vi.stubGlobal('fetch', fetchMock);

    const unserved = renderHook(() => useProcessTrace({ traces: { runs: { 'run-1': { unavailable: 'pruned' } }, process_run: 'run-1', process: ended.process }, done: true }));
    expect(fetchMock).not.toHaveBeenCalled();
    unserved.unmount();

    const { result } = renderHook(() => useProcessTrace({ traces: ended, done: true }));
    await waitFor(() => expect(result.current.nodes).toEqual([root]));
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it('reads a stream that has committed nothing yet as an empty tree', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response('stream "tree-1" not found', { status: 404 })));

    const { result } = renderHook(() => useProcessTrace({ traces: ended, done: true }));

    await waitFor(() => expect(result.current.nodes).toEqual([]));
    expect(result.current.error).toBeUndefined();
  });

  it('refuses a checkpoint once the run has ended', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => ndjson(root)));

    const { result } = renderHook(() => useProcessTrace({ traces: ended, done: true }));
    await act(async () => { await result.current.collectStack(); });

    expect(result.current.collectMessage).toBe(PROCESS_TRACE_ENDED);
  });

  it('checkpoints the running capture by its session, then shows the stack it dumped onto the root', async () => {
    let checkpointed = false;
    vi.stubGlobal('fetch', vi.fn(async (url: string, init?: RequestInit) => {
      if (init?.method === 'POST') {
        expect(url).toBe('/query/run-1/api/v1/sessions/session-1/checkpoint');
        checkpointed = true;
        return new Response('{}');
      }
      return ndjson(checkpointed ? { ...root, stack: 'goroutine 1 [running]:\nmain.first()\n' } : root);
    }));

    const { result } = renderHook(() => useProcessTrace({ traces: running, done: true }));
    await waitFor(() => expect(result.current.nodes).toHaveLength(1));
    expect(result.current.stack).toBeUndefined();
    await act(async () => { await result.current.collectStack(); });

    expect(result.current.stack?.text).toContain('main.first()');
    expect(result.current.collecting).toBe(false);
  });
});
