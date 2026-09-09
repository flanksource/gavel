import { render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Project, SessionStats, TodoItem, TodoListResponse } from '../types';
import { emptyCounts } from './todos/format';
import { defaultTodoFilters } from './todos/todoFilter';
import type { TodoSelection } from './todos/todoSelection';
import type { WorkspaceTodos } from './todos/useWorkspaceTodos';

vi.mock('./todos/TodoToolbar', () => ({
  TodoToolbar: () => <div data-testid="todo-toolbar" />,
}));

const useSessionStats = vi.hoisted(() => vi.fn(() => ({ stats: null as SessionStats | null, elapsedMs: 0, error: '' })));
vi.mock('./todos/TodoSessionTimer', async (importOriginal) => ({
  ...(await importOriginal<object>()),
  useSessionStats,
}));

// The table is stubbed down to a prop recorder: what is under test is the array
// TodoFullPane hands it, not what a table does with it.
const seenRows: unknown[] = [];
vi.mock('./todos/TodoTable', () => ({
  TodoTable: ({ rows }: { rows: unknown[] }) => {
    seenRows.push(rows);
    return <div data-testid="todo-table" />;
  },
  todoTableColumns: () => [],
}));
vi.mock('./todos/TodoDetail', () => ({
  TodoDetail: () => <div data-testid="todo-detail" />,
}));

const { TodoFullPane } = await import('./TodoView');

const gavel: Project = { name: 'gavel', dir: '/repos/gavel', repos: [] } as Project;

function todos(): WorkspaceTodos {
  const items: TodoItem[] = Array.from({ length: 5 }, (_, i) => ({
    ref: `g${i}`,
    title: `todo ${i}`,
    status: 'pending',
    priority: 'medium',
  }));
  const byDir: Record<string, TodoListResponse> = {
    [gavel.dir]: {
      dir: gavel.dir,
      counts: { ...emptyCounts, total: items.length, open: items.length, pending: items.length },
      items,
    },
  };
  return {
    workspaces: [gavel],
    byDir,
    filters: defaultTodoFilters(),
    toggleStatus: vi.fn(),
    density: 'comfortable',
    groupBy: 'workspace',
    sortBy: { column: 'priority', dir: 'desc' },
    timeRange: null,
    selected: null,
    select: vi.fn(),
    loadingList: false,
    error: '',
    selection: undefined as unknown as TodoSelection,
    tagsByDir: undefined,
  } as unknown as WorkspaceTodos;
}

describe('TodoFullPane rows identity', () => {
  // DataTable keys its row window on `data` identity: a new array — even one
  // holding exactly the same rows — resets the window to the first batch and
  // brings back the "Loading more…" sentinel. TodoFullPane used to allocate one
  // on every render via `matched.map(e => e as TodoTableRow)`, a cast that does
  // nothing at runtime, so every keystroke and every todos state change threw
  // away every row the reader had scrolled to.
  it('hands the table the same array across re-renders when nothing changed', () => {
    seenRows.length = 0;
    const props = todos();
    const { rerender } = render(<TodoFullPane todos={props} projectsLoaded />);
    rerender(<TodoFullPane todos={props} projectsLoaded />);
    rerender(<TodoFullPane todos={props} projectsLoaded />);

    expect(seenRows.length).toBeGreaterThanOrEqual(3);
    for (const rows of seenRows) expect(rows).toBe(seenRows[0]);
  });

  it('still hands over a new array when the todos actually change', () => {
    seenRows.length = 0;
    const { rerender } = render(<TodoFullPane todos={todos()} projectsLoaded />);
    const before = seenRows[seenRows.length - 1];

    const changed = todos();
    (changed.byDir[gavel.dir] as TodoListResponse).items = [
      { ref: 'only', title: 'only one left', status: 'pending', priority: 'medium' },
    ];
    rerender(<TodoFullPane todos={changed} projectsLoaded />);

    expect(seenRows[seenRows.length - 1]).not.toBe(before);
    expect((seenRows[seenRows.length - 1] as unknown[]).length).toBe(1);
  });
});
