import { describe, expect, it } from 'vitest';
import type { VerifyReport } from '@flanksource/clicky-ui/data';
import type { TodoSessionAttempt, TodoSessionDetailResponse } from '../../types';
import { verificationAttempts, verificationBadge } from './verificationReport';

function attempt(ordinal: number, overrides: Partial<TodoSessionAttempt> = {}): TodoSessionAttempt {
  return {
    promptRunId: `run-${ordinal}`,
    ordinal,
    step: 'verify',
    requested: {},
    resolved: {},
    status: 'succeeded',
    processActive: false,
    state: 'succeeded',
    phase: 'done',
    queuedAt: `2026-07-30T09:0${ordinal}:00Z`,
    startedAt: `2026-07-30T09:0${ordinal}:00Z`,
    admissionSessionId: `admission-${ordinal}`,
    createdAt: `2026-07-30T09:0${ordinal}:00Z`,
    updatedAt: `2026-07-30T09:0${ordinal}:30Z`,
    verification: null,
    ...overrides,
  };
}

function detail(attempts: TodoSessionAttempt[]): TodoSessionDetailResponse {
  return { attempts, diagnostics: [] };
}

function report(overrides: Partial<VerifyReport> = {}): VerifyReport {
  return {
    kind: 'todo',
    ran: true,
    passed: true,
    state: 'passed',
    summary: { total: 1, passed: 1, failed: 0, warned: 0, skipped: 0, pending: 0, running: 0, timedout: 0 },
    ...overrides,
  };
}

describe('verificationAttempts', () => {
  it('reads attempt.verification directly once the server populates it', () => {
    const set = verificationAttempts(detail([attempt(1, { verification: report({ passed: false, state: 'failed' }) })]));
    expect(set.attempts).toHaveLength(1);
    expect(set.attempts[0]?.outcome).toBe('failed');
    expect(set.malformed).toHaveLength(0);
  });

  it('reads a running snapshot the same way, distinguished only by report.state', () => {
    const running = report({ state: 'running', passed: false });
    const set = verificationAttempts(detail([
      attempt(1, { state: 'running', status: 'running', verification: running }),
    ]));
    expect(set.attempts[0]?.report).toEqual(running);
    expect(set.attempts[0]?.live).toBe(true);
    expect(set.attempts[0]?.outcome).toBe('running');
  });

  it('lists attempts newest-first regardless of the order the endpoint returns', () => {
    const set = verificationAttempts(detail([
      attempt(1, { verification: report() }),
      attempt(2, { verification: report({ passed: false, state: 'failed' }) }),
    ]));
    expect(set.attempts.map(entry => entry.attempt.ordinal)).toEqual([2, 1]);
  });

  it('excludes a run-step attempt that never evaluated a definition of done', () => {
    const set = verificationAttempts(detail([attempt(1, { step: 'run' })]));
    expect(set.attempts).toHaveLength(0);
    expect(set.malformed).toHaveLength(0);
  });
});

describe('verificationBadge', () => {
  it('is neutral with no attempts', () => {
    expect(verificationBadge(verificationAttempts(detail([])))).toEqual({
      count: 0,
      failing: false,
      title: 'No verification attempts yet',
    });
  });

  it('tints red when the newest attempt failed', () => {
    const badge = verificationBadge(verificationAttempts(detail([
      attempt(1, { verification: report() }),
      attempt(2, { verification: report({ passed: false, state: 'failed' }) }),
    ])));
    expect(badge).toEqual({ count: 2, failing: true, title: 'Latest verification failed — 1 of 2 attempts failed' });
  });

  it('stays neutral when the newest attempt passed even though an older one failed', () => {
    const badge = verificationBadge(verificationAttempts(detail([
      attempt(2, { verification: report() }),
      attempt(1, { verification: report({ passed: false, state: 'failed' }) }),
    ])));
    expect(badge).toEqual({ count: 2, failing: false, title: 'Latest verification passed — 1 of 2 attempts failed' });
  });

  it('reports a running attempt without flagging it as failing', () => {
    const badge = verificationBadge(verificationAttempts(detail([
      attempt(1, { state: 'running', status: 'running', verification: report({ state: 'running' }) }),
    ])));
    expect(badge).toEqual({ count: 1, failing: false, title: 'Verification attempt #1 is running' });
  });
});
