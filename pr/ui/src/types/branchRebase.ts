export type RebaseStrategy = 'rebase' | 'merge';

// RebaseStart is how a new operation brings the branch up to date: rebasing
// onto the target, or merging the target in.
export interface RebaseStart {
  target: string;
  strategy: RebaseStrategy;
  rebaseMerges?: boolean;
  noFastForward?: boolean;
}

export interface RebaseOperation {
  id: string;
  worktree: string;
  branch: string;
  strategy: RebaseStrategy;
  rebaseMerges?: boolean;
  noFastForward?: boolean;
  target: string;
  targetSha: string;
  originalHead: string;
  stashSha: string;
  phase: string;
  error: string;
  conflicts: string[];
  aborted: boolean;
}

export interface RebaseContext {
  executable: string;
  branch: string;
  head: string;
  dirtyCount: number;
  targets: { ref: string; label: string }[];
  operation: RebaseOperation | null;
}
