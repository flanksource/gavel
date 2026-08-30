import { render, screen } from '@testing-library/react';
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

const { TodoWorkspaceList } = await import('./TodoView');

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
