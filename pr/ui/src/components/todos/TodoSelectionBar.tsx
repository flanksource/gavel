import { SelectionActionBar } from '@flanksource/clicky-ui/data';
import { useTodoBulkContext, useTodoBulkToolbar } from './todoActions';
import type { WorkspaceTodos } from './useWorkspaceTodos';

/**
 * The bulk-action cluster for the split (sidebar) layout.
 *
 * The full-width table gets the identical cluster from DataTable, which renders
 * `selectionActions` itself. This exists because the sidebar is a hand-rolled
 * list rather than a table — but it renders the *same* descriptors through the
 * *same* exported component, so the two layouts cannot offer different actions.
 *
 * It appears with the first checked row and disappears with the last: there is
 * no mode to enter, and an empty bar advertises nothing.
 */
export function TodoSelectionBar({ todos, onApplied }: {
  todos: WorkspaceTodos;
  onApplied?: () => void;
}) {
  // Split in two so the catalog fetch and the toast hook mount only once
  // something is checked: the toolbar renders this on every layout, and an idle
  // filter row should not require a ToastProvider or a request it will never use.
  if (todos.selection.selection.size === 0) return null;
  return <SelectionBar todos={todos} onApplied={onApplied} />;
}

function SelectionBar({ todos, onApplied }: {
  todos: WorkspaceTodos;
  onApplied?: () => void;
}) {
  const selection = todos.selection;
  const bulkContext = useTodoBulkContext(todos);
  // No early return on an empty catalog: the count and Clear are worth showing
  // the moment something is ticked, and the actions fill in behind them. A bar
  // that rendered nothing until the catalog loaded would leave the row it has
  // taken over blank.
  const actions = useTodoBulkToolbar({
    selection: selection.selection,
    ...bulkContext,
    onApplied,
  });

  return (
    <SelectionActionBar
      actions={actions}
      context={{
        selectedRowIds: [...selection.selection],
        // The sidebar renders its own rows, so the bar is given the ids it
        // acts on and nothing else; nothing in these descriptors reads rows.
        selectedRows: [],
        clearSelection: selection.clearSelection,
      }}
      // Without this the count is a block element above the buttons rather
      // than beside them: SelectionActionBar renders a fragment, so the row is
      // the caller's to declare.
      className="flex flex-wrap items-center gap-density-2"
    />
  );
}
