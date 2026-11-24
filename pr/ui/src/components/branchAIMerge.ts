import { useRef } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useProjectRebaseContext } from '../projectGitQueries';
import { queryKeys } from '../query';
import type { BranchMergeMode, BranchMergeResult } from '../types';
import { projectDiffQueryKey } from './projectMutations';
import { usePromptRun } from './promptRun';

/** The runtime-discovered prompt (prompts/library/branch-merge.prompt) that rebases, resolves and lands. */
export const BRANCH_MERGE_PROMPT = 'branch-merge';

export interface BranchAIMergeInput {
  /** The primary checkout the branch lands into. */
  repo: string;
  /** The linked worktree the branch is checked out in; the agent works there. */
  worktree: string;
  branch: string;
  base: string;
  mode: BranchMergeMode;
  /** Optional squash commit message. */
  message: string;
  /** Paths the failed land conflicted on, listed for the agent. */
  conflicts: string[];
  failure?: string;
  /** Commits ahead of base when the run started, reported once it lands. */
  ahead: number;
}

export function branchAIMergeVariables(input: BranchAIMergeInput, gavelExecutable: string) {
  return {
    repo: input.repo,
    worktree: input.worktree,
    branch: input.branch,
    gavelExecutable,
    base: input.base,
    mode: input.mode,
    message: input.message.trim(),
    conflicts: input.conflicts,
    ...(input.failure && { failure: input.failure }),
  };
}

/**
 * Runs the branch-merge prompt. Its own post-verify steps land the branch and
 * remove its worktree, so success reports a merge.
 */
export function useBranchAIMerge({ projectName, worktree, enabled, onMerged }: {
  projectName: string;
  worktree?: string;
  enabled: boolean;
  onMerged?: (result: BranchMergeResult) => void;
}) {
  const client = useQueryClient();
  const context = useProjectRebaseContext({ projectName, worktree, enabled });
  const executable = context.data?.executable;
  const unavailableReason = executable ? '' : context.error?.message || (context.isPending ? 'Loading the merge prompt…' : 'The server did not report its Gavel executable.');
  const input = useRef<BranchAIMergeInput | null>(null);
  const run = usePromptRun({
    prompt: BRANCH_MERGE_PROMPT,
    label: 'merge run',
    // The prompt's after steps either all ran or failed the run, so a success
    // means the branch landed and its worktree and branch are gone.
    onSuccess: () => {
      const landing = input.current;
      if (!landing) return;
      onMerged?.({
        targetBranch: landing.base, landedSha: '', mode: landing.mode, commits: landing.ahead,
        worktreeRemoved: true, branchDeleted: true,
      });
    },
    onSettled: () => void Promise.all([
      client.invalidateQueries({ queryKey: queryKeys.projectGitSummary() }),
      client.invalidateQueries({ queryKey: queryKeys.projectGit(projectName) }),
      client.invalidateQueries({ queryKey: queryKeys.projectStatusScope(projectName) }),
      client.invalidateQueries({ queryKey: projectDiffQueryKey(projectName) }),
    ]),
  });

  const start = (next: BranchAIMergeInput) => {
    if (!executable) return;
    input.current = next;
    return run.start(branchAIMergeVariables(next, executable));
  };

  return { state: run.state, start, stop: run.stop, unavailableReason };
}
