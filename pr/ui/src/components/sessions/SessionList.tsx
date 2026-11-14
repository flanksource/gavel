import { useEffect, useMemo, useState } from 'react';
import { ContextMeter, providerIcon, providerIconColor } from '@flanksource/clicky-ui/chat';
import { Button, FilterBar } from '@flanksource/clicky-ui/components';
import { Badge, type BadgeStatus, type StaticIconComponent } from '@flanksource/clicky-ui/data';
import { cn } from '@flanksource/clicky-ui/utils';
import {
  UiCheck,
  UiChevronDown,
  UiChevronRight,
  UiCircleX,
  UiCollapseAll,
  UiExpandAll,
  UiFolderGit2,
  UiGitBranch,
  UiLoader,
  UiQuestion,
  UiSearch,
  UiShieldCheck,
} from '@flanksource/clicky-ui/icons';
import type { AgentSession, TodoRunEffort } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { timeAgo } from '../../utils';
import { todoRunEffortPresentation } from '../todos/TodoRunEffortBadge';
import {
  DEFAULT_SESSION_FILTERS,
  SESSION_SORTS,
  SESSION_STATE_LABELS,
  flattenFamilies,
  hasNarrowingFilters,
  latestActivity,
  parentSessionIds,
  sessionDirtyCount,
  sessionFailed,
  sessionState,
  visibleFamilies,
  type SessionFamily,
  type SessionFilters,
  type SessionSort,
  type SessionState,
} from './sessionView';

export function SessionActivityIcon({ session, className }: { session: AgentSession; className?: string }) {
  const base = cn('size-3.5 shrink-0', className);
  if (session.activityState === 'approval') return <UiShieldCheck className={cn(base, 'text-amber-500')} aria-label="Needs approval" />;
  if (session.activityState === 'ask') return <UiQuestion className={cn(base, 'text-amber-500')} aria-label="Asked a question" />;
  if (session.processActive) return <Spinner className={cn(base, 'text-blue-500')} aria-label="Running" />;
  if (sessionFailed(session)) return <UiCircleX className={cn(base, 'text-red-500')} aria-label="Failed" />;
  return (
    <span className="inline-flex size-3.5 shrink-0 items-center justify-center" aria-label="Finished">
      <span className="size-2 rounded-full bg-emerald-500" />
    </span>
  );
}

const STATE_BADGES: Record<SessionState, { status: BadgeStatus; icon: StaticIconComponent }> = {
  working: { status: 'info', icon: UiLoader },
  approval: { status: 'warning', icon: UiShieldCheck },
  ask: { status: 'warning', icon: UiQuestion },
  completed: { status: 'success', icon: UiCheck },
  failed: { status: 'error', icon: UiCircleX },
};

const STATE_OPTIONS = Object.entries(SESSION_STATE_LABELS).map(([value, label]) => ({ value, label }));
// Desktop columns: context · session / project, state, last activity, cost.
const COLUMNS = '@min-[760px]:grid-cols-[minmax(0,1fr)_100px_92px_64px]';

function SessionStateBadge({ session, iconOnly = false }: { session: AgentSession; iconOnly?: boolean }) {
  const state = sessionState(session);
  const { status, icon } = STATE_BADGES[state];
  const label = SESSION_STATE_LABELS[state];
  return (
    <span className="inline-flex" title={session.stateReason || label} {...(iconOnly ? { role: 'img', 'aria-label': label } : {})}>
      <Badge
        variant="status"
        status={status}
        size="xs"
        icon={icon}
        className={cn('whitespace-nowrap', state === 'working' && '[&_svg]:motion-safe:animate-spin')}
      >
        {iconOnly ? undefined : label}
      </Badge>
    </span>
  );
}

function sessionProvider(session: AgentSession): string {
  return session.provider || session.source;
}

/** Top-level sessions get the gauge, subagents the compact bar so they read as secondary. */
function SessionContext({ session, parent }: { session: AgentSession; parent?: AgentSession }) {
  const provider = sessionProvider(session);
  const Glyph = providerIcon(provider);
  return (
    // The meter's hover card holds copy buttons; keep their clicks from selecting the row.
    <span className={cn('inline-flex shrink-0', parent ? 'w-10' : 'size-8')} onClick={event => event.stopPropagation()}>
      {session.context ? (
        <ContextMeter
          mode={parent ? 'bar' : 'gauge'}
          className={parent ? 'w-full' : 'p-0'}
          usedPercent={100 - session.context.freePercent}
          usedTokens={session.context.usedTokens}
          windowTokens={session.context.windowTokens}
          sessionId={session.id}
          providerSessionId={session.providerSessionId}
          provider={provider}
          modelIcon={Glyph}
          modelIconClassName={providerIconColor(provider) ?? ''}
          model={session.model}
          effort={session.effort}
          tokens={session.tokens}
          cost={{ total: session.costUsd }}
        />
      ) : Glyph ? (
        <span className="inline-flex size-full items-center justify-center" title={`${session.model ?? provider} · context unknown`}>
          <Glyph className={cn('size-4', providerIconColor(provider))} />
        </span>
      ) : null}
    </span>
  );
}

/** Model and effort after the title; a subagent shows only the parts that differ from its parent. */
function ModelBadge({ session, parent }: { session: AgentSession; parent?: AgentSession }) {
  const model = session.model && (!parent || parent.model !== session.model) ? session.model : undefined;
  const effort = session.effort && (!parent || parent.effort !== session.effort) ? session.effort : undefined;
  if (!model && !effort) return null;
  const presentation = effort ? todoRunEffortPresentation(effort as TodoRunEffort) : undefined;
  const EffortIcon = presentation?.icon;
  return (
    <span
      className="inline-flex shrink-0 items-center gap-0.5 rounded-full bg-muted px-1.5 py-px text-[10px] font-medium text-muted-foreground"
      title={[session.model, effort && `effort ${effort}`].filter(Boolean).join(' · ')}
    >
      {model}
      {EffortIcon ? <EffortIcon className="size-3" aria-label={`Effort ${effort}`} /> : null}
    </span>
  );
}

function sameCheckout(a: AgentSession, b: AgentSession): boolean {
  return (a.worktree?.path ?? a.cwd) === (b.worktree?.path ?? b.cwd) && a.worktree?.branch === b.worktree?.branch;
}

/** Project and checkout; a subagent shows them only when they differ from its parent's.
 *  The dirty count belongs to the checkout it describes. */
function GitLocation({ session, parent }: { session: AgentSession; parent?: AgentSession }) {
  const showProject = !parent || parent.project !== session.project;
  if (!showProject && parent && sameCheckout(parent, session)) return null;
  const worktree = session.worktree && !session.worktree.primary ? session.worktree : undefined;
  const branch = session.worktree?.branch ?? session.cwd?.split('/').pop() ?? '';
  const CheckoutIcon = worktree ? UiFolderGit2 : UiGitBranch;
  const changes = session.git?.changes;
  const dirty = sessionDirtyCount(session);
  return (
    <span className="flex min-w-0 items-center gap-1 text-xs text-muted-foreground">
      {showProject && session.project ? (
        <>
          <span className="shrink-0">{session.project}</span>
          {branch ? <span aria-hidden="true">·</span> : null}
        </>
      ) : null}
      {branch ? (
        <span
          className="inline-flex min-w-0 items-center gap-1"
          title={worktree ? `Worktree ${worktree.path} on ${worktree.branch}` : session.worktree ? `Branch ${branch} in the main checkout` : session.cwd}
        >
          <CheckoutIcon className={cn('size-3 shrink-0', worktree && 'text-primary')} aria-label={worktree ? 'Worktree' : 'Branch'} />
          <span className="truncate">{branch}</span>
        </span>
      ) : null}
      {changes && dirty > 0 ? (
        <span className="shrink-0 text-[11px] tabular-nums" title={`${dirty} uncommitted file${dirty === 1 ? '' : 's'}`}>
          <span className="text-green-600">+{changes.adds}</span>{' '}
          <span className="text-red-600">−{changes.dels}</span>{' '}
          <span>· {dirty}f</span>
        </span>
      ) : null}
    </span>
  );
}

function cost(usd: number): string {
  return `$${usd.toFixed(2)}`;
}

interface SessionRowsProps {
  families: SessionFamily[];
  collapsed: Set<string>;
  onToggle: (id: string) => void;
  selectedId: string;
  onSelect: (id: string) => void;
  forceOpen: boolean;
  parent?: AgentSession;
}

function SessionRows({ families, collapsed, onToggle, selectedId, onSelect, forceOpen, parent }: SessionRowsProps) {
  return (
    <ol className={cn('m-0 list-none p-0', parent && 'ml-3 border-l border-border')}>
      {families.map(family => {
        const { session, children } = family;
        const hasChildren = children.length > 0;
        const open = forceOpen || !collapsed.has(session.id);
        const Chevron = open ? UiChevronDown : UiChevronRight;
        // A collapsed family reports its most recent activity; expanded, each row speaks for itself.
        const activity = open ? session.lastActivityAt : new Date(latestActivity(family)).toISOString();
        const activityTitle = hasChildren && !open ? 'Latest activity across this session and its subagents' : "This session's last activity";
        const parentProps = parent ? { parent } : {};
        return (
          <li key={session.id} data-session-id={session.id} className="m-0">
            <div className={cn(
              'group flex min-w-0 items-center gap-1 border-b border-border/50 py-2 pl-1 pr-2 hover:bg-muted/50',
              session.id === selectedId && 'bg-accent',
            )}>
              {hasChildren ? (
                <Button
                  variant="ghost"
                  size="icon"
                  className="size-6 shrink-0"
                  aria-label={`${open ? 'Collapse' : 'Expand'} ${session.title}`}
                  aria-expanded={open}
                  disabled={forceOpen}
                  onClick={() => onToggle(session.id)}
                >
                  <Chevron className="size-3.5" />
                </Button>
              ) : (
                <span className="w-6 shrink-0" />
              )}
              <button
                type="button"
                onClick={() => onSelect(session.id)}
                aria-current={session.id === selectedId ? 'true' : undefined}
                aria-label={`View ${session.title}`}
                className={cn('grid min-w-0 flex-1 gap-2 rounded-sm text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring @min-[760px]:items-center', COLUMNS)}
              >
                <span className="flex min-w-0 items-center gap-2">
                  <SessionContext session={session} {...parentProps} />
                  <span className="min-w-0 flex-1 space-y-1">
                    <span className="flex min-w-0 items-center gap-2">
                      <span className="min-w-0 truncate text-sm font-medium text-foreground" title={session.title}>{session.title}</span>
                      <ModelBadge session={session} {...parentProps} />
                      {hasChildren && !open ? (
                        <span
                          className="shrink-0 rounded-full bg-muted px-1.5 text-[10px] font-medium tabular-nums text-muted-foreground"
                          title={`${children.length} hidden ${children.length === 1 ? 'subagent' : 'subagents'}`}
                        >
                          {children.length}
                        </span>
                      ) : null}
                      <span className="ml-auto shrink-0 whitespace-nowrap text-xs tabular-nums text-muted-foreground @min-[760px]:hidden" title={activityTitle}>
                        {timeAgo(activity)} · {cost(session.costUsd)}
                      </span>
                    </span>
                    <span className="flex min-w-0 items-center gap-2">
                      <span className="inline-flex shrink-0 @min-[760px]:hidden">
                        <SessionStateBadge session={session} iconOnly />
                      </span>
                      <GitLocation session={session} {...parentProps} />
                    </span>
                  </span>
                </span>
                <span className="hidden @min-[760px]:block"><SessionStateBadge session={session} /></span>
                <span className="hidden text-right text-xs tabular-nums text-muted-foreground @min-[760px]:block" title={activityTitle}>
                  {timeAgo(activity)}
                </span>
                <span className="hidden pr-2 text-right text-xs tabular-nums text-muted-foreground @min-[760px]:block">
                  {cost(session.costUsd)}
                </span>
              </button>
            </div>
            {hasChildren && open ? (
              <SessionRows
                families={children}
                collapsed={collapsed}
                onToggle={onToggle}
                selectedId={selectedId}
                onSelect={onSelect}
                forceOpen={forceOpen}
                parent={session}
              />
            ) : null}
          </li>
        );
      })}
    </ol>
  );
}

const FILTER_PARAM = 'sessions.';

function isSort(value: string): value is SessionSort {
  return SESSION_SORTS.some(option => option.value === value);
}

function isState(value: string): value is SessionState {
  return value in SESSION_STATE_LABELS;
}

/** Filters live in the URL (`sessions.search`, …) so a reload or shared link keeps them. */
function initialFilters(): SessionFilters {
  const params = new URLSearchParams(window.location.search);
  const sort = params.get(`${FILTER_PARAM}sort`) || DEFAULT_SESSION_FILTERS.sort;
  const state = params.get(`${FILTER_PARAM}state`) ?? '';
  if (!isSort(sort)) throw new Error(`Invalid session sort ${JSON.stringify(sort)}; expected one of ${SESSION_SORTS.map(o => o.value).join(', ')}`);
  if (state && !isState(state)) throw new Error(`Invalid session state ${JSON.stringify(state)}; expected one of ${Object.keys(SESSION_STATE_LABELS).join(', ')}`);
  return {
    search: params.get(`${FILTER_PARAM}search`) ?? '',
    project: params.get(`${FILTER_PARAM}project`) ?? '',
    state: state as SessionState | '',
    sort,
  };
}

function useUrlFilters() {
  const [filters, setFilters] = useState(initialFilters);
  useEffect(() => {
    const url = new URL(window.location.href);
    for (const [key, value] of Object.entries(filters)) {
      if (value && value !== DEFAULT_SESSION_FILTERS[key as keyof SessionFilters]) url.searchParams.set(FILTER_PARAM + key, value);
      else url.searchParams.delete(FILTER_PARAM + key);
    }
    if (url.href !== window.location.href) window.history.replaceState(window.history.state, '', url);
  }, [filters]);
  return [filters, setFilters] as const;
}

export interface SessionListProps {
  families: SessionFamily[];
  selectedId: string;
  onSelect: (id: string) => void;
}

/**
 * The Sessions tab list: session families — each session with the subagents
 * it delegated to — most recently active first. Below 760px of container width
 * a row stacks into two lines; wider, state, last activity and cost become
 * columns.
 */
export function SessionList({ families, selectedId, onSelect }: SessionListProps) {
  const [filters, setFilters] = useUrlFilters();
  const [collapsed, setCollapsed] = useState<Set<string>>(() => new Set());
  const projects = useMemo(
    () => [...new Set(flattenFamilies(families).map(session => session.project).filter((name): name is string => !!name))].sort(),
    [families],
  );
  const rows = useMemo(() => visibleFamilies(families, filters), [families, filters]);
  const forceOpen = hasNarrowingFilters(filters);

  return (
    <div className="@container flex h-full min-h-0 flex-col">
      <FilterBar
        className="shrink-0 border-b border-border px-3"
        search={{
          value: filters.search,
          onChange: search => setFilters(current => ({ ...current, search })),
          placeholder: 'Title, branch, model…',
          ariaLabel: 'Search sessions',
        }}
        filters={[
          {
            key: 'project',
            kind: 'enum',
            label: 'Project',
            value: filters.project,
            options: projects.map(value => ({ value, label: value })),
            onChange: project => setFilters(current => ({ ...current, project })),
          },
          {
            key: 'state',
            kind: 'enum',
            label: 'State',
            value: filters.state,
            options: STATE_OPTIONS,
            onChange: state => setFilters(current => ({ ...current, state: state && isState(state) ? state : '' })),
          },
          {
            key: 'sort',
            kind: 'enum',
            label: 'Sort',
            value: filters.sort,
            options: [...SESSION_SORTS],
            // Clearing the sort returns to the default order.
            onChange: sort => setFilters(current => ({ ...current, sort: isSort(sort) ? sort : DEFAULT_SESSION_FILTERS.sort })),
          },
        ]}
        trailing={
          <div className="flex items-center">
            <Button variant="ghost" size="icon" className="size-9" aria-label="Expand all" title="Expand all" disabled={forceOpen} onClick={() => setCollapsed(new Set())}>
              <UiExpandAll />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="size-9"
              aria-label="Collapse all"
              title="Collapse all"
              disabled={forceOpen}
              onClick={() => setCollapsed(new Set(parentSessionIds(families)))}
            >
              <UiCollapseAll />
            </Button>
          </div>
        }
      />
      <div
        className={cn('hidden shrink-0 gap-2 border-b border-border bg-muted/40 py-2 pl-8 pr-2 text-xs text-muted-foreground @min-[760px]:grid', COLUMNS)}
        aria-hidden="true"
      >
        <span>Context · session / project</span>
        <span>State</span>
        <span className="text-right">Last activity</span>
        <span className="pr-2 text-right">Cost</span>
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto" aria-label="Session families">
        {rows.length ? (
          <SessionRows
            families={rows}
            collapsed={collapsed}
            onToggle={id => setCollapsed(current => {
              const next = new Set(current);
              if (next.has(id)) next.delete(id);
              else next.add(id);
              return next;
            })}
            selectedId={selectedId}
            onSelect={onSelect}
            forceOpen={forceOpen}
          />
        ) : (
          <div className="flex h-full flex-col items-center justify-center gap-3 px-6 text-center">
            <UiSearch className="size-6 text-muted-foreground" />
            <p className="text-sm font-medium">{families.length ? 'No sessions match' : 'No sessions yet'}</p>
            {families.length ? (
              <>
                <p className="text-xs text-muted-foreground">Try another project, state, or search.</p>
                <Button variant="outline" size="sm" onClick={() => setFilters(current => ({ ...current, search: '', project: '', state: '' }))}>
                  Clear filters
                </Button>
              </>
            ) : null}
          </div>
        )}
      </div>
    </div>
  );
}
