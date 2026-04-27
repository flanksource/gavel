import { useEffect, useRef, useState } from 'preact/hooks';
import type { Test, TestAttempt } from '../types';
import { formatDuration, formatRunTimestamp } from '../utils';
import { apiUrl } from '../config';
import { AnsiHtml } from './AnsiHtml';

// attemptStatus reduces an Attempt to one of the same status strings we use
// in the main tree (passed/failed/timedout/skipped/pending), so icons and
// colors stay consistent with the rest of the UI.
function attemptStatus(a: TestAttempt): 'passed' | 'failed' | 'timedout' | 'skipped' | 'pending' {
  if (a.timed_out) return 'timedout';
  if (a.failed) return 'failed';
  if (a.skipped) return 'skipped';
  if (a.passed) return 'passed';
  return 'pending';
}

function attemptIcon(a: TestAttempt): string {
  switch (attemptStatus(a)) {
    case 'timedout': return 'mdi:clock-alert-outline';
    case 'failed': return 'codicon:error';
    case 'skipped': return 'codicon:circle-slash';
    case 'passed': return 'codicon:pass-filled';
    default: return 'svg-spinners:ring-resize';
  }
}

function attemptColor(a: TestAttempt): string {
  switch (attemptStatus(a)) {
    case 'timedout': return 'text-amber-600';
    case 'failed': return 'text-red-600';
    case 'skipped': return 'text-yellow-600';
    case 'passed': return 'text-green-600';
    default: return 'text-blue-500';
  }
}

interface ProcessMetrics {
  pid: number;
  cpu_percent?: number;
  rss?: number;
  vms?: number;
  open_files?: number;
  status?: string;
  command?: string;
}

interface Props {
  test: Test;
}

// TestAttempts renders the per-test execution history: a tabbed strip of
// attempts (latest selected by default) plus Metrics / Stack trace sections
// for the selected attempt. When the attempt is still running it polls the
// live /api/process/metrics endpoint; when it timed out we render the
// captured stack trace inline.
export function TestAttempts({ test }: Props) {
  const attempts = test.attempts ?? [];
  const latest = attempts.length - 1;
  const [selected, setSelected] = useState<number>(latest >= 0 ? latest : 0);

  useEffect(() => {
    if (attempts.length > 0) setSelected(attempts.length - 1);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [attempts.length]);

  if (attempts.length === 0) return null;

  const a = attempts[Math.min(selected, attempts.length - 1)];
  const isLive = !!a.pending || (!a.passed && !a.failed && !a.skipped && !a.timed_out);

  return (
    <div class="space-y-3">
      {attempts.length > 1 && (
        <div class="flex flex-wrap gap-1">
          {attempts.map((att, i) => (
            <button
              key={i}
              class={`inline-flex items-center gap-1.5 text-xs px-2 py-1 rounded border transition-colors ${i === selected ? 'border-blue-300 bg-blue-50 text-blue-800 font-medium' : 'border-gray-200 text-gray-600 hover:bg-gray-50'}`}
              onClick={() => setSelected(i)}
              title={att.started ? formatRunTimestamp(att.started) : ''}
            >
              <iconify-icon icon={attemptIcon(att)} class={attemptColor(att)} />
              <span>Attempt {att.sequence}</span>
              {att.run_kind === 'rerun' && <span class="opacity-60">rerun</span>}
            </button>
          ))}
        </div>
      )}

      <AttemptMetrics attempt={a} live={isLive} />
      {a.stack_trace && (
        <section>
          <div class="text-xs uppercase tracking-wide text-gray-500 mb-1">Stack trace</div>
          <AnsiHtml text={a.stack_trace} class="text-xs text-gray-700 whitespace-pre-wrap font-mono bg-gray-50 rounded p-3 max-h-[32rem] overflow-y-auto" />
        </section>
      )}
    </div>
  );
}

function AttemptMetrics({ attempt, live }: { attempt: TestAttempt; live: boolean }) {
  const [liveMetrics, setLiveMetrics] = useState<ProcessMetrics | null>(null);
  const [error, setError] = useState<string | null>(null);
  const liveRef = useRef(live);
  liveRef.current = live;

  useEffect(() => {
    if (!live || !attempt.pid) return;
    let cancelled = false;
    const poll = async () => {
      try {
        const res = await fetch(apiUrl(`/api/process/metrics?pid=${attempt.pid}`));
        if (!res.ok) {
          if (!cancelled) setError(res.status === 404 ? 'process exited' : `error ${res.status}`);
          return false;
        }
        const body = (await res.json()) as ProcessMetrics;
        if (!cancelled) {
          setLiveMetrics(body);
          setError(null);
        }
        return true;
      } catch (e: any) {
        if (!cancelled) setError(e?.message || 'fetch failed');
        return false;
      }
    };
    poll();
    const timer = window.setInterval(() => {
      if (!liveRef.current) return;
      poll();
    }, 2000);
    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, [attempt.pid, live]);

  const shown = liveMetrics ?? {
    pid: attempt.pid ?? 0,
    cpu_percent: attempt.cpu_percent,
    rss: attempt.rss,
  };

  const hasAny = shown.pid > 0 || shown.cpu_percent !== undefined || shown.rss !== undefined;

  return (
    <section>
      <div class="flex items-center gap-2 mb-1">
        <div class="text-xs uppercase tracking-wide text-gray-500">Process metrics</div>
        {live && <iconify-icon icon="svg-spinners:pulse" class="text-xs text-blue-500" />}
      </div>
      {!hasAny ? (
        <div class="text-xs text-gray-400 italic">No metrics captured for this attempt.</div>
      ) : (
        <div class="grid grid-cols-2 md:grid-cols-4 gap-2 text-xs">
          {shown.pid > 0 && (
            <MetricCell label="PID" value={String(shown.pid)} />
          )}
          <MetricCell label="CPU" value={shown.cpu_percent !== undefined ? `${shown.cpu_percent.toFixed(1)}%` : '—'} />
          <MetricCell label="RSS" value={formatBytes(shown.rss)} />
          {shown.open_files !== undefined && (
            <MetricCell label="FDs" value={String(shown.open_files)} />
          )}
          {attempt.duration !== undefined && attempt.duration > 0 && (
            <MetricCell label="Duration" value={formatDuration(attempt.duration)} />
          )}
          {attempt.exit_code !== undefined && (
            <MetricCell label="Exit" value={String(attempt.exit_code)} />
          )}
        </div>
      )}
      {error && (
        <div class="text-[11px] text-gray-400 mt-1 italic">{error}</div>
      )}
    </section>
  );
}

function MetricCell({ label, value }: { label: string; value: string }) {
  return (
    <div class="rounded-md border border-gray-200 bg-gray-50 px-2 py-1">
      <div class="text-[10px] uppercase tracking-wide text-gray-500">{label}</div>
      <div class="text-xs font-medium text-gray-800 truncate">{value}</div>
    </div>
  );
}

function formatBytes(n?: number): string {
  if (!n || n <= 0) return '—';
  const units = ['B', 'KB', 'MB', 'GB'];
  let v = n;
  let u = 0;
  while (v >= 1024 && u < units.length - 1) {
    v /= 1024;
    u++;
  }
  return `${v.toFixed(v >= 100 ? 0 : 1)} ${units[u]}`;
}
