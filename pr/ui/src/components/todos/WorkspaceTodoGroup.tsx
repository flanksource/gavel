import { useState } from 'react';
import { Button, ListMenuHeader, ListMenuSection } from '@flanksource/clicky-ui/components';
import { UiChevronDown, UiChevronRight, UiFolder } from '@flanksource/clicky-ui/icons';
import type { Project, TodoDensity, TodoListResponse, TodoSortBy, TodoStatus } from '../../types';
import { RepoIcon } from '../RepoIcon';
import { emptyCounts, TodoCountsBar, TodoRow } from './format';
import { defaultSortBy, todoComparator } from './todoGroup';
import { defaultHiddenStatuses, isTodoVisible } from './todoFilter';
import type { ResolvedRange } from './todoTimeRange';

// WorkspaceTodoGroup is one collapsible workspace section, mirroring the PR
// tab's per-repo grouping: a sticky header with the workspace name and its
// open/failed/total counts, with the workspace's todos listed beneath. The
// Closed/Status filter hides matching rows but leaves the header counts whole.
export function WorkspaceTodoGroup({ workspace, data, selectedRef, onSelect, hiddenStatuses, onToggleStatus, range, density = 'comfortable', sortBy = defaultSortBy() }: {
  workspace: Project;
  data?: TodoListResponse;
  selectedRef: string;
  onSelect: (ref: string) => void;
  hiddenStatuses?: Set<TodoStatus>;
  onToggleStatus?: (status: TodoStatus) => void;
  range?: ResolvedRange | null;
  density?: TodoDensity;
  sortBy?: TodoSortBy;
}) {
  const [open, setOpen] = useState(true);

  const hidden = hiddenStatuses ?? defaultHiddenStatuses();
  const allItems = data?.items ?? [];
  // Order rows by the chosen sort preference (defaults to severity-then-recent).
  const items = allItems.filter(item => isTodoVisible(item, hidden, range)).sort(todoComparator(sortBy));
  const hiddenCount = allItems.length - items.length;
  const counts = data?.counts ?? workspace.todoCounts ?? emptyCounts;

  // Match the PR tab's per-repo header: when the workspace maps to a GitHub repo,
  // show that repo's icon and short name in place of a generic folder + dir name.
  const repo = workspace.repos?.[0];
  const repoShort = repo ? repo.split('/').pop() || repo : '';

  return (
    <ListMenuSection>
      <ListMenuHeader>
        <Button
          variant="ghost"
          type="button"
          onClick={() => setOpen(o => !o)}
          className="flex h-auto min-w-0 flex-1 items-center justify-start gap-2 p-0 text-left hover:opacity-80"
        >
          {open ? <UiChevronDown className="text-muted-foreground text-xs" /> : <UiChevronRight className="text-muted-foreground text-xs" />}
          {repo ? (
            <>
              <RepoIcon repo={repo} size={16} />
              <span className="min-w-0 flex-1 truncate text-sm font-medium text-foreground" title={workspace.dir}>{repoShort}</span>
            </>
          ) : (
            <>
              <UiFolder className="text-muted-foreground text-xs" />
              <span className="min-w-0 flex-1 truncate text-sm font-semibold text-foreground" title={workspace.dir}>{workspace.name}</span>
            </>
          )}
        </Button>
        <TodoCountsBar counts={counts} hidden={hidden} onToggle={onToggleStatus} />
      </ListMenuHeader>
      {open && (items.length > 0 ? (
        items.map(item => (
          <TodoRow
            key={item.ref}
            todo={item}
            active={item.ref === selectedRef}
            onClick={() => onSelect(item.ref)}
            density={density}
            dir={workspace.dir}
          />
        ))
      ) : (
        <div className="px-3 py-2 text-xs text-muted-foreground">
          {hiddenCount > 0 ? `${hiddenCount} todo${hiddenCount === 1 ? '' : 's'} hidden by filter` : 'No todos'}
        </div>
      ))}
    </ListMenuSection>
  );
}
