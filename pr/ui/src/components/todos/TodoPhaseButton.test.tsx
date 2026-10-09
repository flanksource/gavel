import { describe, expect, it } from 'vitest';
import type { TodoItem } from '../../types';
import { LIFECYCLE_MOCK_DRAFT, LIFECYCLE_MOCK_IMPLEMENTED, LIFECYCLE_MOCK_REVIEW } from './lifecycleMock';
import { otherLifecycleSteps, primaryLifecycleAction, primaryRunExtras, runBranchItems } from './TodoPhaseButton';

function todoWith(overrides: Partial<TodoItem> = {}): TodoItem {
  return { ref: 'todo-1', title: 'A todo', status: 'pending', priority: 'medium', ...overrides };
}

describe('primaryLifecycleAction', () => {
  it('renders the server-chosen next step, with its reason as the tooltip', () => {
    const todo = todoWith({ status: 'pending', lifecycle: LIFECYCLE_MOCK_DRAFT });
    expect(primaryLifecycleAction(todo, false)).toEqual({
      kind: 'step',
      name: 'plan',
      label: 'Plan',
      reason: LIFECYCLE_MOCK_DRAFT.reason,
    });
  });

  it('falls back to the step name when no step in the catalog matches `next`', () => {
    const todo = todoWith({
      status: 'pending',
      lifecycle: { steps: [], next: 'triage', reason: 'Server suggested a step the client catalog omits.' },
    });
    expect(primaryLifecycleAction(todo, false)).toEqual({
      kind: 'step',
      name: 'triage',
      label: 'triage',
      reason: 'Server suggested a step the client catalog omits.',
    });
  });

  it('offers none with the server reason when nothing applies and no human decision is pending', () => {
    const todo = todoWith({
      status: 'verified',
      lifecycle: { steps: [], next: null, reason: 'Nothing left to run.' },
    });
    expect(primaryLifecycleAction(todo, false)).toEqual({ kind: 'none', reason: 'Nothing left to run.' });
  });

  it('defaults the reason when a todo has no lifecycle payload at all', () => {
    expect(primaryLifecycleAction(todoWith({ status: 'draft' }), false)).toEqual({
      kind: 'none',
      reason: 'Nothing to run.',
    });
  });

  // A live session outranks the stored status and the server's `next`:
  // whatever either says, the only action that applies while an agent is
  // running is stopping it.
  it('lets a live session outrank every lifecycle verdict', () => {
    const todo = todoWith({ status: 'pending', lifecycle: LIFECYCLE_MOCK_DRAFT });
    expect(primaryLifecycleAction(todo, true)).toEqual({ kind: 'stop' });
  });

  // review/ask must route through the banner's approve/reject/answer flow —
  // even when `next` is present, the two human-decision statuses win.
  it('routes review and ask to their own actions ahead of any step', () => {
    expect(primaryLifecycleAction(todoWith({ status: 'review', lifecycle: LIFECYCLE_MOCK_REVIEW }), false))
      .toEqual({ kind: 'review' });
    expect(primaryLifecycleAction(todoWith({ status: 'ask', lifecycle: LIFECYCLE_MOCK_IMPLEMENTED }), false))
      .toEqual({ kind: 'answer' });
  });

  it('reports none when the server marks nothing as next for a review todo with no lifecycle payload', () => {
    expect(primaryLifecycleAction(todoWith({ status: 'review' }), false)).toEqual({ kind: 'review' });
  });
});

describe('run branch choice', () => {
  const runStep = { name: 'run', label: 'Run', applicable: true, suggested: true, done: false, lastRun: null };
  const planStep = { ...runStep, name: 'plan', label: 'Plan', suggested: false };
  const choice = { branch: 'shell/abc', unresolved: 2 };

  it('continues on the reusable branch when the primary action is run', () => {
    const todo = todoWith({ lifecycle: { steps: [runStep], next: 'run', reason: '' } });
    expect(primaryRunExtras(primaryLifecycleAction(todo, false), choice)).toEqual({ reuseBranch: true });
  });

  it('adds nothing when no branch can be reused or the primary step is not run', () => {
    const runTodo = todoWith({ lifecycle: { steps: [runStep], next: 'run', reason: '' } });
    const planTodo = todoWith({ lifecycle: { steps: [planStep, runStep], next: 'plan', reason: '' } });
    expect(primaryRunExtras(primaryLifecycleAction(runTodo, false), null)).toBeUndefined();
    expect(primaryRunExtras(primaryLifecycleAction(planTodo, false), choice)).toBeUndefined();
  });

  it('offers continue and new-branch runs whenever run is applicable', () => {
    const calls: Array<[string, { reuseBranch?: boolean } | undefined]> = [];
    const todo = todoWith({ lifecycle: { steps: [planStep, runStep], next: 'plan', reason: '' } });
    const items = runBranchItems(todo, choice, (name, extras) => calls.push([name, extras]));

    expect(items.map(item => item.label)).toEqual(['Continue on shell/abc · 2 unresolved comments', 'Run on a new branch']);
    items.forEach(item => item.onSelect?.());
    expect(calls).toEqual([['run', { reuseBranch: true }], ['run', { reuseBranch: false }]]);
  });

  it('offers no branch rows when run is inapplicable or nothing is reusable', () => {
    const blocked = todoWith({ lifecycle: { steps: [{ ...runStep, applicable: false }], next: null, reason: '' } });
    const runnable = todoWith({ lifecycle: { steps: [runStep], next: 'run', reason: '' } });
    expect(runBranchItems(blocked, choice, () => {})).toEqual([]);
    expect(runBranchItems(runnable, null, () => {})).toEqual([]);
  });
});

describe('otherLifecycleSteps', () => {
  it('offers every other applicable step, excluding the one already suggested', () => {
    const todo = todoWith({ status: 'pending', lifecycle: LIFECYCLE_MOCK_IMPLEMENTED });
    const primary = primaryLifecycleAction(todo, false);
    const offered = otherLifecycleSteps(todo, primary).map(entry => entry.name);
    expect(offered).toEqual(['plan', 'run']);
  });

  // A step the server marked inapplicable is left out: offering it would just
  // produce a rejected run — review/ask states mark run/verify inapplicable.
  it('drops steps the server marked inapplicable', () => {
    const todo = todoWith({ status: 'review', lifecycle: LIFECYCLE_MOCK_REVIEW });
    const primary = primaryLifecycleAction(todo, false);
    expect(otherLifecycleSteps(todo, primary).map(entry => entry.name)).toEqual([]);
  });

  it('returns nothing for a todo with no lifecycle payload', () => {
    const todo = todoWith({ status: 'draft' });
    expect(otherLifecycleSteps(todo, primaryLifecycleAction(todo, false))).toEqual([]);
  });
});
