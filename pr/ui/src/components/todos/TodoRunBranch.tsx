import { useState, type ReactNode } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { Badge } from '@flanksource/clicky-ui/data';
import { UiChevronDown, UiChevronRight, UiGitBranch, UiGitCommit, UiWarningTriangle } from '@flanksource/clicky-ui/icons';
import type { TodoRunCommit } from '../../types';
import { CommitFiles } from './TodoCommitFiles';
import { useTodoSessionDetail } from './TodoSessionDetail';
import { commitSubject, latestRunWorkspace, shortSha, type LatestRunWorkspace } from './runWorkspace';

// RUN_BRANCH_POLL_MS matches the detail pane's background attempts poll, so the
// branch panel shares that one query rather than adding a faster one.
const RUN_BRANCH_POLL_MS = 15_000;

// CommitRow renders one of the run's commits with an expand toggle that reveals
// its per-file repomap status (each file revealing its own diff on hover).
function CommitRow({ dir, commit }: { dir: string; commit: TodoRunCommit }) {
  const [open, setOpen] = useState(false);
  const ChevronIcon = open ? UiChevronDown : UiChevronRight;
  const subject = commitSubject(commit.message) || shortSha(commit.sha);
  return (
    <li>
      <div className="flex items-start gap-2 px-3 py-2.5 hover:bg-muted/30">
        <Button
          variant="ghost"
          size="icon"
          type="button"
          onClick={() => setOpen(o => !o)}
          aria-expanded={open}
          title={open ? 'Hide files' : 'Show files'}
          className="mt-0.5 inline-flex h-4 w-4 shrink-0 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <ChevronIcon className="text-xs" />
        </Button>
        <span className="mt-0.5 inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full border border-border bg-muted/30 text-muted-foreground">
          <UiGitCommit className="text-xs" />
        </span>
        <div className="min-w-0 flex-1">
          <Button
            variant="ghost"
            type="button"
            onClick={() => setOpen(o => !o)}
            className="block h-auto w-full truncate p-0 text-left text-sm text-foreground hover:underline"
            title={subject}
          >
            {subject}
          </Button>
          <div className="mt-0.5 font-mono text-[11px] text-muted-foreground" title={commit.sha}>
            {shortSha(commit.sha)}
          </div>
        </div>
      </div>
      {open && <CommitFiles dir={dir} hash={commit.sha} />}
    </li>
  );
}

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
      {commits.length > 0 && (
        <ul className="divide-y divide-border">
          {commits.map(commit => (
            <CommitRow key={commit.sha} dir={dir} commit={commit} />
          ))}
        </ul>
      )}
    </section>
  );
}
