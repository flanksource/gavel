import { Button } from '@flanksource/clicky-ui/components';
import type { SessionStats, TodoCounts, TodoDensity, TodoDiffStat, TodoItem, TodoPriority, TodoStatus } from '../../types';
import { timeAgo } from '../../utils';
import { GavelIcon } from '../GavelIcon';
import { DENSITY_OPTIONS } from './todoDensity';
import { formatCost, formatDuration, useSessionStats } from './TodoSessionTimer';

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

// countsFromItems tallies a TodoCounts from a list of todos, mirroring the
// server's summarizeTodos so severity/age bucket headers read the same as the
// per-workspace counts the list endpoint returns.
export function countsFromItems(items: TodoItem[]): TodoCounts {
  const counts: TodoCounts = { ...emptyCounts };
  for (const item of items) {
    counts.total++;
    switch (item.status) {
      case 'completed': counts.completed++; break;
      case 'draft': counts.open++; counts.draft++; break;
      case 'in_progress': counts.open++; counts.inProgress++; break;
      case 'failed': counts.open++; counts.failed++; break;
      case 'verified': counts.open++; counts.verified++; break;
      case 'skipped': counts.open++; counts.skipped++; break;
      default: counts.open++; counts.pending++; break;
    }
  }
  return counts;
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

interface SessionBadgeView {
  className: string;
  icon: string;
  label: string;
}

// sessionBadgeView is the badge's chrome + icon + label, derived from whether the
// agent's run is still in progress (found && !inProgress) and its high-level
// state. Once the run ends the badge stops reading "In progress": a clean turn
// end settles to Done (green), anything else to Ended (neutral), so the row
// mirrors the real session state instead of a perpetual blue spinner. Before the
// log appears (found=false) the session is still starting, so it stays live.
function sessionBadgeView(stats: SessionStats | null): SessionBadgeView {
  if (stats && stats.found && !stats.inProgress) {
    if (stats.state === 'completed') {
      return { className: statusClass('completed'), icon: 'codicon:pass', label: 'Done' };
    }
    return { className: statusClass('draft'), icon: 'codicon:clock', label: 'Ended' };
  }
  switch (stats?.state) {
    case 'thinking':
      return { className: statusClass('in_progress'), icon: 'codicon:lightbulb', label: 'Thinking' };
    case 'working':
      return { className: statusClass('in_progress'), icon: 'svg-spinners:ring-resize', label: 'Working' };
    case 'ask':
      return { className: statusClass('skipped'), icon: 'codicon:comment-discussion', label: 'Awaiting input' };
    case 'completed':
      return { className: statusClass('completed'), icon: 'codicon:pass', label: 'Done' };
    default:
      return { className: statusClass('in_progress'), icon: 'svg-spinners:ring-resize', label: 'In progress' };
  }
}

// InProgressBadge replaces the static "in progress" status pill for a todo with a
// live agent session: the agent's current state drives the chrome, icon, and
// elapsed time (plus cost, once known), and a finished run settles to Done/Ended
// rather than spinning forever. Render it only for a row that has a session so it
// never polls idle todos.
function InProgressBadge({ dir, provider, sessionId }: { dir: string; provider: string; sessionId?: string }) {
  const { stats, elapsedMs } = useSessionStats(dir, provider, sessionId, true);
  const view = sessionBadgeView(stats);
  const cost = stats ? formatCost(stats.costUsd) : '';
  return (
    <span
      className={`inline-flex shrink-0 items-center gap-1 rounded border px-1.5 py-0.5 text-[10px] font-medium tabular-nums ${view.className}`}
      title={`${view.label} · agent session`}
    >
      <GavelIcon name={view.icon} className="text-[11px]" />
      {stats?.found ? formatDuration(elapsedMs) : view.label}
      {cost && <span className="opacity-80">{cost}</span>}
    </span>
  );
}

function CountBadge({ icon, value, label, className = 'text-muted-foreground', status, hidden, onToggle }: {
  icon: string;
  value: number;
  label: string;
  className?: string;
  status?: TodoStatus;
  hidden?: Set<TodoStatus>;
  onToggle?: (status: TodoStatus) => void;
}) {
  if (!value) return null;
  const content = (
    <>
      <GavelIcon name={icon} className="text-[12px]" />
      {value}
    </>
  );
  // A status-mapped badge in a wired counts bar doubles as a filter pill: clicking
  // it toggles that status into/out of the shared hidden set, dimming when the
  // status is hidden. Aggregate badges (open/total) and unwired bars stay static.
  if (status && onToggle) {
    const active = !hidden?.has(status);
    return (
      <Button
        variant="ghost"
        type="button"
        onClick={() => onToggle(status)}
        aria-pressed={active}
        title={active ? `Hide ${label.toLowerCase()}` : `Show ${label.toLowerCase()}`}
        className={`-mx-0.5 inline-flex h-auto items-center gap-1 rounded px-0.5 text-xs tabular-nums transition-colors hover:bg-muted ${active ? className : 'text-muted-foreground opacity-40 hover:opacity-100'}`}
      >
        {content}
      </Button>
    );
  }
  return (
    <span className={`inline-flex items-center gap-1 text-xs tabular-nums ${className}`} title={label}>
      {content}
    </span>
  );
}

// TodoCountsBar renders a workspace/bucket header's status summary. When `hidden`
// and `onToggle` are supplied, each status-mapped badge becomes a filter pill
// (toggling that status in the shared hidden set); without them the badges are
// static counts (e.g. the aggregate bar in the action header).
export function TodoCountsBar({ counts, hidden, onToggle }: {
  counts: TodoCounts;
  hidden?: Set<TodoStatus>;
  onToggle?: (status: TodoStatus) => void;
}) {
  return (
    <div className="flex items-center gap-3 text-xs">
      <CountBadge icon="codicon:check" value={counts.open} label="Open todos" className="text-blue-600" />
      <CountBadge icon="codicon:clock" value={counts.draft} label="Draft" status="draft" hidden={hidden} onToggle={onToggle} />
      <CountBadge icon="codicon:debug-start" value={counts.inProgress} label="In progress" className="text-blue-600" status="in_progress" hidden={hidden} onToggle={onToggle} />
      <CountBadge icon="codicon:error" value={counts.failed} label="Failed" className="text-red-600" status="failed" hidden={hidden} onToggle={onToggle} />
      <CountBadge icon="octicon:check-circle-fill-16" value={counts.verified} label="Verified" className="text-emerald-600" status="verified" hidden={hidden} onToggle={onToggle} />
      <CountBadge icon="codicon:pass" value={counts.completed} label="Completed" className="text-green-600" status="completed" hidden={hidden} onToggle={onToggle} />
      <span className="text-muted-foreground tabular-nums" title="Total todos">{counts.total}</span>
    </div>
  );
}

// TodoDiffBadge shows the aggregated change footprint of a todo's linked commits
// as `+adds`/`-dels`, with the commit/file totals in the tooltip. Rendered only
// when the todo has commits, so todos with no work attached stay uncluttered.
function TodoDiffBadge({ diff }: { diff: TodoDiffStat }) {
  return (
    <span
      className="inline-flex shrink-0 items-center gap-1 tabular-nums"
      title={`${diff.commits} commit${diff.commits === 1 ? '' : 's'}, ${diff.files} file${diff.files === 1 ? '' : 's'} changed`}
    >
      <GavelIcon name="codicon:git-commit" className="text-[11px]" />
      {diff.adds > 0 && <span className="text-green-600">+{diff.adds}</span>}
      {diff.dels > 0 && <span className="text-red-600">-{diff.dels}</span>}
      {diff.adds === 0 && diff.dels === 0 && <span>{diff.commits}</span>}
    </span>
  );
}

// TodoAges shows the todo's created age and, when it differs, its last-activity
// age — the two relative timestamps the list sorts by. Absolute times sit in the
// tooltips. A todo with neither timestamp (some file-backed todos) renders nothing.
function TodoAges({ todo }: { todo: TodoItem }) {
  const showLast = !!todo.lastRun && todo.lastRun !== todo.created;
  return (
    <>
      {todo.created && (
        <span className="inline-flex shrink-0 items-center gap-1" title={`Created ${new Date(todo.created).toLocaleString()}`}>
          <GavelIcon name="codicon:add" className="text-[11px]" />
          {timeAgo(todo.created)}
        </span>
      )}
      {showLast && (
        <span className="inline-flex shrink-0 items-center gap-1" title={`Last activity ${new Date(todo.lastRun!).toLocaleString()}`}>
          <GavelIcon name="codicon:history" className="text-[11px]" />
          {timeAgo(todo.lastRun!)}
        </span>
      )}
    </>
  );
}

// TodoRow renders one todo in a workspace list. When `selectable` is set it grows
// a leading checkbox for multi-select (run several todos in one agent session);
// the checkbox is a sibling of the open-detail button so toggling selection never
// opens the todo. `density` controls the layout: 'comfortable' (default) stacks
// the metadata on a second line; 'compact' folds id + priority inline with the
// title for a single-line row. `workspace` names the owning workspace in the
// metadata — set when rows mix workspaces (severity/age grouping) so each todo's
// origin stays visible.
//
// `dir`/`provider` locate the row's workspace so an in-progress todo's status
// badge can carry the live agent state + elapsed time; they are omitted by
// callers (e.g. the menubar) that don't surface it.
export function TodoRow({ todo, active, onClick, density = 'comfortable', selectable = false, selected = false, onToggleSelect, workspace, dir, provider }: {
  todo: TodoItem;
  active: boolean;
  onClick: () => void;
  density?: TodoDensity;
  selectable?: boolean;
  selected?: boolean;
  onToggleSelect?: () => void;
  workspace?: string;
  dir?: string;
  provider?: string;
}) {
  const compact = density === 'compact';
  // Only running sessions poll for stats, so the sidebar never fires a request
  // storm across a large list of idle/finished todos.
  const hasLiveSession = !!dir && todo.status === 'in_progress' && !!todo.sessionId;
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
      <Button
        variant="ghost"
        type="button"
        onClick={onClick}
        className={`h-auto min-w-0 flex-1 justify-start px-3 text-left transition-colors hover:bg-muted ${compact ? 'py-1' : 'py-2'} ${active ? 'bg-primary/10' : ''}`}
      >
        <div className="flex min-w-0 items-center gap-2">
          {hasLiveSession ? (
            <InProgressBadge dir={dir!} provider={provider || 'auto'} sessionId={todo.sessionId} />
          ) : (
            <span className={`inline-flex shrink-0 rounded border px-1.5 py-0.5 text-[10px] font-medium uppercase ${statusClass(todo.status)}`}>
              {statusLabel(todo.status)}
            </span>
          )}
          <span className="min-w-0 flex-1 truncate text-sm font-medium text-foreground">{todo.title}</span>
          {compact && (
            <span className="flex shrink-0 items-center gap-2 text-[11px] text-muted-foreground">
              {workspace && <span className="max-w-[8rem] truncate" title={workspace}>{workspace}</span>}
              {todo.shortId && <span className="font-mono">{todo.shortId}</span>}
              {todo.diff && <TodoDiffBadge diff={todo.diff} />}
              <span className={priorityClass(todo.priority)}>{todo.priority}</span>
            </span>
          )}
        </div>
        {!compact && (
          <div className="mt-1 flex items-center gap-2 text-[11px] text-muted-foreground">
            {workspace && (
              <span className="inline-flex min-w-0 items-center gap-1" title={workspace}>
                <GavelIcon name="codicon:folder" className="text-[11px]" />
                <span className="max-w-[10rem] truncate">{workspace}</span>
              </span>
            )}
            {todo.shortId && <span className="font-mono">{todo.shortId}</span>}
            <span className={priorityClass(todo.priority)}>{todo.priority}</span>
            <TodoAges todo={todo} />
            {todo.diff && <TodoDiffBadge diff={todo.diff} />}
            {todo.provider && <span className="uppercase">{todo.provider}</span>}
          </div>
        )}
      </Button>
    </div>
  );
}

// TodoDensityPicker is the segmented toggle that switches the todo lists between
// comfortable (two-line) and compact (single-line) rows. It lives in the Todos
// toolbar and drives the shared density preference.
export function TodoDensityPicker({ density, onChange }: {
  density: TodoDensity;
  onChange: (density: TodoDensity) => void;
}) {
  return (
    <div className="inline-flex items-center gap-0.5 rounded-md border border-border p-0.5" role="group" aria-label="Row density">
      {DENSITY_OPTIONS.map(opt => {
        const active = density === opt.value;
        return (
          <Button
            key={opt.value}
            variant="ghost"
            size="icon"
            type="button"
            onClick={() => onChange(opt.value)}
            aria-pressed={active}
            title={`${opt.label} rows`}
            aria-label={`${opt.label} rows`}
            className={`h-6 w-7 rounded transition-colors ${
              active ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-muted hover:text-foreground'
            }`}
          >
            <GavelIcon name={opt.icon} className="text-sm" />
          </Button>
        );
      })}
    </div>
  );
}
