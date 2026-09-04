import type { VerifyReport, VerifyState } from '@flanksource/clicky-ui/data';
import type { TodoSessionAttempt, TodoSessionDetailResponse } from '../../types';

// The Verification tab's attempt model: each attempt carries the captain
// VerifyReport its `verify` step produced (or is producing, while live), read
// off the persisted prompt run so a reload never loses the evidence of a
// failed check.

export type AttemptOutcome = VerifyState;

export interface VerificationAttemptEntry {
  attempt: TodoSessionAttempt;
  report: VerifyReport | null;
  /** True while the report is a running snapshot rather than a terminal result. */
  live: boolean;
  outcome: AttemptOutcome;
}

export interface MalformedAttempt {
  promptRunId: string;
  ordinal: number;
  reason: string;
}

// VerificationAttemptSet.malformed is never populated now that
// readAttemptReport reads the server-typed attempt.verification instead of
// the untyped legacy resultJson.definitionOfDone shape (the only source that
// could produce an unreadable payload) — see readAttemptReport. It stays on
// the type because TodoVerificationAttempts.tsx still renders it; dropping it
// end-to-end is a follow-up outside this fix.
export interface VerificationAttemptSet {
  attempts: VerificationAttemptEntry[];
  malformed: MalformedAttempt[];
}

const TERMINAL_ATTEMPT_STATES = new Set(['succeeded', 'failed', 'cancelled', 'errored', 'error']);
const FAILING_OUTCOMES = new Set<AttemptOutcome>(['failed', 'errored', 'timed_out']);

/**
 * Reads one attempt's VerifyReport. `attempt.verification` is now always sent
 * by the server (a report, or null for an attempt that was never verified —
 * see todo_session_detail.go's todoAttemptDetail.Verification), including
 * while the attempt is still running: it is "the newest definition-of-done
 * report", not only the terminal one, so the live snapshot and the terminal
 * result share this one field and are told apart by `report.state`.
 *
 * This used to fall back to the legacy `resultJson.definitionOfDone.{report,
 * progress}` shape for older records that predated `attempt.verification`.
 * That fallback (and the malformed-shape detection it needed, since the
 * legacy field was untyped JSON) is gone now that every attempt the server
 * sends carries a properly typed `verification`.
 */
function readAttemptReport(attempt: TodoSessionAttempt): VerifyReport | null {
  return attempt.verification;
}

function outcomeFor(attempt: TodoSessionAttempt, report: VerifyReport | null): AttemptOutcome {
  if (report) return report.state;
  const state = (attempt.state || '').toLowerCase();
  if (state === 'cancelled') return 'cancelled';
  if (state && !TERMINAL_ATTEMPT_STATES.has(state)) return 'running';
  return 'errored';
}

function attemptTime(attempt: TodoSessionAttempt): number {
  const parsed = Date.parse(attempt.startedAt || attempt.queuedAt || attempt.createdAt || '');
  return Number.isFinite(parsed) ? parsed : 0;
}

/**
 * Every verification attempt this todo has made — the explicit `verify` step
 * and the in-loop DoD a `run` step evaluates — newest first. The endpoint
 * orders attempts oldest-first, so they are reversed here (ordinal is per step
 * kind, so it only breaks ties within one timestamp).
 */
export function verificationAttempts(detail: TodoSessionDetailResponse | null): VerificationAttemptSet {
  const set: VerificationAttemptSet = { attempts: [], malformed: [] };
  if (!detail) return set;

  for (const attempt of detail.attempts) {
    const result = readAttemptReport(attempt);
    if (result === null && attempt.step !== 'verify') continue;

    set.attempts.push({
      attempt,
      report: result,
      live: result?.state === 'running',
      outcome: outcomeFor(attempt, result),
    });
  }
  set.attempts.sort((a, b) => {
    const byTime = attemptTime(b.attempt) - attemptTime(a.attempt);
    return byTime !== 0 ? byTime : b.attempt.ordinal - a.attempt.ordinal;
  });
  set.malformed.sort((a, b) => b.ordinal - a.ordinal);
  return set;
}

/**
 * The Verification tab badge. It counts attempts, not acceptance criteria, and
 * turns red when the newest attempt did not pass — a tab whose last check failed
 * must not look like one that has never run.
 */
export function verificationBadge(set: VerificationAttemptSet): { count: number; failing: boolean; title: string } {
  const count = set.attempts.length;
  if (set.malformed.length > 0) {
    return { count, failing: true, title: `${set.malformed.length} verification attempt(s) could not be read` };
  }
  if (count === 0) return { count: 0, failing: false, title: 'No verification attempts yet' };

  const latest = set.attempts[0];
  const failed = set.attempts.filter(entry => FAILING_OUTCOMES.has(entry.outcome)).length;
  if (latest.outcome === 'running' || latest.outcome === 'queued') {
    return { count, failing: false, title: `Verification attempt #${latest.attempt.ordinal} is running` };
  }
  const failing = FAILING_OUTCOMES.has(latest.outcome);
  const headline = failing ? 'Latest verification failed' : 'Latest verification passed';
  return { count, failing, title: `${headline} — ${failed} of ${count} attempts failed` };
}
