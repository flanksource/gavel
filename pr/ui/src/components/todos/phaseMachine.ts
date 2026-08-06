import { WORKFLOW_PHASES, type SessionTone } from '@flanksource/clicky-ui/ai';
import type { StaticIconComponent } from '@flanksource/clicky-ui';
import type { TodoItem, TodoStatus } from '../../types';

/**
 * The plan -> run -> verify machine behind the todo header.
 *
 * The header used to render Plan and Run as two peer buttons, each with its own
 * duplicate runtime combo and Advanced cog, with verification reachable only
 * from its own tab. Nothing said which step came next. This module answers that:
 * given a todo's status and what has already happened, which phase does the
 * header suggest, and which phases have already run.
 *
 * The machine *ranks* phases, it never withholds one — every phase stays
 * runnable from the dropdown regardless of state.
 *
 * Only the three phases gavel can actually execute are modelled: `plan` and
 * `run` are agent runs (`POST /api/todos/run`), `verify` is the fixture-backed
 * definition-of-done check (`POST /api/todos/verification/run`). There is
 * deliberately no `audit` phase — that needs the named-prompt axis from TODO
 * ff053357, which does not exist yet.
 */

export type PhaseId = 'plan' | 'run' | 'verify';

export type Phase = {
  id: PhaseId;
  label: string;
  icon: StaticIconComponent;
  /** Canonical tone name from the library's Agent Action Icons set. */
  tone: SessionTone;
  /** What entering this phase does, for the button title. */
  title: string;
};

// Glyph, tone and label come from the library's `WORKFLOW_PHASES` rather than
// being chosen here, so a phase looks the same in the header as it does in the
// session viewer. Picking locally would drift the moment either side changed.
export const PHASES: Phase[] = [
  {
    id: 'plan',
    label: WORKFLOW_PHASES.plan.label,
    icon: WORKFLOW_PHASES.plan.icon,
    tone: WORKFLOW_PHASES.plan.tone,
    title: 'Produce a reviewable plan',
  },
  {
    id: 'run',
    label: WORKFLOW_PHASES.run.label,
    icon: WORKFLOW_PHASES.run.icon,
    tone: WORKFLOW_PHASES.run.tone,
    title: 'Implement this todo',
  },
  {
    id: 'verify',
    label: WORKFLOW_PHASES.verify.label,
    icon: WORKFLOW_PHASES.verify.icon,
    tone: WORKFLOW_PHASES.verify.tone,
    title: 'Run the definition-of-done fixture',
  },
];

export function phase(id: PhaseId): Phase {
  const match = PHASES.find(entry => entry.id === id);
  if (!match) throw new Error(`Unknown phase: ${id}`);
  return match;
}

/**
 * Where the todo stands, as the header needs to read it.
 *
 * These are not `TodoStatus` values. `pending` alone cannot choose between Plan
 * and Run — that depends on whether a plan exists — and `unverified` vs a failed
 * check are the same status with different verification outcomes. So the raw
 * status plus the surrounding signals collapse into one of these.
 */
export type PhaseState =
  | 'draft'
  | 'planned'
  | 'review'
  | 'ask'
  | 'running'
  | 'implemented'
  | 'failed'
  | 'verified'
  | 'closed';

/** What the primary control should do. Only one of these is ever offered. */
export type PrimaryAction =
  | { kind: 'phase'; phase: PhaseId; label: string }
  | { kind: 'stop' }
  | { kind: 'review' }
  | { kind: 'answer' };

export type PhaseSignals = {
  status: TodoStatus;
  /** A plan exists to approve or implement. */
  hasPlan: boolean;
  /** An agent session is in flight right now. */
  sessionInProgress: boolean;
  /** The most recent verification attempt did not pass. */
  verificationFailing: boolean;
  /** At least one verification attempt has been recorded. */
  verificationAttempted: boolean;
};

export function phaseSignals(
  todo: TodoItem,
  extra: { sessionInProgress: boolean; verificationFailing: boolean; verificationAttempted: boolean },
): PhaseSignals {
  return {
    status: todo.status,
    hasPlan: !!todo.hasPlan || !!todo.planPath,
    ...extra,
  };
}

/**
 * Collapse gavel's status plus its surrounding signals into one machine state.
 *
 * A live session outranks everything: whatever the stored status says, the only
 * action that applies while an agent is running is stopping it.
 */
export function phaseState(signals: PhaseSignals): PhaseState {
  if (signals.sessionInProgress) return 'running';
  switch (signals.status) {
    case 'draft':
      return 'draft';
    case 'review':
      return 'review';
    case 'ask':
      return 'ask';
    case 'in_progress':
      // No live session despite the status: the run ended without updating it.
      // Treat it as landed work whose definition of done is unchecked.
      return 'implemented';
    case 'failed':
      return 'failed';
    case 'verified':
      return 'verified';
    case 'completed':
    case 'skipped':
      return 'closed';
    case 'unverified':
      return signals.verificationFailing ? 'failed' : 'implemented';
    case 'pending':
      if (signals.verificationFailing) return 'failed';
      if (signals.verificationAttempted) return 'verified';
      return signals.hasPlan ? 'planned' : 'draft';
  }
}

type FlowEntry = {
  primary: PrimaryAction;
  /** Phases already carried out, in order — the progress ticks. */
  done: PhaseId[];
  /** One line naming why this state suggests what it suggests. */
  note: string;
};

// `label` overrides the phase's own verb where the state changes its meaning:
// re-entering a phase that already ran is a re-run, not a first run.
const FLOW: Record<PhaseState, FlowEntry> = {
  draft: {
    primary: { kind: 'phase', phase: 'plan', label: 'Plan' },
    done: [],
    note: 'Nothing has run yet, so planning comes first.',
  },
  planned: {
    primary: { kind: 'phase', phase: 'run', label: 'Run' },
    done: ['plan'],
    note: 'A plan exists and is waiting to be implemented.',
  },
  review: {
    primary: { kind: 'review' },
    done: ['plan'],
    note: 'A plan is waiting on a human — approve or reject it below.',
  },
  ask: {
    primary: { kind: 'answer' },
    done: ['plan'],
    note: 'The agent asked a blocking question; nothing may start until it is answered.',
  },
  running: {
    primary: { kind: 'stop' },
    done: [],
    note: 'A run is in flight; the only action that applies is stopping it.',
  },
  implemented: {
    primary: { kind: 'phase', phase: 'verify', label: 'Verify' },
    done: ['plan', 'run'],
    note: 'Work landed but the definition of done has not been checked.',
  },
  failed: {
    primary: { kind: 'phase', phase: 'plan', label: 'Replan' },
    done: ['plan', 'run', 'verify'],
    note: 'Verification failed — replanning is the default, but re-running is one click away.',
  },
  verified: {
    primary: { kind: 'phase', phase: 'run', label: 'Re-run' },
    done: ['plan', 'run', 'verify'],
    note: 'Done. Any phase can still be re-entered.',
  },
  closed: {
    primary: { kind: 'phase', phase: 'run', label: 'Re-run' },
    done: ['plan', 'run', 'verify'],
    note: 'Closed. Re-running reopens the work.',
  },
};

export function primaryAction(state: PhaseState): PrimaryAction {
  return FLOW[state].primary;
}

export function completedPhases(state: PhaseState): PhaseId[] {
  return FLOW[state].done;
}

export function stateNote(state: PhaseState): string {
  return FLOW[state].note;
}

/**
 * Every phase except the suggested one, in flow order. The dropdown always
 * carries all of them: the machine ranks phases, it never withholds one.
 */
export function otherPhases(state: PhaseState): Phase[] {
  const suggested = FLOW[state].primary;
  const skip = suggested.kind === 'phase' ? suggested.phase : undefined;
  return PHASES.filter(entry => entry.id !== skip);
}

/**
 * The verb for entering `id` from `state` — "Replan" and "Re-run" rather than
 * "Plan" and "Run" once that phase has already happened.
 */
export function phaseVerb(state: PhaseState, id: PhaseId): string {
  const target = phase(id);
  if (!completedPhases(state).includes(id)) return target.label;
  return id === 'plan' ? 'Replan' : `Re-${target.label.toLowerCase()}`;
}

export type StepStatus = 'done' | 'current' | 'todo';

export function stepStatus(state: PhaseState, id: PhaseId): StepStatus {
  const suggested = FLOW[state].primary;
  if (suggested.kind === 'phase' && suggested.phase === id) return 'current';
  return completedPhases(state).includes(id) ? 'done' : 'todo';
}
