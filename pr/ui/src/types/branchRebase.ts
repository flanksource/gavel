export interface RebaseOperation {
  id: string;
  worktree: string;
  branch: string;
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
