import type React from 'react';
import { render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { FilterBarFilter } from '@flanksource/clicky-ui/components';
import type { Project, TodoItem, TodoListResponse } from '../../types';
import { addCounts, emptyCounts } from './format';
import { defaultTodoFilters, type TodoFilters } from './todoFilter';
import type { TodoSelection } from './todoSelection';
import type { WorkspaceTodos } from './useWorkspaceTodos';

// clicky-ui's FilterBar drags in @floating-ui/react, which is unstable under
// jsdom here (see TodoSession.test.tsx). The toolbar's contract is the filter
// descriptors it builds, so the bar is captured rather than rendered.
const captured: { props: Record<string, unknown> } = { props: {} };

vi.mock('@flanksource/clicky-ui/components', () => ({
  FilterBar: (props: Record<string, unknown>) => {
    captured.props = props;
    return <div data-testid="filter-bar" />;
  },
  // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky Button itself.
  Button: ({ children, variant: _variant, size: _size, ...rest }: React.ButtonHTMLAttributes<HTMLButtonElement> & { variant?: string; size?: string }) => (
    <button type="button" {...rest}>{children}</button>
  ),
  ListMenu: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
  ListMenuHeader: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
  ListMenuSection: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
}));

vi.mock('@flanksource/clicky-ui/icons', async (importOriginal) => {
  const Icon = (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />;
  return { ...(await importOriginal<object>()), UiClose: Icon, UiLinkExternal: Icon, UiRefresh: Icon, UiWarningTriangle: Icon };
});

const { TodoToolbar } = await import('./TodoToolbar');

const workspace: Project = { name: 'gavel', dir: '/repos/gavel' } as Project;
const captainWorkspace: Project = { name: 'captain', dir: '/repos/captain' } as Project;

function todo(ref: string, overrides: Partial<TodoItem> = {}): TodoItem {
  return { ref, title: ref, status: 'pending', priority: 'medium', ...overrides };
}

function workspaceData(ws: Project, items: TodoItem[]): TodoListResponse {
  return { dir: ws.dir, counts: { ...emptyCounts, total: items.length, open: items.length, pending: items.length }, items };
}

function toolbarProps(loaded: [Project, TodoItem[]][], filters: TodoFilters): WorkspaceTodos {
  const byDir: Record<string, TodoListResponse> = Object.fromEntries(
    loaded.map(([ws, items]) => [ws.dir, workspaceData(ws, items)]),
  );
  return {
    workspaces: loaded.map(([ws]) => ws),
    byDir,
    aggregate: loaded.reduce((acc, [ws]) => addCounts(acc, byDir[ws.dir].counts), emptyCounts),
    filters,
    setFilters: vi.fn(),
    groupBy: 'workspace',
    setGroupBy: vi.fn(),
    sortBy: { column: 'priority', dir: 'desc' },
    setSortBy: vi.fn(),
    timeRange: null,
    setTimeRange: vi.fn(),
    loadingList: false,
    refresh: vi.fn(),
    selection: idleSelection(),
  } as unknown as WorkspaceTodos;
}

// Nothing checked: the toolbar's contract here is the filter descriptors, and
// TodoSelectionBar renders nothing (and fetches no catalog) on an empty
// selection.
function idleSelection(): TodoSelection {
  return {
    selection: new Set<string>(),
    isSelected: () => false,
    toggleSelected: vi.fn(),
    setGroupSelected: vi.fn(),
    clearSelection: vi.fn(),
    replaceSelection: vi.fn(),
    targets: [],
  };
}

function renderWorkspaces(loaded: [Project, TodoItem[]][], filters = defaultTodoFilters()) {
  captured.props = {};
  render(<TodoToolbar todos={toolbarProps(loaded, filters)} />);
  return {
    filters: (captured.props.filters ?? []) as FilterBarFilter[],
    className: String(captured.props.className ?? ''),
  };
}

function renderToolbar(items: TodoItem[], filters = defaultTodoFilters()) {
  return renderWorkspaces([[workspace, items]], filters);
}

const linkedIssue = { kind: 'github', repo: 'acme/api', number: 11, url: 'https://github.com/acme/api/issues/11' };

describe('TodoToolbar', () => {
  it('never lets the bar wrap', () => {
    expect(renderToolbar([todo('a')]).className).toContain('flex-nowrap');
  });

  it('offers status and priority facets with their counts in the option labels', () => {
    const { filters } = renderToolbar([
      todo('a', { priority: 'high' }),
      todo('b', { priority: 'high' }),
      todo('c', { priority: 'low' }),
    ]);
    const priority = filters.find(f => f.key === 'priority');
    expect(priority?.kind).toBe('multi');
    expect((priority as { options: { value: string; label: string }[] }).options).toEqual([
      { value: 'high', label: 'High (2)' },
      { value: 'low', label: 'Low (1)' },
    ]);

    const status = filters.find(f => f.key === 'status');
    expect((status as { options: { label: string }[] }).options).toContainEqual({ value: 'pending', label: 'Pending (3)' });
  });

  it('omits the external-issue facet while nothing is linked', () => {
    expect(renderToolbar([todo('a'), todo('b')]).filters.map(f => f.key)).toEqual(['status', 'priority', 'activity']);
  });

  it('puts the activity range last so it is the first control to fold away', () => {
    const keys = renderToolbar([todo('a', { externalIssue: linkedIssue })]).filters.map(f => f.key);
    expect(keys[keys.length - 1]).toBe('activity');
  });

  it('offers linked and not-linked counts once a todo carries an external issue', () => {
    const { filters } = renderToolbar([todo('a', { externalIssue: linkedIssue }), todo('b'), todo('c')]);
    const external = filters.find(f => f.key === 'external');
    expect((external as { options: { value: string; label: string }[] }).options).toEqual([
      { value: 'linked', label: 'Linked (1)' },
      { value: 'unlinked', label: 'Not linked (2)' },
    ]);
  });

  it('offers labels as an include/exclude facet bound to label filter state', () => {
    const active = { ...defaultTodoFilters(), tags: { bug: 'include' as const, area: 'exclude' as const } };
    const facet = renderToolbar([
      todo('a', { labels: ['bug', 'area:ui'] }),
      todo('b', { labels: ['area:api'] }),
    ], active).filters.find(f => f.key === 'tags');

    expect(facet).toMatchObject({
      kind: 'multi',
      label: 'Labels',
      value: { bug: 'include', area: 'exclude' },
    });
    expect((facet as { options: { value: string; label: string }[] }).options).toEqual([
      { value: 'area', label: 'area (2)' },
      { value: 'area:api', label: 'area:api (1)' },
      { value: 'area:ui', label: 'area:ui (1)' },
      { value: 'bug', label: 'bug (1)' },
    ]);
  });

  it('keeps the external facet visible while a stale exclusion is still applied', () => {
    const stale = { ...defaultTodoFilters(), external: { linked: 'exclude' as const } };
    expect(renderToolbar([todo('a')], stale).filters.map(f => f.key)).toContain('external');
  });

  it('omits the workspace facet when only one workspace is configured', () => {
    expect(renderToolbar([todo('a')]).filters.map(f => f.key)).not.toContain('workspace');
  });

  it('leads with a workspace facet counting each workspace once more than one is configured', () => {
    const { filters } = renderWorkspaces([
      [workspace, [todo('a'), todo('b')]],
      [captainWorkspace, [todo('c')]],
    ]);
    expect(filters[0].key).toBe('workspace');
    expect((filters[0] as { options: { value: string; label: string }[] }).options).toEqual([
      { value: '/repos/gavel', label: 'gavel (2)' },
      { value: '/repos/captain', label: 'captain (1)' },
    ]);
  });

  // Keyed by directory: two checkouts named the same must stay separately
  // selectable.
  it('keys each workspace option by its directory, not its display name', () => {
    const fork: Project = { name: 'gavel', dir: '/repos/gavel-fork' } as Project;
    const { filters } = renderWorkspaces([[workspace, [todo('a')]], [fork, [todo('b')]]]);
    expect((filters[0] as { options: { value: string }[] }).options.map(o => o.value))
      .toEqual(['/repos/gavel', '/repos/gavel-fork']);
  });

  // Dropping the selected workspace from projects.json would otherwise empty the
  // list with no control left to undo it.
  it('keeps a selection on a no-longer-configured workspace clearable', () => {
    const stale = { ...defaultTodoFilters(), workspaces: { '/repos/captain': 'include' as const } };
    const facet = renderWorkspaces([[workspace, [todo('a')]]], stale).filters.find(f => f.key === 'workspace');
    expect((facet as { options: { value: string; label: string }[] }).options).toEqual([
      { value: '/repos/gavel', label: 'gavel (1)' },
      { value: '/repos/captain', label: '/repos/captain (0)' },
    ]);
    expect((facet as { value: unknown }).value).toEqual({ '/repos/captain': 'include' });
  });

  it('binds each facet to its own slice of the shared filter state', () => {
    const filters = { statuses: { failed: 'include' as const }, priorities: { high: 'exclude' as const }, external: {} , tags: {}, workspaces: {} };
    const built = renderToolbar([todo('a', { priority: 'high', status: 'failed' })], filters).filters;
    expect((built.find(f => f.key === 'status') as { value: unknown }).value).toEqual({ failed: 'include' });
    expect((built.find(f => f.key === 'priority') as { value: unknown }).value).toEqual({ high: 'exclude' });
  });
});
