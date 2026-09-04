import type { AISpecRuntimeValue, SessionUIMessage } from '@flanksource/clicky-ui/ai';
import type { Test, VerifyReport } from '@flanksource/clicky-ui/data';
import type { LinterResult } from '@flanksource/gavel/testrunner';

// Gavel artifact details ride the wire in gavel's own shapes so the dashboard
// renders them with the testrunner's TestNode / LintView rather than a parallel
// set of row components. Test/TestSummary come from clicky-ui/data (the shared
// test-runner model the Verification tab's VerificationResults also renders),
// not gavel's own testrunner package, so both surfaces read one shape.
export type { Test, TestSummary } from '@flanksource/clicky-ui/data';
export type { LinterResult, Violation } from '@flanksource/gavel/testrunner';

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
  // Absent when this workspace's todo store could not be reached — see error.
  // A project is still listed (and its processes still managed) in that case, so
  // never read a missing todoCounts as "zero todos".
  todoCounts?: TodoCounts;
  error?: string;
}

export type TodoStatus = 'draft' | 'pending' | 'in_progress' | 'review' | 'ask' | 'completed' | 'failed' | 'unverified' | 'verified' | 'skipped';
export type TodoPriority = 'high' | 'medium' | 'low';
// Row density for the todo lists: 'comfortable' is the two-line default,
// 'compact' collapses each todo onto a single line.
export type TodoDensity = 'comfortable' | 'compact';
// Grouping dimension for the todo lists: 'workspace' is the default per-workspace
// grouping; 'severity' buckets by priority and 'age' by last activity, both
// across all workspaces. 'none' is a single flat list.
export type TodoGroupBy = 'workspace' | 'severity' | 'age' | 'none';
// Layout for the dashboard's Todos tab: 'split' is the master-detail default
// (list in the AppShell body sidebar, detail beside it); 'full' drops the
// sidebar for a full-width table and a full-page detail.
export type TodoLayout = 'split' | 'full';
export interface TodoCounts {
  total: number;
  open: number;
  draft: number;
  pending: number;
  inProgress: number;
  // review/ask are absent from server counts until Phase 6 lands; client-side
  // aggregation defaults them to 0 (see addCounts).
  review: number;
  ask: number;
  failed: number;
  unverified: number;
  verified: number;
  completed: number;
  skipped: number;
}

export interface TodoEvent {
  id?: string;
  short_id?: string;
  kind?: string;
  actor?: string;
  timestamp?: string;
  title?: string;
  body?: string;
  label?: string;
  old_label?: string;
  new_label?: string;
}

export interface TodoItem {
  ref: string;
  id?: string;
  shortId?: string;
  version?: number;
  workspaceId?: string;
  executionState?: string;
  title: string;
  status: TodoStatus;
  priority: TodoPriority;
  cwd?: string;
  // Raw label tokens. Their colour and icon are NOT here: presentation is
  // resolved against the separately-cached tag definitions (see tagQueries.ts),
  // so the dictionary is fetched once instead of repeated on every row.
  labels?: string[];
  attempts?: number;
  // ISO timestamp the native issue was created.
  created?: string;
  lastRun?: string;
  // Agent session id of the most recent run, used to follow the session live
  // and to resume it. Recorded in the native issue execution state.
  sessionId?: string;
  // Present only when this detail was resolved from a session UUID. It selects
  // that exact historical transcript while sessionId continues to identify the
  // Todo's active/latest run for lifecycle actions.
  lookupSessionId?: string;
  body?: string;
  implementation?: string;
  events?: TodoEvent[];
  // Aggregated git diff footprint of the todo's commits (those carrying its
  // Gavel-Issue-Id trailer); absent when no commit references the todo.
  diff?: TodoDiffStat;
  // The external tracker issue this todo was pushed to; absent when it has
  // never been pushed. Present on list responses so the list can filter on it.
  externalIssue?: TodoExternalIssue;
  // hasPlan/hasVerification are lightweight availability flags for the list
  // row's indicators — present on both list and detail responses (unlike
  // planPath/verificationMarkdown, which carry full content and are detail-only).
  hasPlan?: boolean;
  hasVerification?: boolean;
  // Latest run per lifecycle phase, keyed by phase. Present on list responses:
  // the server reads a whole workspace's phases in one query, so the phase
  // columns cost no request per row. A phase that has never run is ABSENT
  // rather than zero-valued, which is what distinguishes "not started" from
  // "started and produced nothing".
  phases?: Partial<Record<TodoPhase, TodoPhaseRun>>;
  // Editable acceptance criteria parsed from the todo's "## Acceptance Criteria"
  // section; present on detail responses.
  criteria?: AcceptanceCriterion[];
  // Raw fixture markdown from the todo's "## Verification" section (excluding
  // the heading); present on detail responses, edited via the Verification
  // tab's FixtureEditor and saved through /api/todos/verification/fixture.
  verificationMarkdown?: string;
  // Plan-run bookkeeping (present on detail responses once a plan run finishes):
  // the native plan file's absolute path and whether the last plan was new/
  // updated/unchanged. The plan CONTENT is fetched lazily by the Plan tab via
  // /api/todos/session/plan.
  planPath?: string;
  planStatus?: TodoPlanStatus;
  // Free-text summary the agent reported in its final-result envelope.
  lastRunSummary?: string;
  // Questions blocking an `ask` todo — the agent needs a human decision before
  // it can continue. Answered via /api/todos/answer, which resumes the session.
  questions?: TodoQuestion[];
  // The server's verdict on where this todo stands in its lifecycle: every
  // step's status, which one to run next, and why. Supersedes the client-side
  // plan/run/verify machine that used to derive this from status + hasPlan +
  // verification signals — the server now owns that decision. Present on both
  // list and detail responses.
  lifecycle?: TodoLifecycle;
}

// TodoPhase is one step of a todo's lifecycle, as recorded rather than as it
// behaves: triage runs as a plan-class run but records itself separately so a
// triage pass is distinguishable from a planning pass.
export type TodoPhase = 'plan' | 'triage' | 'run' | 'verify';

// TODO_PHASES is pipeline order — what you plan, you run; what you run, you
// verify. Triage sits with plan because it is the other read-only pass.
export const TODO_PHASES: TodoPhase[] = ['plan', 'triage', 'run', 'verify'];

// TodoPhaseProgress is how far through its own work a phase got. The unit
// differs by phase and is deliberately not normalised: plan, run and triage
// count agent iterations, verification counts the checks in its fixture.
export interface TodoPhaseProgress {
  done: number;
  failed?: number;
  total: number;
}

// TodoPhaseRun is the latest run a todo has for one phase.
export interface TodoPhaseRun {
  phase: TodoPhase;
  // The run's own outcome — NOT the todo's status, which folds every phase into
  // one value.
  state: 'pending' | 'running' | 'waiting' | 'succeeded' | 'failed' | 'cancelled';
  progress?: TodoPhaseProgress;
  started_at?: string;
  finished_at?: string;
  // Elapsed at the moment the row was read. A running phase keeps accruing, so
  // render it by ticking from started_at rather than re-reading this.
  duration_ms?: number;
  cost_usd?: number;
  // The phase executing right now, as opposed to the one that ran most recently.
  active?: boolean;
}

// TodoLifecycleRun is the most recent recorded run for one lifecycle step —
// the run.state values mirror TodoSessionAttempt.state (a live snapshot while
// the attempt runs, terminal once it finishes), not TodoPhaseRun's fixed enum.
export interface TodoLifecycleRun {
  promptRunId: string;
  state: string;
  startedAt?: string;
  finishedAt?: string;
}

// TodoLifecycleStep is one step of the server-computed lifecycle: whether it
// currently applies to this todo, whether the server would suggest it,
// whether it has ever completed, and its most recent run. `suggested` is
// carried per-step for completeness; `TodoLifecycle.next` is the single
// authoritative choice the header renders as primary.
export interface TodoLifecycleStep {
  name: string;
  label: string;
  applicable: boolean;
  suggested: boolean;
  done: boolean;
  lastRun: TodoLifecycleRun | null;
}

// TodoLifecycle is the server's verdict on where a todo stands in its
// plan/run/verify pipeline, from GET /api/todos/item. `next` is null when
// nothing applies — chiefly the two human-decision statuses (review/ask),
// which the client still renders as its own review/answer actions rather than
// a step. `reason` explains the choice (or the absence of one) as a tooltip.
export interface TodoLifecycle {
  steps: TodoLifecycleStep[];
  next: string | null;
  reason: string;
}

// TodoExternalIssue is the tracker issue a todo has been pushed to, mirroring
// the server's types.ExternalIssue. `state` carries the upstream issue's own
// status once something fetches it — it is absent today and must be read as
// "unknown", never as open.
export interface TodoExternalIssue {
  kind: string;
  repo: string;
  number: number;
  url: string;
  state?: string;
}

// TodoPlanStatus is how a plan run classified its plan relative to any prior one.
export type TodoPlanStatus = 'new' | 'updated' | 'unchanged';

// TodoQuestion is one blocking question from an agent that parked in `ask`.
export interface TodoQuestion {
  text: string;
  context?: string;
  options?: string[];
}

// AcceptanceCriterion is one done-ness criterion. checkId is set when the line
// maps to a static verify check (rendered as "<id>: text"); empty checkId marks
// a custom, functionality-specific criterion.
export interface AcceptanceCriterion {
  text: string;
  checkId?: string;
  done?: boolean;
}

// TodoDiffStat is the aggregated change footprint of a todo's linked commits,
// mirroring the server's git.DiffStat.
export interface TodoDiffStat {
  commits: number;
  files: number;
  adds: number;
  dels: number;
}

// The verification payload mirrors (fixture results, checklist, definition of
// done) live in components/todos/verificationAttempts.ts, next to the selectors
// that read them: this module must not import from components/.

// Rolled-up stats for a TODO's agent session (see /api/todos/session/stats):
// identity (agent/model/effort), elapsed time, token usage and derived cost.
// Mirrors cmux.SessionStats. found=false means the session produced no log yet.
export interface SessionStats {
  sessionId?: string;
  agent?: string;
  model?: string;
  effort?: string;
  startedAt?: string;
  updatedAt?: string;
  durationMs: number;
  inputTokens: number;
  outputTokens: number;
  cacheReadTokens: number;
  cacheCreationTokens: number;
  totalTokens: number;
  // Live context-window occupancy (latest turn's input + cache), reset by each
  // compaction — surfaced as the token figure instead of the ever-growing total.
  contextTokens: number;
  // Total context-window size (tokens) for the model, from captain's pricing
  // registry; 0 when the model is unknown. Denominator for the context bar.
  contextWindow: number;
  turns: number;
  // Number of context compactions seen so far; each shrinks contextTokens.
  compactions: number;
  costUsd: number;
  inProgress: boolean;
  found: boolean;
  // High-level agent state from the latest session-log event: thinking | working
  // | ask | completed | error. 'approval' is set when a tool-permission request
  // is pending (see approval). Empty before the first event.
  state?: 'thinking' | 'working' | 'ask' | 'approval' | 'completed' | 'error';
  // API/network failure reason when state === 'error' (the "API Error: …" message).
  error?: string;
  // Pending tool-permission requests awaiting the user's approve/deny/respond.
  // Non-empty (and state === 'approval') only while a driver is blocked on at
  // least one. Also served standalone by GET /api/todos/session/approvals.
  approvals?: TodoSessionApproval[];
}

export interface SessionRuntimeSelection {
  provider?: string;
  // mode is the runtime mechanism: api | agent | cli | cmux.
  mode?: string;
  model?: string;
  effort?: string;
}

export interface TodoSessionAttempt {
  promptRunId: string;
  ordinal: number;
  step: string;
  mode?: string;
  driver?: string;
  requested: SessionRuntimeSelection;
  resolved: SessionRuntimeSelection;
  provider?: string;
  // runtimeMode is the mechanism the model ran on; `mode` above is the run mode
  // (run/plan) — two different axes that share a word.
  runtimeMode?: string;
  model?: string;
  effort?: string;
  status: string;
  pid?: number;
  processActive: boolean;
  state: string;
  phase: string;
  queuedAt: string;
  startedAt?: string;
  finishedAt?: string;
  durationMs?: number;
  error?: string;
  resultText?: string;
  resultJson?: Record<string, unknown>;
  // The step's captain VerifyReport — a running snapshot while the attempt is
  // live, the terminal result once it finishes. Always sent now (null for an
  // attempt that was never verified) — see todo_session_detail.go's
  // todoAttemptDetail.Verification.
  verification: VerifyReport | null;
  admissionSessionId: string;
  executionSessionId?: string;
  providerSessionId?: string;
  canStop?: boolean;
  stopping?: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface TodoSessionDiagnostic {
  severity: 'warning' | 'error';
  code: string;
  message: string;
  details?: unknown;
}

export interface TodoSessionOverview {
  id: string;
  providerSessionId?: string;
  source: string;
  provider?: string;
  hostId: string;
  parentSessionId?: string;
  rootSessionId?: string;
  path?: string;
  historyFile?: string;
  project?: string;
  cwd?: string;
  title?: string;
  agentType?: string;
  description?: string;
  lifecycleStatus: string;
  activityState: string;
  healthState: string;
  startedAt?: string;
  endedAt?: string;
  lastActivityAt?: string;
  processActive: boolean;
  processStatus?: string;
  pid?: number;
  model?: string;
  modelProvider?: string;
  modelMode?: string;
  effort?: string;
  messageCount: number;
  turnCount: number;
  inputTokens: number;
  outputTokens: number;
  totalTokens: number;
  costUsd: number;
}

export interface TodoSessionTurn {
  id: string;
  sessionId: string;
  providerTurnId?: string;
  turnIndex: number;
  status: string;
  stopReason?: string;
  error?: string;
  startedAt?: string;
  endedAt?: string;
  model?: string;
  modelProvider?: string;
  modelMode?: string;
  effort?: string;
  inputTokens: number;
  outputTokens: number;
  reasoningTokens: number;
  cacheReadTokens: number;
  cacheWriteTokens: number;
  totalTokens: number;
  costUsd: number;
  messageCount: number;
}

export interface TodoSessionAgent {
  id: string;
  sessionId: string;
  parentSessionId?: string;
  rootSessionId?: string;
  isRoot: boolean;
  agentType?: string;
  description?: string;
  historyFile?: string;
  source: string;
  provider?: string;
  lifecycleStatus: string;
  activityState: string;
  healthState: string;
  startedAt?: string;
  endedAt?: string;
  inputTokens: number;
  outputTokens: number;
  reasoningTokens: number;
  cacheReadTokens: number;
  cacheWriteTokens: number;
  totalTokens: number;
  costUsd: number;
}

export interface TodoSessionCost {
  id: string;
  sessionId: string;
  model: string;
  provider?: string;
  modelMode?: string;
  effort?: string;
  currency: string;
  modelCallCount: number;
  inputTokens: number;
  outputTokens: number;
  reasoningTokens: number;
  cacheReadTokens: number;
  cacheWriteTokens: number;
  totalTokens: number;
  totalCost: number;
  firstCallAt?: string;
  lastCallAt?: string;
}

export interface TodoProviderThread {
  id: string;
  providerSessionId?: string;
  status: string;
  root: TodoSessionOverview;
  sessions: TodoSessionOverview[];
  turns: TodoSessionTurn[];
  agents: TodoSessionAgent[];
  costs: TodoSessionCost[];
  messages: SessionUIMessage[];
  startedAt?: string;
  lastActivityAt?: string;
  durationMs: number;
  inputTokens: number;
  outputTokens: number;
  totalTokens: number;
  costUsd: number;
}

export interface TodoSessionDetailResponse {
  attempts: TodoSessionAttempt[];
  selectedPromptRunId?: string;
  selectedExecutionSessionId?: string;
  thread?: TodoProviderThread;
  diagnostics: TodoSessionDiagnostic[];
  // Set when the request asked for attempts only: the provider thread was
  // deliberately not resolved, so a missing `thread` is not a missing session.
  attemptsOnly?: boolean;
}

// TodoSessionApproval is a tool-permission request a driver surfaced for human
// review; the dashboard answers it via POST /api/todos/session/approve with
// `{approvalId, action: "approve" | "deny" | "respond", message?, input?}`.
export interface TodoSessionApproval {
  approvalId: string;
  sessionId: string;
  toolUseId?: string;
  tool: string;
  input?: Record<string, unknown>;
}

export type TodoRunAgent = 'claude' | 'codex';
export type TodoRunEffort = 'low' | 'medium' | 'high' | 'xhigh' | 'max' | 'ultra';
// TodoRunDriver is the canonical runtime mode. Provider identity comes from model.
export type TodoRunDriver = 'api' | 'agent' | 'cli' | 'cmux';

// TodoRunOptions is the run POST body's options: the api.Spec under its own
// `spec` key plus the run-orchestration extras below. Dirty worktree,
// auto-commit, dry-run, and checks all live in the spec now
// (setup.checkout.dirty, workflow.commits[].on/dryRun, workflow.verify); only the
// prompt/driver selection and the resume decision sit alongside it.
//
// The spec is nested rather than inlined because the Go payload stopped
// embedding api.Spec: api.Spec declares a value-receiver MarshalJSON, so an
// embedding struct inherited it and marshalled to a bare spec, silently dropping
// driver/runMode/resume. Nesting also dissolves the wire collision between this
// payload's `mode`/`resume` and api.Model's own.
export interface TodoRunOptions {
  // Spec is the captain api.Spec: model/mode/effort flat inside it;
  // prompt/budget/permissions/setup/workflow/sessionId nested, mirroring
  // clicky's AISpecRuntimeValue.
  spec?: AISpecRuntimeValue;
  // step is the lifecycle step name the request dispatches — the sole
  // behaviour-selecting field POST /api/todos/run accepts now (the endpoint
  // decodes strictly and rejects runMode/driver/prompt at the top level).
  // driver/runMode/plan/prompt below stay for the client's own bookkeeping
  // (storage, dropdown labels, the advanced dialog's editors) and are folded
  // into `step` — or into spec.mode for the runtime mechanism — before a
  // request is built; see run.tsx's requestStepFor/buildTodoRunPayload.
  step?: string;
  // Driver is the authoritative canonical runtime mode.
  driver?: TodoRunDriver;
  // runMode is the behaviour class the run executes as (run/plan). Supersedes
  // the `plan` bool; the server accepts both (plan:true is treated as
  // runMode:plan).
  runMode?: 'run' | 'plan';
  // prompt names the template to run — run, plan, triage, or a name declared in
  // .gavel.yaml todos.prompts. It is a separate axis from runMode: each prompt
  // declares the class it runs as, so several prompts share one. Send one or the
  // other; asserting a class that disagrees with the prompt is rejected.
  prompt?: string;
  // Plan-only run: the agent proposes an implementation plan without changing
  // code. Requires cmux mode. Legacy — prefer runMode.
  plan?: boolean;
  // Resume the todo's prior agent session (claude --resume) instead of starting a
  // fresh one. Stays a sibling flag rather than a spec field: a fresh run also
  // carries a (minted) sessionId, so resume can't be inferred from spec.sessionId.
  resume?: boolean;
  // Dispatch even though the todo already has a live run owned by a running
  // process: the two runs proceed in parallel. Set only in answer to the
  // server's 409 run_owned_elsewhere, never as a default.
  force?: boolean;
}

export interface TodoRunResponse {
  status: 'started' | 'skipped' | 'dry_run';
  ref: string;
  dir: string;
  provider?: string;
  driver?: TodoRunDriver;
  // runtimeMode is the mechanism the run dispatched on (api | agent | cli | cmux).
  runtimeMode?: string;
  model?: string;
  effort?: TodoRunEffort;
  plan?: boolean;
  resume?: boolean;
  // Session id the run uses; lets the UI follow the session log immediately.
  sessionId?: string;
  timeout: string;
  maxBudget?: number;
  maxTurns?: number;
  // Whether the run will auto-commit the agent's changes when it finishes.
  commit?: boolean;
  message: string;
}

// Preview of the exact prompt a run would dispatch, shown in the advanced run
// dialog before the user starts the run.
export interface TodoRunPreviewResponse {
  prompt: string;
  specYaml: string;
  provider?: string;
  runtimeMode?: string;
  effort?: TodoRunEffort;
  plan?: boolean;
  count: number;
}

export interface TodoListResponse {
  dir?: string;
  counts: TodoCounts;
  items: TodoItem[];
}

// Where a tag definition came from. 'builtin' is a well-known default that has
// never been stored — editing one writes a row that shadows it; 'derived' never
// appears in a listing because nothing is stored for it.
export type TodoTagScope = 'workspace' | 'global' | 'builtin' | 'derived';

// TodoTagDef is the presentation of one label: which colour and glyph it
// renders as. Definitions live in the gavel database and are edited in
// Settings → Tags. They are fetched once and joined client-side against each
// todo's raw `labels`.
export interface TodoTagDef {
  name: string;
  /** A palette hue token (see TAG_PALETTE), not a hex value. */
  color: string;
  /** The stored clicky icon-registry key, e.g. "debug". */
  icon?: string;
  /** The Iconify name the server resolved from `icon`, e.g. "ph:bug". */
  iconify?: string;
  description?: string;
  scope: TodoTagScope;
  /** Set when the definition matched the label's namespace key, not its name. */
  matchedKey?: string;
}

export interface TodoTagListResponse {
  definitions: TodoTagDef[];
  /** How many todos carry each label, including labels nothing defines. */
  counts?: Record<string, number>;
  /** The hues a definition may use, served so the editor cannot offer an invalid one. */
  palette?: string[];
  /** Present on a removal: what it actually did. */
  removed?: TodoTagRemoval;
}

/**
 * TodoTagRemoval is the result of removing a tag. Removing one from a project
 * also strips it from that project's todos; removing a global definition is
 * presentation only and leaves `todos` at zero.
 */
export interface TodoTagRemoval {
  name: string;
  /** Whether a stored definition was deleted, as opposed to a built-in default. */
  definition: boolean;
  /** How many todos the tag was stripped from. */
  todos: number;
}

// One git commit linked to a todo via its Gavel-Issue-Id trailer. url is the
// commit's page on the origin remote, absent for a local-only repo.
export interface TodoCommit {
  hash: string;
  shortHash: string;
  subject: string;
  author?: string;
  date?: string;
  url?: string;
}

export interface TodoCommitsResponse {
  // issueId is the todo's id that commits were matched against; absent for
  // malformed or incomplete responses that carry no id.
  issueId?: string;
  commits: TodoCommit[];
}

// One commit's rendered diff (ANSI-colored `git show` output). truncated is set
// when the diff exceeded the server's size cap. When the request scopes to a
// single file the diff is just that path's patch.
export interface TodoCommitDiffResponse {
  hash: string;
  diff: string;
  truncated?: boolean;
}

// One file changed in a commit, mirroring git.CommitFile: its path (and previous
// path for a rename), change kind, line counts, and repomap classification.
export interface TodoCommitFile {
  path: string;
  previousPath?: string;
  status: 'added' | 'modified' | 'deleted' | 'renamed';
  adds: number;
  dels: number;
  binary?: boolean;
  language?: string;
  scopes?: string[];
}

export interface TodoCommitFilesResponse {
  hash: string;
  files: TodoCommitFile[];
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
  taskRunId?: string;
  ports?: number[];
  // Live resource sample of the process group. openFiles is -1 where the
  // platform cannot report it. All omitted/zero for a stopped process.
  cpuPercent?: number;
  memoryRss?: number;
  memoryVms?: number;
  openFiles?: number;
  peakCpuPercent?: number;
  peakMemoryRss?: number;
  peakMemoryVms?: number;
  peakOpenFiles?: number;
  // Per-process breakdown of the process group (leader + descendants).
  tree?: ProcNode[];
}

// ProcNode mirrors procfile.ProcNode — one process in a supervised group's tree.
export interface ProcNode {
  pid: number;
  ppid: number;
  command: string;
  status?: string;
  root?: boolean;
  cpuPercent?: number;
  memoryRss?: number;
  memoryVms?: number;
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
  // Uncommitted changes (staged, unstaged, and untracked) in the project's
  // directory. Absent when the directory is not a git work tree.
  gitChanges?: number;
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
  // The server's current closed-PR fetch state; the UI only posts a change when
  // the Closed/Merged State chips disagree with this.
  showClosed?: boolean;
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
  // nodeId is the GraphQL global node ID, required by the merge / approve /
  // auto-merge actions. Present once PR detail has loaded.
  nodeId?: string;
  title: string;
  body?: string;
  author: { login: string; name?: string; avatarUrl?: string };
  headRefName: string;
  baseRefName: string;
  state: string;
  isDraft: boolean;
  reviewDecision: string;
  mergeable: string;
  url: string;
  additions?: number;
  deletions?: number;
  changedFiles?: number;
  statusCheckRollup?: StatusCheck[];
  prCommits?: PRCommitInfo[];
  prFiles?: PRFileInfo[];
}

// PRCommitInfo is a commit in the PR.
export interface PRCommitInfo {
  oid: string;
  messageHeadline: string;
  messageBody?: string;
  committedDate: string;
  authorName?: string;
  authorLogin?: string;
  authorAvatarUrl?: string;
  additions: number;
  deletions: number;
  changedFiles: number;
}

// PRFileInfo is a changed file in the PR.
export interface PRFileInfo {
  path: string;
  additions: number;
  deletions: number;
  changeType: string;
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
  // error is either a crash reported by gavel itself (the run died before it
  // could produce results) or a failure to read the artifact at all.
  error?: string;
  exitCode?: number;
  logTail?: string;
  // duration is the summed leaf-test duration in nanoseconds.
  duration?: number;
  // failures / lint carry a bounded slice of the run itself (up to 5 failing
  // tests and 5 violations); the counts above stay exact.
  failures?: Test[];
  lint?: LinterResult[];
  // commands are the local `gavel test --pr` / `gavel lint --pr` invocations
  // that re-run this shard's failures. PR-scoped, not shard-scoped. Absent on
  // a shard with nothing to reproduce.
  commands?: string[];
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
