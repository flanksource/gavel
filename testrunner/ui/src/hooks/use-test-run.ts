import { useCallback, useEffect, useRef, useState } from 'react';
import type {
  BenchComparison,
  LinterResult,
  RerunRequest,
  RunMeta,
  Snapshot,
  SnapshotStatus,
  Test,
  TestEditRequest,
} from '../types';
import { isLintOnlyPhase, isLintOnlyRun, lintStatusCounts } from '../utils';

// useTestRun is the reusable client for gavel's testrunner HTTP handlers. It is
// SSE-first: it subscribes to /api/tests/stream (event: message / event: done)
// and falls back to polling /api/tests only when EventSource is unavailable. It
// returns the latest full Snapshot plus derived fields and the rerun/stop
// actions, modelled on clicky-ui's useTaskRun.

// The subset of the DOM `EventSource` surface this hook actually uses. A host
// that needs to route this connection through its own transport (e.g. to
// multiplex several SSE streams over one connection, working around the
// six-connections-per-origin cap HTTP/1.1 imposes) can inject a factory that
// returns anything satisfying this — never the full `EventSource` interface.
// This type is deliberately local rather than imported from clicky-ui: this
// package is consumed by both React 18 and 19 hosts, and must not pull in
// clicky-ui's React context.
export interface EventSourceLike {
  addEventListener(type: string, listener: EventListenerOrEventListenerObject): void;
  removeEventListener(type: string, listener: EventListenerOrEventListenerObject): void;
  close(): void;
  onerror: ((event: Event) => void) | null;
}

export interface UseTestRunOptions {
  /** API prefix, e.g. "" or "https://host" or "/results/{repo}/{id}". */
  baseUrl?: string;
  /** Alias for baseUrl kept for parity with useTaskRun. Defaults to
   *  window.__gavelBasePath ?? "". */
  basePath?: string;
  /** Scope every request to a specific run id, hitting
   *  `${base}/${runId}/api/...` instead of `${base}/api/...`. Set when the host
   *  serves many concurrent runs (gavel MultiServer); omit for a single-run
   *  server mounted at the root. */
  runId?: string;
  /** Disable the subscription entirely (e.g. before a target is known). */
  enabled?: boolean;
  /** Poll interval (ms) for the fallback transport. */
  pollMs?: number;
  /** Force the polling transport even when EventSource exists (mainly tests). */
  forcePoll?: boolean;
  /** Builds the SSE connection for a URL. Defaults to `(url) => new
   *  EventSource(url)`. A host embedding this hook behind a connection
   *  multiplexer (e.g. pr/ui's eventHub) passes its own factory here instead
   *  of letting this hook open a raw `EventSource`. */
  createEventSource?: (url: string) => EventSourceLike;
}

export interface UseTestRunResult {
  snapshot: Snapshot | undefined;
  tests: Test[];
  lint: LinterResult[] | undefined;
  bench: BenchComparison | undefined;
  status: SnapshotStatus;
  /** Human-readable run state, e.g. "Running tests...", "Test run complete". */
  statusText: string;
  runMeta: RunMeta | undefined;
  done: boolean;
  error: string | undefined;
  rerun: (body?: RerunRequest) => Promise<ActionResult>;
  stop: (taskId?: string) => Promise<ActionResult>;
  editTest: (body: TestEditRequest) => Promise<ActionResult>;
  /** Reconnect the stream / restart polling. */
  refetch: () => void;
}

/**
 * How a POSTed action ended. A failed one is also reported in `error`; the
 * result lets the caller tell a rerun refused because another is in progress
 * (`conflict`) from one that failed.
 */
export interface ActionResult {
  ok: boolean;
  /** The server's HTTP status. */
  status: number;
  /** Another rerun is in progress (409). */
  conflict: boolean;
  /** The server's own message: the failure text, or a test edit's summary. */
  message?: string;
}

const DEFAULT_POLL_MS = 2_000;
const IDLE_STATUS: SnapshotStatus = { running: false };

// Resolves after ms, or as soon as signal aborts.
function sleep(ms: number, signal: AbortSignal): Promise<void> {
  return new Promise((resolve) => {
    const onAbort = () => {
      clearTimeout(timer);
      resolve();
    };
    const timer = setTimeout(() => {
      signal.removeEventListener('abort', onAbort);
      resolve();
    }, ms);
    signal.addEventListener('abort', onAbort, { once: true });
  });
}

function defaultCreateEventSource(url: string): EventSourceLike {
  return new EventSource(url);
}

function defaultBase(): string {
  if (typeof window !== 'undefined') {
    return ((window as unknown as { __gavelBasePath?: string }).__gavelBasePath) || '';
  }
  return '';
}

function hasEventSource(): boolean {
  return typeof globalThis !== 'undefined' && typeof globalThis.EventSource !== 'undefined';
}

export function deriveStatusText(snap: Snapshot): string {
  const status = snap.status || IDLE_STATUS;
  const tests = snap.tests || [];
  if (!status.running) {
    if (status.stopped) return status.stop_message || 'Stopped by user';
    if (snap.metadata?.kind === 'rerun') return 'Rerun complete';
    if (isLintOnlyRun(tests, !!status.lint_run)) {
      const { done, failed } = lintStatusCounts(snap.lint);
      return `Lint run complete (${done} done · ${failed} failed)`;
    }
    return 'Test run complete';
  }
  if (snap.metadata?.kind === 'rerun') {
    return `Running rerun #${snap.metadata.sequence || 1}...`;
  }
  // A fixture benchmark or profile run has no lint phase.
  if (snap.performance) return 'Running fixtures...';
  if (isLintOnlyPhase(tests, status.running)) {
    if (!snap.lint?.length) return 'Running linters...';
    const { running, done, failed } = lintStatusCounts(snap.lint);
    return `Running linters... (${running} running · ${done} done · ${failed} failed)`;
  }
  return 'Running tests...';
}

export function useTestRun(options: UseTestRunOptions = {}): UseTestRunResult {
  const {
    baseUrl,
    basePath,
    runId,
    enabled = true,
    pollMs = DEFAULT_POLL_MS,
    forcePoll = false,
    createEventSource,
  } = options;
  const root = baseUrl ?? basePath ?? defaultBase();
  // When a run id is supplied, every endpoint is scoped under it so the host can
  // route to that run's isolated stream; otherwise hit the root single-run paths.
  const base = runId ? `${root}/${encodeURIComponent(runId)}` : root;

  const [snapshot, setSnapshot] = useState<Snapshot | undefined>(undefined);
  const [statusText, setStatusText] = useState('Loading...');
  const [done, setDone] = useState(false);
  const [error, setError] = useState<string | undefined>(undefined);
  const [streamToken, setStreamToken] = useState(0);
  const doneRef = useRef(false);
  // Held in a ref (rather than an effect dependency) so an inline
  // `createEventSource` function passed by the host doesn't cause the
  // subscription effect below to tear down and re-open on every render.
  const createEventSourceRef = useRef(createEventSource ?? defaultCreateEventSource);
  useEffect(() => {
    createEventSourceRef.current = createEventSource ?? defaultCreateEventSource;
  });

  const apply = useCallback((snap: Snapshot) => {
    const running = !!(snap.status && snap.status.running);
    setSnapshot(snap);
    setStatusText(deriveStatusText(snap));
    setError(undefined);
    if (!running) {
      doneRef.current = true;
      setDone(true);
    } else {
      doneRef.current = false;
      setDone(false);
    }
  }, []);

  useEffect(() => {
    if (!enabled) return;
    doneRef.current = false;
    setDone(false);

    // Polling fallback transport.
    if (forcePoll || !hasEventSource()) {
      const controller = new AbortController();
      void (async () => {
        while (!controller.signal.aborted) {
          try {
            const res = await fetch(`${base}/api/tests`, { headers: { Accept: 'application/json' } });
            if (res.ok) {
              const snap = (await res.json()) as Snapshot;
              apply(snap);
              if (!snap.status?.running) return; // terminal: stop polling
            }
          } catch {
            if (!doneRef.current) setError('Connection lost — retrying...');
          }
          await sleep(pollMs, controller.signal);
        }
      })();
      return () => controller.abort();
    }

    // SSE transport (default): seed once via /api/tests then follow the stream.
    fetch(`${base}/api/tests`, { headers: { Accept: 'application/json' } })
      .then((r) => r.json())
      .then((snap: Snapshot) => apply(snap))
      .catch(() => {});

    const es = createEventSourceRef.current(`${base}/api/tests/stream`);
    const onMessage = (e: Event) => {
      const snap = JSON.parse((e as MessageEvent).data) as Snapshot;
      apply(snap);
      if (!snap.status?.running) es.close();
    };
    const onDone = () => {
      doneRef.current = true;
      setDone(true);
      setSnapshot((prev) => (prev ? { ...prev, status: { ...prev.status, running: false } } : prev));
      es.close();
    };
    es.addEventListener('message', onMessage);
    es.addEventListener('done', onDone);
    es.onerror = () => {
      if (!doneRef.current) setError('Connection lost — retrying...');
    };
    return () => {
      es.removeEventListener('message', onMessage);
      es.removeEventListener('done', onDone);
      es.close();
    };
  }, [base, enabled, pollMs, forcePoll, streamToken, apply]);

  const rerun = useCallback(
    async (body: RerunRequest = {}): Promise<ActionResult> => {
      doneRef.current = false;
      setDone(false);
      setStreamToken((n) => n + 1);
      const res = await fetch(`${base}/api/rerun`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      if (res.status === 409) return { ok: false, status: 409, conflict: true, message: 'Another rerun is in progress' };
      if (!res.ok) {
        const message = (await res.text()).trim();
        setError(`Rerun failed: ${message}`);
        return { ok: false, status: res.status, conflict: false, message };
      }
      return { ok: true, status: res.status, conflict: false };
    },
    [base],
  );

  const stop = useCallback(
    async (taskId?: string): Promise<ActionResult> => {
      const payload = taskId ? { scope: 'task', task_id: taskId } : { scope: 'global' };
      const res = await fetch(`${base}/api/stop`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      if (!res.ok) {
        const message = (await res.text()).trim();
        setError(`Stop failed: ${message}`);
        return { ok: false, status: res.status, conflict: false, message };
      }
      return { ok: true, status: res.status, conflict: false };
    },
    [base],
  );

  const editTest = useCallback(
    async (body: TestEditRequest): Promise<ActionResult> => {
      const res = await fetch(`${base}/api/tests/edit`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      if (!res.ok) {
        const message = (await res.text()).trim();
        setError(`Test edit failed: ${message}`);
        return { ok: false, status: res.status, conflict: false, message };
      }
      const edited = (await res.json()) as { message?: string };
      setStreamToken((n) => n + 1);
      return { ok: true, status: res.status, conflict: false, ...(edited.message ? { message: edited.message } : {}) };
    },
    [base],
  );

  const refetch = useCallback(() => setStreamToken((n) => n + 1), []);

  return {
    snapshot,
    tests: snapshot?.tests ?? [],
    lint: snapshot?.lint,
    bench: snapshot?.bench,
    status: snapshot?.status ?? IDLE_STATUS,
    statusText,
    runMeta: snapshot?.metadata,
    done,
    error,
    rerun,
    stop,
    editTest,
    refetch,
  };
}
