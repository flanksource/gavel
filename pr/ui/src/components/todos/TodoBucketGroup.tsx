import { useState } from 'react';
import type { TodoDensity, TodoStatus } from '../../types';
import { GavelIcon } from '../GavelIcon';
import { countsFromItems, TodoCountsBar, TodoRow } from './format';
import { defaultHiddenStatuses, isTodoVisible } from './todoFilter';
import type { ResolvedRange } from './todoTimeRange';
import type { SelectedTodo } from './useWorkspaceTodos';
import type { TodoBucket, TodoEntry } from './todoGroup';

// TodoBucketGroup is one collapsible severity/age section of the flattened todo
// list. Unlike WorkspaceTodoGroup it spans workspaces, so each row names its
// owning workspace and there is no batch-run control (a run targets a single
// workspace dir/provider). The Closed/Status filter hides matching rows but
// leaves the header counts whole, mirroring the workspace grouping.
export function TodoBucketGroup({ bucket, selected, onSelect, hiddenStatuses, range, density = 'comfortable' }: {
  bucket: TodoBucket;
  selected: SelectedTodo | null;
  onSelect: (entry: TodoEntry) => void;
  hiddenStatuses?: Set<TodoStatus>;
  range?: ResolvedRange | null;
  density?: TodoDensity;
}) {
  const [open, setOpen] = useState(true);
  const hidden = hiddenStatuses ?? defaultHiddenStatuses();
  const visible = bucket.entries.filter(e => isTodoVisible(e.todo, hidden, range));
  const hiddenCount = bucket.entries.length - visible.length;
  const counts = countsFromItems(bucket.entries.map(e => e.todo));

  return (
    <div className="border-b border-border">
      <div className="sticky top-0 z-10 flex w-full items-center gap-2 bg-background/95 px-3 py-1.5 backdrop-blur">
        <button
          type="button"
          onClick={() => setOpen(o => !o)}
          className="flex min-w-0 flex-1 items-center gap-2 text-left hover:opacity-80"
        >
          <GavelIcon name={open ? 'codicon:chevron-down' : 'codicon:chevron-right'} className="text-muted-foreground text-xs" />
          <GavelIcon name={bucket.icon} className={`text-xs ${bucket.tone}`} />
          <span className={`min-w-0 flex-1 truncate text-sm font-semibold ${bucket.tone}`}>{bucket.label}</span>
        </button>
        <TodoCountsBar counts={counts} />
      </div>
      {open && (visible.length > 0 ? (
        visible.map(entry => (
          <TodoRow
            key={`${entry.workspace.dir}\t${entry.todo.ref}`}
            todo={entry.todo}
            active={selected?.dir === entry.workspace.dir && selected?.ref === entry.todo.ref}
            onClick={() => onSelect(entry)}
            density={density}
            workspace={entry.workspace.name}
            dir={entry.workspace.dir}
            provider={entry.workspace.todoProvider || 'auto'}
          />
        ))
      ) : (
        <div className="px-3 py-2 text-xs text-muted-foreground">
          {hiddenCount > 0 ? `${hiddenCount} todo${hiddenCount === 1 ? '' : 's'} hidden by filter` : 'No todos'}
        </div>
      ))}
    </div>
  );
}
