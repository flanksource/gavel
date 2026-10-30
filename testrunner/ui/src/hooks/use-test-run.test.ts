import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { LinterResult, Snapshot, Test } from '../types';
import { deriveStatusText, useTestRun } from './use-test-run';

// A controllable fake EventSource so tests can drive `message`/`done`/`error`
// frames deterministically without a real network connection.
class FakeEventSource {
  static instances: FakeEventSource[] = [];
  url: string;
  closed = false;
  onerror: ((ev: unknown) => void) | null = null;
  private listeners: Record<string, ((ev: MessageEvent) => void)[]> = {};

  constructor(url: string) {
    this.url = url;
    FakeEventSource.instances.push(this);
  }
  addEventListener(type: string, fn: (ev: MessageEvent) => void) {
    (this.listeners[type] ??= []).push(fn);
  }
  removeEventListener(type: string, fn: (ev: MessageEvent) => void) {
    this.listeners[type] = (this.listeners[type] ?? []).filter((listener) => listener !== fn);
  }
  emit(type: string, data?: unknown) {
    const ev = { data: data === undefined ? undefined : JSON.stringify(data) } as MessageEvent;
    for (const fn of this.listeners[type] ?? []) fn(ev);
  }
  error() {
    this.onerror?.({});
  }
  close() {
    this.closed = true;
  }
  static last(): FakeEventSource {
    const es = FakeEventSource.instances.at(-1);
    if (!es) throw new Error('no EventSource constructed');
    return es;
  }
  static reset() {
    FakeEventSource.instances = [];
  }
}

function snapshot(over: Partial<Snapshot> = {}): Snapshot {
  return {
    status: { running: true },
    tests: [],
    ...over,
  };
}

const runningSnap = snapshot({
  status: { running: true },
  tests: [{ name: 'TestFoo', passed: true }],
  metadata: { sequence: 1, kind: 'initial', started: '2026-06-02T10:00:00Z' },
});

const doneSnap = snapshot({
  status: { running: false },
  tests: [{ name: 'TestFoo', passed: true }],
  metadata: { sequence: 1, kind: 'initial', ended: '2026-06-02T10:00:05Z' },
});

describe('useTestRun (SSE transport)', () => {
  beforeEach(() => {
    FakeEventSource.reset();
    vi.stubGlobal('EventSource', FakeEventSource as unknown as typeof EventSource);
    vi.stubGlobal('fetch', vi.fn(async () => ({ ok: true, json: async () => snapshot() })));
  });
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it('opens the stream at <baseUrl>/api/tests/stream and seeds via /api/tests', async () => {
    renderHook(() => useTestRun({ baseUrl: '/results/x' }));
    await waitFor(() => expect(FakeEventSource.instances.length).toBe(1));
    expect(FakeEventSource.last().url).toBe('/results/x/api/tests/stream');
    expect(fetch).toHaveBeenCalledWith('/results/x/api/tests', expect.anything());
  });

  it('scopes the stream under <baseUrl>/<runId> when runId is set', async () => {
    renderHook(() => useTestRun({ baseUrl: '/live', runId: 'run a/b' }));
    await waitFor(() => expect(FakeEventSource.instances.length).toBe(1));
    expect(FakeEventSource.last().url).toBe('/live/run%20a%2Fb/api/tests/stream');
    expect(fetch).toHaveBeenCalledWith('/live/run%20a%2Fb/api/tests', expect.anything());
  });

  it('exposes tests and status from streamed message frames', async () => {
    const { result } = renderHook(() => useTestRun());
    await waitFor(() => expect(FakeEventSource.instances.length).toBe(1));
    act(() => FakeEventSource.last().emit('message', runningSnap));
    await waitFor(() => expect(result.current.tests).toHaveLength(1));
    expect(result.current.tests[0]?.name).toBe('TestFoo');
    expect(result.current.status.running).toBe(true);
    // statusText delegates to isLintOnlyPhase (covered in utils.test.ts); here
    // we only assert it reflects an in-flight run, not the precise phase.
    expect(result.current.statusText).toMatch(/^Running /);
    expect(result.current.done).toBe(false);
  });

  it('says a fixture benchmark or profile run is running fixtures, not linters', async () => {
    const { result } = renderHook(() => useTestRun());
    await waitFor(() => expect(FakeEventSource.instances.length).toBe(1));
    act(() => FakeEventSource.last().emit('message', snapshot({
      status: { running: true },
      performance: { mode: 'benchmark', status: 'running', duration_ms: 0 },
    })));
    await waitFor(() => expect(result.current.statusText).toBe('Running fixtures...'));
  });

  it('marks done and closes the stream on a non-running snapshot', async () => {
    const { result } = renderHook(() => useTestRun());
    await waitFor(() => expect(FakeEventSource.instances.length).toBe(1));
    act(() => FakeEventSource.last().emit('message', doneSnap));
    await waitFor(() => expect(result.current.done).toBe(true));
    expect(result.current.statusText).toBe('Test run complete');
    expect(FakeEventSource.last().closed).toBe(true);
  });

  it('marks done on an explicit done event', async () => {
    const { result } = renderHook(() => useTestRun());
    await waitFor(() => expect(FakeEventSource.instances.length).toBe(1));
    act(() => FakeEventSource.last().emit('done'));
    await waitFor(() => expect(result.current.done).toBe(true));
    expect(FakeEventSource.last().closed).toBe(true);
  });

  it('reports a connection-lost status on error only while running', async () => {
    const { result } = renderHook(() => useTestRun());
    await waitFor(() => expect(FakeEventSource.instances.length).toBe(1));
    act(() => FakeEventSource.last().error());
    await waitFor(() => expect(result.current.error).toMatch(/Connection lost/i));
  });

  it('rerun() POSTs the body to <baseUrl>/api/rerun', async () => {
    const { result } = renderHook(() => useTestRun({ baseUrl: '/b' }));
    await waitFor(() => expect(FakeEventSource.instances.length).toBe(1));
    await act(async () => {
      await result.current.rerun({ package_paths: ['./pkg'], framework: 'go' });
    });
    expect(fetch).toHaveBeenCalledWith(
      '/b/api/rerun',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ package_paths: ['./pkg'], framework: 'go' }),
      }),
    );
  });

  it('stop() with a taskId scopes to that task, without one scopes global', async () => {
    const { result } = renderHook(() => useTestRun());
    await waitFor(() => expect(FakeEventSource.instances.length).toBe(1));
    await act(async () => {
      await result.current.stop('task-7');
    });
    expect(fetch).toHaveBeenCalledWith(
      '/api/stop',
      expect.objectContaining({ body: JSON.stringify({ scope: 'task', task_id: 'task-7' }) }),
    );
    await act(async () => {
      await result.current.stop();
    });
    expect(fetch).toHaveBeenCalledWith(
      '/api/stop',
      expect.objectContaining({ body: JSON.stringify({ scope: 'global' }) }),
    );
  });

  it('opens the stream through an injected createEventSource factory and closes it on unmount', async () => {
    const closeSpy = vi.fn();
    const injected = {
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      close: closeSpy,
      onerror: null,
    };
    const createEventSource = vi.fn(() => injected);

    const { unmount } = renderHook(() => useTestRun({ baseUrl: '/live', createEventSource }));
    await waitFor(() => expect(createEventSource).toHaveBeenCalledWith('/live/api/tests/stream'));
    // The native EventSource constructor must never be reached when a factory
    // is injected — only the factory opens the connection.
    expect(FakeEventSource.instances).toHaveLength(0);

    unmount();
    expect(closeSpy).toHaveBeenCalledTimes(1);
  });

  it('editTest() POSTs the body to <baseUrl>/api/tests/edit', async () => {
    const { result } = renderHook(() => useTestRun({ baseUrl: '/b' }));
    await waitFor(() => expect(FakeEventSource.instances.length).toBe(1));
    await act(async () => {
      await result.current.editTest({
        action: 'skip',
        scope: 'test',
        framework: 'vitest',
        file: 'sum.test.ts',
        test_name: 'works',
      });
    });
    expect(fetch).toHaveBeenCalledWith(
      '/b/api/tests/edit',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({
          action: 'skip',
          scope: 'test',
          framework: 'vitest',
          file: 'sum.test.ts',
          test_name: 'works',
        }),
      }),
    );
  });
});

describe('useTestRun (polling fallback)', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.stubGlobal('fetch', vi.fn(async () => ({ ok: true, json: async () => runningSnap })));
  });
  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it('polls /api/tests on the configured interval when forcePoll is set', async () => {
    renderHook(() => useTestRun({ forcePoll: true, pollMs: 500 }));
    await vi.advanceTimersByTimeAsync(0);
    expect(fetch).toHaveBeenCalledWith('/api/tests', expect.anything());
    const callsAfterFirst = (fetch as ReturnType<typeof vi.fn>).mock.calls.length;
    await vi.advanceTimersByTimeAsync(500);
    expect((fetch as ReturnType<typeof vi.fn>).mock.calls.length).toBeGreaterThan(callsAfterFirst);
  });

  it('stops polling once a non-running snapshot arrives', async () => {
    (fetch as ReturnType<typeof vi.fn>).mockResolvedValue({ ok: true, json: async () => doneSnap });
    renderHook(() => useTestRun({ forcePoll: true, pollMs: 500 }));
    await vi.advanceTimersByTimeAsync(0);
    const calls = (fetch as ReturnType<typeof vi.fn>).mock.calls.length;
    await vi.advanceTimersByTimeAsync(2000);
    expect((fetch as ReturnType<typeof vi.fn>).mock.calls.length).toBe(calls);
  });

  it('aborts the active request and clears the polling wait on unmount', async () => {
    const fetchMock = vi.mocked(fetch);
    const { unmount } = renderHook(() => useTestRun({ forcePoll: true, pollMs: 500 }));
    await vi.advanceTimersByTimeAsync(0);
    const signal = fetchMock.mock.calls[0]?.[1]?.signal;
    expect(signal?.aborted).toBe(false);
    unmount();
    expect(signal?.aborted).toBe(true);
    await vi.advanceTimersByTimeAsync(2000);
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it('ignores a previous run response whose JSON finishes after the run changes', async () => {
    let resolvePrevious!: (value: Snapshot) => void;
    const previous = new Promise<Snapshot>((resolve) => { resolvePrevious = resolve; });
    vi.mocked(fetch).mockResolvedValueOnce({ ok: true, json: () => previous } as Response);
    const { result, rerender } = renderHook(({ runId }) => useTestRun({ forcePoll: true, runId }), {
      initialProps: { runId: 'previous' },
    });
    await act(async () => { await vi.advanceTimersByTimeAsync(0); });
    rerender({ runId: 'current' });
    await act(async () => { await vi.advanceTimersByTimeAsync(0); });
    expect(result.current.snapshot).toEqual(runningSnap);
    await act(async () => { resolvePrevious(doneSnap); });
    expect(result.current.snapshot).toEqual(runningSnap);
    expect(result.current.done).toBe(false);
  });
});

describe('deriveStatusText', () => {
  const linter = (over: Partial<LinterResult>): LinterResult => ({ linter: 'l', success: true, duration: 1, violations: [], ...over });
  const lintTasks: Test[] = [{ name: 'lint prompt', framework: 'task', context: { status: 'running' } }];

  it('counts running, done and failed linters while a lint-only run is in flight', () => {
    expect(deriveStatusText(snapshot({
      status: { running: true, lint_run: true },
      tests: lintTasks,
      lint: [
        linter({ status: 'running', success: false }),
        linter({ status: 'running', success: false }),
        linter({ status: 'pending', success: false }),
        linter({}),
        linter({}),
        linter({ success: false, error: 'uir executable not found' }),
      ],
    }))).toBe('Running linters... (3 running · 2 done · 1 failed)');
  });

  it('reads Running linters with no counts before any linter has reported', () => {
    expect(deriveStatusText(snapshot({ status: { running: true, lint_run: true }, tests: lintTasks }))).toBe('Running linters...');
  });

  it('reports a finished lint-only run with its done and failed counts', () => {
    expect(deriveStatusText(snapshot({
      status: { running: false, lint_run: true },
      tests: [],
      lint: [linter({}), linter({}), linter({ success: false, error: 'boom' })],
    }))).toBe('Lint run complete (2 done · 1 failed)');
  });

  it('keeps the test wording for a live test group and a finished test+lint run', () => {
    const testGroup: Test = { name: 'Running tests across packages', framework: 'task', context: { status: 'running' } };
    expect(deriveStatusText(snapshot({ status: { running: true, lint_run: true }, tests: [testGroup], lint: [linter({ status: 'pending', success: false })] }))).toBe('Running tests...');
    expect(deriveStatusText(snapshot({
      status: { running: false, lint_run: true },
      tests: [{ name: 'TestFoo', framework: 'go test', passed: true }],
      lint: [linter({})],
    }))).toBe('Test run complete');
  });

  it('keeps the rerun wording', () => {
    const metadata = { sequence: 3, kind: 'rerun' as const };
    expect(deriveStatusText(snapshot({ status: { running: true, lint_run: true }, tests: lintTasks, metadata }))).toBe('Running rerun #3...');
    expect(deriveStatusText(snapshot({ status: { running: false, lint_run: true }, tests: [], metadata }))).toBe('Rerun complete');
  });
});
