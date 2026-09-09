import type { TodoCounts } from '../../types';
import type { TodoEntry } from './todoGroup';

// Plan Review mode cycles through the todos that need a human decision — those
// awaiting plan approval (`review`) or answers to blocking questions (`ask`).
// This module is the pure queue logic behind that UX (no React), so it can be
// reasoned about and unit-tested independently of the bar component.

// reviewableCount is how many todos a set of counts has waiting for review or an
// answer — the badge number on the "Review · N" navbar pill.
export function reviewableCount(counts: TodoCounts | undefined): number {
  if (!counts) return 0;
  return (counts.review ?? 0) + (counts.ask ?? 0);
}

// isReviewable reports whether a todo is in a state Plan Review mode acts on.
export function isReviewable(status: string | undefined): boolean {
  return status === 'review' || status === 'ask';
}

// buildReviewQueue flattens every workspace's reviewable todos into one
// oldest-first queue (the ones waiting longest surface first), with a stable
// tiebreak on ref so the order never jitters between refreshes.
export function buildReviewQueue(entries: TodoEntry[]): TodoEntry[] {
  return entries
    .filter(entry => isReviewable(entry.todo.status))
    .sort((a, b) => {
      const ageDelta = reviewSortKey(a) - reviewSortKey(b);
      if (ageDelta !== 0) return ageDelta;
      return a.todo.ref.localeCompare(b.todo.ref);
    });
}

// reviewSortKey is the epoch-ms a todo entered its waiting state, approximated by
// the last run (when it parked in review/ask) and falling back to creation. A
// missing timestamp sorts last (treated as "just now") so dated todos lead.
function reviewSortKey(entry: TodoEntry): number {
  const stamp = entry.todo.lastRun || entry.todo.created;
  if (!stamp) return Number.MAX_SAFE_INTEGER;
  const ms = Date.parse(stamp);
  return Number.isNaN(ms) ? Number.MAX_SAFE_INTEGER : ms;
}

// nextIndexAfterRemoval picks which queue position to focus after the todo at
// `current` is acted on and drops out. `nextLength` is the queue length AFTER
// removal. It keeps the same slot (the item that shifted up into it) unless the
// removed item was last, in which case it steps back to the new last item.
// Returns -1 when the queue is now empty.
export function nextIndexAfterRemoval(current: number, nextLength: number): number {
  if (nextLength <= 0) return -1;
  return Math.min(current, nextLength - 1);
}
