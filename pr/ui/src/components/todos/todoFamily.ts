import type { TodoItem, TodoListResponse } from '../../types';
import { TODO_PHASES } from '../../types';
import { isChildTodo } from './todoFilter';
import { phaseElapsedMs, phaseRunning } from './TodoPhaseCell';

export interface ChildRollup {
  done: number;
  total: number;
  durationMs: number;
  costUsd: number;
}

const DONE_STATUSES = new Set<TodoItem['status']>(['completed', 'verified']);

export function childrenOf(items: TodoItem[], id: string | undefined): TodoItem[] {
  if (!id) return [];
  return items.filter(item => item.parentId === id);
}

// A child still open when its parent is archived: an archived child reads as
// completed, so completed is the only closed state that matters here.
export function openChildrenOf(children: TodoItem[]): TodoItem[] {
  return children.filter(child => child.status !== 'completed');
}

export function parentOf(items: TodoItem[], todo: TodoItem): TodoItem | null {
  if (!todo.parentId) return null;
  return items.find(item => item.id === todo.parentId) ?? null;
}

// Only a top-level todo can be a parent, and never of itself. `items` is one
// workspace's list, which is what keeps a child in the same workspace as its parent.
export function parentCandidates(items: TodoItem[], todo: TodoItem): TodoItem[] {
  return items.filter(item => !!item.id && !isChildTodo(item) && item.ref !== todo.ref && item.status !== 'completed');
}

export interface ChildCount {
  done: number;
  total: number;
}

// Every parent's children, counted by its full id. A todo without children has
// no entry, so a list row can badge parents without a lookup per child.
export function childCountsById(items: TodoItem[]): Map<string, ChildCount> {
  const counts = new Map<string, ChildCount>();
  for (const item of items) {
    if (!item.parentId) continue;
    const count = counts.get(item.parentId) ?? { done: 0, total: 0 };
    count.total++;
    if (DONE_STATUSES.has(item.status)) count.done++;
    counts.set(item.parentId, count);
  }
  return counts;
}

// Titles of the todos that have children, keyed by full id — all a child's
// "↳ parent" hint needs. Keeping to parents means an unrelated title edit does
// not change the index.
export function parentTitlesById(byDir: Record<string, TodoListResponse>): Map<string, string> {
  const lists = Object.values(byDir);
  const parentIds = new Set(lists.flatMap(list => list.items.flatMap(item => (item.parentId ? [item.parentId] : []))));
  const titles = new Map<string, string>();
  for (const list of lists) {
    for (const item of list.items) if (item.id && parentIds.has(item.id)) titles.set(item.id, item.title);
  }
  return titles;
}

// Whether the family list is still on its way, failed, or in hand — so a parent
// whose children have not arrived is not shown as having none.
export type TodoFamilyStatus = { state: 'ready' } | { state: 'loading' } | { state: 'error'; message: string };

// A phase whose elapsed time is still growing.
export function childHasLivePhase(child: TodoItem): boolean {
  return TODO_PHASES.some(phase => {
    const run = child.phases?.[phase];
    return !!run && phaseRunning(run) && !!run.started_at;
  });
}

// A todo's phases hold the latest run of each phase, so this is the cost and time
// of the current attempt at every step — not of every attempt ever made. A run
// that verified in-run is reported under both run and verify with the same
// prompt run id and the run's whole duration and cost: each id is summed once.
export function childPhaseTotals(child: TodoItem, nowMs: number): { durationMs: number; costUsd: number } {
  let durationMs = 0;
  let costUsd = 0;
  const counted = new Set<string>();
  for (const phase of TODO_PHASES) {
    const run = child.phases?.[phase];
    if (!run) continue;
    if (run.prompt_run_id) {
      if (counted.has(run.prompt_run_id)) continue;
      counted.add(run.prompt_run_id);
    }
    durationMs += phaseElapsedMs(run, nowMs);
    costUsd += run.cost_usd ?? 0;
  }
  return { durationMs, costUsd };
}

export function childRollup(children: TodoItem[], nowMs: number): ChildRollup {
  const rollup: ChildRollup = { done: 0, total: children.length, durationMs: 0, costUsd: 0 };
  for (const child of children) {
    if (DONE_STATUSES.has(child.status)) rollup.done++;
    const totals = childPhaseTotals(child, nowMs);
    rollup.durationMs += totals.durationMs;
    rollup.costUsd += totals.costUsd;
  }
  return rollup;
}
