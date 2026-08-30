import type { ComponentType } from 'react';
import type { IconProps } from '@flanksource/clicky-ui/icons';
import { UiFolder, UiHistory, UiListFlat, UiWarningTriangle } from '@flanksource/clicky-ui/icons';
import type { Project, TodoGroupBy, TodoItem, TodoListResponse } from '../../types';
import type { TodoSort } from './todoSort';
import { defaultTodoSort, todoComparator } from './todoSort';

// Group-by is a per-user view preference for the todo lists, persisted alongside
// density and the status filter so it survives reloads. 'workspace' keeps the
// per-workspace grouping (the default, the only mode that supports batch runs);
// 'severity' and 'age' re-bucket todos across every workspace, and 'none' is one
// flat list. The full-width table drives the same preference through DataTable's
// native grouping picker, so both layouts group by the same dimension.
export const GROUP_BY_OPTIONS: { value: TodoGroupBy; label: string; icon: ComponentType<IconProps> }[] = [
  { value: 'workspace', label: 'Workspace', icon: UiFolder },
  { value: 'severity', label: 'Severity', icon: UiWarningTriangle },
  { value: 'age', label: 'Age', icon: UiHistory },
  { value: 'none', label: 'None', icon: UiListFlat },
];

const STORAGE_KEY = 'gavel.pr-ui.todoGroupBy.v1';

export function defaultGroupBy(): TodoGroupBy {
  return 'workspace';
}

// Persistence is best-effort: localStorage can throw (private mode / disabled),
// so a failure falls back to the workspace default rather than breaking.
export function loadGroupBy(): TodoGroupBy {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw === 'workspace' || raw === 'severity' || raw === 'age' || raw === 'none' ? raw : defaultGroupBy();
  } catch {
    return defaultGroupBy();
  }
}

export function saveGroupBy(groupBy: TodoGroupBy): void {
  try {
    localStorage.setItem(STORAGE_KEY, groupBy);
  } catch {
    // best-effort: storage unavailable — skip persisting.
  }
}

// A todo tagged with its owning workspace so severity/age buckets can mix todos
// across workspaces while still routing a click back to the right directory.
export interface TodoEntry {
  todo: TodoItem;
  workspace: Project;
}

// One severity/age bucket: a labelled, colour-toned section of the flattened
// todo list. The label/tone drive the sticky header; the view maps bucket keys
// to the bundled issue-tracking icon set.
export interface TodoBucket {
  key: string;
  label: string;
  // Tailwind text-color class for the bucket's header label and icon.
  tone: string;
  entries: TodoEntry[];
}

// ageMs is the todo's age anchor in epoch millis: its creation time when known,
// otherwise its last activity. null when neither timestamp is recorded.
export function ageMs(todo: TodoItem): number | null {
  const raw = todo.created ?? todo.lastRun;
  if (!raw) return null;
  const ms = Date.parse(raw);
  return Number.isNaN(ms) ? null : ms;
}

// flattenTodos tags every workspace's todos with their owning workspace, in
// workspace order, so the result can be bucketed by any dimension while each
// entry still knows where it came from.
export function flattenTodos(workspaces: Project[], byDir: Record<string, TodoListResponse>): TodoEntry[] {
  const entries: TodoEntry[] = [];
  for (const workspace of workspaces) {
    for (const todo of byDir[workspace.dir]?.items ?? []) {
      entries.push({ todo, workspace });
    }
  }
  return entries;
}

// Exported so the full-width table's DataTable grouping modes bucket with the
// same definitions the list uses, rather than a parallel copy that can drift.
export const SEVERITY_BUCKETS: { key: string; label: string; tone: string }[] = [
  { key: 'high', label: 'High priority', tone: 'text-red-600' },
  { key: 'medium', label: 'Medium priority', tone: 'text-yellow-600' },
  { key: 'low', label: 'Low priority', tone: 'text-green-600' },
];

// severityKey maps an entry's priority onto a bucket, defaulting unknown values
// to medium the same way the providers do.
export function severityKey(entry: TodoEntry): string {
  const priority = entry.todo.priority;
  return priority === 'high' || priority === 'low' ? priority : 'medium';
}

export const AGE_BUCKETS: { key: string; label: string; maxDays: number }[] = [
  { key: 'today', label: 'Today', maxDays: 1 },
  { key: 'week', label: 'This week', maxDays: 7 },
  { key: 'month', label: 'This month', maxDays: 30 },
  { key: 'older', label: 'Older', maxDays: Infinity },
];

// lastActivityMs is the todo's most recent activity timestamp in epoch millis,
// or null when unknown.
function lastActivityMs(todo: TodoItem): number | null {
  if (!todo.lastRun) return null;
  const ms = Date.parse(todo.lastRun);
  return Number.isNaN(ms) ? null : ms;
}

// ageKey buckets an item by how long ago it was last active; todos with no
// recorded activity sort into a trailing "no activity" bucket.
export function ageKey(item: TodoEntry, now: number): string {
  const ms = lastActivityMs(item.todo);
  if (ms === null) return 'none';
  const days = (now - ms) / 86_400_000;
  return (AGE_BUCKETS.find(b => days < b.maxDays) ?? AGE_BUCKETS[AGE_BUCKETS.length - 1]).key;
}

// bucketTodos splits the flattened entries into ordered, non-empty buckets for
// the given non-workspace grouping. Severity buckets follow high→medium→low; age
// buckets run today→older then a trailing "no activity"; 'none' is one bucket
// holding everything. Rows within each bucket follow the caller's sort
// preference (defaulting to the priority order).
export function bucketTodos(entries: TodoEntry[], groupBy: 'severity' | 'age' | 'none', now: number, sort: TodoSort = defaultTodoSort()): TodoBucket[] {
  const cmp = todoComparator(sort);
  if (groupBy === 'none') {
    if (entries.length === 0) return [];
    return [{
      key: 'all',
      label: 'All todos',
      tone: 'text-muted-foreground',
      entries: [...entries].sort((a, b) => cmp(a.todo, b.todo)),
    }];
  }
  if (groupBy === 'severity') {
    return SEVERITY_BUCKETS
      .map(def => ({
        key: def.key,
        label: def.label,
        tone: def.tone,
        entries: entries.filter(e => severityKey(e) === def.key).sort((a, b) => cmp(a.todo, b.todo)),
      }))
      .filter(bucket => bucket.entries.length > 0);
  }

  const byAge = new Map<string, TodoEntry[]>();
  for (const entry of entries) {
    const key = ageKey(entry, now);
    let bucket = byAge.get(key);
    if (!bucket) {
      bucket = [];
      byAge.set(key, bucket);
    }
    bucket.push(entry);
  }
  const ordered: TodoBucket[] = [];
  for (const def of AGE_BUCKETS) {
    const bucket = byAge.get(def.key);
    if (!bucket) continue;
    bucket.sort((a, b) => cmp(a.todo, b.todo));
    ordered.push({ key: def.key, label: def.label, tone: 'text-muted-foreground', entries: bucket });
  }
  const none = byAge.get('none');
  if (none) {
    ordered.push({ key: 'none', label: 'No activity', tone: 'text-muted-foreground', entries: none });
  }
  return ordered;
}
