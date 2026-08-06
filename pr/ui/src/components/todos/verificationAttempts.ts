import type { TodoSessionAttempt, TodoSessionDetailResponse } from '../../types';
import type { Test } from '@flanksource/clicky-ui/data';
import type { RunArtifact } from '../tests/types';

export type ExecutionState = 'queued' | 'running' | 'passed' | 'failed' | 'errored' | 'warned' | 'skipped' | 'cancelled' | 'timed_out';

export interface ExecutionOrigin {
  file?: string;
  section_path?: string;
  kind?: string;
  table_index?: number;
  row_index?: number;
  line?: number;
}

export interface ExecutionNode {
  key: string;
  name: string;
  kind: string;
  state: ExecutionState;
  origin?: ExecutionOrigin;
  started_at?: string;
  finished_at?: string;
  duration?: number;
  done?: number;
  total?: number;
  error?: string;
  children?: ExecutionNode[];
}

export interface ExecutionSummary {
  total: number;
  queued?: number;
  running?: number;
  passed?: number;
  failed?: number;
  errored?: number;
  warned?: number;
  skipped?: number;
  cancelled?: number;
  timed_out?: number;
}

export interface ExecutionSnapshot {
  version: number;
  iteration: number;
  state: ExecutionState;
  started_at?: string;
  ended_at?: string;
  summary: ExecutionSummary;
  root: ExecutionNode;
}

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
  cel_trace?: string;
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
  ran?: boolean;
  passed?: boolean;
  output?: VerificationOutput;
  progress?: ExecutionSnapshot;
}

export type AttemptOutcome = 'passed' | 'failed' | 'running' | 'cancelled' | 'errored';

export type AttemptTab = 'fixtures' | 'test' | 'lint' | 'session' | 'output';

export interface VerificationAttempt {
  attempt: TodoSessionAttempt;
  dod: DefinitionOfDone | null;
  outcome: AttemptOutcome;
  progress?: ExecutionSnapshot;
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

function readExecutionNode(value: unknown, path: string): ExecutionNode | string {
  if (!isRecord(value)) return `${path} is not an object`;
  for (const field of ['key', 'name', 'kind', 'state'] as const) {
    if (typeof value[field] !== 'string') return `${path}.${field} is not a string`;
  }
  if (value.children !== undefined) {
    if (!Array.isArray(value.children)) return `${path}.children is not an array`;
    for (let index = 0; index < value.children.length; index += 1) {
      const child = readExecutionNode(value.children[index], `${path}.children[${index}]`);
      if (typeof child === 'string') return child;
    }
  }
  return value as unknown as ExecutionNode;
}

function readExecutionSnapshot(value: unknown): ExecutionSnapshot | undefined | string {
  if (value === undefined || value === null) return undefined;
  if (!isRecord(value)) return 'definitionOfDone.progress is not an object';
  if (value.version !== 1) return 'definitionOfDone.progress has an unsupported version';
  if (typeof value.iteration !== 'number' || typeof value.state !== 'string') {
    return 'definitionOfDone.progress is missing iteration/state';
  }
  if (!isRecord(value.summary) || typeof value.summary.total !== 'number') {
    return 'definitionOfDone.progress.summary is missing total';
  }
  const root = readExecutionNode(value.root, 'definitionOfDone.progress.root');
  if (typeof root === 'string') return root;
  return value as unknown as ExecutionSnapshot;
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
  const progress = readExecutionSnapshot(raw.progress);
  if (typeof progress === 'string') return progress;
  if (typeof raw.ran !== 'boolean' || typeof raw.passed !== 'boolean') {
    const state = (attempt.state || attempt.status || '').toLowerCase();
    if (progress && !TERMINAL_STATES.has(state)) return raw as unknown as DefinitionOfDone;
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
      progress: dod?.progress,
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

function executionNodeTest(node: ExecutionNode): Test {
  const test: Test = {
    name: node.name,
    framework: node.kind,
    task_id: node.key,
    file: node.origin?.file,
    line: node.origin?.line,
    duration: node.duration,
    message: node.error,
    detail: node,
    children: node.children?.map(executionNodeTest),
  };
  if (node.total && (node.state === 'running' || node.state === 'queued')) {
    test.progress = { phase: node.kind, done: node.done ?? 0, total: node.total };
  }
  applyExecutionState(test, node.state);
  return test;
}

function applyExecutionState(test: Test, state: ExecutionState): void {
  if (state === 'queued') test.pending = true;
  else if (state === 'running') test.running = true;
  else if (state === 'passed') test.passed = true;
  else if (state === 'failed' || state === 'errored') test.failed = true;
  else if (state === 'warned') test.warned = true;
  else if (state === 'timed_out') test.timed_out = true;
  else test.skipped = true;
}

function resultTest(result: VerificationFixtureResult, index: number, parentKey = 'result'): Test {
  const key = `${parentKey}:${index}:${result.name || result.type || 'fixture'}`;
  const test: Test = {
    name: result.name || result.type || 'Fixture',
    framework: result.type || 'fixture',
    task_id: key,
    duration: result.duration,
    message: result.error,
    command: result.command,
    work_dir: result.cwd,
    stdout: result.stdout,
    stderr: result.stderr,
    detail: result,
    children: result.children
      ?.filter(isRecord)
      .map((child, childIndex) => resultTest(child as unknown as VerificationFixtureResult, childIndex, key)),
  };
  applyResultStatus(test, result.status);
  return test;
}

function applyResultStatus(test: Test, status?: string): void {
  const normalized = (status || '').toLowerCase();
  if (normalized === 'pass' || normalized === 'passed') test.passed = true;
  else if (normalized === 'fail' || normalized === 'failed' || normalized === 'error' || normalized === 'errored') test.failed = true;
  else if (normalized === 'warn' || normalized === 'warned') test.warned = true;
  else if (normalized === 'skip' || normalized === 'skipped' || normalized === 'cancelled') test.skipped = true;
  else if (normalized === 'timeout' || normalized === 'timed_out') test.timed_out = true;
  else if (normalized === 'running') test.running = true;
  else if (normalized === 'queued' || normalized === 'pending') test.pending = true;
}

function checklistTest(checklist: VerificationChecklistItem[]): Test | null {
  if (checklist.length === 0) return null;
  const children = checklist.map((item, index): Test => ({
    name: item.item || `Criterion ${index + 1}`,
    framework: 'checklist',
    task_id: `acceptance-criteria:${index}`,
    message: item.message,
    passed: item.passed === true,
    failed: item.passed === false,
    pending: typeof item.passed !== 'boolean',
    detail: item,
  }));
  return {
    name: 'Acceptance criteria',
    framework: 'checklist',
    task_id: 'acceptance-criteria',
    children,
    passed: children.every((child) => child.passed),
    failed: children.some((child) => child.failed),
    pending: children.some((child) => child.pending),
  };
}

/** The shared fixture execution tree, live while running and rebuilt from terminal evidence afterward. */
export function attemptFixtureTests(entry: VerificationAttempt): Test[] {
  if (entry.progress) return entry.progress.root.children?.map(executionNodeTest) ?? [];
  const tests = entry.steps.map((step, index) => resultTest(step, index));
  const checklist = checklistTest(entry.checklist);
  if (checklist) tests.push(checklist);
  return tests;
}

export function attemptTabs(entry: VerificationAttempt): Record<AttemptTab, AttemptTabState> {
  const artifacts = attemptRunArtifacts(entry);
  const fixtures = attemptFixtureTests(entry);
  const tests = artifacts.filter((item) => item.kind === 'test');
  const lints = artifacts.filter((item) => item.kind === 'lint');
  const outputs = attemptOutputSteps(entry);
  const sessionId = entry.attempt.executionSessionId || entry.attempt.providerSessionId;

  return {
    fixtures: fixtures.length > 0
      ? { available: true, count: entry.progress?.summary.total ?? fixtures.length }
      : { available: false, reason: 'This attempt recorded no fixture execution tree' },
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
  if (entry.outcome === 'running' && tabs.fixtures.available) return 'fixtures';
  const failed = entry.outcome === 'failed' || entry.outcome === 'errored';
  const byFailure: AttemptTab[] = failed
    ? ['test', 'lint', 'output', 'fixtures', 'session']
    : ['test', 'output', 'lint', 'fixtures', 'session'];
  return byFailure.find((tab) => tabs[tab].available) ?? 'output';
}
