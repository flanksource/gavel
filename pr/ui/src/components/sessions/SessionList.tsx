import { useEffect, useMemo, useState } from 'react';
import { ContextMeter, providerIcon, providerIconColor } from '@flanksource/clicky-ui/chat';
import { Button, DropdownMenu, FilterBar } from '@flanksource/clicky-ui/components';
import { Badge, type BadgeStatus, type StaticIconComponent } from '@flanksource/clicky-ui/data';
import { cn } from '@flanksource/clicky-ui/utils';
import {
  UiCheck,
  UiCircleX,
  UiFolderGit,
  UiFolderGit2,
  UiGitBranch,
  UiLoader,
  UiQuestion,
  UiSearch,
  UiShieldCheck,
} from '@flanksource/clicky-ui/icons';
import { Icon } from '@iconify/react/offline';
import sortDescendingThin from '@iconify-icons/ph/sort-descending-thin';
import type { AgentSession, TodoRunEffort } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { buildFacet, parseFacet } from '../../routes';
import { ageShort, timeAgo, type FacetModes } from '../../utils';
import { todoRunEffortPresentation } from '../todos/TodoRunEffortBadge';
import {
  DEFAULT_SESSION_FILTERS,
  SESSION_SORTS,
  SESSION_STATE_LABELS,
  flattenFamilies,
  latestActivity,
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
// Desktop columns: context · session / project (cost ends its meta line), state, last activity.
const COLUMNS = '@min-[760px]:grid-cols-[minmax(0,1fr)_100px_92px]';
// Breakpoint visibility goes on a plain wrapper with no display class: clicky-ui's
// styles.css re-emits base display utilities after this app's variants, so a
// variant overriding one on the same element never applies.
const NARROW_ONLY = '@min-[760px]:hidden';
const WIDE_ONLY = '@max-[760px]:hidden';

function SessionStateBadge({ session }: { session: AgentSession }) {
  const state = sessionState(session);
  const { status, icon } = STATE_BADGES[state];
  const label = SESSION_STATE_LABELS[state];
  return (
    <span className="inline-flex" title={session.stateReason || label}>
      <Badge
        variant="status"
        status={status}
        size="xs"
        icon={icon}
        className={cn('whitespace-nowrap', state === 'working' && '[&_svg]:motion-safe:animate-spin')}
      >
        {label}
      </Badge>
    </span>
  );
}

/** The narrow layout's one state indicator, pinned to the row's leading icon. */
function RowStatus({ session }: { session: AgentSession }) {
  const label = SESSION_STATE_LABELS[sessionState(session)];
  return (
    <span className={cn('absolute -bottom-1 -right-1', NARROW_ONLY)}>
      <span role="img" aria-label={label} title={session.stateReason || label} className="flex rounded-full bg-background p-px">
        <SessionActivityIcon session={session} className="size-3" />
      </span>
    </span>
  );
}

function sessionProvider(session: AgentSession): string {
  return session.provider || session.source;
}

function SessionContext({ session }: { session: AgentSession }) {
  const provider = sessionProvider(session);
  const Glyph = providerIcon(provider);
  return (
    // The meter's hover card holds copy buttons; keep their clicks from selecting the row.
    <span className="inline-flex size-8 shrink-0" onClick={event => event.stopPropagation()}>
      {session.context ? (
        <ContextMeter
          mode="gauge"
          className="p-0"
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

/** Model and effort after the title. Narrow rows drop it: the provider icon already names the model family. */
function ModelBadge({ session }: { session: AgentSession }) {
  if (!session.model && !session.effort) return null;
  const presentation = session.effort ? todoRunEffortPresentation(session.effort as TodoRunEffort) : undefined;
  const EffortIcon = presentation?.icon;
  return (
    <span className={cn('shrink-0', WIDE_ONLY)}>
      <span
        className="inline-flex items-center gap-0.5 rounded-full bg-muted px-1.5 py-px text-[10px] font-medium text-muted-foreground"
        title={[session.model, session.effort && `effort ${session.effort}`].filter(Boolean).join(' · ')}
      >
        {session.model}
        {EffortIcon ? <EffortIcon className="size-3" aria-label={`Effort ${session.effort}`} /> : null}
      </span>
    </span>
  );
}

/** Project and checkout. The dirty count belongs to the checkout it describes;
 *  narrow rows keep only its file count. */
function GitLocation({ session }: { session: AgentSession }) {
  const worktree = session.worktree && !session.worktree.primary ? session.worktree : undefined;
  const branch = session.worktree?.branch ?? session.cwd?.split('/').pop() ?? '';
  const CheckoutIcon = worktree ? UiFolderGit2 : UiGitBranch;
  const changes = session.git?.changes;
  const dirty = sessionDirtyCount(session);
  return (
    <span data-testid="session-location" className="flex min-w-0 items-center gap-1 overflow-hidden text-xs text-muted-foreground">
      {session.project ? (
        <>
          <span className="inline-flex min-w-0 items-center gap-1" title={session.project}>
            {/* -ml-px: the folder glyph sits 1.6px inside its box; this lines its ink up with the title above. */}
            <UiFolderGit className="-ml-px size-3 shrink-0" aria-label="Repository" />
            <span className="truncate">{session.project}</span>
          </span>
          {branch ? <span className="shrink-0" aria-hidden="true">·</span> : null}
        </>
      ) : null}
      {branch ? (
        <span
          className="inline-flex min-w-0 items-center gap-1"
          title={worktree ? `Worktree ${worktree.path} on ${worktree.branch}` : session.worktree ? `Branch ${branch} in the main checkout` : session.cwd}
        >
          <CheckoutIcon className={cn('size-3 shrink-0', !session.project && '-ml-px', worktree && 'text-primary')} aria-label={worktree ? 'Worktree' : 'Branch'} />
          <span className="truncate">{branch}</span>
        </span>
      ) : null}
      {changes && dirty > 0 ? (
        <span
          className="shrink-0 whitespace-nowrap text-[11px] tabular-nums"
          title={`${dirty} uncommitted file${dirty === 1 ? '' : 's'}, +${changes.adds} −${changes.dels} lines`}
        >
          <span className={WIDE_ONLY}>
            <span className="text-green-600">+{changes.adds}</span>{' '}
            <span className="text-red-600">−{changes.dels}</span>{' '}
          </span>
          <span>· {dirty}f</span>
        </span>
      ) : null}
    </span>
  );
}

function cost(usd: number): string {
  return `$${usd.toFixed(2)}`;
}

/** The session's spend, a badge following the git changes on the meta line. */
function CostBadge({ usd }: { usd: number }) {
  return (
    <span data-testid="session-cost" className="flex shrink-0 text-xs">
      <span className="inline-flex items-center rounded-full bg-muted px-1.5 py-px text-[10px] font-medium leading-3.5 tabular-nums text-muted-foreground">
        {cost(usd)}
      </span>
    </span>
  );
}

/** Narrow rows end their meta line with the age the wide layout gives a column. */
function NarrowTrailing({ activity, title }: { activity: string; title: string }) {
  return (
    <span className={cn('ml-auto shrink-0 text-xs', NARROW_ONLY)}>
      <span className="whitespace-nowrap text-xs tabular-nums text-muted-foreground" title={title}>
        {ageShort(activity)}
      </span>
    </span>
  );
}

/** One top-level session. Its sub-sessions are not rows: they are reached from
 *  the session inspector's ⋮ menu, so the row only dates the session by the
 *  latest activity anywhere in its family. */
function SessionRow({ family, selectedId, onSelect }: {
  family: SessionFamily;
  selectedId: string;
  onSelect: (id: string) => void;
}) {
  const { session, children } = family;
  const activity = new Date(latestActivity(family)).toISOString();
  const activityTitle = children.length > 0 ? 'Latest activity across this session and its sub-sessions' : "This session's last activity";
  return (
    <li data-session-id={session.id} className="m-0">
      <div className={cn(
        'group flex min-w-0 items-center border-b border-border/50 py-2 pl-3 pr-2 hover:bg-muted/50',
        session.id === selectedId && 'bg-accent',
      )}>
        <button
          type="button"
          onClick={() => onSelect(session.id)}
          aria-current={session.id === selectedId ? 'true' : undefined}
          aria-label={`View ${session.title}`}
          className={cn('grid min-w-0 flex-1 gap-2 rounded-sm text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring @min-[760px]:items-center', COLUMNS)}
        >
          <span className="flex min-w-0 items-center gap-2">
            <span data-testid="session-context" className="relative inline-flex shrink-0">
              <SessionContext session={session} />
              <RowStatus session={session} />
            </span>
            <span className="min-w-0 flex-1 space-y-1">
              <span data-testid="session-title-line" className="flex min-w-0 items-center gap-2">
                <span
                  className="min-w-0 truncate text-sm font-medium text-foreground"
                  title={[session.title, session.model].filter(Boolean).join(' · ')}
                >
                  {session.title}
                </span>
                <ModelBadge session={session} />
              </span>
              <span data-testid="session-meta" className="flex min-w-0 items-center gap-2">
                <GitLocation session={session} />
                <CostBadge usd={session.costUsd} />
                <NarrowTrailing activity={activity} title={activityTitle} />
              </span>
            </span>
          </span>
          <span data-testid="session-wide-state" className={WIDE_ONLY}><SessionStateBadge session={session} /></span>
          <span className={cn('pr-2 text-right text-xs tabular-nums text-muted-foreground', WIDE_ONLY)} title={activityTitle}>
            {timeAgo(activity)}
          </span>
        </button>
      </div>
    </li>
  );
}

const FILTER_PARAM = 'sessions.';

function isSort(value: string): value is SessionSort {
  return SESSION_SORTS.some(option => option.value === value);
}

/**
 * Filters live in the URL (`sessions.search`, …) so a reload or shared link
 * keeps them. Project and state use the PR filters' facet encoding: a CSV
 * where an excluded key carries a leading `-` (`sessions.state=working,-failed`).
 */
function initialFilters(): SessionFilters {
  const params = new URLSearchParams(window.location.search);
  const sort = params.get(`${FILTER_PARAM}sort`) || DEFAULT_SESSION_FILTERS.sort;
  if (!isSort(sort)) throw new Error(`Invalid session sort ${JSON.stringify(sort)}; expected one of ${SESSION_SORTS.map(o => o.value).join(', ')}`);
  const state = parseFacet(params.get(`${FILTER_PARAM}state`));
  for (const value of Object.keys(state)) {
    if (!Object.hasOwn(SESSION_STATE_LABELS, value)) throw new Error(`Invalid session state ${JSON.stringify(value)}; expected one of ${Object.keys(SESSION_STATE_LABELS).join(', ')}`);
  }
  return {
    search: params.get(`${FILTER_PARAM}search`) ?? '',
    project: parseFacet(params.get(`${FILTER_PARAM}project`)),
    state,
    sort,
  };
}

function useUrlFilters() {
  const [filters, setFilters] = useState(initialFilters);
  useEffect(() => {
    const url = new URL(window.location.href);
    for (const [key, value] of Object.entries(filters) as [keyof SessionFilters, string | FacetModes][]) {
      const encoded = typeof value === 'string' ? (value === DEFAULT_SESSION_FILTERS[key] ? '' : value) : buildFacet(value);
      if (encoded) url.searchParams.set(FILTER_PARAM + key, encoded);
      else url.searchParams.delete(FILTER_PARAM + key);
    }
    if (url.href !== window.location.href) window.history.replaceState(window.history.state, '', url);
  }, [filters]);
  return [filters, setFilters] as const;
}

/** Sort is a choice, not a filter: an icon button whose menu checks the current order. */
function SortMenu({ sort, onChange }: { sort: SessionSort; onChange: (sort: SessionSort) => void }) {
  const current = SESSION_SORTS.find(option => option.value === sort)!;
  return (
    <DropdownMenu
      align="right"
      menuLabel="Sort sessions"
      trigger={(
        <Button variant="ghost" size="icon" className="size-8 shrink-0" aria-label={`Sort: ${current.label}`} title={`Sort: ${current.label}`}>
          <Icon icon={sortDescendingThin} className="size-4" />
        </Button>
      )}
      items={SESSION_SORTS.map(option => ({
        label: option.label,
        onSelect: () => onChange(option.value),
        icon: option.value === sort ? UiCheck : undefined,
      }))}
    />
  );
}

export interface SessionListProps {
  families: SessionFamily[];
  selectedId: string;
  onSelect: (id: string) => void;
}

/**
 * The Sessions tab list: top-level sessions, most recently active first. A
 * session matches the filters when it or any of its sub-sessions does; the
 * sub-sessions themselves are reached from the session inspector. Below 760px
 * of container width a row stacks into two lines; wider, state, last activity
 * and cost become columns.
 */
export function SessionList({ families, selectedId, onSelect }: SessionListProps) {
  const [filters, setFilters] = useUrlFilters();
  const projects = useMemo(() => {
    const names = new Set(Object.keys(filters.project));
    for (const { project } of flattenFamilies(families)) {
      if (project) names.add(project);
    }
    return [...names].sort();
  }, [families, filters.project]);
  const rows = useMemo(() => visibleFamilies(families, filters), [families, filters]);

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
            kind: 'multi',
            label: 'Project',
            placeholder: 'All projects',
            value: filters.project,
            options: projects.map(value => ({ value, label: value })),
            onChange: project => setFilters(current => ({ ...current, project })),
          },
          {
            key: 'state',
            kind: 'multi',
            label: 'State',
            placeholder: 'All states',
            value: filters.state,
            options: STATE_OPTIONS,
            onChange: state => setFilters(current => ({ ...current, state })),
          },
        ]}
        trailing={<SortMenu sort={filters.sort} onChange={sort => setFilters(current => ({ ...current, sort }))} />}
      />
      <div className={cn('shrink-0', WIDE_ONLY)} aria-hidden="true">
        <div className={cn('grid gap-2 border-b border-border bg-muted/40 py-2 pl-3 pr-2 text-xs text-muted-foreground', COLUMNS)}>
          <span>Context · session / project</span>
          <span>State</span>
          <span className="pr-2 text-right">Last activity</span>
        </div>
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto" aria-label="Sessions">
        {rows.length ? (
          <ol className="m-0 list-none p-0">
            {rows.map(family => (
              <SessionRow key={family.session.id} family={family} selectedId={selectedId} onSelect={onSelect} />
            ))}
          </ol>
        ) : (
          <div className="flex h-full flex-col items-center justify-center gap-3 px-6 text-center">
            <UiSearch className="size-6 text-muted-foreground" />
            <p className="text-sm font-medium">{families.length ? 'No sessions match' : 'No sessions yet'}</p>
            {families.length ? (
              <>
                <p className="text-xs text-muted-foreground">Try another project, state, or search.</p>
                <Button variant="outline" size="sm" onClick={() => setFilters(current => ({ ...DEFAULT_SESSION_FILTERS, sort: current.sort }))}>
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
