import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { readLocalCache, writeLocalCache } from './localQueryCache';
import { fetchJSON, queryKeys } from './query';
import type { ProjectGit, ProjectGitSummary } from './types';

// The last good git summary seeds the first render of the next page load. Git
// queries never poll: the server pushes each project's generation over
// /api/git/stream (useGitStream), which invalidates them when it changes.
export const PROJECT_GIT_SUMMARY_CACHE_KEY = 'gavel.pr-ui.cache.project-git-summary.v1';

export interface ProjectGitSummaryView {
  byProject: ReadonlyMap<string, ProjectGitSummary>;
  // Non-empty when the request failed and there is no earlier response to show.
  error: string;
  // True until the first response (or cached copy) arrives.
  loading: boolean;
}

export function useProjectGitSummary({ enabled }: { enabled: boolean }): ProjectGitSummaryView {
  const query = useQuery<ProjectGitSummary[]>({
    queryKey: queryKeys.projectGitSummary(),
    queryFn: async ({ signal }) => {
      const summaries = parseProjectGitSummaries(await fetchJSON<unknown>({
        url: '/api/projects/git-summary',
        signal,
        context: 'Load project git summary',
      }));
      writeLocalCache(PROJECT_GIT_SUMMARY_CACHE_KEY, summaries);
      return summaries;
    },
    placeholderData: () => readLocalCache(PROJECT_GIT_SUMMARY_CACHE_KEY, parseProjectGitSummaries),
    enabled,
    // Failing fast shows the error state at once; the next generation change retries.
    retry: false,
    staleTime: Infinity,
  });
  const byProject = useMemo(
    () => new Map((query.data ?? []).map(summary => [summary.name, summary])),
    [query.data],
  );
  return {
    byProject,
    error: query.data === undefined && query.error ? query.error.message : '',
    loading: query.data === undefined && !query.error,
  };
}

// The git stream refetches on every change; a page showing how fresh the state
// is also re-reads it every refetchIntervalMs, since an unchanged scan advances
// no generation.
export function useProjectGit(projectName: string, { refetchIntervalMs }: { refetchIntervalMs?: number } = {}) {
  return useQuery<ProjectGit>({
    refetchInterval: refetchIntervalMs ?? false,
    queryKey: queryKeys.projectGit(projectName),
    queryFn: async ({ signal }) => parseProjectGit(await fetchJSON<unknown>({
      url: `/api/projects/${encodeURIComponent(projectName)}/git`,
      signal,
      context: `Load git state of ${projectName}`,
    })),
    staleTime: Infinity,
  });
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function hasNumbers(value: unknown, keys: readonly string[]): boolean {
  return isRecord(value) && keys.every(key => typeof value[key] === 'number');
}

// The payloads are read field by field by the sidebar and the detail pane, so
// their shape is checked at the boundary: a malformed one fails the query
// loudly instead of rendering NaN or throwing mid-render.
export function parseProjectGitSummaries(payload: unknown): ProjectGitSummary[] {
  if (!Array.isArray(payload)) throw new Error('Load project git summary: invalid response, expected an array');
  for (const entry of payload) {
    if (!isRecord(entry)
      || typeof entry.name !== 'string'
      || typeof entry.base !== 'string'
      || !hasNumbers(entry, ['adds', 'dels', 'worktrees', 'branches'])
      || (entry.error !== undefined && typeof entry.error !== 'string')) {
      throw new Error(`Load project git summary: invalid entry ${JSON.stringify(entry)}`);
    }
  }
  return payload as ProjectGitSummary[];
}

const CHANGE_KEYS = ['staged', 'unstaged', 'both', 'untracked', 'conflict', 'adds', 'dels'] as const;

export function parseProjectGit(payload: unknown): ProjectGit {
  if (!isRecord(payload)
    || typeof payload.base !== 'string'
    || typeof payload.currentBranch !== 'string'
    || typeof payload.baseCheckedOut !== 'boolean'
    || !Array.isArray(payload.worktrees)
    || !Array.isArray(payload.branches)
    || typeof payload.computedAt !== 'string'
    || typeof payload.generation !== 'number'
    || (payload.error !== undefined && typeof payload.error !== 'string')) {
    throw new Error('Load project git: invalid response');
  }
  for (const worktree of payload.worktrees) {
    if (!isRecord(worktree)
      || typeof worktree.path !== 'string'
      || typeof worktree.branch !== 'string'
      || typeof worktree.primary !== 'boolean'
      || typeof worktree.ahead !== 'number'
      || typeof worktree.lastCommitAt !== 'string'
      || (worktree.touchedAt !== undefined && typeof worktree.touchedAt !== 'string')
      || (worktree.statusScannedAt !== undefined && typeof worktree.statusScannedAt !== 'string')
      || (worktree.statusError !== undefined && typeof worktree.statusError !== 'string')
      || !hasNumbers(worktree.changes, CHANGE_KEYS)) {
      throw new Error(`Load project git: invalid worktree ${JSON.stringify(worktree)}`);
    }
  }
  for (const branch of payload.branches) {
    if (!isRecord(branch)
      || typeof branch.name !== 'string'
      || typeof branch.worktree !== 'string'
      || typeof branch.lastCommitAt !== 'string'
      || !hasNumbers(branch, ['ahead', 'behind'])
      || !hasNumbers(branch.diff, ['commits', 'files', 'adds', 'dels'])) {
      throw new Error(`Load project git: invalid branch ${JSON.stringify(branch)}`);
    }
  }
  return payload as unknown as ProjectGit;
}
