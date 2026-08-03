import { useEffect, useState, type ComponentType } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import { Button, DropdownMenu } from '@flanksource/clicky-ui/components';
import { UiChevronDown, UiClock, UiCollapseAll, UiDebugStepOver, UiError, UiEye, UiHubot, UiUnknown, type IconProps } from '@flanksource/clicky-ui/icons';
import type { SessionStats } from '../../types';
import { fetchJSON } from '../../query';
import { Spinner } from '../../icons/Spinner';
import { todoQuery } from './format';
import { todoMutationJSON } from './todoMutations';
import { sessionStatsQueryOptions, todoQueryKeys } from './todoQueries';

// useSessionStats polls /api/todos/session/stats for one agent session. The
// server serves live runs from the cmux tailer's in-memory cache and reads
// finished/old sessions from the on-disk log, so this hook simply polls: fast
// while a run is in progress, slower while waiting for the log to appear, and it
// stops once a finished session's totals are final. Between polls of a running
// session the displayed clock ticks locally so the timer advances smoothly.
export function useSessionStats(dir: string, sessionId: string | undefined, active: boolean) {
  const [nowMs, setNowMs] = useState(() => Date.now());
  const enabled = active && !!sessionId;
  const query = useQuery({
    ...sessionStatsQueryOptions(dir, sessionId ?? ''),
    enabled,
  });
  const stats = enabled && !query.error ? query.data ?? null : null;

  useEffect(() => {
    if (enabled && query.dataUpdatedAt) setNowMs(Date.now());
  }, [enabled, query.dataUpdatedAt]);

  useEffect(() => {
    if (!stats?.inProgress) return;
    const id = setInterval(() => setNowMs(Date.now()), 1000);
    return () => clearInterval(id);
  }, [stats?.inProgress]);

  const elapsedMs = stats
    ? stats.inProgress
      ? stats.durationMs + Math.max(0, nowMs - query.dataUpdatedAt)
      : stats.durationMs
    : 0;

  return {
    stats,
    elapsedMs,
    error: enabled && query.error ? query.error.message : '',
  };
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

// TodoSessionTimer is the inline session readout for the Session header:
// identity (model + effort), live elapsed time, a context-window usage bar,
// estimated cost, and a control to open the session's cmux terminal (focus or
// resume). It renders nothing until the session has produced a log (found), and
// returns a bare run of flex items so the caller's header owns the surrounding row.
export function TodoSessionTimer({ dir, sessionId, active = true, onResume, resumeDisabled }: {
  dir: string;
  sessionId?: string;
  active?: boolean;
  // onResume continues this session in a fresh cmux run (claude --resume). Wired
  // from the todo detail's run flow; the "Resume in cmux" menu item hides without it.
  onResume?: () => void;
  resumeDisabled?: boolean;
}) {
  const { stats, elapsedMs } = useSessionStats(dir, sessionId, active);

  if (!sessionId || !stats?.found) return null;

  const cost = formatCost(stats.costUsd);
  const ctxPct = stats.contextWindow > 0 ? (stats.contextTokens / stats.contextWindow) * 100 : 0;
  return (
    <>
      <span className="inline-flex items-center gap-1" title={`${stats.agent || 'claude'} session`}>
        <UiHubot className="text-[12px]" />
        <span className="font-medium text-foreground">{modelLabel(stats)}</span>
        {stats.effort && (
          <span className="rounded bg-border/60 px-1 uppercase tracking-wide">{stats.effort}</span>
        )}
      </span>
      <span className="inline-flex items-center gap-1 tabular-nums" title="Session time">
        {stats.inProgress ? <Spinner className="text-[12px]" /> : <UiClock className="text-[12px]" />}
        {formatDuration(elapsedMs)}
      </span>
      {stats.contextTokens > 0 && (
        <span
          className="inline-flex items-center gap-1.5 tabular-nums"
          title={`Context: ${stats.contextTokens.toLocaleString()}${stats.contextWindow > 0 ? ` / ${stats.contextWindow.toLocaleString()} tokens (${Math.round(ctxPct)}%)` : ' tokens'}${stats.compactions ? ` · compacted ${stats.compactions}×` : ''}\n${stats.inputTokens.toLocaleString()} in / ${stats.outputTokens.toLocaleString()} out${stats.turns ? ` · ${stats.turns} turns` : ''} · ${formatTokens(stats.totalTokens)} total`}
        >
          <UiUnknown className="text-[12px]" />
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
              <UiCollapseAll className="text-[11px]" />
              {stats.compactions}
            </span>
          )}
        </span>
      )}
      {cost && (
        <span className="inline-flex items-center gap-1 tabular-nums" title="Estimated cost">
          <UiUnknown className="text-[12px]" />
          {cost}
        </span>
      )}
      {stats.state === 'error' && (
        <span
          className="inline-flex max-w-full items-center gap-1 rounded border border-red-500/30 bg-red-500/15 px-1.5 py-0.5 font-medium text-red-300"
          title={stats.error || 'Session ended on an API/network error'}
        >
          <UiError className="text-[12px]" />
          <span className="truncate">{stats.error || 'API error'}</span>
        </span>
      )}
      <CmuxSessionButton
        dir={dir}
        sessionId={sessionId}
        agent={stats.agent}
        onResume={onResume}
        resumeDisabled={resumeDisabled}
      />
    </>
  );
}

// cmuxSurfaceLabel renders the cmux terminal reference for a session, e.g.
// "Cmux workspace:2 surface:1". Missing parts are dropped so a workspace with no
// resolved surface still labels its terminal ("Cmux workspace:2").
export function cmuxSurfaceLabel(workspace?: string, surface?: string): string {
  return ['Cmux', workspace?.trim(), surface?.trim()].filter(Boolean).join(' ');
}

interface CmuxSurface {
  found: boolean;
  workspace?: string;
  surface?: string;
  reason?: string;
}

// useCmuxSurface resolves which cmux workspace/surface a session's terminal sits
// on. It only fetches while `enabled` (the menu is open) because each call shells
// out to cmux; a stopped cmux or a closed terminal resolves to found=false with a
// reason rather than throwing, so the control still offers Resume.
function useCmuxSurface(dir: string, agent: string | undefined, enabled: boolean) {
  const query = useQuery({
    queryKey: todoQueryKeys.cmuxSurface(dir, agent),
    queryFn: async ({ signal }) => {
      const params = new URLSearchParams(todoQuery(dir));
      if (agent) params.set('agent', agent);
      return fetchJSON<CmuxSurface>({
        url: `/api/todos/session/cmux?${params.toString()}`,
        signal,
        context: 'Cmux session request failed',
      });
    },
    enabled,
    staleTime: 5_000,
  });

  // An unreachable endpoint is inconclusive, not proof the terminal is gone —
  // leave the surface unknown (null) so Focus stays enabled and fails loudly on
  // its own if the terminal really is closed, rather than pre-disabling it.
  return { surface: query.error ? null : query.data ?? null, loading: query.isFetching };
}

// CmuxSessionButton is the session header's cmux control: a dropdown that either
// focuses the live cmux terminal for this session or resumes it in a fresh cmux
// run, captioned with the resolved workspace/surface the session maps to.
export function CmuxSessionButton({ dir, sessionId, agent, onResume, resumeDisabled }: {
  dir: string;
  sessionId: string;
  agent?: string;
  onResume?: () => void;
  resumeDisabled?: boolean;
}) {
  const [open, setOpen] = useState(false);
  const { surface, loading } = useCmuxSurface(dir, agent, open);
  const focusMutation = useMutation({
    mutationKey: ['todos', 'session', 'focus', { dir: dir.trim(), agent: agent ?? '' }],
    mutationFn: () => {
      const params = new URLSearchParams(todoQuery(dir));
      if (agent) params.set('agent', agent);
      return todoMutationJSON<{ focused: boolean }>(
        `/api/todos/session/focus?${params.toString()}`,
        { method: 'POST' },
        'Could not focus cmux session',
      );
    },
  });
  const focusBusy = focusMutation.isPending;
  const focusError = focusMutation.error?.message ?? '';

  // A resolved-but-absent workspace means the terminal was closed, so Focus can't
  // land there — only Resume can reopen it.
  const focusUnavailable = surface?.found === false;

  return (
    <DropdownMenu
      align="right"
      menuLabel="cmux session actions"
      menuClassName="w-64 max-w-[calc(100vw-24px)]"
      onOpenChange={setOpen}
      trigger={
        <Button
          variant="ghost"
          type="button"
          title={focusError || 'Open this session in cmux'}
          className={`inline-flex h-auto min-h-8 items-center gap-1 rounded border px-2 py-1 hover:bg-muted ${focusError ? 'border-red-500/40 text-red-400' : 'border-border'}`}
        >
          {focusBusy ? <Spinner className="text-[12px]" /> : focusError ? <UiError className="text-[12px]" /> : <UiEye className="text-[12px]" />}
          cmux
          <UiChevronDown className="text-[10px] opacity-70" />
        </Button>
      }
    >
      {close => (
        <div className="p-1 text-xs">
          {/* The cmux surface reference for this session — the "comment" mapping
              the session id to its cmux workspace/surface (e.g. workspace:2
              surface:1) so the user knows which terminal Focus/Resume targets. */}
          <div className="px-2 pb-1.5 pt-1 leading-tight text-muted-foreground">
            {loading ? (
              <span className="inline-flex items-center gap-1 text-[10px]"><Spinner className="text-[10px]" /> Locating cmux terminal…</span>
            ) : (
              <>
                <span className="block font-mono text-[10px] text-foreground/80">{cmuxSurfaceLabel(surface?.workspace, surface?.surface)}</span>
                <span className="block break-all font-mono text-[10px]">for session: {sessionId}</span>
                {focusUnavailable && (
                  <span className="mt-0.5 block text-[10px] text-amber-600">{surface?.reason || 'Terminal not found — resume to reopen'}</span>
                )}
              </>
            )}
          </div>
          <div className="border-t border-border pt-1">
            <CmuxMenuItem
              icon={UiEye}
              label="Focus in cmux"
              detail={focusUnavailable ? 'Terminal closed' : 'Bring the terminal to the front'}
              disabled={focusBusy || focusUnavailable}
              onClick={() => { close(); focusMutation.mutate(); }}
            />
            {onResume && (
              <CmuxMenuItem
                icon={UiDebugStepOver}
                label="Resume in cmux"
                detail="Continue this session in a new cmux run"
                disabled={resumeDisabled}
                onClick={() => { close(); onResume(); }}
              />
            )}
          </div>
          {focusError && <div className="px-2 pt-1 text-[10px] text-red-500">{focusError}</div>}
        </div>
      )}
    </DropdownMenu>
  );
}

function CmuxMenuItem({ icon: Icon, label, detail, disabled, onClick }: {
  icon: ComponentType<IconProps>;
  label: string;
  detail?: string;
  disabled?: boolean;
  onClick: () => void;
}) {
  return (
    <Button
      variant="ghost"
      type="button"
      disabled={disabled}
      onClick={onClick}
      className="flex h-auto w-full items-start justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted disabled:opacity-50"
    >
      <Icon className="mt-0.5 shrink-0 text-sm text-muted-foreground" />
      <span className="min-w-0 flex-1">
        <span className="block truncate font-medium text-foreground">{label}</span>
        {detail && <span className="block truncate text-[10px] text-muted-foreground">{detail}</span>}
      </span>
    </Button>
  );
}
