import { Button } from '@flanksource/clicky-ui/components';
import { UiClock, UiCollapseAll, UiError, UiHubot, UiStop } from '@flanksource/clicky-ui/icons';
import { Spinner } from '../../icons/Spinner';
import { formatCost, formatDuration, formatTokens, useSessionStats } from './TodoSessionTimer';

/**
 * What the header becomes while a run is in flight.
 *
 * The cluster it replaces went the other way: every control greyed out and Run
 * flipped to a *disabled* "Stop" whose own tooltip admitted it did nothing,
 * while elapsed time, model, context and cost stayed buried in the Session tab.
 * So the header said least exactly when there was most to say.
 *
 * This is the same readout `TodoSessionTimer` renders in the session header —
 * deliberately, so the two agree — plus the one action that applies.
 */
export function TodoRunStatusStrip({
  dir,
  sessionId,
  phaseLabel,
  stopping,
  onStop,
}: {
  dir: string;
  sessionId: string | undefined;
  /** Which phase is running, from the active attempt's step. */
  phaseLabel?: string | undefined;
  stopping?: boolean;
  /** Absent when the run in flight cannot be interrupted. */
  onStop?: (() => void) | undefined;
}) {
  const { stats, elapsedMs } = useSessionStats({ dir, sessionId, active: !!sessionId });
  // A session that has not priced any turns yet reports no cost at all, so this
  // reads the field's presence rather than assuming the payload is complete.
  const cost = typeof stats?.costUsd === 'number' ? formatCost(stats.costUsd) : '';
  const contextPct = stats && stats.contextWindow > 0 ? (stats.contextTokens / stats.contextWindow) * 100 : 0;

  return (
    <div className="inline-flex min-w-0 flex-wrap items-center gap-x-3 gap-y-1 rounded-md border border-border bg-muted/20 px-2 py-1 text-xs text-muted-foreground">
      <span className="inline-flex shrink-0 items-center gap-1.5 font-medium text-foreground">
        <Spinner className="text-[12px]" />
        {phaseLabel ?? 'Running'}
      </span>

      {stats?.found && (
        <>
          <span className="inline-flex min-w-0 items-center gap-1" title={`${stats.agent || 'claude'} session`}>
            <UiHubot className="shrink-0 text-[12px]" />
            <span className="truncate font-medium text-foreground">{stats.model || stats.agent || 'claude'}</span>
            {stats.effort && <span className="shrink-0 rounded bg-border/60 px-1 uppercase tracking-wide">{stats.effort}</span>}
          </span>

          <span className="inline-flex shrink-0 items-center gap-1 tabular-nums" title="Session time">
            <UiClock className="text-[12px]" />
            {formatDuration(elapsedMs)}
          </span>

          {stats.contextTokens > 0 && (
            <span
              className="inline-flex shrink-0 items-center gap-1.5 tabular-nums"
              title={`Context: ${stats.contextTokens.toLocaleString()}${stats.contextWindow > 0 ? ` / ${stats.contextWindow.toLocaleString()} tokens (${Math.round(contextPct)}%)` : ' tokens'}${stats.compactions ? ` · compacted ${stats.compactions}×` : ''}`}
            >
              {stats.contextWindow > 0 ? (
                <>
                  <span className="h-1.5 w-16 overflow-hidden rounded-full bg-border/60">
                    <span
                      className={`block h-full rounded-full transition-all duration-300 ${contextBarColor(contextPct)}`}
                      style={{ width: `${Math.min(100, contextPct)}%` }}
                    />
                  </span>
                  <span>{Math.round(contextPct)}%</span>
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
            <span className="shrink-0 tabular-nums" title="Estimated cost">
              {cost}
            </span>
          )}

          {stats.state === 'error' && (
            <span
              className="inline-flex min-w-0 items-center gap-1 rounded border border-red-500/30 bg-red-500/15 px-1.5 py-0.5 font-medium text-red-600 [[data-theme=dark]_&]:text-red-300"
              title={stats.error || 'Session ended on an API/network error'}
            >
              <UiError className="shrink-0 text-[12px]" />
              <span className="truncate">{stats.error || 'API error'}</span>
            </span>
          )}
        </>
      )}

      <Button
        variant="ghost"
        size="sm"
        type="button"
        onClick={onStop}
        disabled={!onStop || stopping}
        // A run with no stoppable attempt says so, rather than offering a
        // control that looks live and does nothing.
        title={onStop ? 'Stop this run' : 'This run cannot be interrupted'}
        className="ml-auto inline-flex h-6 shrink-0 items-center gap-1 rounded border border-border px-2 text-xs font-medium text-red-600 hover:bg-red-500/10 disabled:opacity-50 [[data-theme=dark]_&]:text-red-400"
      >
        <UiStop className="text-xs" />
        {stopping ? 'Stopping…' : 'Stop'}
      </Button>
    </div>
  );
}

// Shades the context bar from healthy through filling to nearly exhausted, the
// same thresholds TodoSessionTimer uses.
function contextBarColor(pct: number): string {
  if (pct >= 85) return 'bg-red-500';
  if (pct >= 60) return 'bg-amber-500';
  return 'bg-emerald-500';
}
