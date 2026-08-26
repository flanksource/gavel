import type React from 'react';
import { render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { FilterBarFilter } from '@flanksource/clicky-ui/components';
import type { Project, TodoItem, TodoListResponse } from '../../types';
import { emptyCounts } from './format';
import { defaultTodoFilters } from './todoFilter';
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

function todo(ref: string, overrides: Partial<TodoItem> = {}): TodoItem {
  return { ref, title: ref, status: 'pending', priority: 'medium', ...overrides };
}

function toolbarProps(items: TodoItem[], filters = defaultTodoFilters()): WorkspaceTodos {
  const byDir: Record<string, TodoListResponse> = {
    [workspace.dir]: { dir: workspace.dir, counts: { ...emptyCounts, total: items.length, open: items.length, pending: items.length }, items },
  };
  return {
    workspaces: [workspace],
    byDir,
    aggregate: byDir[workspace.dir].counts,
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

// Bulk-edit mode off: the toolbar's contract here is the filter descriptors, and
// TodoBulkBar renders nothing (and mounts no mutation) while bulkMode is false.
function idleSelection(): TodoSelection {
  return {
    bulkMode: false,
    setBulkMode: vi.fn(),
    selection: new Set<string>(),
    isSelected: () => false,
    toggleSelected: vi.fn(),
    setGroupSelected: vi.fn(),
    clearSelection: vi.fn(),
    targets: [],
  };
}

function renderToolbar(items: TodoItem[], filters = defaultTodoFilters()) {
  captured.props = {};
  render(<TodoToolbar todos={toolbarProps(items, filters)} />);
  return {
    filters: (captured.props.filters ?? []) as FilterBarFilter[],
    className: String(captured.props.className ?? ''),
  };
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

  it('keeps the external facet visible while a stale exclusion is still applied', () => {
    const stale = { ...defaultTodoFilters(), external: { linked: 'exclude' as const } };
    expect(renderToolbar([todo('a')], stale).filters.map(f => f.key)).toContain('external');
  });

  it('binds each facet to its own slice of the shared filter state', () => {
    const filters = { statuses: { failed: 'include' as const }, priorities: { high: 'exclude' as const }, external: {} };
    const built = renderToolbar([todo('a', { priority: 'high', status: 'failed' })], filters).filters;
    expect((built.find(f => f.key === 'status') as { value: unknown }).value).toEqual({ failed: 'include' });
    expect((built.find(f => f.key === 'priority') as { value: unknown }).value).toEqual({ high: 'exclude' });
  });
});
