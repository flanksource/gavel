// Wire shapes of /api/projects/{name}/checkout/switch (pr/ui/project_checkout_switch.go).

export type CheckoutSync = 'none' | 'fast-forward' | 'reset';

export interface CheckoutSwitchCommit {
  sha: string;
  subject: string;
}

// GET — read from the last-fetched upstream; the POST fetches before syncing.
export interface CheckoutSwitchPlan {
  repo: string;
  // Branch the primary checkout is on; "" when detached.
  current: string;
  head: string;
  branch: string;
  branchHead: string;
  // Absent when the branch tracks no upstream; upstreamHead is absent when it is gone.
  upstream?: string;
  upstreamHead?: string;
  ahead: number;
  behind: number;
  // Branch commits no other ref holds and the upstream has not landed: a reset orphans these.
  unpreserved: CheckoutSwitchCommit[];
  dirtyCount: number;
  checkedOutIn?: string;
}

export interface CheckoutSwitchRequest {
  branch: string;
  expectedHead: string;
  sync: CheckoutSync;
}

// POST — the previous* fields say where to go back to.
export interface CheckoutSwitchResult {
  branch: string;
  head: string;
  sync: CheckoutSync;
  previousBranch: string;
  previousHead: string;
  previousBranchHead: string;
}
