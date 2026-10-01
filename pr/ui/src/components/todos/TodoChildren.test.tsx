import type React from 'react';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { TodoItem, TodoPhase, TodoPhaseRun } from '../../types';
import { TodoChildren, TodoRouteLink } from './TodoChildren';

const dialog = vi.hoisted(() => ({ props: [] as Array<{ open: boolean; parent: { ref: string; title: string; dir: string }; onCreated: (dir: string, todo: unknown) => void }> }));

vi.mock('./CreateTodoDialog', () => ({
  CreateTodoDialog: (props: (typeof dialog.props)[number]) => {
    dialog.props.push(props);
    return props.open ? <div data-testid="create-child-dialog">{props.parent.title}</div> : null;
  },
}));

// The shared one-second clock, fixed, so the tests can see who subscribes to it
// and what a live phase reads against.
const clock = vi.hoisted(() => ({ useNow: vi.fn(() => Date.parse('2026-09-30T12:00:00Z')) }));
vi.mock('../../useNow', () => ({ useNow: clock.useNow }));

const PARENT_ID = '7a1d5c3e-9b2f-4e60-8d14-0c5b7a9e2f31';
const DIR = '/repos/billing';

function run(phase: TodoPhase, overrides: Partial<TodoPhaseRun> = {}): TodoPhaseRun {
  return { phase, state: 'succeeded', ...overrides };
}

function child(ref: string, overrides: Partial<TodoItem> = {}): TodoItem {
  return { ref, id: `id-${ref}`, parentId: PARENT_ID, title: `Child ${ref}`, status: 'pending', priority: 'medium', ...overrides };
}

const parent: TodoItem = { ref: 'parent-ref', id: PARENT_ID, title: 'Migrate ledger', status: 'in_progress', priority: 'medium' };

const finished = child('finished', {
  title: 'Backfill rows',
  status: 'completed',
  lastRun: '2026-09-01T10:00:00Z',
  phases: {
    plan: run('plan', { duration_ms: 90_000, cost_usd: 0.5 }),
    run: run('run', { duration_ms: 150_000, cost_usd: 1.5 }),
  },
  diff: { commits: 2, files: 3, adds: 10, dels: 4 },
});
const verifying = child('verifying', {
  title: 'Reconcile totals',
  status: 'in_progress',
  phases: { verify: run('verify', { state: 'failed', duration_ms: 30_000, cost_usd: 0.25 }) },
});

// Reports whether the component claimed the click (preventDefault) and then
// cancels it itself, so jsdom does not attempt the navigation it cannot perform.
function clickClaimed(link: HTMLElement, init: MouseEventInit = {}): boolean {
  let claimed = false;
  const observe = (event: Event) => {
    claimed = event.defaultPrevented;
    event.preventDefault();
  };
  document.addEventListener('click', observe);
  fireEvent.click(link, init);
  document.removeEventListener('click', observe);
  return claimed;
}

function renderChildren(overrides: Partial<React.ComponentProps<typeof TodoChildren>> = {}) {
  return render(<TodoChildren todo={parent} dir={DIR} childTodos={[finished, verifying]} {...overrides} />);
}

describe('TodoChildren', () => {
  it('summarises the rollup in the header: done of total, total duration and total cost', () => {
    renderChildren();
    const header = screen.getByRole('heading', { name: /children/i }).parentElement!;

    expect(within(header).getByText('1/2')).toBeTruthy();
    expect(within(header).getByText('4m 30s')).toBeTruthy();
    expect(within(header).getByText('$2.25')).toBeTruthy();
    expect(within(header).getByTitle('Latest run of each phase, summed over every child')).toBeTruthy();
  });

  it('renders one row per child with a route link, summed cost, activity and diff', () => {
    renderChildren();
    const rows = screen.getAllByRole('listitem');
    expect(rows).toHaveLength(2);

    const first = within(rows[0]);
    expect(first.getByRole('link', { name: 'Backfill rows' }).getAttribute('href')).toBe('/todos/finished');
    expect(first.getByText('$2.00')).toBeTruthy();
    expect(first.getByText('4m 00s')).toBeTruthy();
    expect(first.getByText('+10')).toBeTruthy();
    expect(first.getByText(/ago$/)).toBeTruthy();
    expect(first.getByLabelText('Plan: succeeded')).toBeTruthy();
    expect(first.getByLabelText('Run: succeeded')).toBeTruthy();

    const second = within(rows[1]);
    expect(second.getByRole('link', { name: 'Reconcile totals' }).getAttribute('href')).toBe('/todos/verifying');
    expect(second.getByText('$0.25')).toBeTruthy();
    expect(second.getByLabelText('Verify: failed')).toBeTruthy();
    expect(second.queryByText(/ago$/)).toBeNull();
  });

  it('opens a child in place on a plain click and leaves modified clicks to the browser', () => {
    const onOpenTodo = vi.fn();
    renderChildren({ onOpenTodo });
    const link = screen.getByRole('link', { name: 'Backfill rows' });

    expect(clickClaimed(link)).toBe(true);
    expect(onOpenTodo).toHaveBeenCalledExactlyOnceWith('finished');

    expect(clickClaimed(link, { metaKey: true })).toBe(false);
    expect(onOpenTodo).toHaveBeenCalledTimes(1);
  });

  it('says there are no children yet and still offers Add child on a top-level todo', () => {
    renderChildren({ childTodos: [] });

    expect(screen.getByText('No children yet')).toBeTruthy();
    expect(screen.queryAllByRole('listitem')).toHaveLength(0);
    expect(screen.getByRole('button', { name: /add child/i })).toBeTruthy();
  });

  it('opens the create flow with this todo as the parent and closes it once the child is created', () => {
    renderChildren({ childTodos: [] });
    expect(screen.queryByTestId('create-child-dialog')).toBeNull();

    fireEvent.click(screen.getByRole('button', { name: /add child/i }));
    expect(screen.getByTestId('create-child-dialog').textContent).toBe('Migrate ledger');
    expect(dialog.props.at(-1)!.parent).toEqual({ ref: 'parent-ref', title: 'Migrate ledger', dir: DIR });

    act(() => dialog.props.at(-1)!.onCreated(DIR, child('new')));
    expect(screen.queryByTestId('create-child-dialog')).toBeNull();
  });

  it('offers no Add child on a todo that is itself a child', () => {
    renderChildren({ todo: child('nested'), childTodos: [finished] });
    expect(screen.queryByRole('button', { name: /add child/i })).toBeNull();
  });
});

describe('TodoChildren while the family is not known', () => {
  it('says the children are loading instead of claiming there are none', () => {
    renderChildren({ childTodos: [], familyStatus: { state: 'loading' } });

    expect(screen.getByRole('status').textContent).toBe('Loading children…');
    expect(screen.queryByText('No children yet')).toBeNull();
    expect(screen.queryByText(/\d\/\d/)).toBeNull();
  });

  it('shows why the children could not be listed instead of claiming there are none', () => {
    renderChildren({ childTodos: [], familyStatus: { state: 'error', message: 'workspace list unavailable' } });

    expect(screen.getByRole('alert').textContent).toContain('workspace list unavailable');
    expect(screen.queryByText('No children yet')).toBeNull();
  });

  it('still lets a child be added while the list cannot be shown', () => {
    renderChildren({ childTodos: [], familyStatus: { state: 'error', message: 'workspace list unavailable' } });

    expect(screen.getByRole('button', { name: /add child/i })).toBeTruthy();
  });

  it('shows the genuinely empty state only once the list is ready', () => {
    renderChildren({ childTodos: [], familyStatus: { state: 'ready' } });

    expect(screen.getByText('No children yet')).toBeTruthy();
    expect(screen.queryByRole('alert')).toBeNull();
    expect(screen.queryByRole('status')).toBeNull();
  });
});

describe('TodoChildren clock', () => {
  const settled = child('settled', { phases: { run: run('run', { duration_ms: 60_000, cost_usd: 0.5 }) } });
  const live = child('live', {
    status: 'in_progress',
    phases: { run: run('run', { state: 'running', started_at: '2026-09-30T11:59:15Z', duration_ms: 10_000 }) },
  });

  it('does not subscribe to the one-second clock while every phase is settled', () => {
    clock.useNow.mockClear();
    renderChildren({ childTodos: [settled] });

    expect(clock.useNow).not.toHaveBeenCalled();
  });

  it('ticks the rollup while a child has a live phase', () => {
    clock.useNow.mockClear();
    renderChildren({ childTodos: [settled, live] });

    expect(clock.useNow).toHaveBeenCalled();
    expect(screen.getByRole('heading', { name: /children/i }).parentElement!.textContent).toContain('1m 45s');
  });
});

describe('TodoRouteLink', () => {
  it('links to the todo route and leaves the browser to follow it when no handler is given', () => {
    render(<TodoRouteLink todoRef="a/b" title="Open">Parent</TodoRouteLink>);
    const link = screen.getByRole('link', { name: 'Parent' });

    expect(link.getAttribute('href')).toBe('/todos/a/b');
    expect(clickClaimed(link)).toBe(false);
  });
});
