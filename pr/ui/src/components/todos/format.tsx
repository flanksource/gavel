import type { TodoCounts, TodoItem, TodoPriority, TodoStatus } from '../../types';
import { GavelIcon } from '../GavelIcon';

export const inputClass = 'w-full rounded-md border border-input bg-background px-2.5 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring';

export const statuses: TodoStatus[] = ['pending', 'in_progress', 'failed', 'completed', 'skipped'];
export const priorities: TodoPriority[] = ['high', 'medium', 'low'];

export const emptyCounts: TodoCounts = {
  total: 0,
  open: 0,
  pending: 0,
  inProgress: 0,
  failed: 0,
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
    pending: a.pending + b.pending,
    inProgress: a.inProgress + b.inProgress,
    failed: a.failed + b.failed,
    completed: a.completed + b.completed,
    skipped: a.skipped + b.skipped,
  };
}

export function statusLabel(status: TodoStatus | string) {
  return status.replace('_', ' ');
}

export function statusClass(status: TodoStatus | string) {
  switch (status) {
    case 'completed':
      return 'text-green-600 bg-green-500/10 border-green-500/20';
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
      <CountBadge icon="codicon:debug-start" value={counts.inProgress} label="In progress" className="text-blue-600" />
      <CountBadge icon="codicon:error" value={counts.failed} label="Failed" className="text-red-600" />
      <CountBadge icon="codicon:pass" value={counts.completed} label="Completed" className="text-green-600" />
      <span className="text-muted-foreground tabular-nums" title="Total todos">{counts.total}</span>
    </div>
  );
}

export function TodoRow({ todo, active, onClick }: { todo: TodoItem; active: boolean; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`w-full px-3 py-2 text-left border-b border-border hover:bg-muted transition-colors ${active ? 'bg-primary/10' : ''}`}
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
  );
}
