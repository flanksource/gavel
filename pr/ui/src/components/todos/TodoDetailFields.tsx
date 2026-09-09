import { useEffect, useState, type ComponentType, type ReactNode } from 'react';
import { Button, DropdownMenu } from '@flanksource/clicky-ui/components';
import { UiCheck, UiChevronDown, UiChevronRight, UiEdit, UiLinkExternal, type IconProps } from '@flanksource/clicky-ui/icons';
import type { TodoExternalIssue, TodoPriority, TodoStatus } from '../../types';
import { RepoIcon } from '../RepoIcon';
import { priorities, priorityBadgeClass, priorityIcon, statusClass, statusIcon, statuses, statusLabel } from './format';

export function StatusMenu({
  value,
  disabled,
  compact = false,
  onSelect,
}: {
  value: TodoStatus;
  disabled?: boolean;
  compact?: boolean;
  onSelect: (status: TodoStatus) => void;
}) {
  const ValueIcon = statusIcon(value);
  return (
    <DropdownMenu
      align="left"
      menuLabel="Update todo status"
      menuClassName="w-56"
      trigger={
        <Button
          variant="ghost"
          type="button"
          disabled={disabled}
          className={compact
            ? 'inline-flex h-7 items-center gap-1.5 rounded-md border border-border bg-muted/20 px-2 text-[11px] font-medium text-muted-foreground hover:bg-muted disabled:opacity-50'
            : `inline-flex h-8 items-center gap-1.5 rounded-full border px-2.5 text-xs font-semibold capitalize ${statusClass(value)} disabled:opacity-50`}
          title="Update status"
          aria-label="Update todo status"
        >
          {compact && <span>Status</span>}
          <span className={compact ? `inline-flex h-4 w-4 items-center justify-center rounded-full border ${statusClass(value)}` : ''}>
            <ValueIcon className="text-xs" />
          </span>
          <span className={compact ? 'capitalize text-foreground' : ''}>{statusLabel(value)}</span>
          <UiChevronDown className="text-[11px] opacity-70" />
        </Button>
      }
    >
      {close => (
        <div className="p-1 text-xs">
          {statuses.map(status => {
            const StatusItemIcon = statusIcon(status);
            return (
            <Button
              key={status}
              variant="ghost"
              type="button"
              disabled={disabled}
              onClick={() => {
                close();
                if (status !== value) onSelect(status);
              }}
              className="flex h-auto w-full items-center justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted"
            >
              <span className={`inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full border ${statusClass(status)}`}>
                <StatusItemIcon className="text-xs" />
              </span>
              <span className="min-w-0 flex-1 capitalize text-foreground">{statusLabel(status)}</span>
              {status === value && <UiCheck className="text-xs text-primary" />}
            </Button>
            );
          })}
        </div>
      )}
    </DropdownMenu>
  );
}

export function PriorityMenu({
  value,
  disabled,
  compact = false,
  onSelect,
}: {
  value: TodoPriority;
  disabled?: boolean;
  compact?: boolean;
  onSelect: (priority: TodoPriority) => void;
}) {
  const ValueIcon = priorityIcon(value);
  return (
    <DropdownMenu
      align="left"
      menuLabel="Update todo priority"
      menuClassName="w-44"
      trigger={
        <Button
          variant="ghost"
          type="button"
          disabled={disabled}
          className={compact
            ? 'inline-flex h-7 items-center gap-1.5 rounded-md border border-border bg-muted/20 px-2 text-[11px] font-medium text-muted-foreground hover:bg-muted disabled:opacity-50'
            : `inline-flex h-8 items-center gap-1.5 rounded-full border px-2.5 text-xs font-semibold capitalize ${priorityBadgeClass(value)} disabled:opacity-50`}
          title="Update priority"
          aria-label="Update todo priority"
        >
          {compact && <span>Severity</span>}
          <span className={compact ? `inline-flex h-4 w-4 items-center justify-center rounded-full border ${priorityBadgeClass(value)}` : ''}>
            <ValueIcon className="text-xs" />
          </span>
          <span className={compact ? 'capitalize text-foreground' : ''}>{value}</span>
          <UiChevronDown className="text-[11px] opacity-70" />
        </Button>
      }
    >
      {close => (
        <div className="p-1 text-xs">
          {priorities.map(priority => {
            const PriorityItemIcon = priorityIcon(priority);
            return (
            <Button
              key={priority}
              variant="ghost"
              type="button"
              disabled={disabled}
              onClick={() => {
                close();
                if (priority !== value) onSelect(priority);
              }}
              className="flex h-auto w-full items-center justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted"
            >
              <span className={`inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full border ${priorityBadgeClass(priority)}`}>
                <PriorityItemIcon className="text-xs" />
              </span>
              <span className="min-w-0 flex-1 capitalize text-foreground">{priority}</span>
              {priority === value && <UiCheck className="text-xs text-primary" />}
            </Button>
            );
          })}
        </div>
      )}
    </DropdownMenu>
  );
}

// ExternalIssueLink names the tracker issue this todo was pushed to. It is what
// makes the list's Linked/Not-linked filter legible: the filter says a link
// exists, this says which issue.
export function ExternalIssueLink({ issue }: { issue: TodoExternalIssue }) {
  return (
    <a
      href={issue.url}
      target="_blank"
      rel="noreferrer"
      title={`Open ${issue.repo}#${issue.number} on GitHub`}
      className="inline-flex h-auto min-w-0 items-center gap-1.5 rounded border border-border bg-muted/20 px-2 py-1 text-[11px] text-muted-foreground hover:bg-muted hover:text-foreground"
    >
      <RepoIcon repo={issue.repo} size={12} />
      <span className="min-w-0 truncate">{issue.repo}#{issue.number}</span>
      <UiLinkExternal className="shrink-0 text-[10px]" />
    </a>
  );
}


export function EditPencil({ label, onClick, disabled }: { label: string; onClick: () => void; disabled?: boolean }) {
  return (
    <Button
      variant="ghost"
      size="icon"
      type="button"
      onClick={onClick}
      disabled={disabled}
      title={label}
      aria-label={label}
      className="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
    >
      <UiEdit className="text-xs" />
    </Button>
  );
}

export function TodoSection({
  title,
  icon: Icon,
  count,
  defaultOpen = false,
  resetKey,
  action,
  children,
}: {
  title: string;
  icon: ComponentType<IconProps>;
  count?: number;
  defaultOpen?: boolean;
  resetKey: string;
  // action renders a control (e.g. an edit pencil) to the right of the header,
  // outside the toggle button so it doesn't nest interactive elements.
  action?: ReactNode;
  children: ReactNode;
}) {
  const [open, setOpen] = useState(defaultOpen);

  useEffect(() => {
    setOpen(defaultOpen);
  }, [defaultOpen, resetKey]);

  return (
    <section className="overflow-hidden rounded-lg border border-border bg-card shadow-sm">
      <div className="flex w-full min-w-0 items-center gap-2 border-b border-border bg-muted/30 pr-2">
        <Button
          variant="ghost"
          type="button"
          onClick={() => setOpen(o => !o)}
          className="flex h-auto min-w-0 flex-1 items-center justify-start gap-2 rounded-none px-3 py-2.5 text-left hover:bg-muted/70"
          aria-expanded={open}
        >
          {open ? <UiChevronDown className="shrink-0 text-xs text-muted-foreground" /> : <UiChevronRight className="shrink-0 text-xs text-muted-foreground" />}
          <span className="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-md border border-border bg-background text-muted-foreground">
            <Icon className="text-xs" />
          </span>
          <span className="min-w-0 flex-1 truncate text-xs font-semibold uppercase tracking-wide text-muted-foreground">{title}</span>
          {typeof count === 'number' && (
            <span className="rounded-full border border-border bg-background px-1.5 py-0.5 text-[11px] tabular-nums text-muted-foreground">
              {count}
            </span>
          )}
        </Button>
        {action}
      </div>
      {open && <div className="px-3 py-3">{children}</div>}
    </section>
  );
}
