import type { Snapshot } from '@flanksource/clicky-ui/data';

// TestRunView mirrors the Go pr/ui.testRunView served by /api/tests.
export interface TestRunView {
  runId: string;
  kind: 'test' | 'lint' | 'test+lint';
  started?: string;
  ended?: string;
  repo?: string;
  sha?: string;
  frameworks?: string[];
  passed: number;
  failed: number;
  skipped: number;
  warned: number;
  total: number;
  lintViolations: number;
  lintLinters: number;
}

export interface ProjectRuns {
  name: string;
  dir: string;
  runs: TestRunView[];
}

export interface TestRunsResponse {
  projects: ProjectRuns[];
}

// LintViolation / LinterResult mirror gavel's models.Violation and
// linters.LinterResult JSON, which /api/tests/run returns under `lint`.
export interface LintViolation {
  file?: string;
  line?: number;
  column?: number;
  message?: string;
  code?: string;
  source?: string;
  rule?: { pattern?: string; package?: string; method?: string } | null;
  Severity?: string;
}

export interface LinterResult {
  linter: string;
  work_dir?: string;
  success?: boolean;
  skipped?: boolean;
  // null when the linter errored before producing findings (Go marshals a nil slice as null).
  violations: LintViolation[] | null;
  error?: string;
  file_count?: number;
  rule_count?: number;
}

// RunSnapshot is the clicky test-runner Snapshot plus gavel's lint section,
// which the clicky type omits.
export interface RunSnapshot extends Omit<Snapshot, 'status' | 'tests'> {
  status: Snapshot['status'] & { lint_run?: boolean };
  // Go marshals a nil tests slice as null for lint-only runs.
  tests: Snapshot['tests'] | null;
  lint?: LinterResult[];
}

// RunFailure / RunArtifact mirror gavel's fixtures.RunFailure and
// fixtures.RunArtifact: a fixture runner step records its engine output in the
// .gavel store and carries only these counts, the head of the failure list, and
// the run id inline. Fetch the full tree with fetchRunSnapshot(runId).
export interface RunFailure {
  name: string;
  suite?: string;
  status?: string;
  message?: string;
}

export interface RunArtifact {
  run_id: string;
  path?: string;
  kind: 'test' | 'lint' | 'test+lint';
  total: number;
  passed: number;
  failed: number;
  warned: number;
  skipped: number;
  lint_violations?: number;
  linters?: string[];
  failures?: RunFailure[];
  /** Failures beyond the recorded cap; the full list lives in the snapshot. */
  truncated?: number;
  /** Set when the engine failed before producing results. */
  error?: string;
}

/**
 * Reads one run snapshot from the .gavel store. The Tests tab addresses a
 * workspace by project name, the TODO surfaces by directory — exactly one. An
 * empty `dir` is meaningful (the server's default workspace), so the mode is
 * chosen by which key is supplied, not by whether its value is blank.
 */
export async function fetchRunSnapshot({ project, dir, runId }: { project?: string; dir?: string; runId: string }): Promise<RunSnapshot> {
  if ((project === undefined) === (dir === undefined)) {
    throw new Error('fetchRunSnapshot requires exactly one of project or dir');
  }
  const params = new URLSearchParams({ runId });
  if (project !== undefined) params.set('project', project);
  if (dir !== undefined) params.set('dir', dir);
  const response = await fetch(`/api/tests/run?${params.toString()}`);
  const body = await response.json().catch(() => null);
  if (!response.ok) {
    throw new Error((body && typeof body.error === 'string' && body.error) || `Failed to load run ${runId}`);
  }
  return body as RunSnapshot;
}
