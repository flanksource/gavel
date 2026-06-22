import { useState } from 'react';
import type { Project, TodoListResponse } from '../../types';
import { GavelIcon } from '../GavelIcon';
import { emptyCounts, TodoCountsBar, TodoRow } from './format';

// WorkspaceTodoGroup is one collapsible workspace section, mirroring the PR
// tab's per-repo grouping: a sticky header with the workspace name and its
// open/failed/total counts, with the workspace's todos listed beneath.
export function WorkspaceTodoGroup({ workspace, data, selectedRef, onSelect }: {
  workspace: Project;
  data?: TodoListResponse;
  selectedRef: string;
  onSelect: (ref: string) => void;
}) {
  const [open, setOpen] = useState(true);
  const items = data?.items ?? [];
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
        <div className="px-3 py-2 text-xs text-muted-foreground">No todos</div>
      ))}
    </div>
  );
}
