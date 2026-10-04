import type { ReactNode } from 'react';
import { Badge } from '@flanksource/clicky-ui/data';
import { UiGitBranch, UiWarningTriangle } from '@flanksource/clicky-ui/icons';
import { TodoRunChanges } from './TodoRunChanges';
import { useTodoSessionDetail } from './TodoSessionDetail';
import { latestRunWorkspace, shortSha, type LatestRunWorkspace } from './runWorkspace';

// RUN_BRANCH_POLL_MS matches the detail pane's background attempts poll, so the
// branch panel shares that one query rather than adding a faster one.
const RUN_BRANCH_POLL_MS = 15_000;

function KeptWorktreeWarning({ path, reason, dirty }: { path?: string; reason?: string; dirty: number }) {
  return (
    <div role="alert" className="flex items-start gap-2 border-b border-border bg-amber-500/10 px-3 py-2 text-xs text-amber-700 [[data-theme=dark]_&]:text-amber-300">
      <UiWarningTriangle className="mt-0.5 shrink-0" aria-hidden="true" />
      <span className="min-w-0">
        Worktree kept at <span className="break-all font-mono">{path || 'an unrecorded path'}</span>
        {reason ? `: ${reason}` : ''}
        {dirty > 0 ? ` · ${dirty} uncommitted path${dirty === 1 ? '' : 's'}` : ''}
      </span>
    </div>
  );
}

export interface TodoRunBranchProps {
  dir: string;
  todoRef: string;
  /** Actions for the run's branch (e.g. landing it), rendered in the header. */
  renderActions?: (latest: LatestRunWorkspace) => ReactNode;
}

/**
 * The branch the todo's latest run worked on, read from the run's recorded
 * workspace: the branch chip, the agent's own setup…head range, a warning when
 * teardown kept the worktree, and the commits it made. Renders nothing until a
 * run records a worktree or commits.
 */
export function TodoRunBranch({ dir, todoRef, renderActions }: TodoRunBranchProps) {
  const { detail, error } = useTodoSessionDetail(dir, todoRef, !!todoRef, { intervalMs: RUN_BRANCH_POLL_MS });
  const latest = detail ? latestRunWorkspace(detail.attempts) : null;
  if (!latest && !error) return null;
  const worktree = latest?.workspace.worktree;
  const commits = latest?.workspace.commits ?? [];

  return (
    <section className="overflow-hidden rounded-lg border border-border bg-card shadow-sm">
      <div className="flex flex-wrap items-center gap-2 border-b border-border bg-muted/30 px-3 py-2.5">
        <span className="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-md border border-border bg-background text-muted-foreground">
          <UiGitBranch className="text-xs" />
        </span>
        <span className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Run branch</span>
        {worktree?.branch && <Badge variant="outline" size="sm" icon={UiGitBranch} clickToCopy>{worktree.branch}</Badge>}
        {worktree?.setup && worktree.head && (
          <span className="font-mono text-[11px] text-muted-foreground" title={`${worktree.setup}..${worktree.head}`}>
            {`${shortSha(worktree.setup)}…${shortSha(worktree.head)}`}
          </span>
        )}
        {worktree?.branchDeleted && <span className="text-[11px] text-muted-foreground">branch deleted</span>}
        {latest && <span className="text-[11px] text-muted-foreground">Attempt #{latest.attempt.ordinal}</span>}
        <span className="ml-auto flex items-center gap-2">
          {commits.length > 0 && (
            <span className="rounded-full border border-border bg-background px-1.5 py-0.5 text-[11px] tabular-nums text-muted-foreground">{commits.length}</span>
          )}
          {latest && renderActions?.(latest)}
        </span>
      </div>
      {error && <div className="whitespace-pre-wrap px-3 py-2 text-xs text-red-600">{error}</div>}
      {worktree?.kept && <KeptWorktreeWarning path={worktree.path} reason={worktree.keptReason} dirty={worktree.dirty?.length ?? 0} />}
      {(commits.length > 0 || (worktree?.setup && worktree.head)) && (
        <TodoRunChanges dir={dir} worktree={worktree} commits={commits} landing={latest?.attempt.landing} />
      )}
    </section>
  );
}
