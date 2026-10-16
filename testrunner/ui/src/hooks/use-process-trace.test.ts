import { afterEach, describe, expect, it, vi } from 'vitest';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { GoProfileRecord, StackRecord } from '@flanksource/clicky-ui/data';
import { PROCESS_TRACE_ENDED, SAMPLE_WINDOW, useProcessTrace } from './use-process-trace';
import type { ProcessSample, ProcessTraceNode, TracesStatus } from '../types';

const root: ProcessTraceNode = {
  key: '4242@1', pid: 4242, ppid: 1, name: 'gavel', command: 'gavel test', status: 'sleeping', isRoot: true,
  startedAt: '2026-01-01T00:00:00Z', lastSeenAt: '2026-01-01T00:00:09Z',
  cpuPercent: null, rssBytes: 0, peakCpuPercent: null, peakRssBytes: 0,
  stackTransports: [{ transport: 'inprocess', destructive: false }],
};
const child: ProcessTraceNode = {
  key: '4300@2', parentKey: root.key, pid: 4300, ppid: 4242, name: 'pkg.test', command: 'pkg.test -test.v', status: 'active',
  startedAt: '2026-01-01T00:00:02Z', lastSeenAt: '2026-01-01T00:00:09Z',
  cpuPercent: 12.5, rssBytes: 1024, peakCpuPercent: 12.5, peakRssBytes: 1024,
  stackTransports: [{ transport: 'native', destructive: false }, { transport: 'sigquit', destructive: true }],
};

const sample = (key: string, tick: number, extra: Partial<ProcessSample> = {}): ProcessSample => ({
  sampledAt: `2026-01-01T00:00:0${tick}Z`, tick, key, pid: Number(key.split('@')[0]), memoryPercent: 1, rssBytes: 2048, cpuPercent: 5, ...extra,
});
const stack = (key: string, checkpointId: string, extra: Partial<StackRecord> = {}): StackRecord => ({
  checkpointId, takenAt: '2026-01-01T00:00:05Z', reason: 'manual', transport: 'native', key, pid: Number(key.split('@')[0]),
  format: 'native', raw: 'Thread 1:\n  poll()\n', ...extra,
});
const profile = (key: string): GoProfileRecord => ({
  checkpointId: 'p1', takenAt: '2026-01-01T00:00:06Z', reason: 'manual', source: 'pprof', key, pid: 4300, profile: 'cpu',
  sampleType: 'cpu', unit: 'nanoseconds', value: 10, frames: [{ func: 'main.main', file: 'main.go', line: 3 }], stackHash: 'h1',
});

const ended: TracesStatus = {
  runs: { 'run-1': { mount: '/query/run-1' } },
  process_run: 'run-1',
  process: {
    tree: { kind: 'process_tree', stream: 'tree-1' },
    samples: { kind: 'process_samples', stream: 'samples-1' },
    stacks: { kind: 'stacks', stream: 'stacks-1' },
  },
};
const withGoprof: TracesStatus = {
  ...ended,
  process: { ...ended.process!, goprof: { kind: 'goprof', stream: 'goprof-1' } },
};
const running: TracesStatus = {
  ...ended,
  sessions: { tree: 'tree-session', samples: 'samples-session', stacks: 'stacks-session' },
};

const treeExport = '/query/run-1/api/v1/traces/process_tree/export?stream=tree-1&format=ndjson';
const samplesProfile = '/query/run-1/api/v1/profile/profile-traces-process-samples';
const samplesExport = '/query/run-1/api/v1/traces/process_samples/export?stream=samples-1&format=ndjson';
const stacksExport = '/query/run-1/api/v1/traces/stacks/export?stream=stacks-1&format=ndjson';
const goprofExport = '/query/run-1/api/v1/traces/goprof/export?stream=goprof-1&format=ndjson';

function ndjson(...records: unknown[]): Response {
  return new Response(records.map(record => JSON.stringify(record)).join('\n') + '\n');
}

interface Store {
  tree: unknown[];
  samples: unknown[];
  stacks: unknown[];
  goprof: unknown[];
}

// serve answers the query mount from store, recording every URL asked and every POST made.
function serve(store: Store, onPost?: (url: string, body: unknown) => Response | void) {
  const urls: string[] = [];
  const mock = vi.fn(async (url: string, init?: RequestInit) => {
    urls.push(url);
    if (init?.method === 'POST') {
      const response = onPost?.(url, JSON.parse(init.body as string));
      return response ?? new Response('{}');
    }
    if (url.startsWith(samplesProfile)) return new Response(JSON.stringify(store.samples.length ? [{ seq: store.samples.length }] : []));
    if (url === treeExport) return ndjson(...store.tree);
    if (url === samplesExport) return ndjson(...store.samples);
    if (url === stacksExport) return ndjson(...store.stacks);
    if (url === goprofExport) return ndjson(...store.goprof);
    throw new Error(`unexpected ${url}`);
  });
  vi.stubGlobal('fetch', mock);
  return { mock, urls };
}

const empty = (): Store => ({ tree: [], samples: [], stacks: [], goprof: [] });

describe('useProcessTrace', () => {
  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it('reads the tree and samples once through the query mount for a finished run, and nothing while the store is not served', async () => {
    const { mock, urls } = serve({ ...empty(), tree: [root], samples: [sample(root.key, 1)] });

    const unserved = renderHook(() => useProcessTrace({ traces: { runs: { 'run-1': { unavailable: 'pruned' } }, process_run: 'run-1', process: ended.process }, done: true, base: '' }));
    expect(mock).not.toHaveBeenCalled();
    expect(unserved.result.current.loading).toBe(false);
    unserved.unmount();

    const { result } = renderHook(() => useProcessTrace({ traces: ended, done: true, base: '' }));
    expect(result.current.loading).toBe(true);
    await waitFor(() => expect(result.current.samples).toHaveLength(1));
    expect(result.current.nodes.map(node => node.key)).toEqual([root.key]);
    expect(result.current.loading).toBe(false);
    expect(urls.filter(url => !url.startsWith(samplesProfile)).sort()).toEqual([samplesExport, treeExport].sort());
  });

  it('reads the mount under the base of the page and run it is served from', async () => {
    const base = '/results/o/r/7/run-9';
    const mock = vi.fn(async (url: string) => {
      if (url === `${base}${treeExport}`) return ndjson(root);
      if (url.startsWith(`${base}${samplesProfile}`)) return new Response('[]');
      throw new Error(`unexpected ${url}`);
    });
    vi.stubGlobal('fetch', mock);

    const { result } = renderHook(() => useProcessTrace({ traces: ended, done: true, base }));
    await waitFor(() => expect(result.current.nodes).toHaveLength(1));
  });

  it('reads a stream that has committed nothing yet as an empty tree', async () => {
    vi.stubGlobal('fetch', vi.fn(async (url: string) => (url.startsWith(samplesProfile)
      ? new Response('[]')
      : new Response('stream "tree-1" not found', { status: 404 }))));

    const { result } = renderHook(() => useProcessTrace({ traces: ended, done: true, base: '' }));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.nodes).toEqual([]);
    expect(result.current.error).toBeUndefined();
  });

  it('reads the stacks and goprof windows only for the selected process, goprof only when the run has one', async () => {
    const { urls } = serve({
      ...empty(), tree: [root, child], stacks: [stack(child.key, 'c1')], goprof: [profile(child.key)],
    });

    const idle = renderHook(() => useProcessTrace({ traces: withGoprof, done: true, base: '' }));
    await waitFor(() => expect(idle.result.current.nodes).toHaveLength(2));
    expect(urls).not.toContain(stacksExport);
    expect(urls).not.toContain(goprofExport);
    idle.unmount();

    const selected = renderHook(() => useProcessTrace({ traces: withGoprof, done: true, base: '', selectedKey: child.key }));
    await waitFor(() => expect(selected.result.current.stacks).toHaveLength(1));
    expect(selected.result.current.stacks[0]?.checkpointId).toBe('c1');
    await waitFor(() => expect(selected.result.current.goprof).toHaveLength(1));
    expect(selected.result.current.goprofExportUrl).toBe('/query/run-1/api/v1/traces/goprof/export?stream=goprof-1&format=pprof');
    selected.unmount();

    const urlsWithoutGoprof = serve({ ...empty(), tree: [root], stacks: [stack(root.key, 'c9')] }).urls;
    const plain = renderHook(() => useProcessTrace({ traces: ended, done: true, base: '', selectedKey: root.key }));
    await waitFor(() => expect(plain.result.current.stacks).toHaveLength(1));
    expect(plain.result.current.goprof).toEqual([]);
    expect(urlsWithoutGoprof).not.toContain(goprofExport);
  });

  it('reads only the latest SAMPLE_WINDOW samples of the stream', async () => {
    const urls: string[] = [];
    vi.stubGlobal('fetch', vi.fn(async (url: string) => {
      urls.push(url);
      if (url === treeExport) return ndjson(root);
      if (url.startsWith(samplesProfile)) return new Response(JSON.stringify([{ seq: SAMPLE_WINDOW + 500 }]));
      return ndjson(sample(root.key, 1));
    }));

    const { result } = renderHook(() => useProcessTrace({ traces: ended, done: true, base: '' }));
    await waitFor(() => expect(result.current.samples).toHaveLength(1));

    expect(urls).toContain(`${samplesExport.replace('&format', '&afterSeq=500&format')}`);
  });

  it('keeps polling the tree, samples and the selected stacks while the run is live', async () => {
    vi.useFakeTimers();
    const store: Store = { ...empty(), tree: [root, child], samples: [sample(child.key, 1)], stacks: [] };
    serve(store);

    const { result } = renderHook(() => useProcessTrace({ traces: running, done: false, base: '', selectedKey: child.key }));
    await act(async () => { await vi.advanceTimersByTimeAsync(10); });
    expect(result.current.samples).toHaveLength(1);
    expect(result.current.stacks).toHaveLength(0);

    store.samples = [sample(child.key, 1), sample(child.key, 2)];
    store.stacks = [stack(child.key, 'c1')];
    await act(async () => { await vi.advanceTimersByTimeAsync(1100); });

    expect(result.current.samples).toHaveLength(2);
    expect(result.current.stacks.map(record => record.checkpointId)).toEqual(['c1']);
  });

  it('checkpoints stacks of one node by its key and transport, then reads the new dump', async () => {
    const store: Store = { ...empty(), tree: [root, child] };
    let posted: { url: string; body: unknown } | undefined;
    serve(store, (url, body) => {
      posted = { url, body };
      store.stacks = [stack(child.key, 'c2', { reason: 'manual', transport: 'native' })];
    });

    const { result } = renderHook(() => useProcessTrace({ traces: running, done: false, base: '', selectedKey: child.key }));
    await waitFor(() => expect(result.current.nodes).toHaveLength(2));
    expect(result.current.canCollect).toBe(true);
    expect(result.current.stacks).toEqual([]);

    await act(async () => { await result.current.collectStack(child.key, 'native'); });

    expect(posted).toEqual({
      url: '/query/run-1/api/v1/sessions/stacks-session/checkpoint',
      body: { key: child.key, transport: 'native', reason: 'manual' },
    });
    expect(result.current.stacks.map(record => record.checkpointId)).toEqual(['c2']);
    expect(result.current.collecting).toBe(false);
    expect(result.current.collectError).toBeUndefined();
  });

  it('forces a full facet reading of one node through the samples session, then re-reads the samples', async () => {
    const store: Store = { ...empty(), tree: [root, child], samples: [sample(child.key, 1)] };
    let posted: { url: string; body: unknown } | undefined;
    serve(store, (url, body) => {
      posted = { url, body };
      store.samples = [sample(child.key, 1), sample(child.key, 2, { openFiles: 3, openFileList: [{ fd: '3', type: 'REG', name: '/tmp/x' }] })];
    });

    const { result } = renderHook(() => useProcessTrace({ traces: running, done: false, base: '', selectedKey: child.key }));
    await waitFor(() => expect(result.current.samples).toHaveLength(1));

    await act(async () => { await result.current.refreshFacets(child.key); });

    expect(posted).toEqual({ url: '/query/run-1/api/v1/sessions/samples-session/checkpoint', body: { key: child.key } });
    expect(result.current.samples).toHaveLength(2);
    expect(result.current.refreshing).toBe(false);
  });

  it('offers no capture once the run has ended, and refuses a checkpoint', async () => {
    const { urls } = serve({ ...empty(), tree: [root, child] });

    const finished = renderHook(() => useProcessTrace({ traces: ended, done: true, base: '' }));
    await waitFor(() => expect(finished.result.current.nodes).toHaveLength(2));
    expect(finished.result.current.canCollect).toBe(false);
    await act(async () => { await finished.result.current.collectStack(root.key, 'inprocess'); });
    expect(finished.result.current.collectError).toBe(PROCESS_TRACE_ENDED);
    await act(async () => { await finished.result.current.refreshFacets(root.key); });
    expect(finished.result.current.refreshError).toBe(PROCESS_TRACE_ENDED);
    expect(urls.filter(url => url.includes('/sessions/'))).toEqual([]);
    finished.unmount();

    // A stored run served by a dashboard that never ran it can still carry session ids; done keeps them from being offered.
    const stored = renderHook(() => useProcessTrace({ traces: running, done: true, base: '' }));
    await waitFor(() => expect(stored.result.current.nodes).toHaveLength(2));
    expect(stored.result.current.canCollect).toBe(false);
  });

  it('reports a checkpoint the server refused, with the server\'s own text', async () => {
    serve({ ...empty(), tree: [root, child] }, () => new Response('transport sigquit is not available for pid 4300', { status: 409 }));

    const { result } = renderHook(() => useProcessTrace({ traces: running, done: false, base: '', selectedKey: child.key }));
    await waitFor(() => expect(result.current.nodes).toHaveLength(2));
    await act(async () => { await result.current.collectStack(child.key, 'sigquit'); });

    expect(result.current.collectError).toBe('checkpoint failed (409): transport sigquit is not available for pid 4300');
    expect(result.current.collecting).toBe(false);
  });
});
