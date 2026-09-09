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
