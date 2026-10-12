import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Project, SessionStats, TodoItem, TodoListResponse } from '../types';
import { emptyCounts } from './todos/format';
import { defaultTodoFilters, type TodoFilters } from './todos/todoFilter';
import type { TodoSelection } from './todos/todoSelection';
import type { WorkspaceTodos } from './todos/useWorkspaceTodos';

// The toolbar is clicky-ui's FilterBar, which drags in @floating-ui/react and is
// unstable under jsdom (see TodoSession.test.tsx). What is under test here is the
// workspace sections the list renders beneath it, so the row is stubbed out.
vi.mock('./todos/TodoToolbar', () => ({
  TodoToolbar: () => <div data-testid="todo-toolbar" />,
}));

const useSessionStats = vi.hoisted(() => vi.fn(() => ({ stats: null as SessionStats | null, elapsedMs: 0, error: '' })));
vi.mock('./todos/TodoSessionTimer', async (importOriginal) => ({
  ...(await importOriginal<object>()),
  useSessionStats,
}));

// The full-width pane's two halves are stubbed: DataTable and the detail's
// tabs both pull in @floating-ui/react, and what is under test here is which of
// the two is mounted, not what either renders.
vi.mock('./todos/TodoTable', () => ({
  TodoTable: () => <div data-testid="todo-table" />,
  todoTableColumns: () => [],
}));
vi.mock('./todos/TodoDetail', () => ({
  TodoDetail: ({ navigation }: { navigation?: { position: number; total: number; onNext: () => void } }) => (
    <div data-testid="todo-detail" data-navigation={navigation ? `${navigation.position}/${navigation.total}` : ''}>
      {navigation && <button type="button" onClick={navigation.onNext}>Next todo</button>}
    </div>
  ),
}));

const { TodoFullPane, TodoWorkspaceList } = await import('./TodoView');

const gavel: Project = { name: 'gavel', dir: '/repos/gavel', repos: [] } as Project;
const captain: Project = { name: 'captain', dir: '/repos/captain', repos: [] } as Project;

function todo(ref: string): TodoItem {
  return { ref, title: `todo ${ref}`, status: 'pending', priority: 'medium' };
}

function listProps(filters: TodoFilters): WorkspaceTodos {
  const loaded: [Project, TodoItem[]][] = [[gavel, [todo('g1')]], [captain, [todo('c1')]]];
  const byDir: Record<string, TodoListResponse> = Object.fromEntries(loaded.map(([ws, items]) => [
    ws.dir,
    { dir: ws.dir, counts: { ...emptyCounts, total: items.length, open: items.length, pending: items.length }, items },
  ] as [string, TodoListResponse]));
  return {
    workspaces: loaded.map(([ws]) => ws),
    byDir,
    filters,
    toggleStatus: vi.fn(),
    density: 'comfortable',
    groupBy: 'workspace',
    sortBy: { column: 'priority', dir: 'desc' },
    timeRange: null,
    selected: null,
    select: vi.fn(),
    loadingList: false,
    listReady: true,
    errorsByDir: {},
    detail: null,
    error: '',
    selection: undefined as unknown as TodoSelection,
    tagsByDir: undefined,
  } as unknown as WorkspaceTodos;
}

function renderList(filters = defaultTodoFilters()) {
  return render(<TodoWorkspaceList todos={listProps(filters)} projectsLoaded />);
}

describe('TodoWorkspaceList workspace sections', () => {
  it('lists every configured workspace while the workspace facet is neutral', () => {
    renderList();
    expect(screen.getByText('gavel')).toBeTruthy();
    expect(screen.getByText('captain')).toBeTruthy();
  });

  // The whole section goes, header and counts included: a workspace left showing
  // "0 open" would read as one that had gone quiet, not one filtered away.
  it('drops an excluded workspace section entirely', () => {
    renderList({ ...defaultTodoFilters(), workspaces: { [captain.dir]: 'exclude' } });
    expect(screen.getByText('gavel')).toBeTruthy();
    expect(screen.queryByText('captain')).toBeNull();
    expect(screen.queryByText('todo c1')).toBeNull();
  });

  it('narrows to the included workspace and hides the rest', () => {
    renderList({ ...defaultTodoFilters(), workspaces: { [gavel.dir]: 'include' } });
    expect(screen.getByText('gavel')).toBeTruthy();
    expect(screen.queryByText('captain')).toBeNull();
  });

  // Excluding every workspace must say so rather than render a bare empty list
  // that looks like a failed load.
  it('explains an empty list when the facet excludes every workspace', () => {
    renderList({
      ...defaultTodoFilters(),
      workspaces: { [gavel.dir]: 'exclude', [captain.dir]: 'exclude' },
    });
    expect(screen.getByText('No workspaces match the filter')).toBeTruthy();
  });
});

describe('TodoFullPane', () => {
  const paneProps = (selected: WorkspaceTodos['selected']): WorkspaceTodos => ({
    ...listProps(defaultTodoFilters()),
    selected,
  });

  it('shows the table alone until a todo is selected', () => {
    render(<TodoFullPane todos={paneProps(null)} projectsLoaded />);
    expect(screen.getByTestId('todo-table')).toBeTruthy();
    expect(screen.queryByTestId('todo-detail')).toBeNull();
  });

  // Opening a todo used to swap the table out for the detail, which destroyed
  // the table's scroll offset: Back put the reader back at the top of the list
  // however far down it they had been.
  it('keeps the table mounted behind an open todo so Back returns to the same row', () => {
    render(<TodoFullPane todos={paneProps({ dir: gavel.dir, ref: 'g1' })} projectsLoaded />);
    expect(screen.getByTestId('todo-detail')).toBeTruthy();
    expect(screen.getByTestId('todo-table')).toBeTruthy();
  });

  // Mounted is not the same as reachable: the covered table must leave the tab
  // order and the accessibility tree, or the detail is read through it.
  it('takes the covered table out of the tab order', () => {
    render(<TodoFullPane todos={paneProps({ dir: gavel.dir, ref: 'g1' })} projectsLoaded />);
    const layer = screen.getByTestId('todo-table').parentElement!;
    expect(layer.hasAttribute('inert')).toBe(true);
  });

  it('passes the current list position to the selected todo', () => {
    render(<TodoFullPane todos={paneProps({ dir: gavel.dir, ref: 'g1' })} projectsLoaded />);
    expect(screen.getByTestId('todo-detail').getAttribute('data-navigation')).toBe('1/2');
  });

  it('suppresses general navigation while plan review owns the queue', () => {
    render(
      <TodoFullPane
        todos={paneProps({ dir: gavel.dir, ref: 'g1' })}
        projectsLoaded
        navigationEnabled={false}
      />,
    );
    expect(screen.getByTestId('todo-detail').getAttribute('data-navigation')).toBe('');
  });

  // Editing the open todo refetches the list. Navigation used to re-sort and
  // re-filter it by the edited values, so J jumped to the neighbour of the
  // todo's new slot, or lost the todo entirely once a filter hid it.
  it.each([
    ['raising its priority re-sorts it to the top', { priority: 'high' }],
    ['completing it hides it from the default filter', { status: 'completed' }],
  ] as const)('keeps browsing from the open todo\'s slot when %s', (_, edit) => {
    const base = listProps(defaultTodoFilters());
    const [g1, g2, g3] = [todo('g1'), todo('g2'), todo('g3')];
    const withItems = (items: TodoItem[]): WorkspaceTodos => ({
      ...base,
      workspaces: [gavel],
      byDir: { [gavel.dir]: { dir: gavel.dir, counts: emptyCounts, items } },
      selected: { dir: gavel.dir, ref: g2.ref },
    });
    const { rerender } = render(<TodoFullPane todos={withItems([g1, g2, g3])} projectsLoaded />);
    expect(screen.getByTestId('todo-detail').getAttribute('data-navigation')).toBe('2/3');

    rerender(<TodoFullPane todos={withItems([g1, { ...g2, ...edit }, g3])} projectsLoaded />);
    expect(screen.getByTestId('todo-detail').getAttribute('data-navigation')).toBe('2/3');
    fireEvent.click(screen.getByRole('button', { name: 'Next todo' }));
    expect(base.select).toHaveBeenCalledWith({ dir: gavel.dir, ref: g3.ref });
  });

  // The ordinary queue leaves children out, so a child opened from its parent
  // used to have no position at all and lost its previous/next controls.
  describe('a child todo', () => {
    const parentId = '5e7a9c1b-2d4f-4680-a3b5-7c9e1f3a5b70';
    const parent = { ...todo('p1'), id: parentId };
    const [c1, c2, c3] = ['c1', 'c2', 'c3'].map(ref => ({ ...todo(ref), parentId }));
    const other = { ...todo('g9'), priority: 'high' as const };
    const base = listProps(defaultTodoFilters());
    const withItems = (items: TodoItem[], ref: string): WorkspaceTodos => ({
      ...base,
      workspaces: [gavel],
      byDir: { [gavel.dir]: { dir: gavel.dir, counts: emptyCounts, items } },
      selected: { dir: gavel.dir, ref },
    });
    const position = () => screen.getByTestId('todo-detail').getAttribute('data-navigation');

    it('walks its siblings with previous and next', () => {
      render(<TodoFullPane todos={withItems([parent, c1, c2, other, c3], 'c2')} projectsLoaded />);

      expect(position()).toBe('2/3');
      fireEvent.click(screen.getByRole('button', { name: 'Next todo' }));
      expect(base.select).toHaveBeenLastCalledWith({ dir: gavel.dir, ref: 'c3' });
    });

    // The navigation pin holds the pre-edit copy of the open todo; once it has no
    // parent it belongs to the ordinary queue and must be navigable there.
    it('is navigable in the ordinary queue once it is detached from its parent', () => {
      const { rerender } = render(<TodoFullPane todos={withItems([parent, c1, c2, other, c3], 'c2')} projectsLoaded />);
      expect(position()).toBe('2/3');

      // Detached and demoted, so the ordinary queue (g9, p1, c2) places it last.
      const detached = { ...c2, parentId: undefined, priority: 'low' as const };
      rerender(<TodoFullPane todos={withItems([parent, c1, detached, other, c3], 'c2')} projectsLoaded />);
      expect(position()).toBe('3/3');
    });
  });
});
