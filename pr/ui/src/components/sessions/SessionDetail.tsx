import { useId, useMemo, useState, type ComponentType, type ReactNode } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { SessionInspector, formatCost } from '@flanksource/clicky-ui/ai';
import { cn } from '@flanksource/clicky-ui/utils';
import {
  UiChevronDown,
  UiChevronRight,
  UiExternalLink,
  UiFile,
  UiFolderGit2,
  UiGitBranch,
  UiGitCommit,
  UiGitPr,
  type IconProps,
} from '@flanksource/clicky-ui/icons';
import type { AgentSession } from '../../types';
import { queryKeys } from '../../query';
import { useProjectGit } from '../../projectGitQueries';
import { BranchLandActions } from '../BranchLandActions';
import { CommitRangeChanges } from '../CommitRangeChanges';
import { branchRangeRequests, branchRangeSummary } from '../ProjectBranchChanges';
import { sessionCollection } from './sessionCollection';
import { SessionActivityIcon } from './SessionList';
import { attentionReasons, sessionDirtyCount, sessionFailed, workStateBadge } from './sessionView';

export interface SessionDetailProps {
  session: AgentSession;
  /** Every listed session, so the inspector can offer this one's sub-sessions. */
  sessions: AgentSession[];
  onOpenTodo: (id: string) => void;
  onOpenWorktree: (project: string, path: string) => void;
}

function activityLabel(session: AgentSession): string {
  if (session.activityState === 'approval') return 'Needs approval';
  if (session.activityState === 'ask') return 'Asked a question';
  if (session.processActive) return session.activityState === 'thinking' ? 'Thinking' : 'Running';
  if (sessionFailed(session)) return 'Failed';
  return 'Finished';
}

function formatDuration(ms: number): string {
  const minutes = Math.round(ms / 60_000);
  if (minutes < 60) return `${minutes}m`;
  return `${Math.floor(minutes / 60)}h ${minutes % 60}m`;
}

type WorkTone = 'warning' | 'danger' | 'ok';

const WORK_TONES: Record<WorkTone, string> = {
  warning: 'text-amber-700 dark:text-amber-400',
  danger: 'text-red-700 dark:text-red-400',
  ok: 'text-emerald-700 dark:text-emerald-400',
};

function WorkItem({ icon: Icon, label, tone, title, children }: {
  icon: ComponentType<IconProps>;
  label: string;
  tone?: WorkTone | undefined;
  title?: string | undefined;
  children: ReactNode;
}) {
  return (
    <span data-work-item={label} title={title} className={cn('inline-flex shrink-0 items-center gap-1', tone && WORK_TONES[tone])}>
      <Icon className="size-3.5 shrink-0" aria-label={label} />
      {children}
    </span>
  );
}

/** Merge into base and Create PR for the session's branch, at the header's right while it has commits to land. */
function SessionLandActions({ session }: { session: AgentSession }) {
  const queryClient = useQueryClient();
  const { worktree, git, pr, project } = session;
  if (!project || !worktree?.branch || !git || git.ahead === 0 || pr?.state === 'merged') return null;
  return (
    <div className="shrink-0">
      <BranchLandActions
        projectName={project}
        branch={worktree.branch}
        base={worktree.base || 'base'}
        ahead={git.ahead}
        dirtyCount={sessionDirtyCount(session)}
        baseCheckedOut={git.baseCheckedOut}
        worktreePath={worktree.path}
        onMerged={() => { void queryClient.invalidateQueries({ queryKey: queryKeys.agentSessions() }); }}
      />
    </div>
  );
}

// The model, effort and context meter live in the transcript toolbar; the
// header keeps what that toolbar cannot say: activity, project, time and cost.
function SessionHeader({ session, onOpenTodo, workState }: {
  session: AgentSession;
  onOpenTodo: (id: string) => void;
  workState: { open: boolean; onToggle: () => void; controls: string };
}) {
  // The work-state line below carries conflicts, failing checks and left-behind changes.
  const reasons = attentionReasons(session, { work: false });
  const badge = workStateBadge(session);
  return (
    <header className="shrink-0 border-b border-border bg-background px-4 py-2 max-md:!px-3">
      <div className="flex min-w-0 items-center gap-2">
        <SessionActivityIcon session={session} />
        <h2 className="min-w-0 flex-1 truncate text-sm font-semibold text-foreground" title={session.title}>{session.title}</h2>
        <SessionLandActions session={session} />
      </div>
      <div className="mt-1 flex min-w-0 items-center gap-x-3 overflow-hidden whitespace-nowrap text-[11px] text-muted-foreground">
        <span className="font-medium text-foreground max-md:hidden">{activityLabel(session)}</span>
        {session.project ? <span className="min-w-0 truncate">{session.project}</span> : null}
        {session.todo ? (
          <button type="button" className="min-w-0 truncate text-primary hover:underline" onClick={() => onOpenTodo(session.todo!.id)}>
            todo · {session.todo.title}
          </button>
        ) : null}
        {session.pid ? <span className="max-md:hidden">pid {session.pid}</span> : null}
        {session.durationMs && session.durationMs >= 60_000 ? <span>{formatDuration(session.durationMs)}</span> : null}
        {session.costUsd > 0 ? <span>{formatCost(session.costUsd)}</span> : null}
        {/* A plain wrapper carries the breakpoint: see the note in SessionDetail. */}
        <span className="ml-auto shrink-0 md:hidden">
          <button
            type="button"
            aria-expanded={workState.open}
            aria-controls={workState.controls}
            aria-label={`Work state: ${badge.label}`}
            onClick={workState.onToggle}
            className={cn(
              'inline-flex min-h-6 items-center gap-0.5 rounded border px-1.5 tabular-nums',
              badge.tone === 'warning' ? 'border-amber-500/50 text-amber-700 dark:text-amber-400' : 'border-border',
            )}
          >
            {badge.label}
            {workState.open ? <UiChevronDown className="size-3" /> : <UiChevronRight className="size-3" />}
          </button>
        </span>
      </div>
      {session.stateReason || reasons.length > 0 ? (
        <p className="mt-1.5 text-[11px] font-medium text-amber-700 dark:text-amber-400">
          {session.stateReason || reasons.join(' · ')}
        </p>
      ) : null}
    </header>
  );
}

/**
 * Where the session's work sits, on one line: checkout, uncommitted changes,
 * commits ahead/behind the base, and the PR. It carries the work-state warnings
 * (conflicts, changes left behind, failing checks) the header leaves out.
 */
function WorkStateLine({ session, onOpenWorktree }: { session: AgentSession; onOpenWorktree: SessionDetailProps['onOpenWorktree'] }) {
  const { worktree, git, pr, project } = session;
  const base = worktree?.base || 'base';
  const dirty = sessionDirtyCount(session);
  const linked = worktree && !worktree.primary ? (git && !git.worktreeLive ? 'worktree removed' : 'worktree') : undefined;
  const checksFailing = !!pr && pr.state !== 'merged' && pr.checkStatus === 'failure';
  return (
    <section aria-label="Work state" className="flex min-w-0 items-center gap-x-4 overflow-x-auto whitespace-nowrap text-[11px] text-muted-foreground">
      <WorkItem icon={linked ? UiFolderGit2 : UiGitBranch} label="Checkout" title={worktree?.path ?? session.cwd}>
        {worktree ? (
          <>
            <span className="font-medium text-foreground">{worktree.branch || 'detached'}</span>
            <span>→ {base}</span>
            {linked ? <span className="rounded border border-border px-1 text-[10px]">{linked}</span> : null}
            {project && (!git || git.worktreeLive) ? (
              <button type="button" className="text-primary hover:underline" title="Open in Projects" onClick={() => onOpenWorktree(project, worktree.path)}>
                Open
              </button>
            ) : null}
          </>
        ) : (
          <span className="font-mono">{session.cwd || 'unknown directory'}</span>
        )}
      </WorkItem>
      {git ? (
        <>
          <WorkItem
            icon={UiFile}
            label="Uncommitted"
            tone={git.changes.conflict ? 'danger' : dirty > 0 ? 'warning' : 'ok'}
            title={`${git.changes.staged} staged · ${git.changes.unstaged + git.changes.both} unstaged · ${git.changes.untracked} untracked`}
          >
            {dirty === 0 ? (git.worktreeLive ? 'clean' : '—') : (
              <span className="inline-flex items-center gap-1.5 tabular-nums">
                {git.changes.conflict ? <span className="font-semibold">{git.changes.conflict} conflicted</span> : null}
                <span>{dirty} uncommitted</span>
                <span className="text-emerald-600">+{git.changes.adds}</span>
                <span className="text-red-600">−{git.changes.dels}</span>
                {!session.processActive ? <span className="font-medium">left behind</span> : null}
              </span>
            )}
          </WorkItem>
          <WorkItem
            icon={UiGitCommit}
            label="Unmerged"
            tone={git.ahead > 0 ? undefined : 'ok'}
            title={`${git.ahead} commits ahead of and ${git.behind} behind ${base}`}
          >
            <span className="tabular-nums">↑{git.ahead} ↓{git.behind}</span>
          </WorkItem>
        </>
      ) : null}
      <WorkItem icon={UiGitPr} label="Pull request" tone={checksFailing ? 'danger' : undefined}>
        {pr ? (
          <a href={pr.url} target="_blank" rel="noreferrer" className="inline-flex items-center gap-1 hover:underline">
            <span className={cn('font-medium', pr.state === 'merged' && 'text-violet-600 dark:text-violet-400')}>#{pr.number} {pr.state}</span>
            {pr.state !== 'merged' && pr.checkStatus !== 'none' ? <span>· checks {checksFailing ? 'failing' : pr.checkStatus}</span> : null}
            {pr.reviewDecision ? <span>· {pr.reviewDecision.toLowerCase().replace(/_/g, ' ')}</span> : null}
            <UiExternalLink className="size-3" />
          </a>
        ) : 'no PR'}
      </WorkItem>
    </section>
  );
}

function BranchChanges({ project, branch }: { project: string; branch: string }) {
  const git = useProjectGit(project);
  if (git.error) return <div role="alert" className="p-3 text-xs text-red-600">{git.error.message}</div>;
  if (!git.data) return <div role="status" className="p-3 text-xs text-muted-foreground">Loading {branch}…</div>;
  const info = git.data.branches.find(candidate => candidate.name === branch);
  if (!info) return <div role="status" className="p-3 text-xs text-muted-foreground">{branch} has no commits ahead of {git.data.base}</div>;
  const { filesRequest, diffRequest } = branchRangeRequests(project, info);
  return (
    <CommitRangeChanges
      options={[{ id: 'all', label: `All changes vs ${git.data.base}` }]}
      showPicker={false}
      summary={branchRangeSummary(info.diff)}
      filesRequest={filesRequest}
      diffRequest={diffRequest}
      unreachable={<div role="status" className="px-3 py-3 text-xs text-muted-foreground">Branch no longer exists</div>}
    />
  );
}

/**
 * One session: what it is doing, where its work sits (worktree, uncommitted
 * and unmerged changes, merge into base, PR), then its transcript — read and
 * followed from Captain's session handler — and the branch's changes.
 */
export function SessionDetail({ session, sessions, onOpenTodo, onOpenWorktree }: SessionDetailProps) {
  const [workOpen, setWorkOpen] = useState(false);
  const collection = useMemo(() => sessionCollection(session, sessions), [session, sessions]);
  const workStateId = useId();
  const branch = session.worktree?.branch;
  const canShowChanges = !!(session.project && branch && session.git && session.git.ahead > 0);
  const hostTabs = canShowChanges
    ? [{ id: 'branch-changes', label: 'Branch changes', icon: UiGitCommit, content: <BranchChanges project={session.project!} branch={branch!} /> }]
    : [];

  return (
    <div className="flex h-full min-h-0 flex-col bg-background">
      {/* Narrow screens collapse the work state into a header badge so the transcript keeps the screen.
          A plain wrapper with no display class carries each breakpoint: clicky-ui's styles.css
          re-emits base utilities (hidden, inline-flex) after this app's variants, so a variant
          that overrides one of them on the same element never applies. */}
      <SessionHeader
        session={session}
        onOpenTodo={onOpenTodo}
        workState={{ open: workOpen, onToggle: () => setWorkOpen(open => !open), controls: workStateId }}
      />
      <div
        id={workStateId}
        className={cn(
          'shrink-0 border-b border-border bg-muted/20 px-4 py-1.5 max-md:!px-3',
          !workOpen && 'max-md:hidden',
        )}
      >
        <WorkStateLine session={session} onOpenWorktree={onOpenWorktree} />
      </div>
      <div className="flex min-h-0 flex-1 flex-col overflow-auto">
        <SessionInspector
          key={session.id}
          session={collection}
          follow={session.processActive}
          layout="compact"
          hostTabs={hostTabs}
          className="min-h-0 flex-1 text-xs"
        />
      </div>
    </div>
  );
}
