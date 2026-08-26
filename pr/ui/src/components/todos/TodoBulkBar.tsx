import { useState, type ComponentType } from 'react';
import { Button, DropdownMenu } from '@flanksource/clicky-ui/components';
import type { IconProps } from '@flanksource/clicky-ui/icons';
import { UiCheck, UiChevronDown, UiClose } from '@flanksource/clicky-ui/icons';
import type { TodoPriority, TodoStatus } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { assignableStatuses, priorities, priorityBadgeClass, priorityIcon, statusClass, statusIcon, statusLabel } from './format';
import { useBulkUpdateTodosMutation, type BulkTodoResponse, type BulkTodoUpdate } from './todoMutations';
import type { TodoSelection } from './todoSelection';

// TodoBulkBar is the action row that replaces the filter bar while bulk-edit mode
// is on: it reports how many todos are checked and applies one status or priority
// to all of them in a single request. Title/body are deliberately absent — writing
// the same prose over every selected todo is not an edit anyone wants.
//
// The selection survives a successful apply so a user can re-status and then
// re-prioritize the same batch without re-checking every row.
// The bar is split in two so the mutation hook mounts only while bulk-edit mode
// is on: the toolbar renders this on every layout, and an idle filter row should
// not require a QueryClient to hold a mutation it will never fire.
export function TodoBulkBar({ selection, onApplied }: {
  selection: TodoSelection;
  onApplied?: () => void;
}) {
  if (!selection.bulkMode) return null;
  return <BulkActions selection={selection} onApplied={onApplied} />;
}

function BulkActions({ selection, onApplied }: {
  selection: TodoSelection;
  onApplied?: () => void;
}) {
  const { setBulkMode, targets, clearSelection } = selection;
  const bulkUpdate = useBulkUpdateTodosMutation();
  const [result, setResult] = useState<BulkTodoResponse | null>(null);

  const count = targets.length;
  const apply = (update: Omit<BulkTodoUpdate, 'items'>) => {
    setResult(null);
    bulkUpdate.mutate({ items: targets, ...update }, {
      onSuccess: response => {
        setResult(response);
        onApplied?.();
      },
    });
  };

  const disabled = count === 0 || bulkUpdate.isPending;

  return (
    <div className="shrink-0 border-b border-border bg-primary/5 px-2 py-1.5">
      <div className="flex flex-wrap items-center gap-1.5">
        <span className="mr-auto whitespace-nowrap px-1 text-xs font-medium text-foreground">
          {bulkUpdate.isPending && <Spinner className="mr-1 inline text-xs" />}
          {count} selected
        </span>

        <BulkMenu
          label="Status"
          menuLabel="Set status on selected todos"
          disabled={disabled}
          options={assignableStatuses.map(status => ({
            value: status,
            label: statusLabel(status),
            icon: statusIcon(status),
            badgeClass: statusClass(status),
          }))}
          onSelect={status => apply({ status: status as TodoStatus })}
        />
        <BulkMenu
          label="Severity"
          menuLabel="Set severity on selected todos"
          disabled={disabled}
          options={priorities.map(priority => ({
            value: priority,
            label: priority,
            icon: priorityIcon(priority),
            badgeClass: priorityBadgeClass(priority),
          }))}
          onSelect={priority => apply({ priority: priority as TodoPriority })}
        />

        <Button
          type="button"
          variant="ghost"
          onClick={clearSelection}
          disabled={count === 0 || bulkUpdate.isPending}
          className="h-8 shrink-0 whitespace-nowrap rounded-md border border-border px-2 text-xs text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
        >
          Clear
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={() => { setResult(null); setBulkMode(false); }}
          title="Exit bulk edit"
          aria-label="Exit bulk edit"
          className="h-8 w-8 shrink-0 text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <UiClose className="text-xs" />
        </Button>
      </div>

      {bulkUpdate.error && (
        <p role="alert" className="mt-1 px-1 text-xs text-destructive">{bulkUpdate.error.message}</p>
      )}
      {/* A partial batch answers 200: the failures are per-item, so they are
          reported here rather than thrown away as a success. */}
      {result && <BulkResult result={result} />}
    </div>
  );
}

function BulkResult({ result }: { result: BulkTodoResponse }) {
  const failures = result.results.filter(item => item.error);
  return (
    <p
      role="status"
      className={`mt-1 px-1 text-xs ${failures.length ? 'text-destructive' : 'text-muted-foreground'}`}
    >
      Updated {result.updated} todo{result.updated === 1 ? '' : 's'}
      {failures.length > 0 && `; ${failures.length} failed: ${failures.map(item => `${item.ref} (${item.error})`).join(', ')}`}
    </p>
  );
}

interface BulkMenuOption {
  value: string;
  label: string;
  icon: ComponentType<IconProps>;
  badgeClass: string;
}

// BulkMenu mirrors the detail pane's StatusMenu/PriorityMenu vocabulary — same
// glyphs, same badge colours — so a value reads identically whether it is being
// set on one todo or forty. It has no current value: a selection spans todos that
// disagree, so nothing is ticked and picking an option is always a write.
function BulkMenu({ label, menuLabel, options, disabled, onSelect }: {
  label: string;
  menuLabel: string;
  options: BulkMenuOption[];
  disabled?: boolean;
  onSelect: (value: string) => void;
}) {
  return (
    <DropdownMenu
      align="left"
      menuLabel={menuLabel}
      menuClassName="w-52"
      trigger={
        <Button
          variant="ghost"
          type="button"
          disabled={disabled}
          title={menuLabel}
          aria-label={menuLabel}
          className="inline-flex h-8 shrink-0 items-center gap-1.5 whitespace-nowrap rounded-md border border-border px-2 text-xs font-medium text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
        >
          {label}
          <UiChevronDown className="text-[10px] opacity-70" />
        </Button>
      }
    >
      {close => (
        <div className="p-1 text-xs">
          {options.map(option => {
            const OptionIcon = option.icon;
            return (
              <Button
                key={option.value}
                variant="ghost"
                type="button"
                disabled={disabled}
                onClick={() => { close(); onSelect(option.value); }}
                className="flex h-auto w-full items-center justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted"
              >
                <span className={`inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full border ${option.badgeClass}`}>
                  <OptionIcon className="text-xs" />
                </span>
                <span className="min-w-0 flex-1 capitalize text-foreground">{option.label}</span>
              </Button>
            );
          })}
        </div>
      )}
    </DropdownMenu>
  );
}

// TodoBulkToggle enters and leaves bulk-edit mode. It lives in the filter bar's
// leading cluster beside Group and Sort, so the list's row-shaping controls stay
// together.
export function TodoBulkToggle({ selection }: { selection: TodoSelection }) {
  const { bulkMode, setBulkMode } = selection;
  return (
    <Button
      variant="ghost"
      type="button"
      onClick={() => setBulkMode(!bulkMode)}
      aria-pressed={bulkMode}
      title={bulkMode ? 'Exit bulk edit' : 'Select todos to edit in bulk'}
      className={`inline-flex h-8 shrink-0 items-center gap-1.5 whitespace-nowrap rounded-md border border-border px-2 text-xs transition-colors ${
        bulkMode ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-muted hover:text-foreground'
      }`}
    >
      <UiCheck className="text-xs" />
      <span className="font-medium max-sm:hidden">Select</span>
    </Button>
  );
}
