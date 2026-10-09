import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Project, TodoItem } from '../../types';
import type { TodoSelection } from './todoSelection';
import type { WorkspaceTodos } from './useWorkspaceTodos';

const renderedTable = vi.hoisted(() => ({ onRowClick: undefined as unknown }));

vi.mock('@flanksource/clicky-ui/data', async (importOriginal) => {
  const original = await importOriginal<object>();
  return {
    ...original,
    DataTable: ({ data, getRowHref, onRowClick }: {
      data: unknown[];
      getRowHref?: (row: unknown) => string | undefined;
      onRowClick?: (row: unknown) => void;
    }) => {
      renderedTable.onRowClick = onRowClick;
      return data.map((row, index) => {
        const title = (row as { todo: TodoItem }).todo.title;
        const href = getRowHref?.(row);
        return href
          ? <a key={index} href={href}>{title}</a>
          : <span key={index}>{title}</span>;
      });
    },
  };
});

vi.mock('./todoFilterBar', () => ({
  TODO_ACTIVITY_FILTER_LABEL: 'Active',
  useTodoFilterBar: () => ({ facets: [], range: { onApply: vi.fn() } }),
}));

vi.mock('./todoActions', () => ({
  useTodoBulkContext: () => ({ todos: [], tags: undefined, labelCounts: {} }),
  useTodoBulkToolbar: () => [],
}));

const { TodoTable } = await import('./TodoTable');

function tableFixture() {
  const workspace = { name: 'mission-control-oipa', dir: '/repos/oipa', repos: [] } as Project;
  const todo: TodoItem = {
    ref: 'native/issue 7',
    title: 'Preserve fixture upload paths and import closures',
    status: 'pending',
    priority: 'medium',
  };
  const row = { todo, workspace };
  const selection = {
    selection: new Set<string>(),
    replaceSelection: vi.fn(),
  } as unknown as TodoSelection;
  const todos = {
    workspaces: [workspace],
    byDir: {},
    timeRange: null,
    setTimeRange: vi.fn(),
    density: 'comfortable',
    groupBy: 'none',
    setGroupBy: vi.fn(),
    sortBy: { column: 'priority', dir: 'desc' },
    setSortBy: vi.fn(),
    select: vi.fn(),
    selected: null,
    loadingList: false,
    refresh: vi.fn(),
    error: '',
    selection,
  } as unknown as WorkspaceTodos;
  return { row, todos };
}

describe('TodoTable row links', () => {
  it('renders each todo as a scoped deep link without an imperative row click handler', () => {
    const { row, todos } = tableFixture();

    render(
      <TodoTable
        todos={todos}
        projectsLoaded
        rows={[row]}
        columns={[]}
        query=""
        onQueryChange={vi.fn()}
        scopeProject="mission-control-oipa"
      />,
    );

    expect(screen.getByRole('link', { name: row.todo.title }).getAttribute('href'))
      .toBe('/todos/native/issue%207?project=mission-control-oipa');
    expect(renderedTable.onRowClick).toBeUndefined();
  });
});
