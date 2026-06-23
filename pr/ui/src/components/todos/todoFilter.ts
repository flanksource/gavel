import type { TodoCounts, TodoItem, TodoStatus } from '../../types';

// The todos list hides "closed" (completed) todos by default; the Closed/Status
// pills toggle a status into or out of this hidden set. Filtering is done client
// side so the per-workspace counts stay computed over the full list.
export const CLOSED_STATUS: TodoStatus = 'completed';

// Each pill maps a status to its label and the count field that feeds its badge.
// "Closed" is the user-facing name for completed (grite's `closed` state), kept
// last so it reads as the trailing "hide closed" control.
export const STATUS_FILTER_DEFS: { status: TodoStatus; label: string; countKey: keyof TodoCounts }[] = [
  { status: 'draft', label: 'Draft', countKey: 'draft' },
  { status: 'pending', label: 'Pending', countKey: 'pending' },
  { status: 'in_progress', label: 'In progress', countKey: 'inProgress' },
  { status: 'failed', label: 'Failed', countKey: 'failed' },
  { status: 'verified', label: 'Verified', countKey: 'verified' },
  { status: 'skipped', label: 'Skipped', countKey: 'skipped' },
  { status: 'completed', label: 'Closed', countKey: 'completed' },
];

const STORAGE_KEY = 'gavel.pr-ui.todoFilter.v1';

export function defaultHiddenStatuses(): Set<TodoStatus> {
  return new Set<TodoStatus>([CLOSED_STATUS]);
}

export function isTodoVisible(item: TodoItem, hidden: Set<TodoStatus>): boolean {
  return !hidden.has(item.status);
}

export function toggleHiddenStatus(hidden: Set<TodoStatus>, status: TodoStatus): Set<TodoStatus> {
  const next = new Set(hidden);
  if (next.has(status)) next.delete(status);
  else next.add(status);
  return next;
}

// Persistence is best-effort: localStorage can throw (private mode / disabled),
// so a failure falls back to the default of hiding closed rather than breaking.
export function loadHiddenStatuses(): Set<TodoStatus> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return defaultHiddenStatuses();
    const parsed = JSON.parse(raw) as TodoStatus[];
    if (!Array.isArray(parsed)) return defaultHiddenStatuses();
    return new Set(parsed);
  } catch {
    return defaultHiddenStatuses();
  }
}

export function saveHiddenStatuses(hidden: Set<TodoStatus>): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify([...hidden]));
  } catch {
    // best-effort: storage unavailable — skip persisting.
  }
}
