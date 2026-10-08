import { useMemo } from 'react';
import type { TodoEvent, TodoSessionAttempt } from '../../types';
import { unresolvedLineCommentCount } from './lineComments';
import { latestRunWorkspace } from './runWorkspace';
import { useTodoSessionDetail } from './TodoSessionDetail';

// RUN_BRANCH_POLL_MS matches TodoRunBranch, so both share one attempts poll.
const RUN_BRANCH_POLL_MS = 15_000;

/** The branch a new run may continue on instead of starting a fresh one. */
export interface RunBranchChoice {
  branch: string;
  /** Line comments left on the branch's diffs that nobody resolved yet. */
  unresolved: number;
}

/**
 * The branch the latest run left behind, when it still exists: the newest run
 * attempt that recorded a worktree with a branch teardown did not delete.
 */
export function reusableRunBranch(attempts: TodoSessionAttempt[], events: TodoEvent[]): RunBranchChoice | null {
  const worktree = latestRunWorkspace(attempts)?.workspace.worktree;
  if (!worktree?.branch || worktree.branchDeleted) return null;
  return { branch: worktree.branch, unresolved: unresolvedLineCommentCount(events) };
}

export function useReusableRunBranch(dir: string, todoRef: string, events: TodoEvent[]): RunBranchChoice | null {
  const { detail } = useTodoSessionDetail(dir, todoRef, !!todoRef, { intervalMs: RUN_BRANCH_POLL_MS });
  return useMemo(() => (detail ? reusableRunBranch(detail.attempts, events) : null), [detail, events]);
}

/** "Continue on shell/1 · 2 unresolved comments" */
export function continueBranchLabel({ branch, unresolved }: RunBranchChoice): string {
  const comments = unresolved === 0 ? '' : ` · ${unresolved} unresolved ${unresolved === 1 ? 'comment' : 'comments'}`;
  return `Continue on ${branch}${comments}`;
}

export const NEW_BRANCH_LABEL = 'New branch';
