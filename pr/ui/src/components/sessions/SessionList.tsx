import { useMemo, useState } from 'react';
import { FilterBar } from '@flanksource/clicky-ui/components';
import { cn } from '@flanksource/clicky-ui/utils';
import {
  UiChevronDown,
  UiChevronRight,
  UiCircleX,
  UiGitBranch,
  UiGitCommit,
  UiGitMerge,
  UiGitPr,
  UiQuestion,
  UiShieldCheck,
  UiWarningTriangle,
} from '@flanksource/clicky-ui/icons';
import type { AgentSession } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { timeAgo } from '../../utils';
import {
  SESSION_GROUPS,
  attentionReasons,
  contextTone,
  contextUsedPercent,
  filterSessions,
  groupSessions,
  sessionDirtyCount,
  sessionFailed,
  waitingOnInput,
  type SessionFilters,
  type SessionGroupKey,
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

/** Icon + number only; color only for problems or work waiting on you. */
function QuietCounters({ session }: { session: AgentSession }) {
  const dirty = sessionDirtyCount(session);
  const { git, pr } = session;
  return (
    <span className="flex shrink-0 items-center gap-2 text-[10px] tabular-nums text-muted-foreground">
      {git?.changes.conflict ? (
        <span className="inline-flex items-center gap-0.5 text-red-500" title="Merge conflicts">
          <UiWarningTriangle className="size-3" />
          {git.changes.conflict}
        </span>
      ) : null}
      {dirty > 0 ? (
        <span className="inline-flex items-center gap-0.5 text-amber-600" title={`${dirty} uncommitted (+${git?.changes.adds} −${git?.changes.dels})`}>
          <span className="size-1.5 rounded-full bg-current" />
          {dirty}
        </span>
      ) : null}
      {git && git.ahead > 0 ? (
        <span className="inline-flex items-center gap-0.5" title={`${git.ahead} commits not merged into ${session.worktree?.base || 'base'}`}>
          <UiGitCommit className="size-3" />
          {git.ahead}
        </span>
      ) : null}
      {pr ? (
        <span
          className={cn(
            'inline-flex items-center gap-0.5',
            pr.state === 'merged' && 'text-violet-500',
            pr.state !== 'merged' && pr.checkStatus === 'failure' && 'text-red-500',
          )}
          title={`PR #${pr.number} ${pr.state} · checks ${pr.checkStatus}`}
        >
          {pr.state === 'merged' ? <UiGitMerge className="size-3" /> : <UiGitPr className="size-3" />}
          {pr.number}
        </span>
      ) : null}
    </span>
  );
}

function ContextUnderline({ session }: { session: AgentSession }) {
  const percent = contextUsedPercent(session);
  if (percent === undefined) return null;
  const tone = contextTone(percent);
  return (
    <span className="absolute inset-x-0 bottom-0 h-0.5" aria-hidden>
      <span
        className={cn('block h-full', tone === 'danger' ? 'bg-red-500' : tone === 'warning' ? 'bg-amber-500' : 'bg-emerald-500/60')}
        style={{ width: `${percent}%` }}
      />
    </span>
  );
}

function SessionListRow({ session, selected, onSelect }: { session: AgentSession; selected: boolean; onSelect: () => void }) {
  const branch = session.worktree?.branch ?? session.cwd?.split('/').pop() ?? '';
  const percent = contextUsedPercent(session);
  return (
    <button
      type="button"
      onClick={onSelect}
      aria-current={selected ? 'true' : undefined}
      title={percent !== undefined ? `Context ${Math.round(percent)}% used` : undefined}
      className={cn(
        'relative flex w-full min-w-0 flex-col gap-0.5 px-3 py-2 text-left hover:bg-accent/60',
        selected && 'bg-accent',
      )}
    >
      <span className="flex min-w-0 items-center gap-2">
        <SessionActivityIcon session={session} />
        <span className="min-w-0 flex-1 truncate text-xs font-medium text-foreground">{session.title}</span>
        <span className="shrink-0 whitespace-nowrap text-[10px] tabular-nums text-muted-foreground" title={session.lastActivityAt}>
          {timeAgo(session.lastActivityAt)}
        </span>
      </span>
      <span className="flex min-w-0 items-center gap-2 pl-5">
        <span className="min-w-0 flex-1 truncate">
          {waitingOnInput(session) && session.stateReason ? (
            <span className="text-[10px] font-medium text-amber-700 dark:text-amber-400">{session.stateReason}</span>
          ) : (
            <span className="inline-flex min-w-0 max-w-full items-center gap-1 text-[10px] text-muted-foreground" title={session.worktree?.path ?? session.cwd}>
              {branch ? <UiGitBranch className="size-3 shrink-0 opacity-60" /> : null}
              <span className="truncate">{[session.project, branch].filter(Boolean).join(' · ')}</span>
            </span>
          )}
        </span>
        <QuietCounters session={session} />
      </span>
      <ContextUnderline session={session} />
    </button>
  );
}

export interface SessionListProps {
  sessions: AgentSession[];
  projects: string[];
  selectedId: string;
  onSelect: (id: string) => void;
}

/**
 * The Sessions tab sidebar: sessions grouped Active → Recent (24h) → Older, two
 * quiet lines each — state, title and age; then branch (or the question the
 * session waits on) with icon counters for dirty files, unmerged commits and
 * the PR. A 2px underline shows context-window usage.
 */
export function SessionList({ sessions, projects, selectedId, onSelect }: SessionListProps) {
  const [filters, setFilters] = useState<SessionFilters>({ search: '', project: '', attention: false });
  const [open, setOpen] = useState<Record<SessionGroupKey, boolean>>(
    () => Object.fromEntries(SESSION_GROUPS.map(group => [group.key, group.defaultOpen])) as Record<SessionGroupKey, boolean>,
  );
  const groups = useMemo(() => groupSessions(filterSessions(sessions, filters), Date.now()), [sessions, filters]);
  const needYou = groups.active.filter(session => attentionReasons(session).length > 0).length;

  return (
    <div className="flex h-full min-h-0 flex-col">
      <FilterBar
        className="shrink-0 border-b border-border px-2 py-1.5"
        search={{
          value: filters.search,
          onChange: search => setFilters(current => ({ ...current, search })),
          placeholder: 'Search title, branch, todo…',
          ariaLabel: 'Search sessions',
        }}
        filters={[
          {
            key: 'project',
            kind: 'enum',
            label: 'Project',
            value: filters.project,
            options: [{ value: '', label: 'All' }, ...projects.map(project => ({ value: project }))],
            onChange: project => setFilters(current => ({ ...current, project })),
          },
          {
            key: 'attention',
            kind: 'boolean',
            label: 'Needs attention',
            value: filters.attention,
            onChange: attention => setFilters(current => ({ ...current, attention })),
          },
        ]}
      />
      <div className="min-h-0 flex-1 overflow-y-auto">
        {SESSION_GROUPS.map(group => {
          const rows = groups[group.key];
          if (rows.length === 0 && group.key !== 'active') return null;
          const Chevron = open[group.key] ? UiChevronDown : UiChevronRight;
          return (
            <section key={group.key} aria-label={group.label}>
              <button
                type="button"
                onClick={() => setOpen(current => ({ ...current, [group.key]: !current[group.key] }))}
                className="sticky top-0 z-10 flex w-full items-center gap-1.5 bg-background/95 px-3 pb-1 pt-3 text-xs font-semibold text-foreground backdrop-blur"
              >
                <Chevron className="size-3 text-muted-foreground" />
                {group.label}
                <span className="text-[11px] font-normal tabular-nums text-muted-foreground">{rows.length}</span>
                {group.key === 'active' && needYou > 0 ? (
                  <span className="ml-auto text-[11px] font-medium text-amber-700 dark:text-amber-400">{needYou} need you</span>
                ) : null}
              </button>
              {open[group.key] && rows.length === 0 ? (
                <p className="px-3 pb-2 pl-8 text-[11px] text-muted-foreground">No running sessions</p>
              ) : null}
              {open[group.key]
                ? rows.map(session => (
                  <SessionListRow key={session.id} session={session} selected={session.id === selectedId} onSelect={() => onSelect(session.id)} />
                ))
                : null}
            </section>
          );
        })}
      </div>
    </div>
  );
}
