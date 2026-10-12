import type { TodoRunWorkspace, TodoSessionAttempt } from '../../types';

export interface LatestRunWorkspace {
  attempt: TodoSessionAttempt;
  workspace: TodoRunWorkspace;
}

export function shortSha(sha: string | undefined) {
  return (sha ?? '').slice(0, 7);
}

export function commitSubject(message: string | undefined) {
  return (message ?? '').trim().split('\n', 1)[0]?.trim() ?? '';
}

/** Whether a run's workspace recorded anything worth showing: a worktree or commits. */
export function hasRunWorkspace(workspace: TodoRunWorkspace | undefined): workspace is TodoRunWorkspace {
  return !!workspace && (!!workspace.worktree || (workspace.commits?.length ?? 0) > 0);
}

/**
 * The newest `run` attempt that recorded a worktree or commits. Attempts arrive
 * newest first; verify and plan steps never carry the run's work.
 */
export function latestRunWorkspace(attempts: TodoSessionAttempt[]): LatestRunWorkspace | null {
  const attempt = attempts.find(candidate => candidate.step === 'run' && hasRunWorkspace(candidate.workspace));
  return attempt?.workspace ? { attempt, workspace: attempt.workspace } : null;
}
