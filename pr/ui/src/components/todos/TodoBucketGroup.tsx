import { useState } from 'react';
import { Button, ListMenuHeader, ListMenuSection } from '@flanksource/clicky-ui/components';
import type { TodoDensity } from '../../types';
import { UiChevronDown, UiChevronRight } from '@flanksource/clicky-ui/icons';
import { MetaDueDate, SeverityHigh, SeverityLow, SeverityMedium } from '../../icons/issues';
import { countsFromItems, TodoCountsBar, TodoRow } from './format';
import type { TagIndex } from './tagResolve';
import { defaultTodoFilters, isEntryVisible, type TodoFilters } from './todoFilter';
import { GroupSelectAll } from './TodoGroupSelectAll';
import type { TodoSelection } from './todoSelection';
import type { ResolvedRange } from './todoTimeRange';
import type { SelectedTodo } from './useWorkspaceTodos';
import type { TodoBucket, TodoEntry } from './todoGroup';

function BucketIcon({ bucket }: { bucket: TodoBucket }) {
  const Icon = bucket.key === 'high'
    ? SeverityHigh
    : bucket.key === 'medium'
      ? SeverityMedium
      : bucket.key === 'low'
        ? SeverityLow
        : MetaDueDate;
  return <Icon width={14} height={14} className="shrink-0" />;
}

// TodoBucketGroup is one collapsible severity/age section of the flattened todo
// list. Unlike WorkspaceTodoGroup it spans workspaces, so each row names its
// owning workspace and there is no batch-run control (a run targets a single
// workspace directory). The Closed/Status filter hides matching rows but
// leaves the header counts whole, mirroring the workspace grouping.
export function TodoBucketGroup({ bucket, selected, onSelect, filters, range, density = 'comfortable', selection, tagsByDir }: {
  bucket: TodoBucket;
  selected: SelectedTodo | null;
  onSelect: (entry: TodoEntry) => void;
  filters?: TodoFilters;
  range?: ResolvedRange | null;
  density?: TodoDensity;
  selection?: TodoSelection;
  tagsByDir?: Map<string, TagIndex>;
}) {
  const [open, setOpen] = useState(true);
  const active = filters ?? defaultTodoFilters();
  const visible = bucket.entries.filter(e => isEntryVisible(e, active, range));
  const hiddenCount = bucket.entries.length - visible.length;
  const counts = countsFromItems(bucket.entries.map(e => e.todo));
  const ChevronIcon = open ? UiChevronDown : UiChevronRight;

  // A bucket spans workspaces, so each target carries its own row's dir.
  const bulkTargets = visible.map(entry => ({ dir: entry.workspace.dir, ref: entry.todo.ref }));

  return (
    <ListMenuSection>
      <ListMenuHeader>
        {selection && (
          <GroupSelectAll label={bucket.label} targets={bulkTargets} selection={selection} />
        )}
        <Button
          variant="ghost"
          type="button"
          onClick={() => setOpen(o => !o)}
          className="flex h-auto min-w-0 flex-1 items-center justify-start gap-2 p-0 text-left hover:opacity-80"
        >
          <ChevronIcon className="text-muted-foreground text-xs" />
          <BucketIcon bucket={bucket} />
          <span className={`min-w-0 flex-1 truncate text-sm font-semibold ${bucket.tone}`}>{bucket.label}</span>
        </Button>
        <TodoCountsBar counts={counts} />
      </ListMenuHeader>
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
            selectable={Boolean(selection)}
            selected={selection?.isSelected({ dir: entry.workspace.dir, ref: entry.todo.ref })}
            onToggleSelect={() => selection?.toggleSelected({ dir: entry.workspace.dir, ref: entry.todo.ref })}
            tags={tagsByDir?.get(entry.workspace.dir)}
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
