import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { DataTableGroupingCustomMode } from '@flanksource/clicky-ui/data';
import type { Project, TodoItem, TodoPriority } from '../../types';
import { TODO_PHASES } from '../../types';
import { TODO_SORT_COLUMN_OPTIONS } from './todoSort';
import { todoGroupingModes, todoTableColumns, todoTableRowId, type TodoTableRow } from './TodoTable';

const workspaces: Project[] = [
  { name: 'alpha', dir: '/repos/alpha' } as Project,
  { name: 'beta', dir: '/repos/beta' } as Project,
];

// The selection key joins dir and ref with NUL, which cannot occur in either
// half. Built here rather than pasted so the source carries no control byte.
const NUL = String.fromCharCode(0);

function row(ref: string, priority: TodoPriority, dir = '/repos/alpha'): TodoTableRow {
  const todo: TodoItem = {
    ref,
    title: ref,
    status: 'pending',
    priority,
    created: '2026-07-01T00:00:00Z',
  };
  const workspace = workspaces.find(ws => ws.dir === dir)!;
  return { todo, workspace } as TodoTableRow;
}

const customModes = (now: number) =>
  todoGroupingModes({ workspaces, now })
    .filter((mode): mode is DataTableGroupingCustomMode<TodoTableRow> => mode.type === 'custom');

describe('todoTableColumns', () => {
  // The sort state is the shared TodoSort preference, so a column that is not
  // one of its columns must not be sortable — a sortable header for an unknown
  // key would write a value loadTodoSort later rejects, silently resetting the
  // user's sort on the next reload.
  it('marks exactly the TodoSort columns sortable', () => {
    const sortable = todoTableColumns({ groupBy: 'none' })
      .filter(column => column.sortable)
      .map(column => column.key)
      .sort();
    const expected = TODO_SORT_COLUMN_OPTIONS.map(option => option.value).sort();
    expect(sortable).toEqual(expected);
  });

  // DataTable treats a column as sortable unless it is explicitly `false`, so
  // leaving the flag off is not the same as opting out — asserting only the
  // positive above would pass while every header rendered a live sort control.
  it('opts every other column out of sorting explicitly', () => {
    const sortKeys = new Set<string>(TODO_SORT_COLUMN_OPTIONS.map(option => option.value));
    const notOptedOut = todoTableColumns({ groupBy: 'none' })
      .filter(column => !sortKeys.has(column.key) && column.sortable !== false)
      .map(column => column.key);
    expect(notOptedOut).toEqual([]);
  });

  it('drops the workspace column only while grouping by workspace', () => {
    const keys = (groupBy: 'workspace' | 'severity') =>
      todoTableColumns({ groupBy }).map(column => column.key);
    expect(keys('workspace')).not.toContain('workspace');
    expect(keys('severity')).toContain('workspace');
  });

  it('covers every requested dimension', () => {
    expect(todoTableColumns({ groupBy: 'none' }).map(column => column.key)).toEqual([
      'status', 'title', 'workspace', 'priority',
      'phase.plan', 'phase.triage', 'phase.run', 'phase.verify',
      'tags', 'created', 'updated', 'signals',
    ]);
  });

  // Pipeline order, not map or alphabetical order: what you plan, you run; what
  // you run, you verify. Triage sits with plan as the other read-only pass.
  it('orders the phase columns down the pipeline', () => {
    const phaseKeys = todoTableColumns({ groupBy: 'none' })
      .map(column => column.key)
      .filter(key => key.startsWith('phase.'));
    expect(phaseKeys).toEqual(TODO_PHASES.map(phase => `phase.${phase}`));
  });
});

describe('title cell parent hint', () => {
  const PARENT_ID = '5d0c8f7e-1b2a-4c3d-9e4f-6a7b8c9d0e1f';
  const parentTitles = new Map([[PARENT_ID, 'Migrate billing']]);
  const cell = (candidate: TodoTableRow, titles?: Map<string, string>) => {
    const column = todoTableColumns({ groupBy: 'none', parentTitles: titles }).find(entry => entry.key === 'title')!;
    return render(<>{column.render!(candidate.todo.title, candidate)}</>);
  };
  const child = (): TodoTableRow => {
    const base = row('child-todo', 'medium');
    return { ...base, todo: { ...base.todo, parentId: PARENT_ID } } as TodoTableRow;
  };

  it('names the parent under which a search-surfaced child hangs', () => {
    const view = cell(child(), parentTitles);
    expect(view.getByTitle('Child of Migrate billing').textContent).toBe('↳ Migrate billing');
    expect(view.getByTitle('child-todo').textContent).toBe('child-todo');
  });

  it('shows no hint for a top-level todo', () => {
    const view = cell(row('top-level', 'medium'), parentTitles);
    expect(view.queryByText(/↳/)).toBeNull();
  });

  it('says so when the parent is not in the loaded list', () => {
    const view = cell(child(), new Map());
    expect(view.getByText(/↳/).textContent).toBe('↳ parent not loaded');
  });
});

describe('title cell child count', () => {
  const PARENT_ID = '7e1d2c3b-4a5f-4e6d-8c7b-9a0f1e2d3c4b';
  const childCounts = new Map([[PARENT_ID, { done: 1, total: 3 }]]);
  const cell = (candidate: TodoTableRow) => {
    const column = todoTableColumns({ groupBy: 'none', childCounts }).find(entry => entry.key === 'title')!;
    return render(<>{column.render!(candidate.todo.title, candidate)}</>);
  };
  const parent = (): TodoTableRow => {
    const base = row('parent-todo', 'medium');
    return { ...base, todo: { ...base.todo, id: PARENT_ID } } as TodoTableRow;
  };

  it("badges a parent with its children's done/total", () => {
    expect(cell(parent()).getByTitle('3 child todos, 1 done').textContent).toBe('1/3');
  });

  it('shows no badge on a todo without children', () => {
    expect(cell(row('lonely', 'medium')).queryByTitle(/child todo/)).toBeNull();
  });
});

describe('todoGroupingModes', () => {
  const now = Date.parse('2026-07-10T00:00:00Z');

  it('offers one mode per grouping dimension', () => {
    expect(todoGroupingModes({ workspaces, now }).map(mode => mode.value))
      .toEqual(['workspace', 'severity', 'age', 'none']);
  });

  // DataTable groups after sorting and otherwise keeps first-appearance order,
  // so without compareGroups the severity headers would reshuffle whenever the
  // sort column changed.
  it('orders severity groups high before medium before low', () => {
    const severity = customModes(now).find(mode => mode.value === 'severity')!;
    const groups = ['low', 'high', 'medium'].map(key => ({ key, rows: [] as TodoTableRow[] }));
    expect(groups.sort(severity.compareGroups!).map(group => group.key))
      .toEqual(['high', 'medium', 'low']);
  });

  it('buckets a row by its priority and its workspace', () => {
    const [workspace, severity] = customModes(now);
    expect(workspace.getGroupKey(row('a', 'high', '/repos/beta'))).toBe('/repos/beta');
    expect(workspace.getGroupLabel!('/repos/beta', [])).toBe('beta');
    expect(severity.getGroupKey(row('a', 'high'))).toBe('high');
    // Providers default an unset priority to medium; the grouping must agree.
    expect(severity.getGroupKey(row('b', '' as TodoPriority))).toBe('medium');
  });
});

describe('todoTableRowId', () => {
  // The row id doubles as the bulk-selection key, so a todo checked in the
  // table is the same todo the bulk bar acts on.
  it('matches the bulk selection key for the same todo', () => {
    expect(todoTableRowId(row('abc', 'high'))).toBe(`/repos/alpha${NUL}abc`);
  });

  it('distinguishes the same ref in different workspaces', () => {
    expect(todoTableRowId(row('dup', 'high', '/repos/alpha')))
      .not.toBe(todoTableRowId(row('dup', 'high', '/repos/beta')));
  });
});
