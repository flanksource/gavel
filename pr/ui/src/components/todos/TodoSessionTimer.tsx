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

// TodoSessionTimer is the full session readout for the detail pane: identity
// (model + effort), live elapsed time, total token usage and estimated cost. It
// renders nothing until the session has produced a log (found).
export function TodoSessionTimer({ dir, provider, sessionId, active = true }: {
  dir: string;
  provider: string;
  sessionId?: string;
  active?: boolean;
}) {
  const { stats, elapsedMs } = useSessionStats(dir, provider, sessionId, active);
  if (!sessionId || !stats?.found) return null;

  const cost = formatCost(stats.costUsd);
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
      {stats.totalTokens > 0 && (
        <span
          className="inline-flex items-center gap-1 tabular-nums"
          title={`${stats.inputTokens.toLocaleString()} in / ${stats.outputTokens.toLocaleString()} out${stats.turns ? ` · ${stats.turns} turns` : ''}`}
        >
          <GavelIcon name="codicon:symbol-number" className="text-[12px]" />
          {formatTokens(stats.totalTokens)}
        </span>
      )}
      {cost && (
        <span className="inline-flex items-center gap-1 tabular-nums" title="Estimated cost">
          <GavelIcon name="codicon:credit-card" className="text-[12px]" />
          {cost}
        </span>
      )}
    </div>
  );
}

// TodoSessionTimerCompact is the condensed readout for a sidebar row: elapsed
// time (with a spinner while running) and cost, no chrome. Render it only for a
// row that actually has a session so it never polls idle todos.
export function TodoSessionTimerCompact({ dir, provider, sessionId, active = true }: {
  dir: string;
  provider: string;
  sessionId?: string;
  active?: boolean;
}) {
  const { stats, elapsedMs } = useSessionStats(dir, provider, sessionId, active);
  if (!sessionId || !stats?.found) return null;

  const cost = formatCost(stats.costUsd);
  return (
    <span className="inline-flex shrink-0 items-center gap-1.5 text-[11px] tabular-nums text-muted-foreground" title="Agent session">
      <GavelIcon name={stats.inProgress ? 'svg-spinners:ring-resize' : 'codicon:clock'} className="text-[11px]" />
      <span>{formatDuration(elapsedMs)}</span>
      {cost && <span>{cost}</span>}
    </span>
  );
}
