export interface Test {
  name: string;
  package?: string;
  package_path?: string;
  command?: string;
  suite?: string[];
  message?: string;
  file?: string;
  line?: number;
  framework?: string;
  duration?: number; // nanoseconds
  skipped?: boolean;
  failed?: boolean;
  passed?: boolean;
  pending?: boolean;
  stdout?: string;
  stderr?: string;
  children?: Test[];
  summary?: TestSummary;
  context?: GoTestContext | GinkgoContext | FixtureContext;

  // Synthetic node markers (frontend-only). Used to render lint results as tree nodes.
  kind?: 'lint-root' | 'linter' | 'violation' | 'lint-file' | 'lint-rule';
  violation?: Violation;
  violations?: Violation[];
  linter?: LinterResult;
  linterName?: string;
  ruleName?: string;
}

export type Severity = 'error' | 'warning' | 'info';

export interface Violation {
  file?: string;
  line?: number;
  column?: number;
  message?: string;
  source?: string;
  rule?: { method?: string; description?: string };
  severity?: Severity;
  code?: string;
}

export interface LinterResult {
  linter: string;
  work_dir?: string;
  success: boolean;
  skipped?: boolean;
  timed_out?: boolean;
  duration: number; // nanoseconds
  violations: Violation[];
  raw_output?: string;
  error?: string;
  file_count?: number;
  rule_count?: number;
}

export interface TestSummary {
  Total: number;
  Passed: number;
  Failed: number;
  Skipped: number;
  Pending: number;
  Duration: number;
}

export interface Snapshot {
  tests: Test[];
  lint?: LinterResult[];
  lint_run?: boolean;
  bench?: BenchComparison;
  done: boolean;
}

export interface BenchDelta {
  name: string;
  package?: string;
  base_mean: number;
  base_stddev?: number;
  head_mean: number;
  head_stddev?: number;
  delta_pct: number;
  p_value?: number;
  samples?: number;
  significant?: boolean;
  only_in?: 'base' | 'head';
}

export interface BenchComparison {
  base_label?: string;
  head_label?: string;
  threshold: number;
  deltas: BenchDelta[];
  geomean_delta: number;
  has_regression: boolean;
}

export interface GoTestContext {
  parent_test?: string;
  import_path?: string;
}

export interface GinkgoContext {
  suite_description?: string;
  suite_path?: string;
  failure_location?: string;
}

export interface FixtureContext {
  command?: string;
  exit_code?: number;
  cwd?: string;
  cel_expression?: string;
  cel_vars?: Record<string, any>;
  expected?: any;
  actual?: any;
}
