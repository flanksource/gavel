import { useState, type ComponentType, type ReactNode } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { ContextMeter, SessionInspector, formatCost } from '@flanksource/clicky-ui/ai';
import { Button } from '@flanksource/clicky-ui/components';
import { cn } from '@flanksource/clicky-ui/utils';
import {
  UiComment,
  UiExternalLink,
  UiFile,
  UiFolderGit2,
  UiGitBranch,
  UiGitCommit,
  UiGitMerge,
  UiGitPr,
  UiRobotAi,
  type IconProps,
} from '@flanksource/clicky-ui/icons';
import type { AgentSession } from '../../types';
import { queryKeys } from '../../query';
import { useProjectGit } from '../../projectGitQueries';
import { BranchLandActions } from '../BranchLandActions';
import { CommitRangeChanges } from '../CommitRangeChanges';
import { branchRangeRequests, branchRangeSummary } from '../ProjectBranchChanges';
import { captainSessionUrl } from '../todos/TodoSessionDetail';
import { SessionActivityIcon } from './SessionList';
import { attentionReasons, sessionDirtyCount, sessionFailed } from './sessionView';

type DetailTab = 'transcript' | 'changes';

export interface SessionDetailProps {
  session: AgentSession;
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

function WorkRow({ icon: Icon, label, tone, children }: {
  icon: ComponentType<IconProps>;
  label: string;
  tone?: 'warning' | 'danger' | 'ok';
  children: ReactNode;
}) {
  return (
    <div className="grid grid-cols-[7.5rem_minmax(0,1fr)] items-start gap-2 border-b border-border px-3 py-2 last:border-b-0">
      <span className={cn(
        'inline-flex items-center gap-1.5 text-[11px] font-medium text-muted-foreground',
        tone === 'warning' && 'text-amber-700 dark:text-amber-400',
        tone === 'danger' && 'text-red-700 dark:text-red-400',
        tone === 'ok' && 'text-emerald-700 dark:text-emerald-400',
      )}>
        <Icon className="size-3.5 shrink-0" />
        {label}
      </span>
      <div className="min-w-0 text-xs text-foreground">{children}</div>
    </div>
  );
}

function SessionHeader({ session, onOpenTodo }: { session: AgentSession; onOpenTodo: (id: string) => void }) {
  const reasons = attentionReasons(session);
  return (
    <header className="shrink-0 border-b border-border bg-background px-4 py-2">
      <div className="flex min-w-0 items-center gap-2">
        <SessionActivityIcon session={session} />
        <h2 className="min-w-0 flex-1 truncate text-sm font-semibold text-foreground" title={session.title}>{session.title}</h2>
        {session.context ? (
          <ContextMeter
            mode="bar"
            usedPercent={100 - session.context.freePercent}
            usedTokens={session.context.usedTokens}
            windowTokens={session.context.windowTokens}
            sessionId={session.id}
            providerSessionId={session.providerSessionId}
            provider={session.provider}
            model={session.model}
            effort={session.effort}
            tokens={session.tokens}
            cost={{ total: session.costUsd }}
          />
        ) : null}
      </div>
      <div className="mt-1 flex min-w-0 flex-wrap items-center gap-x-3 gap-y-1 text-[11px] text-muted-foreground">
        <span className="font-medium text-foreground">{activityLabel(session)}</span>
        {session.project ? <span>{session.project}</span> : null}
        {session.todo ? (
          <button type="button" className="truncate text-primary hover:underline" onClick={() => onOpenTodo(session.todo!.id)}>
            todo · {session.todo.title}
          </button>
        ) : null}
        <span className="inline-flex items-center gap-1">
          <UiRobotAi className="size-3" />
          {[session.model || session.source, session.effort].filter(Boolean).join(' · ')}
        </span>
        {session.pid ? <span>pid {session.pid}</span> : null}
        {session.durationMs && session.durationMs >= 60_000 ? <span>{formatDuration(session.durationMs)}</span> : null}
        {session.costUsd > 0 ? <span>{formatCost(session.costUsd)}</span> : null}
      </div>
      {session.stateReason || reasons.length > 0 ? (
        <p className="mt-1.5 text-[11px] font-medium text-amber-700 dark:text-amber-400">
          {session.stateReason || reasons.join(' · ')}
        </p>
      ) : null}
    </header>
  );
}

function WorkStatePanel({ session, onOpenWorktree }: { session: AgentSession; onOpenWorktree: SessionDetailProps['onOpenWorktree'] }) {
  const queryClient = useQueryClient();
  const { worktree, git, pr, project } = session;
  const base = worktree?.base || 'base';
  const dirty = sessionDirtyCount(session);
  return (
    <section aria-label="Work state" className="overflow-hidden rounded-lg border border-border bg-card">
      <WorkRow icon={worktree && !worktree.primary ? UiFolderGit2 : UiGitBranch} label="Worktree">
        {worktree ? (
          <div className="flex min-w-0 flex-col gap-0.5">
            <span className="flex min-w-0 flex-wrap items-center gap-1.5">
              <span className="truncate font-medium">{worktree.branch || 'detached'}</span>
              <span className="shrink-0 text-muted-foreground">→ {base}</span>
              <span className="rounded border border-border px-1 text-[10px] text-muted-foreground">
                {worktree.primary ? 'main checkout' : git && !git.worktreeLive ? 'worktree removed' : 'linked worktree'}
              </span>
              {project && (!git || git.worktreeLive) ? (
                <button type="button" className="text-[11px] text-primary hover:underline" onClick={() => onOpenWorktree(project, worktree.path)}>
                  Open in Projects
                </button>
              ) : null}
            </span>
            <span className="truncate font-mono text-[10px] text-muted-foreground" title={worktree.path}>{worktree.path}</span>
          </div>
        ) : (
          <span className="font-mono text-[11px] text-muted-foreground">{session.cwd || 'unknown directory'}</span>
        )}
      </WorkRow>
      {git ? (
        <>
          <WorkRow icon={UiFile} label="Uncommitted" tone={git.changes.conflict ? 'danger' : dirty > 0 ? 'warning' : 'ok'}>
            {dirty === 0 ? (
              <span className="text-muted-foreground">{git.worktreeLive ? 'Clean' : '—'}</span>
            ) : (
              <span className="flex flex-wrap items-center gap-x-3 gap-y-0.5 tabular-nums">
                {git.changes.conflict ? <span className="font-semibold text-red-600">{git.changes.conflict} conflicted</span> : null}
                <span>{git.changes.staged} staged</span>
                <span>{git.changes.unstaged + git.changes.both} unstaged</span>
                <span>{git.changes.untracked} untracked</span>
                <span className="text-emerald-600">+{git.changes.adds}</span>
                <span className="text-red-600">−{git.changes.dels}</span>
              </span>
            )}
          </WorkRow>
          <WorkRow icon={UiGitCommit} label="Unmerged" tone={git.ahead > 0 ? undefined : 'ok'}>
            <span className="tabular-nums">{git.ahead} ahead · {git.behind} behind {base}</span>
          </WorkRow>
          {project && worktree?.branch && pr?.state !== 'merged' ? (
            <WorkRow icon={UiGitMerge} label={`Merge → ${base}`}>
              <BranchLandActions
                projectName={project}
                branch={worktree.branch}
                base={base}
                ahead={git.ahead}
                dirtyCount={dirty}
                baseCheckedOut={git.baseCheckedOut}
                worktreePath={worktree.path}
                onMerged={() => { void queryClient.invalidateQueries({ queryKey: queryKeys.agentSessions() }); }}
              />
            </WorkRow>
          ) : null}
        </>
      ) : null}
      <WorkRow icon={UiGitPr} label="Pull request">
        {pr ? (
          <span className="flex flex-wrap items-center gap-2">
            <span className={cn(
              'font-medium',
              pr.state === 'merged' && 'text-violet-600 dark:text-violet-400',
              pr.state !== 'merged' && pr.checkStatus === 'failure' && 'text-red-600',
            )}>
              #{pr.number} {pr.state}
            </span>
            {pr.state !== 'merged' && pr.checkStatus !== 'none' ? <span className="text-muted-foreground">checks {pr.checkStatus}</span> : null}
            {pr.reviewDecision ? <span className="text-muted-foreground">{pr.reviewDecision.toLowerCase().replace(/_/g, ' ')}</span> : null}
            <a href={pr.url} target="_blank" rel="noreferrer" className="inline-flex items-center gap-1 text-[11px] text-primary hover:underline">
              Open <UiExternalLink className="size-3" />
            </a>
          </span>
        ) : (
          <span className="text-muted-foreground">No PR for this branch</span>
        )}
      </WorkRow>
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
export function SessionDetail({ session, onOpenTodo, onOpenWorktree }: SessionDetailProps) {
  const [tab, setTab] = useState<DetailTab>('transcript');
  const branch = session.worktree?.branch;
  const canShowChanges = !!(session.project && branch && session.git && session.git.ahead > 0);
  const tabs: { key: DetailTab; label: string; icon: ComponentType<IconProps>; disabled: boolean }[] = [
    { key: 'transcript', label: 'Transcript', icon: UiComment, disabled: false },
    { key: 'changes', label: 'Branch changes', icon: UiGitCommit, disabled: !canShowChanges },
  ];
  const active = tab === 'changes' && !canShowChanges ? 'transcript' : tab;

  return (
    <div className="flex h-full min-h-0 flex-col bg-background">
      <SessionHeader session={session} onOpenTodo={onOpenTodo} />
      <div className="shrink-0 border-b border-border bg-muted/20 p-3">
        <WorkStatePanel session={session} onOpenWorktree={onOpenWorktree} />
      </div>
      <nav aria-label="Session detail" className="flex shrink-0 items-stretch border-b border-border bg-card px-2">
        {tabs.map(({ key, label, icon: Icon, disabled }) => (
          <Button
            key={key}
            variant="ghost"
            type="button"
            disabled={disabled}
            aria-pressed={active === key}
            onClick={() => setTab(key)}
            className={cn(
              'h-9 gap-1.5 rounded-none border-b-2 px-2.5 text-[11px]',
              active === key ? 'border-primary text-foreground' : 'border-transparent text-muted-foreground hover:text-foreground',
            )}
          >
            <Icon className="size-3.5" />
            {label}
          </Button>
        ))}
      </nav>
      <div className="flex min-h-0 flex-1 flex-col overflow-auto">
        {active === 'transcript' ? (
          <SessionInspector
            key={session.id}
            src={captainSessionUrl(session.id)}
            follow={session.processActive}
            layout="compact"
            className="min-h-0 flex-1 text-xs"
          />
        ) : (
          <BranchChanges project={session.project!} branch={branch!} />
        )}
      </div>
    </div>
  );
}
