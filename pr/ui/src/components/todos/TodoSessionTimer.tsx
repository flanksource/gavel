import { useEffect, useRef, useState } from 'react';
import type { SessionStats } from '../../types';
import { GavelIcon } from '../GavelIcon';
import { todoQuery } from './format';

// useSessionStats polls /api/todos/session/stats for one agent session. The
// server serves live runs from the cmux tailer's in-memory cache and reads
// finished/old sessions from the on-disk log, so this hook simply polls: fast
// while a run is in progress, slower while waiting for the log to appear, and it
// stops once a finished session's totals are final. Between polls of a running
// session the displayed clock ticks locally so the timer advances smoothly.
export function useSessionStats(dir: string, provider: string, sessionId: string | undefined, active: boolean) {
  const [stats, setStats] = useState<SessionStats | null>(null);
  const fetchedAtRef = useRef(0);
  const [nowMs, setNowMs] = useState(() => Date.now());

  useEffect(() => {
    setStats(null);
    if (!active || !sessionId) return;
    let cancelled = false;
    let timer: ReturnType<typeof setTimeout>;
    const params = new URLSearchParams(todoQuery(dir, provider));
    params.set('sessionId', sessionId);
    const url = `/api/todos/session/stats?${params.toString()}`;

    const poll = async () => {
      let nextDelay = 5000;
      try {
        const res = await fetch(url);
        const data = (await res.json()) as SessionStats;
        if (cancelled) return;
        setStats(data);
        fetchedAtRef.current = Date.now();
        setNowMs(Date.now());
        if (data.inProgress) nextDelay = 2000;
        else if (!data.found) nextDelay = 4000;
        else return; // finished: totals are final, stop polling
      } catch {
        if (cancelled) return;
      }
      timer = setTimeout(poll, nextDelay);
    };
    poll();
    return () => { cancelled = true; clearTimeout(timer); };
  }, [dir, provider, sessionId, active]);

  useEffect(() => {
    if (!stats?.inProgress) return;
    const id = setInterval(() => setNowMs(Date.now()), 1000);
    return () => clearInterval(id);
  }, [stats?.inProgress]);

  const elapsedMs = stats
    ? stats.inProgress
      ? stats.durationMs + Math.max(0, nowMs - fetchedAtRef.current)
      : stats.durationMs
    : 0;

  return { stats, elapsedMs };
}

export function formatDuration(ms: number): string {
  const totalSec = Math.max(0, Math.floor(ms / 1000));
  const h = Math.floor(totalSec / 3600);
  const m = Math.floor((totalSec % 3600) / 60);
  const s = totalSec % 60;
  if (h > 0) return `${h}h ${String(m).padStart(2, '0')}m`;
  if (m > 0) return `${m}m ${String(s).padStart(2, '0')}s`;
  return `${s}s`;
}

export function formatTokens(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}k`;
  return String(n);
}

export function formatCost(usd: number): string {
  if (usd <= 0) return '';
  if (usd < 0.01) return '<$0.01';
  return `$${usd.toFixed(2)}`;
}

function modelLabel(stats: SessionStats): string {
  return stats.model || stats.agent || 'claude';
}

// ctxBarColor shades the context-usage bar from healthy (green) through filling
// (amber) to nearly-exhausted (red) so an overfull window reads at a glance.
function ctxBarColor(pct: number): string {
  if (pct >= 85) return 'bg-red-500';
  if (pct >= 60) return 'bg-amber-500';
  return 'bg-emerald-500';
}

// TodoSessionTimer is the full session readout for the detail pane: identity
// (model + effort), live elapsed time, a context-window usage bar, estimated
// cost, and a control to focus the session's cmux terminal. It renders nothing
// until the session has produced a log (found).
export function TodoSessionTimer({ dir, provider, sessionId, active = true }: {
  dir: string;
  provider: string;
  sessionId?: string;
  active?: boolean;
}) {
  const { stats, elapsedMs } = useSessionStats(dir, provider, sessionId, active);
  const [focusBusy, setFocusBusy] = useState(false);
  const [focusError, setFocusError] = useState('');

  // focusSession brings the agent's cmux terminal to the front. The workspace is
  // keyed by the run's working directory and agent, so the server can find it
  // without the UI tracking cmux refs. A closed terminal / stopped cmux surfaces
  // its reason inline rather than failing silently.
  const focusSession = async () => {
    if (focusBusy) return;
    setFocusBusy(true);
    setFocusError('');
    try {
      const params = new URLSearchParams(todoQuery(dir, provider));
      if (stats?.agent) params.set('agent', stats.agent);
      const res = await fetch(`/api/todos/session/focus?${params.toString()}`, { method: 'POST' });
      if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || 'Could not focus cmux session');
    } catch (err: any) {
      setFocusError(err?.message || 'Could not focus cmux session');
    } finally {
      setFocusBusy(false);
    }
  };

  if (!sessionId || !stats?.found) return null;

  const cost = formatCost(stats.costUsd);
  const ctxPct = stats.contextWindow > 0 ? (stats.contextTokens / stats.contextWindow) * 100 : 0;
  return (
    <div className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 rounded-md border border-border bg-muted/40 px-2.5 py-1.5 text-[11px] text-muted-foreground">
      <span className="inline-flex items-center gap-1" title={`${stats.agent || 'claude'} session`}>
        <GavelIcon name="codicon:hubot" className="text-[12px]" />
        <span className="font-medium text-foreground">{modelLabel(stats)}</span>
        {stats.effort && (
          <span className="rounded bg-border/60 px-1 uppercase tracking-wide">{stats.effort}</span>
        )}
      </span>
      <span className="inline-flex items-center gap-1 tabular-nums" title="Session time">
        <GavelIcon name={stats.inProgress ? 'svg-spinners:ring-resize' : 'codicon:clock'} className="text-[12px]" />
        {formatDuration(elapsedMs)}
      </span>
      {stats.contextTokens > 0 && (
        <span
          className="inline-flex items-center gap-1.5 tabular-nums"
          title={`Context: ${stats.contextTokens.toLocaleString()}${stats.contextWindow > 0 ? ` / ${stats.contextWindow.toLocaleString()} tokens (${Math.round(ctxPct)}%)` : ' tokens'}${stats.compactions ? ` · compacted ${stats.compactions}×` : ''}\n${stats.inputTokens.toLocaleString()} in / ${stats.outputTokens.toLocaleString()} out${stats.turns ? ` · ${stats.turns} turns` : ''} · ${formatTokens(stats.totalTokens)} total`}
        >
          <GavelIcon name="codicon:symbol-number" className="text-[12px]" />
          {stats.contextWindow > 0 ? (
            <>
              <span className="h-1.5 w-16 overflow-hidden rounded-full bg-border/60">
                <span
                  className={`block h-full rounded-full transition-all duration-300 ${ctxBarColor(ctxPct)}`}
                  style={{ width: `${Math.min(100, ctxPct)}%` }}
                />
              </span>
              <span>{Math.round(ctxPct)}%</span>
            </>
          ) : (
            formatTokens(stats.contextTokens)
          )}
          {stats.compactions > 0 && (
            <span className="inline-flex items-center gap-0.5" title={`Context compacted ${stats.compactions}×`}>
              <GavelIcon name="codicon:fold" className="text-[11px]" />
              {stats.compactions}
            </span>
          )}
        </span>
      )}
      {cost && (
        <span className="inline-flex items-center gap-1 tabular-nums" title="Estimated cost">
          <GavelIcon name="codicon:credit-card" className="text-[12px]" />
          {cost}
        </span>
      )}
      {stats.state === 'error' && (
        <span
          className="inline-flex max-w-full items-center gap-1 rounded border border-red-500/30 bg-red-500/15 px-1.5 py-0.5 font-medium text-red-300"
          title={stats.error || 'Session ended on an API/network error'}
        >
          <GavelIcon name="codicon:error" className="text-[12px]" />
          <span className="truncate">{stats.error || 'API error'}</span>
        </span>
      )}
      <button
        type="button"
        onClick={focusSession}
        disabled={focusBusy}
        title={focusError || 'Focus this session’s terminal in cmux'}
        className={`ml-auto inline-flex items-center gap-1 rounded border px-1.5 py-0.5 hover:bg-muted disabled:opacity-50 ${focusError ? 'border-red-500/40 text-red-400' : 'border-border'}`}
      >
        <GavelIcon name={focusBusy ? 'svg-spinners:ring-resize' : focusError ? 'codicon:error' : 'codicon:eye'} className="text-[12px]" />
        cmux
      </button>
    </div>
  );
}
