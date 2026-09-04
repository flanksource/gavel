import { beforeEach, describe, expect, it } from 'vitest';
import type { Project, TodoItem, TodoPriority, TodoStatus } from '../../types';
import type { TodoEntry } from './todoGroup';
import {
  defaultTodoFilters,
  externalKey,
  isEntryVisible,
  isStatusShown,
  isTodoVisible,
  isWorkspaceShown,
  loadTodoFilters,
  priorityKey,
  todoMatchesQuery,
  toggleStatusFilter,
  LEGACY_HIDDEN_STATUS_KEY,
  LEGACY_TODO_FILTERS_V2_KEY,
  TODO_FILTERS_KEY,
} from './todoFilter';

const todo: TodoItem = {
  ref: '3785b0a4-0bf6-4f65-b1c2-41eab73e9f6b',
  id: '3785b0a4-0bf6-4f65-b1c2-41eab73e9f6b',
  title: 'Parse user shell commands',
  status: 'completed',
  priority: 'medium',
  sessionId: '019f5b2e-75b7-7de2-911b-de8b70266479',
};

function item(overrides: Partial<TodoItem> & { ref: string }): TodoItem {
  return { title: `todo ${overrides.ref}`, status: 'pending', priority: 'medium', ...overrides };
}

const linked = item({
  ref: 'linked-high',
  priority: 'high',
  externalIssue: { kind: 'github', repo: 'acme/api', number: 11, url: 'https://github.com/acme/api/issues/11' },
});
const unlinked = item({ ref: 'unlinked-low', priority: 'low' });

describe('todoMatchesQuery', () => {
  it('matches both Todo and current session UUID fragments', () => {
    expect(todoMatchesQuery(todo, '3785b0a4')).toBe(true);
    expect(todoMatchesQuery(todo, '019F5B2E-75B7')).toBe(true);
  });

  it('does not treat an unrelated UUID fragment as a match', () => {
    expect(todoMatchesQuery(todo, 'aaaaaaaa-bbbb')).toBe(false);
  });
});

describe('facet keys', () => {
  it('reads an absent priority as medium, matching the severity buckets', () => {
    expect(priorityKey(item({ ref: 'a' }))).toBe('medium');
    expect(priorityKey(item({ ref: 'b', priority: 'high' }))).toBe('high');
    expect(priorityKey(item({ ref: 'c', priority: 'nonsense' as TodoPriority }))).toBe('medium');
  });

  it('splits todos on whether they carry an external issue', () => {
    expect(externalKey(linked)).toBe('linked');
    expect(externalKey(unlinked)).toBe('unlinked');
  });
});

describe('isTodoVisible', () => {
  it('hides closed todos by default and shows open ones', () => {
    const filters = defaultTodoFilters();
    expect(isTodoVisible(item({ ref: 'a', status: 'completed' }), filters)).toBe(false);
    expect(isTodoVisible(item({ ref: 'b', status: 'pending' }), filters)).toBe(true);
  });

  it('includes only the selected priorities when any priority is included', () => {
    const filters = { ...defaultTodoFilters(), priorities: { high: 'include' as const } };
    expect(isTodoVisible(linked, filters)).toBe(true);
    expect(isTodoVisible(unlinked, filters)).toBe(false);
  });

  it('excludes a priority without narrowing the others', () => {
    const filters = { ...defaultTodoFilters(), priorities: { low: 'exclude' as const } };
    expect(isTodoVisible(linked, filters)).toBe(true);
    expect(isTodoVisible(unlinked, filters)).toBe(false);
    expect(isTodoVisible(item({ ref: 'mid' }), filters)).toBe(true);
  });

  it('includes only todos carrying an included label', () => {
    const labeled = item({ ref: 'labeled', labels: ['area:ui', 'bug'] });
    const unlabeled = item({ ref: 'unlabeled' });
    const filters = { ...defaultTodoFilters(), tags: { bug: 'include' as const } };

    expect(isTodoVisible(labeled, filters)).toBe(true);
    expect(isTodoVisible(unlabeled, filters)).toBe(false);
  });

  it('excludes matching labels without narrowing other todos', () => {
    const ui = item({ ref: 'ui', labels: ['area:ui'] });
    const api = item({ ref: 'api', labels: ['area:api'] });
    const unrelated = item({ ref: 'unrelated', labels: ['bug'] });
    const filters = { ...defaultTodoFilters(), tags: { area: 'exclude' as const } };

    expect(isTodoVisible(ui, filters)).toBe(false);
    expect(isTodoVisible(api, filters)).toBe(false);
    expect(isTodoVisible(unrelated, filters)).toBe(true);
  });

  it('filters on whether a todo is linked to an external issue', () => {
    const onlyLinked = { ...defaultTodoFilters(), external: { linked: 'include' as const } };
    expect(isTodoVisible(linked, onlyLinked)).toBe(true);
    expect(isTodoVisible(unlinked, onlyLinked)).toBe(false);

    const onlyUnlinked = { ...defaultTodoFilters(), external: { unlinked: 'include' as const } };
    expect(isTodoVisible(linked, onlyUnlinked)).toBe(false);
    expect(isTodoVisible(unlinked, onlyUnlinked)).toBe(true);
  });

  it('requires every facet to pass, not just one', () => {
    const filters = {
      statuses: {},
      priorities: { high: 'include' as const },
      external: { unlinked: 'include' as const },
      tags: {},
      workspaces: {},
    };
    // linked is high priority but linked; unlinked is unlinked but low priority.
    expect(isTodoVisible(linked, filters)).toBe(false);
    expect(isTodoVisible(unlinked, filters)).toBe(false);
    expect(isTodoVisible(item({ ref: 'both', priority: 'high' }), filters)).toBe(true);
  });

  it('still applies the activity range alongside the facets', () => {
    const recent = item({ ref: 'recent', priority: 'high', lastRun: '2026-08-11T00:00:00Z' });
    const range = { from: Date.parse('2026-08-10T00:00:00Z'), to: Date.parse('2026-08-12T00:00:00Z') };
    const stale = { from: Date.parse('2026-01-01T00:00:00Z'), to: Date.parse('2026-01-02T00:00:00Z') };
    expect(isTodoVisible(recent, defaultTodoFilters(), range)).toBe(true);
    expect(isTodoVisible(recent, defaultTodoFilters(), stale)).toBe(false);
  });
});

describe('the workspace facet', () => {
  const gavel: Project = { name: 'gavel', dir: '/repos/gavel', repos: [] };
  const captain: Project = { name: 'captain', dir: '/repos/captain', repos: [] };
  const inGavel: TodoEntry = { todo: item({ ref: 'g1' }), workspace: gavel };
  const inCaptain: TodoEntry = { todo: item({ ref: 'c1' }), workspace: captain };

  it('shows every workspace while the facet is neutral', () => {
    expect(isWorkspaceShown(defaultTodoFilters(), gavel.dir)).toBe(true);
    expect(isWorkspaceShown(defaultTodoFilters(), captain.dir)).toBe(true);
  });

  it('narrows to the included workspace and hides the rest', () => {
    const filters = { ...defaultTodoFilters(), workspaces: { [gavel.dir]: 'include' as const } };
    expect(isWorkspaceShown(filters, gavel.dir)).toBe(true);
    expect(isWorkspaceShown(filters, captain.dir)).toBe(false);
  });

  it('excludes one workspace without narrowing the others', () => {
    const filters = { ...defaultTodoFilters(), workspaces: { [captain.dir]: 'exclude' as const } };
    expect(isWorkspaceShown(filters, gavel.dir)).toBe(true);
    expect(isWorkspaceShown(filters, captain.dir)).toBe(false);
  });

  // Keyed on the directory, not the display name: two checkouts of the same repo
  // share a name, and selecting one must not drag in the other.
  it('keys on the directory rather than the workspace name', () => {
    const fork: Project = { name: 'gavel', dir: '/repos/gavel-fork', repos: [] };
    const filters = { ...defaultTodoFilters(), workspaces: { [gavel.dir]: 'include' as const } };
    expect(isWorkspaceShown(filters, fork.dir)).toBe(false);
  });

  it('applies alongside the todo facets in a flattened list', () => {
    const onlyGavel = { ...defaultTodoFilters(), workspaces: { [gavel.dir]: 'include' as const } };
    expect(isEntryVisible(inGavel, onlyGavel)).toBe(true);
    expect(isEntryVisible(inCaptain, onlyGavel)).toBe(false);

    // The workspace matches but the status facet does not: both must pass.
    const closedInGavel: TodoEntry = { todo: item({ ref: 'g2', status: 'completed' }), workspace: gavel };
    expect(isEntryVisible(closedInGavel, onlyGavel)).toBe(false);
  });
});

describe('status badge helpers', () => {
  it('treats neutral as shown and excluded as hidden', () => {
    expect(isStatusShown({}, 'pending')).toBe(true);
    expect(isStatusShown({ pending: 'exclude' }, 'pending')).toBe(false);
    expect(isStatusShown({ pending: 'include' }, 'pending')).toBe(true);
  });

  it('treats a status as hidden when another status is the only include', () => {
    expect(isStatusShown({ failed: 'include' }, 'pending')).toBe(false);
  });

  it('round-trips a status through exclude and back to neutral', () => {
    const once = toggleStatusFilter({}, 'failed' as TodoStatus);
    expect(once).toEqual({ failed: 'exclude' });
    expect(toggleStatusFilter(once, 'failed' as TodoStatus)).toEqual({});
  });
});

describe('loadTodoFilters', () => {
  beforeEach(() => localStorage.clear());

  it('defaults to hiding closed todos', () => {
    expect(loadTodoFilters()).toEqual({
      statuses: { completed: 'exclude' }, priorities: {}, external: {}, tags: {}, workspaces: {},
    });
  });

  it('migrates the v1 hidden-status array into excludes and drops the old key', () => {
    localStorage.setItem(LEGACY_HIDDEN_STATUS_KEY, JSON.stringify(['completed', 'skipped']));
    expect(loadTodoFilters()).toEqual({
      statuses: { completed: 'exclude', skipped: 'exclude' },
      priorities: {},
      external: {},
      tags: {},
      workspaces: {},
    });
    expect(localStorage.getItem(LEGACY_HIDDEN_STATUS_KEY)).toBeNull();
  });

  it('prefers stored v3 filters over the legacy keys', () => {
    localStorage.setItem(LEGACY_HIDDEN_STATUS_KEY, JSON.stringify(['completed']));
    localStorage.setItem(TODO_FILTERS_KEY, JSON.stringify({
      statuses: {}, priorities: { high: 'include' }, external: {}, tags: { bug: 'include' },
      workspaces: { '/repos/gavel': 'include' },
    }));
    expect(loadTodoFilters()).toEqual({
      statuses: {}, priorities: { high: 'include' }, external: {}, tags: { bug: 'include' },
      workspaces: { '/repos/gavel': 'include' },
    });
  });

  // v2 predates the tag facet. Widening it must keep every selection the user
  // already had, not reset them to the defaults.
  it('widens a stored v2 payload into v3 and drops the v2 key', () => {
    localStorage.setItem(LEGACY_TODO_FILTERS_V2_KEY, JSON.stringify({
      statuses: { failed: 'include' }, priorities: { high: 'exclude' }, external: { linked: 'include' },
    }));
    expect(loadTodoFilters()).toEqual({
      statuses: { failed: 'include' },
      priorities: { high: 'exclude' },
      external: { linked: 'include' },
      tags: {},
      workspaces: {},
    });
    expect(localStorage.getItem(LEGACY_TODO_FILTERS_V2_KEY)).toBeNull();
  });

  it('prefers a v3 payload over a still-present v2 one', () => {
    localStorage.setItem(LEGACY_TODO_FILTERS_V2_KEY, JSON.stringify({
      statuses: { failed: 'include' }, priorities: {}, external: {},
    }));
    localStorage.setItem(TODO_FILTERS_KEY, JSON.stringify({
      statuses: { draft: 'include' }, priorities: {}, external: {}, tags: {},
    }));
    expect(loadTodoFilters().statuses).toEqual({ draft: 'include' });
  });

  // A payload written before either facet existed has no tags/workspaces key; it
  // must widen rather than be rejected as malformed, keeping the selections it
  // does carry.
  it('tolerates a stored payload with no tags or workspaces key', () => {
    localStorage.setItem(TODO_FILTERS_KEY, JSON.stringify({
      statuses: { draft: 'include' }, priorities: {}, external: {},
    }));
    expect(loadTodoFilters()).toEqual({
      statuses: { draft: 'include' }, priorities: {}, external: {}, tags: {}, workspaces: {},
    });
  });

  it('falls back to the defaults when the stored value is malformed', () => {
    localStorage.setItem(TODO_FILTERS_KEY, '{"statuses":"nope"}');
    expect(loadTodoFilters()).toEqual(defaultTodoFilters());
  });
});
