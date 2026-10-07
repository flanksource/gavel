// Wire shapes of the project git endpoints (pr/ui/project_git.go).

// GET /api/projects/git-summary — one entry per project. adds/dels cover all
// work not yet in the base branch (uncommitted changes plus unmerged commits).
export interface ProjectGitSummary {
  name: string;
  base: string;
  adds: number;
  dels: number;
  worktrees: number;
  branches: number;
  // Set when this project's git state could not be read; the other fields are
  // then not meaningful and must not be shown as zero work.
  error?: string;
}

export interface GitWorktreeChanges {
  staged: number;
  unstaged: number;
  both: number;
  untracked: number;
  conflict: number;
  adds: number;
  dels: number;
}

export interface GitWorktree {
  path: string;
  branch: string;
  head: string;
  primary: boolean;
  detached: boolean;
  prunable: boolean;
  changes: GitWorktreeChanges;
  // Commits on the worktree's branch that are not in the base branch.
  ahead: number;
  // Committer date of head (RFC 3339); Go's zero time for a prunable worktree.
  lastCommitAt: string;
  // Newest mtime among the uncommitted files; absent when there are none.
  touchedAt?: string;
}

export interface GitBranchDiff {
  commits: number;
  files: number;
  adds: number;
  dels: number;
}

export interface GitBranchInfo {
  name: string;
  head: string;
  ahead: number;
  behind: number;
  // Path of the worktree this branch is checked out in, "" when it has none.
  worktree: string;
  diff: GitBranchDiff;
  // Committer date of head (RFC 3339).
  lastCommitAt: string;
}

// GET /api/projects/{name}/git
export interface ProjectGit {
  base: string;
  currentBranch: string;
  // Whether the primary checkout is on the base branch; merging needs it.
  baseCheckedOut: boolean;
  worktrees: GitWorktree[];
  branches: GitBranchInfo[];
  // When the server computed this state (RFC 3339); it is memoized, so this
  // can be seconds older than the response.
  computedAt: string;
}

export type BranchMergeMode = 'squash' | 'incremental';

export interface BranchMergeRequest {
  branch: string;
  mode: BranchMergeMode;
  message?: string;
}

// POST /api/projects/{name}/branch/merge
export interface BranchMergeResult {
  targetBranch: string;
  landedSha: string;
  mode: BranchMergeMode;
  commits: number;
  worktreeRemoved: boolean;
  branchDeleted: boolean;
}

export interface BranchPRRequest {
  branch: string;
  draft: boolean;
}

// POST /api/projects/{name}/branch/pr
export interface BranchPRResult {
  number: number;
  url: string;
  topicBranch: string;
}
