import { UiCancel, UiCheckFilled, UiCircleOutline, UiCircleXFilled, UiQuestion } from '@flanksource/clicky-ui/icons';
import type { IconProps } from '@flanksource/clicky-ui/icons';
import type { ComponentType } from 'react';
import { Spinner } from '../../icons/Spinner';
import type { TodoItem, TodoPhase, TodoPhaseRun } from '../../types';
import { useNow } from '../../useNow';
import { phase as phaseMeta } from './phaseMachine';
import { formatDuration } from './TodoSessionTimer';

// A phase's own outcome, styled like the todo statuses in format.tsx so a
// failed phase reads red wherever it appears. `waiting` is amber rather than
// red: the agent is blocked on a person, which is not a failure.
const STATE_VIEWS: Record<TodoPhaseRun['state'], { icon: ComponentType<IconProps>; className: string }> = {
  pending: { icon: UiCircleOutline, className: 'text-muted-foreground' },
  running: { icon: Spinner, className: 'text-blue-600 dark:text-blue-400' },
  waiting: { icon: UiQuestion, className: 'text-amber-600 dark:text-amber-400' },
  succeeded: { icon: UiCheckFilled, className: 'text-emerald-600 dark:text-emerald-400' },
  failed: { icon: UiCircleXFilled, className: 'text-red-600 dark:text-red-400' },
  cancelled: { icon: UiCancel, className: 'text-muted-foreground' },
};

export function phaseRunning(run: TodoPhaseRun): boolean {
  return run.state === 'running' || run.state === 'pending';
}

// A verification run that executed cleanly but produced failing checks is a
// failure, even though its own state is `succeeded`.
export function phaseFailed(run: TodoPhaseRun): boolean {
  return run.state === 'failed' || run.state === 'cancelled' || (run.progress?.failed ?? 0) > 0;
}

// phaseElapsedMs is the elapsed time to display. `duration_ms` is a snapshot
// from when the row was read, so a still-running phase is measured against the
// shared clock instead — otherwise its timer would freeze between list polls.
export function phaseElapsedMs(run: TodoPhaseRun, nowMs: number): number {
  if (phaseRunning(run) && run.started_at) {
    const started = Date.parse(run.started_at);
    if (!Number.isNaN(started)) return Math.max(0, nowMs - started);
  }
  return run.duration_ms ?? 0;
}

// phaseProgressLabel renders progress in the unit the phase actually counts.
// A single iteration is not worth showing — "1/1" is noise on every row that
// succeeded first try — but any failure is.
export function phaseProgressLabel(run: TodoPhaseRun): string {
  const progress = run.progress;
  if (!progress) return '';
  if (progress.total > 1) return `${progress.done}/${progress.total}`;
  if ((progress.failed ?? 0) > 0) return `${progress.failed} failed`;
  return '';
}

// PhaseTimer subscribes to the shared clock only while its phase is live, so a
// table of settled todos mounts no timers at all. useSessionStats keeps a
// private setInterval per instance; four of those across hundreds of rows is
// the thing this deliberately avoids.
function PhaseTimer({ run }: { run: TodoPhaseRun }) {
  const now = useNow();
  return <>{formatDuration(phaseElapsedMs(run, now))}</>;
}

// A settled phase's elapsed time cannot change, so it never subscribes to the
// clock. Separate components rather than one branch because hooks cannot be
// called conditionally.
function StaticTimer({ run }: { run: TodoPhaseRun }) {
  return <>{formatDuration(run.duration_ms ?? 0)}</>;
}

// TodoPhaseCell renders one phase of one todo: its status, its progress and how
// long it took. A phase that has never run gets an em-dash rather than an empty
// cell, so "not started" stays visibly different from a cell that failed to
// render.
export function TodoPhaseCell({ todo, phase }: { todo: TodoItem; phase: TodoPhase }) {
  const run = todo.phases?.[phase];
  if (!run) {
    return <span className="text-muted-foreground/50" title={`${phaseMeta(phase).label}: not started`}>—</span>;
  }

  const view = STATE_VIEWS[run.state] ?? STATE_VIEWS.pending;
  const Icon = phaseFailed(run) ? STATE_VIEWS.failed.icon : view.icon;
  const className = phaseFailed(run) ? STATE_VIEWS.failed.className : view.className;
  const progress = phaseProgressLabel(run);
  const live = phaseRunning(run);
  // A queued phase has no start time yet, so there is no elapsed time to show —
  // "0s" would read as "finished instantly" rather than "not started yet".
  const timed = live ? !!run.started_at : (run.duration_ms ?? 0) > 0;

  return (
    <span
      className={`inline-flex items-center gap-1 whitespace-nowrap text-xs ${className}`}
      title={`${phaseMeta(phase).label}: ${run.state}${progress ? ` · ${progress}` : ''}`}
    >
      <Icon className="shrink-0 text-[11px]" />
      {progress && <span className="tabular-nums">{progress}</span>}
      {timed && (
        <span className="tabular-nums text-muted-foreground">
          {live ? <PhaseTimer run={run} /> : <StaticTimer run={run} />}
        </span>
      )}
    </span>
  );
}

// TodoPhaseStrip is the glyph-only form for the split layout's narrow rows:
// the same four phases, no progress and no timer, sized like the plan and
// verification indicators it sits beside.
export function TodoPhaseStrip({ todo, phases }: { todo: TodoItem; phases: TodoPhase[] }) {
  const runs = phases
    .map(phase => ({ phase, run: todo.phases?.[phase] }))
    .filter((entry): entry is { phase: TodoPhase; run: TodoPhaseRun } => !!entry.run);
  if (runs.length === 0) return null;

  return (
    <span className="inline-flex shrink-0 items-center gap-1">
      {runs.map(({ phase, run }) => {
        const failed = phaseFailed(run);
        const view = failed ? STATE_VIEWS.failed : (STATE_VIEWS[run.state] ?? STATE_VIEWS.pending);
        const Icon = view.icon;
        return (
          <Icon
            key={phase}
            className={`text-[11px] ${view.className}`}
            title={`${phaseMeta(phase).label}: ${run.state}`}
          />
        );
      })}
    </span>
  );
}
