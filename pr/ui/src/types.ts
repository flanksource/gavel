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
}

export interface PRDetail {
  pr?: PRInfo;
  runs?: Record<string, WorkflowRun>;
  comments?: PRComment[];
  error?: string;
}

// Activity API types

export type ActivityKind = 'rest' | 'graphql' | 'search';

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
