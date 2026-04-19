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
  author?: string;
  any?: boolean;
  bots?: boolean;
  // GitHub org logins the user has chosen to hide from the chooser and
  // exclude from default-org resolution. Persists across daemon restarts.
  ignoredOrgs?: string[];
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
  rateLimit?: RateLimit;
  // Sparse map keyed by `${repo}#${number}`. A PR is unread iff its key
  // appears here. Absent key = read. Server omits the field entirely when
  // every PR is read.
  unread?: Record<string, boolean>;
  syncStatus?: Record<string, PRSyncStatus>;
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
}

export interface PRDetail {
  pr?: PRInfo;
  runs?: Record<string, WorkflowRun>;
  comments?: PRComment[];
  gavelResults?: GavelResultsSummary;
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
