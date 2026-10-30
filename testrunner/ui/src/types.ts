import type { ProcessDescriptor as ClickyProcessDescriptor, ProcessNodeRow, ProcessSampleRow } from '@flanksource/clicky-ui/data';

export interface FixtureExecution {
  key: string;
  kind: string;
  state: string;
  command_ms?: number;
  violations?: string[];
  // pids of processes this fixture step started: services as they launch,
  // commands once they exit.
  pids?: number[];
}

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
  end_line?: number;
  framework?: string;
  duration?: number; // nanoseconds
  fixture_profile?: FixtureProfile;
  sql_profile?: FixtureSQLProfile;
  fixture?: FixtureExecution;
  skipped?: boolean;
  // aborted marks a test that never ran because the run stopped after an
  // earlier failure — distinct from an intentional skip.
  aborted?: boolean;
  failed?: boolean;
  passed?: boolean;
  // warned marks a node that completed with a non-blocking warning — amber,
  // shown to the operator but never a failure (a warned child never fails its
  // parent). Distinct from failed.
  warned?: boolean;
  pending?: boolean;
  // running marks a node whose execution has started but not yet finished —
  // distinct from pending (queued/not-yet-started). Renderers show running with
  // an active spinner and pending with a static hollow icon.
  running?: boolean;
  timed_out?: boolean;
  cached?: boolean;
  task_id?: string;
  can_stop?: boolean;
  stdout?: string;
  stderr?: string;
  children?: Test[];
  summary?: TestSummary;
  attempts?: TestAttempt[];
  context?: GoTestContext | GinkgoContext | FixtureContext | TaskContext;
  // detail is provider-owned structured JSON. Legacy snapshots may still carry
  // a ClickyDocument shape and are detected by the detail panel.
  detail?: unknown;
  failure_detail?: FailureDetail;
  // progress is live in-flight progress for a still-running node (e.g. an
  // intake step being consumed). Cleared on completion.
  progress?: { phase?: string; done: number; total: number };

  // route_path is the host-assigned stable path used for selection and the URL.
  route_path?: string;
  // own_name is the node's name before single-child chains were joined into it
  // ("pkg/ > TestFoo" is TestFoo's row); frontend-only, set by the host.
  own_name?: string;
  // export_path is the path the server's export of this node is requested at
  // (its uncollapsed ancestry); frontend-only, set by the host.
  export_path?: string;
}

export type Severity = 'error' | 'warning' | 'info';

export interface Violation {
  file?: string;
  raw_file?: string;
  line?: number;
  column?: number;
  end_line?: number;
  end_column?: number;
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
  // status is set only while the linter is in flight; a final result omits it.
  status?: 'pending' | 'running';
  // run_id is the Captain prompt run producing the result; session_id is the
  // agent session behind it, kept so a replayed snapshot still names it.
  run_id?: string;
  session_id?: string;
}

export interface TestAttempt {
  sequence: number;
  run_kind?: string;
  started?: string;
  ended?: string;
  duration?: number;
  pid?: number;
  // process_key names the attempt's process in the run's process_tree
  // (`<pid>@<startUnix>`); its live metrics are that key's process_samples.
  process_key?: string;
  command?: string;
  framework?: string;
  exit_code?: number;
  passed?: boolean;
  failed?: boolean;
  skipped?: boolean;
  pending?: boolean;
  timed_out?: boolean;
  message?: string;
  stdout?: string;
  stderr?: string;
  stack_trace?: string;
  cpu_percent?: number;
  rss?: number;
  goroutine_count?: number;
  // shared_process marks a pid, CPU and RSS that belong to the go test batch
  // the package ran in, not to the package alone.
  shared_process?: boolean;
}

export interface TestSummary {
  Total: number;
  Passed: number;
  Failed: number;
  Warned?: number;
  Aborted?: number;
  Skipped: number;
  Pending: number;
  Running?: number;
  Duration: number;
}

export interface Snapshot {
  metadata?: RunMeta;
  git?: SnapshotGit;
  status: SnapshotStatus;
  tests: Test[];
  lint?: LinterResult[];
  bench?: BenchComparison;
  fixture_benchmark?: FixtureBenchmarkState;
  performance?: FixturePerformance;
  error?: string;
}

export interface FixtureExecutionNode {
  key: string;
  name: string;
  kind: string;
  state: string;
  duration?: number;
  children?: FixtureExecutionNode[];
}

export interface FixtureExecutionSnapshot {
  version: number;
  iteration: number;
  state: string;
  root: FixtureExecutionNode;
}

export interface FixtureBenchmarkEntry {
  key: string;
  name: string;
  kind: string;
  status: string;
  duration_ms: number;
  command_ms?: number;
  profile_samples?: number;
  profile?: FixtureProfile;
  sql_profile?: FixtureSQLProfile;
  violations?: string[];
}

export interface FixtureSQLProfile {
  path: string;
  query_count: number;
  slow_query_count: number;
  total_duration_ms: number;
  max_query_ms: number;
  statements?: FixtureSQLStatement[];
}

export interface FixtureSQLStatement {
  sql: string;
  params: string[];
  duration_ms: number;
  rows: number;
  slow: boolean;
  error: boolean;
}

export interface ProfileDiskIO {
  disk_read_bytes: number;
  disk_write_bytes: number;
  sample_count: number;
  missing_processes?: number;
}

export interface FixtureProfile {
  scope: 'process_tree' | 'run_tree';
  pid?: number;
  sample_count: number;
  peak_cpu_percent: number;
  peak_memory_percent: number;
  peak_rss_bytes: number;
  disk_io?: ProfileDiskIO;
}

export interface FixtureBenchmarkDelta {
  key: string;
  name: string;
  kind: string;
  baseline_ms: number;
  current_ms: number;
  delta_ms: number;
  delta_pct?: number;
  disk_read_bytes_delta?: number;
  disk_write_bytes_delta?: number;
}

export interface FixtureProcess {
  pid: number;
  ppid?: number;
  command?: string;
  status?: string;
  cpuPercent?: number;
  memoryPercent?: number;
  rssBytes?: number;
  io?: { diskReadBytes: number; diskWriteBytes: number };
}

export interface FixtureProfileSample {
  sampled_at: string;
  cpu_percent: number;
  memory_percent: number;
  rss_bytes: number;
  processes: FixtureProcess[];
}

export interface FixtureBenchmarkReport {
  version: number;
  mode: string;
  started_at: string;
  finished_at: string;
  duration_ms: number;
  status: string;
  files: string[];
  phases: FixtureBenchmarkEntry[];
  fixtures: FixtureBenchmarkEntry[];
  limits?: {
    max_deviation_pct?: number;
    max_duration_ms?: number;
    max_rss_bytes?: number;
    max_disk_read_bytes?: number;
    max_disk_write_bytes?: number;
    max_sql_duration_ms?: number;
    max_sql_query_ms?: number;
    max_sql_queries?: number;
    max_slow_sql?: number;
  };
  violations?: { key: string; name: string; reasons: string[] }[];
  comparison?: {
    baseline: string;
    deltas: FixtureBenchmarkDelta[];
    added?: string[];
    missing?: string[];
    unmeasured?: string[];
    peak_cpu_percent_delta?: number;
    peak_rss_bytes_delta?: number;
  };
  profile?: {
    interval_ms: number;
    peak_cpu_percent: number;
    peak_memory_percent: number;
    peak_rss_bytes: number;
    disk_io?: ProfileDiskIO;
    samples: FixtureProfileSample[];
  };
}

export interface FixtureBenchmarkState {
  mode: string;
  progress?: FixtureExecutionSnapshot;
  report?: FixtureBenchmarkReport;
  artifact_path?: string;
}

export interface FixturePerformance {
  mode: string;
  status: string;
  duration_ms: number;
  artifact_path?: string;
  profile?: FixtureBenchmarkReport['profile'];
  comparison?: FixtureBenchmarkReport['comparison'];
  limits?: FixtureBenchmarkReport['limits'];
  violations?: FixtureBenchmarkReport['violations'];
}

export interface RunMeta {
  version?: string;
  sequence: number;
  kind?: 'initial' | 'rerun' | string;
  started?: string;
  ended?: string;
  args?: Record<string, unknown>;
  pid?: number;
  command?: string;
  frameworks?: string[];
  exit_code?: number;
  timed_out?: boolean;
  // run_id names the run's trace store, .gavel/traces/<run_id>.
  run_id?: string;
  // process_trace names the run's process capture in that store.
  process_trace?: ProcessTraceRefs;
}

export interface SnapshotGit {
  repo?: string;
  root?: string;
  sha?: string;
}

export interface SnapshotStatus {
  running: boolean;
  lint_run?: boolean;
  // traces is the run's trace store as the UI reaches it; absent when the run kept none.
  traces?: TracesStatus;
  stop_supported?: boolean;
  test_edit_supported?: boolean;
  stopped?: boolean;
  stop_message?: string;
}

// ProcessTraceNode is one process_tree record (commons-db tracing/process):
// cpuPercent and peakCpuPercent are absent until a full sampling window has
// been observed, and stackTransports lists the ways a stack of the process can
// be collected.
export type ProcessTraceNode = ProcessNodeRow;

// ProcessDescriptor is one file or socket a process holds open
// (clicky process.Descriptor), as process_samples stores it.
export type ProcessDescriptor = ClickyProcessDescriptor;

// ProcessSample is one process_samples record: a process's usage at one tick.
export type ProcessSample = ProcessSampleRow;

// TraceRef names a window of a trace stream: the records after seq afterSeq
// through seq toSeq (0 = unbounded) of kind's stream (commons-db tracing.TraceRef).
export interface TraceRef {
  kind: string;
  stream: string;
  afterSeq?: number;
  toSeq?: number;
}

// TraceResult is one armed trace's window of a fixture, with its record count,
// the kind's summary of it, or the error that kept it from being read.
export interface TraceResult extends TraceRef {
  // run_id names the run whose store holds the window, .gavel/traces/<run_id>.
  run_id: string;
  count: number;
  summary?: Record<string, unknown>;
  error?: string;
}

// ProcessTraceRefs are the windows of the run's process capture: the process
// tree, the per-tick samples (resources, disk IO, facet counts and lists), the
// stack dumps and, when the run ingests Go profiles, goprof.
export interface ProcessTraceRefs {
  tree: TraceRef;
  samples: TraceRef;
  stacks: TraceRef;
  goprof?: TraceRef;
}

// ProcessTraceSessions are the session ids of the same four streams while the
// capture runs: a checkpoint of stacks takes a stack dump, one of samples
// forces a full facet reading.
export interface ProcessTraceSessions {
  tree: string;
  samples: string;
  stacks: string;
  goprof?: string;
}

// RunTraces is one run's store as the UI reaches it.
export interface RunTraces {
  // mount is where the store is served through query ("/query/<run id>"); absent when it is not.
  mount?: string;
  // unavailable says why a run that kept a store is not serving it (pruned since).
  unavailable?: string;
}

// TracesStatus is the stores of the runs the server shows: a replay of several
// snapshots shows several runs, and each trace window names its run.
export interface TracesStatus {
  runs: Record<string, RunTraces>;
  // process_run is the run whose process capture (process) the Processes tab reads.
  process_run?: string;
  process?: ProcessTraceRefs;
  // sessions are the running capture's session ids, which a checkpoint names;
  // absent once the run has ended (and for a stored run), when nothing can be
  // checkpointed any more.
  sessions?: ProcessTraceSessions;
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
  labels?: string[];
  num_attempts?: number;
  parallel_process?: number;
  report_entries?: GinkgoReportEntry[];
}

export interface GinkgoReportEntry {
  name: string;
  value?: unknown;
  representation?: string;
  file?: string;
  line?: number;
}

// TaskContext mirrors the shape emitted by the Go server for virtual
// clicky-task pseudo-tests (testrunner/ui/handler.go: taskSnapshotToTest).
export interface TaskContext {
  status?: string;
  type?: string;
  duration?: number;
}

export interface FixtureContext {
  command?: string;
  exit_code?: number;
  cwd?: string;
  cel_expression?: string;
  cel_trace?: string;
  cel_vars?: Record<string, any>;
  expected?: any;
  actual?: any;
  // traces are the windows of the traces the fixture armed.
  traces?: TraceResult[];
}

export type FailureKind = 'gomega' | 'panic' | 'go_test' | 'raw';

export interface FailureDetail {
  kind?: FailureKind;
  summary?: string;
  matcher?: string;
  expected?: string;
  actual?: string;
  location?: string;
  stack?: string;
}

// RerunRequest mirrors the Go testui.RerunRequest payload accepted by
// POST /api/rerun.
export interface RerunRequest {
  package_paths?: string[];
  work_dir?: string;
  test_name?: string;
  suite?: string[];
  framework?: string;
  lint?: boolean;
  lint_files?: string[];
  lint_linters?: string[];
}

export type TestEditAction = 'skip' | 'delete';
export type TestEditScope = 'test' | 'file';

export interface TestEditRequest {
  action: TestEditAction;
  scope: TestEditScope;
  framework?: string;
  work_dir?: string;
  package_path?: string;
  file?: string;
  line?: number;
  test_name?: string;
  suite?: string[];
}

export interface TestEditResponse {
  file: string;
  action: TestEditAction;
  scope: TestEditScope;
  changed: boolean;
  edited?: number;
  removed?: number;
  message?: string;
}
