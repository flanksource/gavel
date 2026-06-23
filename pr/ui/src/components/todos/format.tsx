import type { TodoCounts, TodoItem, TodoPriority, TodoStatus } from '../../types';
import { GavelIcon } from '../GavelIcon';

export const inputClass = 'w-full rounded-md border border-input bg-background px-2.5 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring';

export const statuses: TodoStatus[] = ['draft', 'pending', 'in_progress', 'failed', 'verified', 'completed', 'skipped'];
export const priorities: TodoPriority[] = ['high', 'medium', 'low'];

export const emptyCounts: TodoCounts = {
  total: 0,
  open: 0,
  draft: 0,
  pending: 0,
  inProgress: 0,
  failed: 0,
  verified: 0,
  completed: 0,
  skipped: 0,
};

// Listing/detail/mutation requests carry the workspace dir and its provider.
// Provider defaults to 'auto' (server resolves Grite-or-.todos); a workspace
// pinned to 'grite'/'todos' passes that instead.
export function todoQuery(dir: string, provider: string = 'auto') {
  const params = new URLSearchParams({ provider: provider || 'auto' });
  if (dir.trim()) params.set('dir', dir.trim());
  return params.toString();
}

export function addCounts(a: TodoCounts, b: TodoCounts): TodoCounts {
  return {
    total: a.total + b.total,
    open: a.open + b.open,
    draft: a.draft + b.draft,
    pending: a.pending + b.pending,
    inProgress: a.inProgress + b.inProgress,
    failed: a.failed + b.failed,
    verified: a.verified + b.verified,
    completed: a.completed + b.completed,
    skipped: a.skipped + b.skipped,
  };
}

export function statusLabel(status: TodoStatus | string) {
  return status.replace('_', ' ');
}

export function statusClass(status: TodoStatus | string) {
  switch (status) {
    case 'draft':
      return 'text-muted-foreground bg-muted border-border';
    case 'completed':
      return 'text-green-600 bg-green-500/10 border-green-500/20';
    case 'verified':
      return 'text-emerald-700 dark:text-emerald-400 bg-emerald-500/10 border-emerald-500/20';
    case 'in_progress':
      return 'text-blue-600 bg-blue-500/10 border-blue-500/20';
    case 'failed':
      return 'text-red-600 bg-red-500/10 border-red-500/20';
    case 'skipped':
      return 'text-yellow-700 dark:text-yellow-400 bg-yellow-500/10 border-yellow-500/20';
    default:
      return 'text-muted-foreground bg-muted border-border';
  }
}

export function priorityClass(priority: TodoPriority | string) {
  switch (priority) {
    case 'high':
      return 'text-red-600';
    case 'low':
      return 'text-green-600';
    default:
      return 'text-yellow-600';
  }
}

function CountBadge({ icon, value, label, className = 'text-muted-foreground' }: { icon: string; value: number; label: string; className?: string }) {
  if (!value) return null;
  return (
    <span className={`inline-flex items-center gap-1 text-xs tabular-nums ${className}`} title={label}>
      <GavelIcon name={icon} className="text-[12px]" />
      {value}
    </span>
  );
}

export function TodoCountsBar({ counts }: { counts: TodoCounts }) {
  return (
    <div className="flex items-center gap-3 text-xs">
      <CountBadge icon="codicon:check" value={counts.open} label="Open todos" className="text-blue-600" />
      <CountBadge icon="codicon:clock" value={counts.draft} label="Draft" />
      <CountBadge icon="codicon:debug-start" value={counts.inProgress} label="In progress" className="text-blue-600" />
      <CountBadge icon="codicon:error" value={counts.failed} label="Failed" className="text-red-600" />
      <CountBadge icon="octicon:check-circle-fill-16" value={counts.verified} label="Verified" className="text-emerald-600" />
      <CountBadge icon="codicon:pass" value={counts.completed} label="Completed" className="text-green-600" />
      <span className="text-muted-foreground tabular-nums" title="Total todos">{counts.total}</span>
    </div>
  );
}

// TodoRow renders one todo in a workspace list. When `selectable` is set it grows
// a leading checkbox for multi-select (run several todos in one agent session);
// the checkbox is a sibling of the open-detail button so toggling selection never
// opens the todo.
export function TodoRow({ todo, active, onClick, selectable = false, selected = false, onToggleSelect }: {
  todo: TodoItem;
  active: boolean;
  onClick: () => void;
  selectable?: boolean;
  selected?: boolean;
  onToggleSelect?: () => void;
}) {
  return (
    <div className={`flex items-stretch border-b border-border ${selected ? 'bg-primary/5' : ''}`}>
      {selectable && (
        <label className="flex shrink-0 cursor-pointer items-center pl-3" title="Select for batch run">
          <input
            type="checkbox"
            checked={selected}
            onChange={onToggleSelect}
            aria-label={`Select ${todo.title}`}
            className="h-3.5 w-3.5 cursor-pointer accent-primary"
          />
        </label>
      )}
      <button
        type="button"
        onClick={onClick}
        className={`min-w-0 flex-1 px-3 py-2 text-left transition-colors hover:bg-muted ${active ? 'bg-primary/10' : ''}`}
      >
        <div className="flex min-w-0 items-center gap-2">
          <span className={`inline-flex shrink-0 rounded border px-1.5 py-0.5 text-[10px] font-medium uppercase ${statusClass(todo.status)}`}>
            {statusLabel(todo.status)}
          </span>
          <span className="min-w-0 flex-1 truncate text-sm font-medium text-foreground">{todo.title}</span>
        </div>
        <div className="mt-1 flex items-center gap-2 text-[11px] text-muted-foreground">
          {todo.shortId && <span className="font-mono">{todo.shortId}</span>}
          <span className={priorityClass(todo.priority)}>{todo.priority}</span>
          {todo.provider && <span className="uppercase">{todo.provider}</span>}
        </div>
      </button>
    </div>
  );
}
