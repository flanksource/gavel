import { useCallback, useEffect, useRef, useState } from 'react';
import type {
  BenchComparison,
  LinterResult,
  RerunRequest,
  RunMeta,
  Snapshot,
  SnapshotStatus,
  Test,
} from '../types';
import { isLintOnlyPhase } from '../utils';

// useTestRun is the reusable client for gavel's testrunner HTTP handlers. It is
// SSE-first: it subscribes to /api/tests/stream (event: message / event: done)
// and falls back to polling /api/tests only when EventSource is unavailable. It
// returns the latest full Snapshot plus derived fields and the rerun/stop
// actions, modelled on clicky-ui's useTaskRun.

export interface UseTestRunOptions {
  /** API prefix, e.g. "" or "https://host" or "/results/{repo}/{id}". */
  baseUrl?: string;
  /** Alias for baseUrl kept for parity with useTaskRun. Defaults to
   *  window.__gavelBasePath ?? "". */
  basePath?: string;
  /** Disable the subscription entirely (e.g. before a target is known). */
  enabled?: boolean;
  /** Poll interval (ms) for the fallback transport. */
  pollMs?: number;
  /** Force the polling transport even when EventSource exists (mainly tests). */
  forcePoll?: boolean;
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
  rerun: (body?: RerunRequest) => Promise<void>;
  stop: (taskId?: string) => Promise<void>;
  /** Reconnect the stream / restart polling. */
  refetch: () => void;
}

const DEFAULT_POLL_MS = 2_000;
const IDLE_STATUS: SnapshotStatus = { running: false };

function defaultBase(): string {
  if (typeof window !== 'undefined') {
    return ((window as unknown as { __gavelBasePath?: string }).__gavelBasePath) || '';
  }
  return '';
}

function hasEventSource(): boolean {
  return typeof globalThis !== 'undefined' && typeof globalThis.EventSource !== 'undefined';
}

function deriveStatusText(snap: Snapshot): string {
  const status = snap.status || IDLE_STATUS;
  if (!status.running) {
    if (status.stopped) return status.stop_message || 'Stopped by user';
    return snap.metadata?.kind === 'rerun' ? 'Rerun complete' : 'Test run complete';
  }
  if (snap.metadata?.kind === 'rerun') {
    return `Running rerun #${snap.metadata.sequence || 1}...`;
  }
  if (isLintOnlyPhase(snap.tests || [], status.running, !!status.lint_run)) {
    return 'Running linters...';
  }
  return 'Running tests...';
}

export function useTestRun(options: UseTestRunOptions = {}): UseTestRunResult {
  const {
    baseUrl,
    basePath,
    enabled = true,
    pollMs = DEFAULT_POLL_MS,
    forcePoll = false,
  } = options;
  const base = baseUrl ?? basePath ?? defaultBase();

  const [snapshot, setSnapshot] = useState<Snapshot | undefined>(undefined);
  const [statusText, setStatusText] = useState('Loading...');
  const [done, setDone] = useState(false);
  const [error, setError] = useState<string | undefined>(undefined);
  const [streamToken, setStreamToken] = useState(0);
  const doneRef = useRef(false);

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
      let stopped = false;
      let timer: ReturnType<typeof setTimeout> | undefined;
      const tick = async () => {
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
        if (!stopped) timer = setTimeout(tick, pollMs);
      };
      void tick();
      return () => {
        stopped = true;
        if (timer) clearTimeout(timer);
      };
    }

    // SSE transport (default): seed once via /api/tests then follow the stream.
    fetch(`${base}/api/tests`, { headers: { Accept: 'application/json' } })
      .then((r) => r.json())
      .then((snap: Snapshot) => apply(snap))
      .catch(() => {});

    const es = new EventSource(`${base}/api/tests/stream`);
    es.addEventListener('message', (e) => {
      const snap = JSON.parse((e as MessageEvent).data) as Snapshot;
      apply(snap);
      if (!snap.status?.running) es.close();
    });
    es.addEventListener('done', () => {
      doneRef.current = true;
      setDone(true);
      setSnapshot((prev) => (prev ? { ...prev, status: { ...prev.status, running: false } } : prev));
      es.close();
    });
    es.onerror = () => {
      if (!doneRef.current) setError('Connection lost — retrying...');
    };
    return () => es.close();
  }, [base, enabled, pollMs, forcePoll, streamToken, apply]);

  const rerun = useCallback(
    async (body: RerunRequest = {}) => {
      doneRef.current = false;
      setDone(false);
      setStreamToken((n) => n + 1);
      const res = await fetch(`${base}/api/rerun`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      if (!res.ok && res.status !== 409) {
        setError(`Rerun failed: ${(await res.text()).trim()}`);
      }
    },
    [base],
  );

  const stop = useCallback(
    async (taskId?: string) => {
      const payload = taskId ? { scope: 'task', task_id: taskId } : { scope: 'global' };
      const res = await fetch(`${base}/api/stop`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      if (!res.ok) setError(`Stop failed: ${(await res.text()).trim()}`);
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
    refetch,
  };
}
