import { useState } from 'react';
import type { Project, TodoListResponse, TodoStatus } from '../../types';
import { GavelIcon } from '../GavelIcon';
import { emptyCounts, TodoCountsBar, TodoRow } from './format';
import { defaultHiddenStatuses, isTodoVisible } from './todoFilter';

// WorkspaceTodoGroup is one collapsible workspace section, mirroring the PR
// tab's per-repo grouping: a sticky header with the workspace name and its
// open/failed/total counts, with the workspace's todos listed beneath. The
// Closed/Status filter hides matching rows but leaves the header counts whole.
export function WorkspaceTodoGroup({ workspace, data, selectedRef, onSelect, hiddenStatuses }: {
  workspace: Project;
  data?: TodoListResponse;
  selectedRef: string;
  onSelect: (ref: string) => void;
  hiddenStatuses?: Set<TodoStatus>;
}) {
  const [open, setOpen] = useState(true);
  const hidden = hiddenStatuses ?? defaultHiddenStatuses();
  const allItems = data?.items ?? [];
  const items = allItems.filter(item => isTodoVisible(item, hidden));
  const hiddenCount = allItems.length - items.length;
  const counts = data?.counts ?? workspace.todoCounts ?? emptyCounts;
  return (
    <div className="border-b border-border">
      <button
        type="button"
        onClick={() => setOpen(o => !o)}
        className="sticky top-0 z-10 flex w-full items-center gap-2 bg-background/95 px-3 py-1.5 backdrop-blur hover:bg-muted"
      >
        <GavelIcon name={open ? 'codicon:chevron-down' : 'codicon:chevron-right'} className="text-muted-foreground text-xs" />
        <GavelIcon name="codicon:folder" className="text-muted-foreground text-xs" />
        <span className="min-w-0 flex-1 truncate text-left text-sm font-semibold text-foreground" title={workspace.dir}>{workspace.name}</span>
        <TodoCountsBar counts={counts} />
      </button>
      {open && (items.length > 0 ? (
        items.map(item => (
          <TodoRow key={item.ref} todo={item} active={item.ref === selectedRef} onClick={() => onSelect(item.ref)} />
        ))
      ) : (
        <div className="px-3 py-2 text-xs text-muted-foreground">
          {hiddenCount > 0 ? `${hiddenCount} todo${hiddenCount === 1 ? '' : 's'} hidden by filter` : 'No todos'}
        </div>
      ))}
    </div>
  );
}
