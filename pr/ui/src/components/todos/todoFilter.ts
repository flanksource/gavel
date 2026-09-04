import type { TodoCounts, TodoItem, TodoPriority, TodoStatus } from '../../types';
import type { FacetModes } from '../../utils';
import { passFacet } from '../../utils';
import type { TodoEntry } from './todoGroup';
import { withinActivityRange, type ResolvedRange } from './todoTimeRange';
import { todoTagTokens } from './tagResolve';

// The todos list filters on five tri-state facets — workspace, status, priority,
// tags, and whether the todo is linked to an external tracker issue. Each is the
// same include/exclude/neutral shape clicky-ui's FilterBar `kind:"multi"` binds
// to, and the same shape the PRs tab's facets use, so one rule (passFacet) covers
// all of them. Filtering stays client side so the per-workspace counts remain
// computed over the full list.
export const CLOSED_STATUS: TodoStatus = 'completed';

// Each pill maps a status to its label and the count field that feeds its badge.
// "Closed" is the user-facing name for completed native issues, kept
// last so it reads as the trailing "hide closed" control.
export const STATUS_FILTER_DEFS: { status: TodoStatus; label: string; countKey: keyof TodoCounts }[] = [
  { status: 'draft', label: 'Draft', countKey: 'draft' },
  { status: 'pending', label: 'Pending', countKey: 'pending' },
  { status: 'in_progress', label: 'In progress', countKey: 'inProgress' },
  { status: 'review', label: 'Review', countKey: 'review' },
  { status: 'ask', label: 'Ask', countKey: 'ask' },
  { status: 'failed', label: 'Failed', countKey: 'failed' },
  { status: 'unverified', label: 'Unverified', countKey: 'unverified' },
  { status: 'verified', label: 'Verified', countKey: 'verified' },
  { status: 'skipped', label: 'Skipped', countKey: 'skipped' },
  { status: 'completed', label: 'Closed', countKey: 'completed' },
];

// Priority options, ordered most severe first so the facet reads the way the
// severity buckets do.
export const PRIORITY_FILTER_DEFS: { priority: TodoPriority; label: string }[] = [
  { priority: 'high', label: 'High' },
  { priority: 'medium', label: 'Medium' },
  { priority: 'low', label: 'Low' },
];

// TodoExternalKey splits todos on whether they have been pushed to an external
// tracker. When the upstream issue's own state is fetched some day it becomes a
// third dimension on TodoItem.externalIssue.state, not another key here.
export type TodoExternalKey = 'linked' | 'unlinked';

export const EXTERNAL_FILTER_DEFS: { key: TodoExternalKey; label: string }[] = [
  { key: 'linked', label: 'Linked' },
  { key: 'unlinked', label: 'Not linked' },
];

export interface TodoFilters {
  statuses: FacetModes;
  priorities: FacetModes;
  external: FacetModes;
  // Keyed by the tokens todoTagTokens emits: each full label plus, for a
  // namespaced label, its bare key — so "only area:ui" and "exclude every
  // area:*" are both expressible from one facet.
  tags: FacetModes;
  // Keyed by normalized workspace directory — the same key byDir and the batch
  // API use. Not the display name: two checkouts of the same repo share a name
  // but are separate workspaces.
  workspaces: FacetModes;
}

export const TODO_FILTERS_KEY = 'gavel.pr-ui.todoFilters.v3';
// The v2 key held the three facets without tags. It is read once, widened, and
// removed — the same one-shot migration the v1 key gets below.
export const LEGACY_TODO_FILTERS_V2_KEY = 'gavel.pr-ui.todoFilters.v2';
// The v1 key held a bare array of hidden statuses. It is read once, migrated
// into excludes, and removed.
export const LEGACY_HIDDEN_STATUS_KEY = 'gavel.pr-ui.todoFilter.v1';

export function defaultTodoFilters(): TodoFilters {
  return { statuses: { [CLOSED_STATUS]: 'exclude' }, priorities: {}, external: {}, tags: {}, workspaces: {} };
}

// todoMatchesQuery is the free-text search over a todo's identity fields — title,
// its short/ref/full ids, current session id, and labels. It backs the ⌘K
// command palette's todo results. An empty query matches everything.
export function todoMatchesQuery(item: TodoItem, query: string): boolean {
  const q = query.trim().toLowerCase();
  if (!q) return true;
  return (
    item.title.toLowerCase().includes(q) ||
    (item.shortId || '').toLowerCase().includes(q) ||
    (item.ref || '').toLowerCase().includes(q) ||
    (item.id || '').toLowerCase().includes(q) ||
    (item.sessionId || '').toLowerCase().includes(q) ||
    (item.labels || []).some(label => label.toLowerCase().includes(q))
  );
}

// priorityKey defaults an unknown or missing priority to medium, the same way
// the providers and the severity buckets do.
export function priorityKey(item: TodoItem): TodoPriority {
  return item.priority === 'high' || item.priority === 'low' ? item.priority : 'medium';
}

export function externalKey(item: TodoItem): TodoExternalKey {
  return item.externalIssue ? 'linked' : 'unlinked';
}

// isTodoVisible applies the four facets carried by the todo itself and, when an
// activity range is active, the time filter. Every facet must pass — they narrow
// together. The workspace facet is deliberately not here: it is keyed on where a
// todo lives, which only its TodoEntry knows (see isEntryVisible).
export function isTodoVisible(item: TodoItem, filters: TodoFilters, range?: ResolvedRange | null): boolean {
  if (!passFacet(filters.statuses, [item.status])) return false;
  if (!passFacet(filters.priorities, [priorityKey(item)])) return false;
  if (!passFacet(filters.external, [externalKey(item)])) return false;
  if (!passFacet(filters.tags ?? {}, todoTagTokens(item))) return false;
  if (range && !withinActivityRange(item, range)) return false;
  return true;
}

// isWorkspaceShown answers the workspace facet on its own, which is what the
// per-workspace sections ask: an excluded workspace drops out whole — header,
// counts and all — rather than rendering as an empty section.
export function isWorkspaceShown(filters: TodoFilters, dir: string): boolean {
  return passFacet(filters.workspaces ?? {}, [dir]);
}

// isEntryVisible is the full rule for a flattened list, where rows from every
// workspace sit side by side: the workspace facet plus everything isTodoVisible
// checks.
export function isEntryVisible(entry: TodoEntry, filters: TodoFilters, range?: ResolvedRange | null): boolean {
  return isWorkspaceShown(filters, entry.workspace.dir) && isTodoVisible(entry.todo, filters, range);
}

// isStatusShown answers the one question the per-workspace count badges ask:
// would a todo of this status survive the status facet? It is passFacet over a
// single category, named for the caller's intent.
export function isStatusShown(statuses: FacetModes, status: TodoStatus): boolean {
  return passFacet(statuses, [status]);
}

// toggleStatusFilter flips one status between neutral and excluded — the badge
// click, which is a "hide this" control rather than the full tri-state cycle the
// FilterBar chips offer.
export function toggleStatusFilter(statuses: FacetModes, status: TodoStatus): FacetModes {
  const next = { ...statuses };
  if (next[status] === 'exclude') delete next[status];
  else next[status] = 'exclude';
  return next;
}

function isFacetModes(value: unknown): value is FacetModes {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return false;
  return Object.values(value).every(mode => mode === 'include' || mode === 'exclude');
}

// migrateHiddenStatuses reads the v1 array of hidden statuses, if it is still
// there, as the equivalent set of excludes.
function migrateHiddenStatuses(): TodoFilters | null {
  const raw = localStorage.getItem(LEGACY_HIDDEN_STATUS_KEY);
  if (!raw) return null;
  localStorage.removeItem(LEGACY_HIDDEN_STATUS_KEY);
  const parsed: unknown = JSON.parse(raw);
  if (!Array.isArray(parsed)) return null;
  const statuses: FacetModes = {};
  for (const status of parsed) {
    if (typeof status === 'string') statuses[status] = 'exclude';
  }
  return { statuses, priorities: {}, external: {}, tags: {}, workspaces: {} };
}

// migrateV2Filters widens a stored v2 payload — the same three facets, before
// tags existed — into v3. A v2 value is a structural subset of v3, so the
// migration is total: nothing the user had selected is lost.
function migrateV2Filters(): TodoFilters | null {
  const raw = localStorage.getItem(LEGACY_TODO_FILTERS_V2_KEY);
  if (!raw) return null;
  localStorage.removeItem(LEGACY_TODO_FILTERS_V2_KEY);
  const parsed: unknown = JSON.parse(raw);
  if (!parsed || typeof parsed !== 'object') return null;
  const { statuses, priorities, external } = parsed as Partial<TodoFilters>;
  if (!isFacetModes(statuses) || !isFacetModes(priorities) || !isFacetModes(external)) return null;
  return { statuses, priorities, external, tags: {}, workspaces: {} };
}

// Persistence is best-effort: localStorage can throw (private mode / disabled),
// so a failure falls back to the defaults rather than breaking the list.
export function loadTodoFilters(): TodoFilters {
  try {
    const raw = localStorage.getItem(TODO_FILTERS_KEY);
    if (raw) {
      const parsed: unknown = JSON.parse(raw);
      if (parsed && typeof parsed === 'object') {
        const { statuses, priorities, external, tags, workspaces } = parsed as Partial<TodoFilters>;
        if (isFacetModes(statuses) && isFacetModes(priorities) && isFacetModes(external)) {
          // tags and workspaces widen in place rather than behind a key bump: a
          // payload written before either facet existed simply has no such key,
          // and reading it as neutral loses nothing the user had selected.
          return {
            statuses,
            priorities,
            external,
            tags: isFacetModes(tags) ? tags : {},
            workspaces: isFacetModes(workspaces) ? workspaces : {},
          };
        }
      }
      return defaultTodoFilters();
    }
    return migrateV2Filters() ?? migrateHiddenStatuses() ?? defaultTodoFilters();
  } catch {
    return defaultTodoFilters();
  }
}

export function saveTodoFilters(filters: TodoFilters): void {
  try {
    localStorage.setItem(TODO_FILTERS_KEY, JSON.stringify(filters));
  } catch {
    // best-effort: storage unavailable — skip persisting.
  }
}
