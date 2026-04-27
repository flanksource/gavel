export interface Test {
  name: string;
  package?: string;
  package_path?: string;
  work_dir?: string;
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
  timed_out?: boolean;
  stdout?: string;
  stderr?: string;
  children?: Test[];
  summary?: TestSummary;
  context?: GoTestContext | GinkgoContext | FixtureContext;

  // Synthetic node markers (frontend-only). Used to render lint results as tree nodes.
  kind?: 'lint-root' | 'lint-folder' | 'linter' | 'violation' | 'lint-file' | 'lint-rule' | 'lint-rule-group';
  violation?: Violation;
  violations?: Violation[];
  noFileViolations?: Violation[];
  linter?: LinterResult;
  linterName?: string;
  ruleName?: string;
  target_path?: string;
  route_path?: string;
}

export type Severity = 'error' | 'warning' | 'info';

export interface Violation {
  file?: string;
  raw_file?: string;
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
  command?: string;   // resolved executable, e.g. "eslint"
  args?: string[];    // argv without the command name
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
  metadata?: RunMeta;
  git?: SnapshotGit;
  status: SnapshotStatus;
  tests: Test[];
  lint?: LinterResult[];
  bench?: BenchComparison;
  diagnostics?: DiagnosticsSnapshot;
}

export interface RunMeta {
  version?: string;
  sequence: number;
  kind?: 'initial' | 'rerun' | string;
  started?: string;
  ended?: string;
  args?: Record<string, unknown>;
}

export interface SnapshotGit {
  repo?: string;
  root?: string;
  sha?: string;
}

export interface SnapshotStatus {
  running: boolean;
  lint_run?: boolean;
  diagnostics_available?: boolean;
}

export interface DiagnosticsSnapshot {
  root?: ProcessNode;
  generated_at?: string;
}

export interface ProcessNode {
  pid: number;
  ppid?: number;
  name?: string;
  command?: string;
  status?: string;
  cpu_percent?: number;
  rss?: number;
  vms?: number;
  open_files?: number;
  is_root?: boolean;
  children?: ProcessNode[];
  stack_capture?: StackCapture;
}

export interface ProcessDetails {
  pid: number;
  ppid?: number;
  name?: string;
  command?: string;
  status?: string;
  cpu_percent?: number;
  rss?: number;
  vms?: number;
  open_files?: number;
  is_root?: boolean;
  stack_capture?: StackCapture;
}

export interface StackCapture {
  status: 'ready' | 'unsupported' | 'error';
  supported: boolean;
  text?: string;
  error?: string;
  collected_at?: string;
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

export interface RunCounts {
  total: number;
  passed: number;
  failed: number;
  skipped: number;
  pending: number;
}

export interface RunIndexEntry {
  name: string;
  path: string;
  pointer?: string;
  modified: string;
  sha?: string;
  started?: string;
  ended?: string;
  counts?: RunCounts;
  lint?: number;
  error?: string;
}
