import { describe, expect, it } from 'vitest';
import { WORKFLOW_PHASES, type WorkflowPhase } from '@flanksource/clicky-ui/ai';
import type { TodoStatus } from '../../types';
import {
  ALL_PHASES,
  PHASES,
  completedPhases,
  otherPhases,
  phase,
  phaseState,
  phaseVerb,
  primaryAction,
  stepStatus,
  type PhaseSignals,
  type PhaseState,
} from './phaseMachine';

const ALL_STATES: PhaseState[] = [
  'draft',
  'planned',
  'review',
  'ask',
  'running',
  'implemented',
  'failed',
  'verified',
  'closed',
];

function signals(overrides: Partial<PhaseSignals> = {}): PhaseSignals {
  return {
    status: 'pending',
    hasPlan: false,
    sessionInProgress: false,
    verificationFailing: false,
    verificationAttempted: false,
    ...overrides,
  };
}

describe('phase catalog', () => {
  // The point of sourcing plan/run/verify from the library's Agent Action Icons
  // set is that the header cannot drift from the session view. A locally
  // re-picked glyph would look fine in isolation and wrong beside a session.
  it('takes each phase glyph, tone and label from the library', () => {
    for (const entry of PHASES) {
      const canonical = WORKFLOW_PHASES[entry.id as WorkflowPhase];
      expect(entry.icon).toBe(canonical.icon);
      expect(entry.tone).toBe(canonical.tone);
      expect(entry.label).toBe(canonical.label);
    }
  });

  // Only the phases gavel can actually execute are modelled. `audit` needs the
  // named-prompt axis from ff053357, so offering it would be a dead control.
  it('models exactly the three runnable phases', () => {
    expect(PHASES.map(entry => entry.id)).toEqual(['plan', 'run', 'verify']);
  });
});

describe('phaseState', () => {
  it('reads a fresh todo as draft', () => {
    expect(phaseState(signals({ status: 'draft' }))).toBe('draft');
  });

  // `pending` alone cannot choose between Plan and Run — a plan's existence is
  // what separates them, which is why the raw status is not enough.
  it('splits pending on whether a plan exists', () => {
    expect(phaseState(signals({ status: 'pending', hasPlan: false }))).toBe('draft');
    expect(phaseState(signals({ status: 'pending', hasPlan: true }))).toBe('planned');
  });

  it('reads unverified work as implemented until a check fails', () => {
    expect(phaseState(signals({ status: 'unverified' }))).toBe('implemented');
    expect(phaseState(signals({ status: 'unverified', verificationFailing: true }))).toBe('failed');
  });

  // A live session outranks the stored status: whatever it says, the only
  // action that applies while an agent is running is stopping it.
  it('lets a live session outrank every stored status', () => {
    for (const status of ['draft', 'pending', 'review', 'unverified', 'verified'] as TodoStatus[]) {
      expect(phaseState(signals({ status, sessionInProgress: true }))).toBe('running');
    }
  });

  // in_progress with no live session is a run that ended without updating the
  // status. Reading it as "running" would offer a Stop with nothing to stop.
  it('treats a stale in_progress as landed work', () => {
    expect(phaseState(signals({ status: 'in_progress' }))).toBe('implemented');
  });

  it('routes the two human-decision statuses to their own states', () => {
    expect(phaseState(signals({ status: 'review' }))).toBe('review');
    expect(phaseState(signals({ status: 'ask' }))).toBe('ask');
  });
});

describe('primaryAction', () => {
  it('suggests plan first and verify after work lands', () => {
    expect(primaryAction('draft')).toEqual({ kind: 'phase', phase: 'plan', label: 'Plan' });
    expect(primaryAction('implemented')).toEqual({ kind: 'phase', phase: 'verify', label: 'Verify' });
  });

  // A failed check defaults to replanning rather than blindly re-running the
  // same implementation that just failed its definition of done.
  it('defaults a failed verification to replanning', () => {
    expect(primaryAction('failed')).toEqual({ kind: 'phase', phase: 'plan', label: 'Replan' });
  });

  // review/ask must route through the banner's approve/reject/answer flow, not
  // be bypassed by re-triggering a run from the header.
  it('never offers a phase while a human decision is pending', () => {
    expect(primaryAction('review').kind).toBe('review');
    expect(primaryAction('ask').kind).toBe('answer');
  });

  it('offers only stop while running', () => {
    expect(primaryAction('running')).toEqual({ kind: 'stop' });
  });
});

describe('otherPhases', () => {
  // The machine ranks phases, it never withholds one: whatever the state, every
  // RUNNABLE phase not already on the primary stays reachable from the dropdown.
  // The universe is ALL_PHASES, not PHASES — triage is runnable but is not a
  // pipeline step, so it is offered without ever appearing in the progress strip.
  it('always covers every phase the primary does not', () => {
    for (const state of ALL_STATES) {
      const suggested = primaryAction(state);
      const offered = otherPhases(state).map(entry => entry.id);
      const expected = ALL_PHASES.map(entry => entry.id).filter(
        id => !(suggested.kind === 'phase' && suggested.phase === id),
      );
      expect(offered).toEqual(expected);
    }
  });

  it('offers triage from every state without ever ticking it as done', () => {
    for (const state of ALL_STATES) {
      const suggested = primaryAction(state);
      // Triage is never a primary suggestion, so it is always in the dropdown.
      expect(suggested).not.toEqual({ kind: 'phase', phase: 'triage', label: 'Triage' });
      expect(otherPhases(state).map(entry => entry.id)).toContain('triage');
      expect(completedPhases(state)).not.toContain('triage');
      expect(phaseVerb(state, 'triage')).toBe('Triage');
    }
  });

  // The progress strip renders PHASES; an auxiliary action in it would read as a
  // pipeline step the todo has not reached yet.
  it('keeps triage out of the pipeline strip', () => {
    expect(PHASES.map(entry => entry.id)).toEqual(['plan', 'run', 'verify']);
  });
});

describe('phaseVerb', () => {
  it('names a first entry plainly and a repeat as a re-run', () => {
    expect(phaseVerb('draft', 'plan')).toBe('Plan');
    expect(phaseVerb('verified', 'plan')).toBe('Replan');
    expect(phaseVerb('verified', 'run')).toBe('Re-run');
    expect(phaseVerb('verified', 'verify')).toBe('Re-verify');
  });
});

describe('stepStatus', () => {
  it('marks completed phases done and the suggestion current', () => {
    expect(stepStatus('implemented', 'plan')).toBe('done');
    expect(stepStatus('implemented', 'run')).toBe('done');
    expect(stepStatus('implemented', 'verify')).toBe('current');
  });

  // After a failed verification the machine loops back: plan becomes current
  // again even though it has already run once.
  it('returns to plan as current after a failed verification', () => {
    expect(stepStatus('failed', 'plan')).toBe('current');
    expect(stepStatus('failed', 'verify')).toBe('done');
  });

  it('leaves nothing current while running', () => {
    for (const entry of PHASES) {
      expect(stepStatus('running', entry.id)).not.toBe('current');
    }
  });
});

describe('completedPhases', () => {
  it('reports nothing done in draft and everything after verification', () => {
    expect(completedPhases('draft')).toEqual([]);
    expect(completedPhases('verified')).toEqual(['plan', 'run', 'verify']);
  });
});

describe('phase', () => {
  it('throws on an unknown id rather than returning a blank phase', () => {
    // @ts-expect-error deliberately outside PhaseId
    expect(() => phase('audit')).toThrow(/Unknown phase/);
  });
});
