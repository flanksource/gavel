// BranchDeletePlan is GET /api/projects/{name}/branch/delete: what deleting a
// branch would remove.
export interface BranchDeletePlan {
  branch: string;
  head: string;
  base: string;
  // The branch's linked worktree, removed with it.
  worktree?: string;
  uncommitted: string[];
  unmerged: { sha: string; subject: string }[];
}

export interface BranchDeleteRequest {
  branch: string;
  expectedHead: string;
  discardUncommitted?: boolean;
  deleteUnmerged?: boolean;
}

export interface BranchDeleteResult {
  worktreeRemoved: boolean;
  branchDeleted: boolean;
}
