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
  rule?: { pattern?: string } | null;
  severity?: string;
}

export interface LinterResult {
  linter: string;
  work_dir?: string;
  success?: boolean;
  skipped?: boolean;
  violations: LintViolation[];
  file_count?: number;
  rule_count?: number;
}

// RunSnapshot is the clicky test-runner Snapshot plus gavel's lint section,
// which the clicky type omits.
export interface RunSnapshot extends Snapshot {
  lint?: LinterResult[];
}
