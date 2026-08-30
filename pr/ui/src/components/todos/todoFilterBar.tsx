import { useMemo } from 'react';
import type { FilterBarFilter, FilterBarRangeProps, MultiSelectOption } from '@flanksource/clicky-ui/components';
import type { FacetModes } from '../../utils';
import { flattenTodos } from './todoGroup';
import { EXTERNAL_FILTER_DEFS, PRIORITY_FILTER_DEFS, STATUS_FILTER_DEFS, externalKey, priorityKey } from './todoFilter';
import { todoTagTokens } from './tagResolve';
import type { WorkspaceTodos } from './useWorkspaceTodos';

// useTodoFilterBar builds the todos' facets once for both layouts: the sidebar
// toolbar and the full-width table's DataTable filter bar. The facets and the
// activity range come back separately because the two hosts place the range
// differently: the narrow sidebar folds it in among the facets so it can
// collapse, while the full-width bar has room for the trailing range slot.
export const TODO_ACTIVITY_FILTER_LABEL = 'Active';

export function useTodoFilterBar(todos: WorkspaceTodos): { facets: FilterBarFilter[]; range: FilterBarRangeProps } {
  const { workspaces, byDir, aggregate, filters, setFilters, timeRange, setTimeRange } = todos;

  // Priority, external-link and per-workspace counts are derived from the loaded
  // items — the server's TodoCounts only aggregates status, and the items are
  // already here.
  const derived = useMemo(() => {
    const priorities: Record<string, number> = {};
    const external: Record<string, number> = {};
    const tags: Record<string, number> = {};
    const dirs: Record<string, number> = {};
    for (const { todo, workspace } of flattenTodos(workspaces, byDir)) {
      priorities[priorityKey(todo)] = (priorities[priorityKey(todo)] ?? 0) + 1;
      external[externalKey(todo)] = (external[externalKey(todo)] ?? 0) + 1;
      dirs[workspace.dir] = (dirs[workspace.dir] ?? 0) + 1;
      for (const token of todoTagTokens(todo)) {
        tags[token] = (tags[token] ?? 0) + 1;
      }
    }
    return { priorities, external, tags, dirs };
  }, [workspaces, byDir]);

  const setFacet = (key: 'statuses' | 'priorities' | 'external' | 'tags' | 'workspaces', value: FacetModes) =>
    setFilters({ ...filters, [key]: value });

  const statusOptions: MultiSelectOption[] = STATUS_FILTER_DEFS
    .filter(def => aggregate[def.countKey] > 0)
    .map(def => ({ value: def.status, label: `${def.label} (${aggregate[def.countKey]})` }));
  const priorityOptions: MultiSelectOption[] = PRIORITY_FILTER_DEFS
    .filter(def => (derived.priorities[def.priority] ?? 0) > 0)
    .map(def => ({ value: def.priority, label: `${def.label} (${derived.priorities[def.priority]})` }));
  const externalOptions: MultiSelectOption[] = EXTERNAL_FILTER_DEFS
    .map(def => ({ value: def.key, label: `${def.label} (${derived.external[def.key] ?? 0})` }));
  // Tags in use, most-used first then alphabetical, so a long taxonomy still
  // opens on the ones that actually slice this backlog. Options are plain
  // strings: clicky's FilterBar drops per-option icons in its tri-state branch,
  // so a glyph here would not render.
  const tagOptions: MultiSelectOption[] = Object.entries(derived.tags)
    .sort(([aName, aCount], [bName, bCount]) => bCount - aCount || aName.localeCompare(bName))
    .map(([token, count]) => ({ value: token, label: `${token} (${count})` }));
  // Keyed by directory, labelled by the workspace's display name — two checkouts
  // can share a name, and it is the dir that identifies which one a row came
  // from. Configured order is kept, matching the sidebar's workspace sections.
  const selectedDirs = filters.workspaces ?? {};
  const workspaceOptions: MultiSelectOption[] = workspaces
    .filter(ws => (derived.dirs[ws.dir] ?? 0) > 0 || selectedDirs[ws.dir])
    .map(ws => ({ value: ws.dir, label: `${ws.name || ws.dir} (${derived.dirs[ws.dir] ?? 0})` }));
  // A selection can outlive the workspace it names — dropped from projects.json,
  // or renamed on disk. Without an option of its own an "only that workspace"
  // include empties the list with no control left to undo it, so the orphan is
  // offered under its bare directory until the user clears it.
  for (const dir of Object.keys(selectedDirs)) {
    if (!workspaces.some(ws => ws.dir === dir)) workspaceOptions.push({ value: dir, label: `${dir} (0)` });
  }

  const facets: FilterBarFilter[] = [];
  // Workspace leads: it is the broadest cut, and with one workspace configured
  // there is nothing to slice, so the control stays out of the way entirely.
  if (workspaceOptions.length > 1 || Object.keys(filters.workspaces ?? {}).length > 0) {
    facets.push({
      key: 'workspace', kind: 'multi', label: 'Workspace', options: workspaceOptions,
      value: filters.workspaces ?? {}, onChange: value => setFacet('workspaces', value),
    });
  }
  if (statusOptions.length) {
    facets.push({
      key: 'status', kind: 'multi', label: 'Status', options: statusOptions,
      value: filters.statuses, onChange: value => setFacet('statuses', value),
    });
  }
  if (priorityOptions.length) {
    facets.push({
      key: 'priority', kind: 'multi', label: 'Priority', options: priorityOptions,
      value: filters.priorities, onChange: value => setFacet('priorities', value),
    });
  }
  // A workspace that has never pushed a todo to GitHub gets no external facet —
  // "Not linked (all of them)" is chrome with nothing behind it. It appears as
  // soon as one todo is linked, or while a stale exclusion is still applied.
  if ((derived.external.linked ?? 0) > 0 || Object.keys(filters.external).length > 0) {
    facets.push({
      key: 'external', kind: 'multi', label: 'Issue',
      options: externalOptions, value: filters.external, onChange: value => setFacet('external', value),
    });
  }

  // A workspace with no labels gets no label facet, the same way the external facet
  // is gated — but a stale exclusion still surfaces its own control so it can be
  // cleared rather than silently narrowing the list forever.
  if (tagOptions.length > 0 || Object.keys(filters.tags ?? {}).length > 0) {
    facets.push({
      key: 'tags', kind: 'multi', label: 'Labels', options: tagOptions,
      value: filters.tags ?? {}, onChange: value => setFacet('tags', value),
    });
  }

  // No key/kind/label here: FilterBarRangeProps is the bar's trailing-range
  // shape. A host that wants it inline among the facets spreads it into a
  // FilterBarDateRangeFilter instead (see TodoToolbar).
  const range: FilterBarRangeProps = {
    from: timeRange?.from,
    to: timeRange?.to,
    // An applied range with both bounds empty is the user clearing it.
    onApply: (from, to) => setTimeRange(from || to ? { from, to } : null),
    presets: ['hr', 'day', 'wk+'],
    emptyLabel: 'Any time',
  };

  return { facets, range };
}
