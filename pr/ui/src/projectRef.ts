// A project's detail pane shows one git ref at a time, encoded in the URL as
// ?ref=wt:<absolute worktree path> or ?ref=br:<branch name>. No ref means the
// project's main checkout.
export type ProjectRef =
  | { kind: 'main' }
  | { kind: 'worktree'; path: string }
  | { kind: 'branch'; name: string };

const WORKTREE_PREFIX = 'wt:';
const BRANCH_PREFIX = 'br:';

export const MAIN_REF: ProjectRef = { kind: 'main' };

export function worktreeRef(path: string): string {
  return `${WORKTREE_PREFIX}${path}`;
}

export function branchRef(name: string): string {
  return `${BRANCH_PREFIX}${name}`;
}

// isProjectRefParam reports whether a raw ?ref= value is well formed, so the
// router can drop a malformed one instead of carrying it around.
export function isProjectRefParam(raw: string): boolean {
  return (raw.startsWith(WORKTREE_PREFIX) && raw.length > WORKTREE_PREFIX.length)
    || (raw.startsWith(BRANCH_PREFIX) && raw.length > BRANCH_PREFIX.length);
}

export function parseProjectRef(raw: string): ProjectRef {
  if (!raw) return MAIN_REF;
  if (!isProjectRefParam(raw)) throw new Error(`Invalid project ref ${JSON.stringify(raw)}: expected wt:<path> or br:<branch>`);
  return raw.startsWith(WORKTREE_PREFIX)
    ? { kind: 'worktree', path: raw.slice(WORKTREE_PREFIX.length) }
    : { kind: 'branch', name: raw.slice(BRANCH_PREFIX.length) };
}
