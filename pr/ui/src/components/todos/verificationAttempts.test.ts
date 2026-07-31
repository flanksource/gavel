import { describe, expect, it } from 'vitest';
import type { TodoSessionAttempt, TodoSessionDetailResponse } from '../../types';
import type { RunArtifact } from '../tests/types';
import {
  attemptOutputSteps,
  attemptRunArtifacts,
  attemptTabs,
  defaultAttemptTab,
  verificationAttempts,
  verificationBadge,
  type VerificationFixtureResult,
} from './verificationAttempts';

function testArtifact(overrides: Partial<RunArtifact> = {}): RunArtifact {
  return { run_id: 'run-2026-07-30T09-00-00Z-suite', kind: 'test', total: 12, passed: 10, failed: 2, warned: 0, skipped: 0, ...overrides };
}

function lintArtifact(overrides: Partial<RunArtifact> = {}): RunArtifact {
  return { run_id: 'run-2026-07-30T09-00-01Z-lint', kind: 'lint', total: 0, passed: 0, failed: 0, warned: 0, skipped: 0, lint_violations: 3, ...overrides };
}

function attempt(ordinal: number, overrides: Partial<TodoSessionAttempt> = {}): TodoSessionAttempt {
  return {
    promptRunId: `prompt-run-${ordinal}`,
    ordinal,
    step: 'verify',
    requested: {},
    resolved: {},
    status: 'succeeded',
    processActive: false,
    state: 'succeeded',
    phase: 'done',
    queuedAt: '2026-07-30T09:00:00Z',
    admissionSessionId: `admission-${ordinal}`,
    createdAt: '2026-07-30T09:00:00Z',
    updatedAt: '2026-07-30T09:01:00Z',
    ...overrides,
  } as TodoSessionAttempt;
}

function dod(passed: boolean, results: VerificationFixtureResult[] = [], checklist: unknown[] = []) {
  return { definitionOfDone: { ran: true, passed, output: { results, checklist } } };
}

function detail(attempts: TodoSessionAttempt[]): TodoSessionDetailResponse {
  return { attempts, diagnostics: [] };
}

describe('verificationAttempts', () => {
  it('includes verify attempts and run attempts that carry a definition of done', () => {
    const set = verificationAttempts(
      detail([
        attempt(3, { step: 'verify', resultJson: dod(false) }),
        attempt(2, { step: 'run', resultJson: dod(true) }),
        attempt(1, { step: 'plan', resultJson: { plan: { status: 'new' } } }),
      ])
    );

    expect(set.attempts.map((entry) => entry.attempt.ordinal)).toEqual([3, 2]);
    expect(set.attempts.map((entry) => entry.attempt.step)).toEqual(['verify', 'run']);
    expect(set.malformed).toEqual([]);
  });

  // The endpoint sorts attempts by creation time ascending; the tab reads newest first.
  it('reverses the oldest-first order the endpoint returns', () => {
    const set = verificationAttempts(
      detail([
        attempt(1, { queuedAt: '2026-07-30T09:00:00Z', resultJson: dod(true) }),
        attempt(2, { queuedAt: '2026-07-30T10:00:00Z', resultJson: dod(false) }),
        attempt(3, { queuedAt: '2026-07-30T11:00:00Z', resultJson: dod(false) }),
      ])
    );

    expect(set.attempts.map((entry) => entry.attempt.ordinal)).toEqual([3, 2, 1]);
  });

  it('lists a verify attempt that produced no definition of done', () => {
    const set = verificationAttempts(detail([attempt(1, { step: 'verify', state: 'failed', resultJson: {} })]));

    expect(set.attempts).toHaveLength(1);
    expect(set.attempts[0].outcome).toBe('errored');
  });

  it.each([
    ['a failed verdict', { state: 'succeeded', resultJson: dod(false) }, 'failed'],
    ['a passing verdict', { state: 'succeeded', resultJson: dod(true) }, 'passed'],
    ['a non-terminal state', { state: 'running', resultJson: dod(false) }, 'running'],
    ['a cancelled state', { state: 'cancelled', resultJson: dod(false) }, 'cancelled'],
    ['a run that never evaluated', { state: 'failed', resultJson: { definitionOfDone: { ran: false, passed: false } } }, 'errored'],
  ])('maps %s to %s', (_label, overrides, expected) => {
    const set = verificationAttempts(detail([attempt(1, overrides as Partial<TodoSessionAttempt>)]));
    expect(set.attempts[0].outcome).toBe(expected);
  });

  // One unreadable payload must not blank the whole tab.
  it('collects malformed payloads instead of dropping or throwing', () => {
    const set = verificationAttempts(
      detail([
        attempt(2, { resultJson: { definitionOfDone: 'not an object' } }),
        attempt(1, { resultJson: dod(true) }),
      ])
    );

    expect(set.attempts.map((entry) => entry.attempt.ordinal)).toEqual([1]);
    expect(set.malformed).toEqual([{ promptRunId: 'prompt-run-2', ordinal: 2, reason: expect.stringContaining('definitionOfDone') }]);
  });

  it('reports a definition of done without the boolean verdict as malformed', () => {
    const set = verificationAttempts(detail([attempt(1, { resultJson: { definitionOfDone: { ran: 'yes', passed: 1 } } })]));
    expect(set.malformed).toHaveLength(1);
    expect(set.attempts).toEqual([]);
  });
});

describe('verificationBadge', () => {
  it('is neutral with no attempts', () => {
    expect(verificationBadge(verificationAttempts(detail([])))).toEqual({ count: 0, failing: false, title: 'No verification attempts yet' });
  });

  it('turns red when the newest attempt failed and counts every failure', () => {
    const badge = verificationBadge(
      verificationAttempts(detail([attempt(3, { resultJson: dod(false) }), attempt(2, { resultJson: dod(false) }), attempt(1, { resultJson: dod(true) })]))
    );

    expect(badge.count).toBe(3);
    expect(badge.failing).toBe(true);
    expect(badge.title).toBe('Latest verification failed — 2 of 3 attempts failed');
  });

  it('stays neutral when the newest attempt passed after earlier failures', () => {
    const badge = verificationBadge(verificationAttempts(detail([attempt(2, { resultJson: dod(true) }), attempt(1, { resultJson: dod(false) })])));

    expect(badge.failing).toBe(false);
    expect(badge.title).toBe('Latest verification passed — 1 of 2 attempts failed');
  });

  it('does not flag a still-running attempt as a failure', () => {
    const badge = verificationBadge(verificationAttempts(detail([attempt(1, { state: 'running', resultJson: dod(false) })])));
    expect(badge.failing).toBe(false);
    expect(badge.title).toContain('is running');
  });

  it('turns red when an attempt could not be read', () => {
    const badge = verificationBadge(verificationAttempts(detail([attempt(1, { resultJson: { definitionOfDone: [] } })])));
    expect(badge.failing).toBe(true);
    expect(badge.title).toContain('could not be read');
  });
});

describe('attempt sub-tabs', () => {
  const steps: VerificationFixtureResult[] = [
    { name: 'run the suite', type: 'test', run: testArtifact() },
    { name: 'lint the repo', type: 'lint', run: lintArtifact() },
    { name: 'check the binary', type: 'exec', command: 'gavel --version', exit_code: 0 },
  ];

  it('splits runner steps by engine and leaves the rest to Output', () => {
    const entry = verificationAttempts(detail([attempt(1, { resultJson: dod(false, steps) })])).attempts[0];

    expect(attemptRunArtifacts(entry).map((item) => item.kind)).toEqual(['test', 'lint']);
    expect(attemptOutputSteps(entry).map((step) => step.name)).toEqual(['check the binary']);
  });

  it('counts tests and violations on the available tabs', () => {
    const entry = verificationAttempts(detail([attempt(1, { executionSessionId: 'session-1', resultJson: dod(false, steps) })])).attempts[0];
    const tabs = attemptTabs(entry);

    expect(tabs.test).toEqual({ available: true, count: 12 });
    expect(tabs.lint).toEqual({ available: true, count: 3 });
    expect(tabs.session).toEqual({ available: true });
    expect(tabs.output).toEqual({ available: true, count: 1 });
  });

  // An unavailable tab must say why rather than render an empty pane.
  it('explains why each missing tab is unavailable', () => {
    const entry = verificationAttempts(detail([attempt(1, { resultJson: dod(true, [{ name: 'echo', type: 'exec' }]) })])).attempts[0];
    const tabs = attemptTabs(entry);

    expect(tabs.test).toEqual({ available: false, reason: 'This attempt ran no test step' });
    expect(tabs.lint).toEqual({ available: false, reason: 'This attempt ran no lint step' });
    expect(tabs.session).toEqual({ available: false, reason: 'This attempt recorded no agent session' });
  });

  // Steps recorded before run artifacts existed have no `run`: classify them
  // honestly as Output rather than sending them to an empty Test pane.
  it('treats a test-typed step without a run artifact as output', () => {
    const entry = verificationAttempts(detail([attempt(1, { resultJson: dod(false, [{ name: 'legacy suite', type: 'test' }]) })])).attempts[0];

    expect(attemptTabs(entry).test.available).toBe(false);
    expect(attemptOutputSteps(entry).map((step) => step.name)).toEqual(['legacy suite']);
    expect(defaultAttemptTab(entry)).toBe('output');
  });

  it('opens a failed attempt on its test results and a checklist-only attempt on output', () => {
    const failed = verificationAttempts(detail([attempt(2, { resultJson: dod(false, steps) })])).attempts[0];
    expect(defaultAttemptTab(failed)).toBe('test');

    const checklistOnly = verificationAttempts(detail([attempt(1, { resultJson: dod(true, [], [{ item: 'docs updated', passed: true }]) })])).attempts[0];
    expect(defaultAttemptTab(checklistOnly)).toBe('output');
  });
});
