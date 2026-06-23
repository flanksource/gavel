import { useEffect, useRef, useState } from 'react';
import type { Project, TodoListResponse, TodoRunOptions, TodoStatus } from '../../types';
import { GavelIcon } from '../GavelIcon';
import { emptyCounts, TodoCountsBar, TodoRow } from './format';
import { TodoRunAdvancedDialog, TodoRunSplitButton, useTodoRun } from './run';
import { defaultHiddenStatuses, isTodoVisible } from './todoFilter';

// WorkspaceTodoGroup is one collapsible workspace section, mirroring the PR
// tab's per-repo grouping: a sticky header with the workspace name and its
// open/failed/total counts, with the workspace's todos listed beneath. The
// Closed/Status filter hides matching rows but leaves the header counts whole.
//
// With `multiSelect`, each row grows a checkbox and the header swaps its counts
// for a "Run N" control once any are checked, dispatching the whole selection to
// one agent session via /api/todos/run. Selection is per-workspace because a run
// targets a single workspace dir/provider. The menubar omits multiSelect.
export function WorkspaceTodoGroup({ workspace, data, selectedRef, onSelect, hiddenStatuses, multiSelect = false, onRunStarted }: {
  workspace: Project;
  data?: TodoListResponse;
  selectedRef: string;
  onSelect: (ref: string) => void;
  hiddenStatuses?: Set<TodoStatus>;
  multiSelect?: boolean;
  onRunStarted?: () => void;
}) {
  const [open, setOpen] = useState(true);
  const [checked, setChecked] = useState<Set<string>>(new Set());
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const { runBusy, runMessage, runError, run } = useTodoRun(workspace.dir, workspace.todoProvider || 'auto');

  const hidden = hiddenStatuses ?? defaultHiddenStatuses();
  const allItems = data?.items ?? [];
  const items = allItems.filter(item => isTodoVisible(item, hidden));
  const hiddenCount = allItems.length - items.length;
  const counts = data?.counts ?? workspace.todoCounts ?? emptyCounts;

  const checkedRefs = items.filter(item => checked.has(item.ref)).map(item => item.ref);
  const allChecked = items.length > 0 && checkedRefs.length === items.length;

  // Reflect partial selection as the header checkbox's indeterminate state, which
  // is settable only via the DOM node, not a React prop.
  const selectAllRef = useRef<HTMLInputElement>(null);
  useEffect(() => {
    if (selectAllRef.current) {
      selectAllRef.current.indeterminate = checkedRefs.length > 0 && !allChecked;
    }
  }, [checkedRefs.length, allChecked]);

  function toggle(ref: string) {
    setChecked(prev => {
      const next = new Set(prev);
      if (next.has(ref)) next.delete(ref);
      else next.add(ref);
      return next;
    });
  }

  function toggleAll() {
    setChecked(prev => {
      const next = new Set(prev);
      if (allChecked) items.forEach(item => next.delete(item.ref));
      else items.forEach(item => next.add(item.ref));
      return next;
    });
  }

  async function runChecked(options?: TodoRunOptions) {
    if (checkedRefs.length === 0) return;
    const result = await run(checkedRefs, options);
    // A real (non-dry-run) start consumes the selection and refreshes the list;
    // a dry run keeps it so the user can follow up with an actual run.
    if (result && result.status !== 'dry_run') {
      setChecked(new Set());
      onRunStarted?.();
    }
  }

  return (
    <div className="border-b border-border">
      <div className="sticky top-0 z-10 flex w-full items-center gap-2 bg-background/95 px-3 py-1.5 backdrop-blur">
        {multiSelect && items.length > 0 && (
          <input
            ref={selectAllRef}
            type="checkbox"
            checked={allChecked}
            onChange={toggleAll}
            aria-label={`Select all todos in ${workspace.name}`}
            title="Select all"
            className="h-3.5 w-3.5 shrink-0 cursor-pointer accent-primary"
          />
        )}
        <button
          type="button"
          onClick={() => setOpen(o => !o)}
          className="flex min-w-0 flex-1 items-center gap-2 text-left hover:opacity-80"
        >
          <GavelIcon name={open ? 'codicon:chevron-down' : 'codicon:chevron-right'} className="text-muted-foreground text-xs" />
          <GavelIcon name="codicon:folder" className="text-muted-foreground text-xs" />
          <span className="min-w-0 flex-1 truncate text-sm font-semibold text-foreground" title={workspace.dir}>{workspace.name}</span>
        </button>
        {multiSelect && checkedRefs.length > 0 ? (
          <div className="flex shrink-0 items-center gap-1.5">
            <span className="text-[11px] tabular-nums text-muted-foreground">{checkedRefs.length} selected</span>
            <TodoRunSplitButton
              label={`Run ${checkedRefs.length}`}
              title="Run selected todos in one agent session"
              loading={runBusy}
              disabled={runBusy}
              onRun={runChecked}
              onAdvanced={() => setAdvancedOpen(true)}
            />
            <button
              type="button"
              onClick={() => setChecked(new Set())}
              title="Clear selection"
              aria-label="Clear selection"
              className="inline-flex h-8 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <GavelIcon name="codicon:close" className="text-xs" />
            </button>
          </div>
        ) : (
          <TodoCountsBar counts={counts} />
        )}
      </div>
      {(runError || runMessage) && (
        <div className={`px-3 py-1 text-[11px] ${runError ? 'text-red-600' : 'text-emerald-600'}`}>{runError || runMessage}</div>
      )}
      {multiSelect && (
        <TodoRunAdvancedDialog
          open={advancedOpen}
          onClose={() => setAdvancedOpen(false)}
          onRun={options => {
            setAdvancedOpen(false);
            runChecked(options);
          }}
          loading={runBusy}
          title={`Run ${checkedRefs.length} todo${checkedRefs.length === 1 ? '' : 's'}`}
        />
      )}
      {open && (items.length > 0 ? (
        items.map(item => (
          <TodoRow
            key={item.ref}
            todo={item}
            active={item.ref === selectedRef}
            onClick={() => onSelect(item.ref)}
            selectable={multiSelect}
            selected={checked.has(item.ref)}
            onToggleSelect={() => toggle(item.ref)}
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
