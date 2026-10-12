import { useMutation, useQueryClient } from '@tanstack/react-query';
import { HttpError, mutationJSON, queryKeys } from '../query';
import type { BranchMergeRequest, BranchMergeResult, BranchPRRequest, BranchPRResult } from '../types';
import { projectDiffQueryKey } from './projectMutations';

export type BranchLandRequest =
  | ({ via: 'merge' } & BranchMergeRequest)
  | ({ via: 'pr' } & BranchPRRequest);

export type BranchLandResult =
  | ({ via: 'merge' } & BranchMergeResult)
  | ({ via: 'pr' } & BranchPRResult);

/**
 * Lands a project branch: `POST /api/projects/{name}/branch/merge` squashes or
 * cherry-picks it onto the base branch, `.../branch/pr` opens a pull request
 * for it. A 409 (conflict, dirty worktree, …) rejects with the server's reason
 * as an HttpError whose body may list the conflicting paths. A merge removes
 * the worktree and branch, so the summary, git state, status and diffs are
 * refreshed on every outcome, never only on success.
 *
 * onMerged runs as a hook-level callback, before the refresh: the refresh drops
 * the merged branch, which unmounts the caller, and React Query skips
 * per-mutate callbacks of an unmounted observer.
 */
export function useBranchLandMutation({ projectName, onMerged }: {
  projectName: string;
  onMerged?: (result: BranchMergeResult) => void;
}) {
  const client = useQueryClient();
  const base = `/api/projects/${encodeURIComponent(projectName)}/branch`;
  return useMutation({
    mutationKey: ['projects', projectName, 'branch', 'land'],
    mutationFn: async (request: BranchLandRequest): Promise<BranchLandResult> => {
      if (request.via === 'merge') {
        const { via, ...body } = request;
        return { via, ...await mutationJSON<BranchMergeResult>({
          url: `${base}/merge`,
          method: 'POST',
          body,
          context: `Could not merge ${request.branch} in ${projectName}`,
        }) };
      }
      const { via, ...body } = request;
      return { via, ...await mutationJSON<BranchPRResult>({
        url: `${base}/pr`,
        method: 'POST',
        body,
        context: `Could not open a PR for ${request.branch} in ${projectName}`,
      }) };
    },
    onSuccess: result => {
      if (result.via !== 'merge') return;
      const { via: _via, ...merged } = result;
      onMerged?.(merged);
    },
    onSettled: () => Promise.all([
      client.invalidateQueries({ queryKey: queryKeys.projectGitSummary() }),
      client.invalidateQueries({ queryKey: queryKeys.projectGit(projectName) }),
      client.invalidateQueries({ queryKey: queryKeys.projectStatusScope(projectName) }),
      client.invalidateQueries({ queryKey: projectDiffQueryKey(projectName) }),
    ]),
  });
}

/** The conflicting paths a rejected land reported, empty for any other failure. */
export function landConflicts(error: unknown): string[] {
  if (!(error instanceof HttpError) || typeof error.body !== 'object' || error.body === null) return [];
  const { conflicts } = error.body as { conflicts?: unknown };
  return Array.isArray(conflicts) ? conflicts.filter((path): path is string => typeof path === 'string') : [];
}
