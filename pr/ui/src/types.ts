export interface FailedCheck {
  name: string;
  detailsUrl?: string;
  failedSteps?: string[];
  logTail?: string;
}

export interface CheckSummary {
  passed: number;
  failed: number;
  running: number;
  pending: number;
  failures?: FailedCheck[];
}

export interface PRItem {
  number: number;
  title: string;
  author: string;
  authorAvatarUrl?: string;
  // True when the author is a GitHub App bot whose login lacks a "bot" suffix
  // (e.g. renovate); lets the @bots chip group it without the suffix heuristic.
  authorIsApp?: boolean;
  repo: string;
  repoAvatarUrl?: string;
  repoHomepageUrl?: string;
  source: string;
  target: string;
  state: string;
  isDraft: boolean;
  reviewDecision?: string;
  mergeable?: string;
  url: string;
  updatedAt: string;
  isCurrent?: boolean;
  ahead?: number;
  behind?: number;
  checkStatus?: CheckSummary;
  route_path?: string;
}

export interface SearchConfig {
  repos: string[];
  all?: boolean;
  org?: string;
  // GitHub org logins the user has chosen to hide from the chooser and
  // exclude from default-org resolution. Persists across daemon restarts.
  ignoredOrgs?: string[];
}

// Project associates one or more repos with a local workspace directory where
// Gavel discovers a Procfile. Mirrors pr/ui.Project / projectInfo.
export interface Project {
  name: string;
  dir: string;
  repos: string[];
  hasProcfile?: boolean;
}

// ProcProcess mirrors procfile.ProcState — one supervised process.
export interface ProcProcess {
  name: string;
  command: string;
  pid?: number;
  status: string;
  started?: string;
  restarts: number;
  exitCode?: number;
  logFile: string;
  ports?: number[];
  // Live resource sample of the process group. openFiles is -1 where the
  // platform cannot report it. All omitted/zero for a stopped process.
  cpuPercent?: number;
  memoryRss?: number;
  openFiles?: number;
  // Per-process breakdown of the process group (leader + descendants).
  tree?: ProcNode[];
}

// ProcNode mirrors procfile.ProcNode — one process in a supervised group's tree.
export interface ProcNode {
  pid: number;
  ppid: number;
  command: string;
  cpuPercent?: number;
  memoryRss?: number;
  openFiles?: number;
}

// ProcStatus mirrors pr/ui.procStatus — a project's Procfile supervision state.
// hasProcfile=false is the normal "no Procfile here" state, not an error.
export interface ProcStatus {
  hasProcfile: boolean;
  running: boolean;
  supervisorPid?: number;
  processes?: ProcProcess[];
  // profiles declared in the Procfile; profile is the active one (running
  // supervisor's, else the .gavel.yaml default).
  profiles?: string[];
  profile?: string;
  error?: string;
}

export interface RateLimit {
  limit: number;
  remaining: number;
  used: number;
  reset: number;
  resource: string;
}

export interface Snapshot {
  prs: PRItem[];
  fetchedAt: string;
  nextFetchIn: number;
  incremental: boolean;
  paused: boolean;
  error?: string;
  config: SearchConfig;
  // Login of the authenticated GitHub user, used to resolve the @me author
  // filter client-side. Empty until the auth probe completes.
  viewer?: string;
  // True once the server has learned of any bot author, so the @bots chip stays
  // available even while bots are excluded from the fetch.
  botsAvailable?: boolean;
  // The server's current bot-fetch state; the UI only posts a change when the
  // @bots chip disagrees with this.
  includeBots?: boolean;
  rateLimit?: RateLimit;
  // Sparse map keyed by `${repo}#${number}`. A PR is unread iff its key
  // appears here. Absent key = read. Server omits the field entirely when
  // every PR is read.
  unread?: Record<string, boolean>;
  syncStatus?: Record<string, PRSyncStatus>;
  gavelResults?: Record<string, GavelResultsSummary>;
}

// Sync status types

export type SyncState = 'queued' | 'syncing' | 'up-to-date' | 'out-of-date' | 'error';

export interface PRSyncStatus {
  state: SyncState;
  lastSynced?: string;
  error?: string;
  phase?: string;
}

// Detail API types

export interface PRInfo {
  number: number;
  title: string;
  author: { login: string; name?: string; avatarUrl?: string };
  headRefName: string;
  baseRefName: string;
  state: string;
  isDraft: boolean;
  reviewDecision: string;
  mergeable: string;
  url: string;
  statusCheckRollup?: StatusCheck[];
}

export interface StatusCheck {
  name: string;
  status: string;
  conclusion: string;
  workflowName?: string;
  detailsUrl?: string;
}

export interface Step {
  name: string;
  status: string;
  conclusion: string;
  number: number;
  logs?: string;
}

export interface Job {
  databaseId: number;
  name: string;
  status: string;
  conclusion: string;
  startedAt?: string;
  completedAt?: string;
  url?: string;
  steps?: Step[];
  logs?: string;
}

export interface WorkflowRun {
  databaseId: number;
  name: string;
  status: string;
  conclusion: string;
  url?: string;
  jobs?: Job[];
}

export interface PRComment {
  id: number;
  body: string;
  author: string;
  avatarUrl?: string;
  url: string;
  createdAt: string;
  path?: string;
  line?: number;
  isResolved?: boolean;
  isOutdated?: boolean;
  severity?: string;
  botType?: string;
}

export interface GavelResultsSummary {
  // stickyId is the gavel sticky-comment id, e.g. "gavel-test-pg15".
  // Empty for the legacy single-artifact path or for an aggregate.
  stickyId?: string;
  artifactId: number;
  artifactUrl: string;
  testsPassed: number;
  testsFailed: number;
  testsSkipped: number;
  testsTotal: number;
  lintViolations: number;
  lintLinters: number;
  hasBench: boolean;
  benchRegressions?: number;
  error?: string;
  topFailures?: TestFailure[];
  topLintViolations?: LintViolation[];
}

export interface TestFailure {
  name: string;
  suite?: string;
  file?: string;
  line?: number;
  message?: string;
  details?: string;
}

export interface LintViolation {
  linter: string;
  file?: string;
  line?: number;
  rule?: string;
  message?: string;
}

export interface PRDetail {
  pr?: PRInfo;
  runs?: Record<string, WorkflowRun>;
  comments?: PRComment[];
  // One summary per gavel sticky comment on the PR (typically one per
  // matrix shard). Order matches the order of the sticky comments.
  gavelResults?: GavelResultsSummary[];
  error?: string;
  // Progressive loading state (set by frontend, not backend)
  runsLoading?: boolean;
  gavelLoading?: boolean;
}

// Activity API types

export type ActivityKind = 'rest' | 'graphql' | 'search' | 'favicon';

export interface ActivityEntry {
  timestamp: string;
  method: string;
  url: string;
  kind: ActivityKind;
  statusCode: number;
  durationNs: number;
  sizeBytes: number;
  fromCache: boolean;
  error?: string;
}

export interface ActivityKindStats {
  total: number;
  cacheHits: number;
  errors: number;
  totalBytes: number;
  totalNs: number;
}

export interface ActivityStats {
  total: number;
  cacheHits: number;
  errors: number;
  totalBytes: number;
  totalNs: number;
  byKind: Record<string, ActivityKindStats>;
}

export interface ActivitySnapshot {
  entries: ActivityEntry[];
  stats: ActivityStats;
}

export interface CacheStatus {
  enabled: boolean;
  driver: string;
  dsnSource: string;
  dsnMasked: string;
  retentionSec: number;
  counts: Record<string, number>;
  error?: string;
}

export type Severity = 'ok' | 'degraded' | 'down';

// ComponentStatus matches pr/ui.ComponentStatus — one component (db / github)
// of the aggregated /api/status response. `detail` is component-specific
// extra data the UI can surface in a tooltip.
export interface ComponentStatus {
  severity: Severity;
  message: string;
  detail?: unknown;
}

// HealthStatus is the /api/status payload. Drives both the CLI
// (`gavel system status`) and the PR UI's header status indicator from a
// single source of truth.
export interface HealthStatus {
  overall: Severity;
  database: ComponentStatus;
  github: ComponentStatus;
  checkedAt: string;
}

// Org matches github.Org — a lightweight entry in the header's org chooser
// dropdown. AvatarURL comes straight from the GitHub API so it can be used
// as an <img src>.
export interface Org {
  login: string;
  avatarUrl: string;
}
