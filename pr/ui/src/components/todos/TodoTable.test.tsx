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
