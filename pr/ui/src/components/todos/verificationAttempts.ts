import type { TodoSessionAttempt, TodoSessionDetailResponse } from '../../types';
import type { RunArtifact } from '../tests/types';

/**
 * Mirrors of gavel's fixtures.FixtureResult, types.VerificationOutput and
 * todos.DoDOutcome as they appear under a prompt run's
 * result_json.definitionOfDone. The runtime shape is unvalidated JSON, so every
 * selector here treats a malformed payload as data to report, never as a
 * precondition to assume.
 */
export interface VerificationFixtureResult {
  name?: string;
  type?: string;
  status?: string;
  duration?: number;
  error?: string;
  expected?: unknown;
  actual?: unknown;
  cel_result?: boolean;
  cel_expression?: string;
  cel_vars?: Record<string, unknown>;
  command?: string;
  cwd?: string;
  stdout?: string;
  stderr?: string;
  exit_code?: number;
  metadata?: Record<string, unknown>;
  children?: unknown[];
  run?: RunArtifact;
}

export interface VerificationChecklistItem {
  item?: string;
  passed?: boolean;
  message?: string;
}

export interface VerificationOutput {
  results?: VerificationFixtureResult[];
  checklist?: VerificationChecklistItem[];
  summary?: Record<string, unknown>;
}

export interface DefinitionOfDone {
  ran: boolean;
  passed: boolean;
  output?: VerificationOutput;
}

export type AttemptOutcome = 'passed' | 'failed' | 'running' | 'cancelled' | 'errored';

export type AttemptTab = 'test' | 'lint' | 'session' | 'output';

export interface VerificationAttempt {
  attempt: TodoSessionAttempt;
  dod: DefinitionOfDone | null;
  outcome: AttemptOutcome;
  steps: VerificationFixtureResult[];
  checklist: VerificationChecklistItem[];
}

export interface MalformedAttempt {
  promptRunId: string;
  ordinal: number;
  reason: string;
}

export interface VerificationAttemptSet {
  attempts: VerificationAttempt[];
  malformed: MalformedAttempt[];
}

export interface AttemptTabState {
  available: boolean;
  count?: number;
  reason?: string;
}

const TERMINAL_STATES = new Set(['succeeded', 'failed', 'cancelled', 'errored', 'error']);

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

/**
 * Reads the definition-of-done payload off an attempt. Returns undefined when
 * the attempt simply has none, and a string when the payload is present but not
 * the shape the contract promises — the caller reports that rather than
 * silently dropping the attempt.
 */
function readDefinitionOfDone(attempt: TodoSessionAttempt): DefinitionOfDone | undefined | string {
  const raw = attempt.resultJson?.definitionOfDone;
  if (raw === undefined || raw === null) return undefined;
  if (!isRecord(raw)) return `definitionOfDone is ${Array.isArray(raw) ? 'an array' : typeof raw}, expected an object`;
  if (typeof raw.ran !== 'boolean' || typeof raw.passed !== 'boolean') {
    return 'definitionOfDone is missing the boolean ran/passed verdict';
  }
  const output = raw.output;
  if (output !== undefined && output !== null) {
    if (!isRecord(output)) return 'definitionOfDone.output is not an object';
    if (output.results !== undefined && output.results !== null && !Array.isArray(output.results)) {
      return 'definitionOfDone.output.results is not an array';
    }
    if (output.checklist !== undefined && output.checklist !== null && !Array.isArray(output.checklist)) {
      return 'definitionOfDone.output.checklist is not an array';
    }
  }
  return raw as unknown as DefinitionOfDone;
}

function attemptOutcome(attempt: TodoSessionAttempt, dod: DefinitionOfDone | null): AttemptOutcome {
  const state = (attempt.state || '').toLowerCase();
  if (state === 'cancelled') return 'cancelled';
  if (state && !TERMINAL_STATES.has(state)) return 'running';
  if (dod?.ran) return dod.passed ? 'passed' : 'failed';
  return 'errored';
}

/**
 * The attempts the Verification tab lists: every prompt run carrying a
 * definition of done — the explicit `verify` step and the in-loop DoD a `run`
 * step evaluates. The endpoint orders attempts oldest-first; the tab reads
 * newest-first, so they are reversed here (ordinal is per step kind, so it only
 * breaks ties within one timestamp).
 */
export function verificationAttempts(detail: TodoSessionDetailResponse | null): VerificationAttemptSet {
  const set: VerificationAttemptSet = { attempts: [], malformed: [] };
  if (!detail) return set;

  for (const attempt of detail.attempts) {
    const parsed = readDefinitionOfDone(attempt);
    if (typeof parsed === 'string') {
      set.malformed.push({ promptRunId: attempt.promptRunId, ordinal: attempt.ordinal, reason: parsed });
      continue;
    }
    if (parsed === undefined && attempt.step !== 'verify') continue;

    const dod = parsed ?? null;
    set.attempts.push({
      attempt,
      dod,
      outcome: attemptOutcome(attempt, dod),
      steps: dod?.output?.results ?? [],
      checklist: dod?.output?.checklist ?? [],
    });
  }
  set.attempts.sort((a, b) => {
    const byTime = attemptTime(b.attempt) - attemptTime(a.attempt);
    return byTime !== 0 ? byTime : b.attempt.ordinal - a.attempt.ordinal;
  });
  set.malformed.sort((a, b) => b.ordinal - a.ordinal);
  return set;
}

function attemptTime(attempt: TodoSessionAttempt): number {
  const parsed = Date.parse(attempt.startedAt || attempt.queuedAt || attempt.createdAt || '');
  return Number.isFinite(parsed) ? parsed : 0;
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
  const failed = set.attempts.filter((entry) => entry.outcome === 'failed' || entry.outcome === 'errored').length;
  if (latest.outcome === 'running') {
    return { count, failing: false, title: `Verification attempt #${latest.attempt.ordinal} is running` };
  }
  const failing = latest.outcome === 'failed' || latest.outcome === 'errored';
  const headline = failing ? 'Latest verification failed' : 'Latest verification passed';
  return { count, failing, title: `${headline} — ${failed} of ${count} attempts failed` };
}

/** The runner steps of an attempt, split by which engine produced them. */
export function attemptRunArtifacts(entry: VerificationAttempt): { kind: 'test' | 'lint'; step: VerificationFixtureResult; artifact: RunArtifact }[] {
  const out: { kind: 'test' | 'lint'; step: VerificationFixtureResult; artifact: RunArtifact }[] = [];
  for (const step of entry.steps) {
    const artifact = step.run;
    if (!artifact?.run_id) continue;
    out.push({ kind: artifact.kind === 'lint' ? 'lint' : 'test', step, artifact });
  }
  return out;
}

/**
 * Steps rendered on the Output tab: everything that is not a runner step.
 * A test-typed step recorded before run artifacts existed has no `run`, so it is
 * honestly classified here rather than sent to an empty Test pane.
 */
export function attemptOutputSteps(entry: VerificationAttempt): VerificationFixtureResult[] {
  return entry.steps.filter((step) => !step.run?.run_id);
}

export function attemptTabs(entry: VerificationAttempt): Record<AttemptTab, AttemptTabState> {
  const artifacts = attemptRunArtifacts(entry);
  const tests = artifacts.filter((item) => item.kind === 'test');
  const lints = artifacts.filter((item) => item.kind === 'lint');
  const outputs = attemptOutputSteps(entry);
  const sessionId = entry.attempt.executionSessionId || entry.attempt.providerSessionId;

  return {
    test: tests.length > 0
      ? { available: true, count: tests.reduce((sum, item) => sum + item.artifact.total, 0) }
      : { available: false, reason: 'This attempt ran no test step' },
    lint: lints.length > 0
      ? { available: true, count: lints.reduce((sum, item) => sum + (item.artifact.lint_violations ?? 0), 0) }
      : { available: false, reason: 'This attempt ran no lint step' },
    session: sessionId
      ? { available: true }
      : { available: false, reason: 'This attempt recorded no agent session' },
    output: outputs.length > 0 || entry.checklist.length > 0
      ? { available: true, count: outputs.length + entry.checklist.length }
      : { available: false, reason: 'This attempt produced no command output or checklist' },
  };
}

/** Opens on the evidence most likely to explain the verdict. */
export function defaultAttemptTab(entry: VerificationAttempt): AttemptTab {
  const tabs = attemptTabs(entry);
  const failed = entry.outcome === 'failed' || entry.outcome === 'errored';
  const byFailure: AttemptTab[] = failed ? ['test', 'lint', 'output', 'session'] : ['test', 'output', 'lint', 'session'];
  return byFailure.find((tab) => tabs[tab].available) ?? 'output';
}
