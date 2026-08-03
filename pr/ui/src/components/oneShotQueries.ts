import { queryOptions } from '@tanstack/react-query';
import type { PromptSpecDetail } from '@flanksource/clicky-ui/ai';
import type { GitDiffPayload } from '@flanksource/clicky-ui/data';
import type { JsonSchemaObject } from '@flanksource/clicky-ui/components';
import type { Org } from '../types';

const oneShotStaleTime = 5 * 60 * 1000;

export type ProjectActionName = 'commit' | 'lint' | 'test';

export interface ProjectActionSchemaResponse {
  schemaVersion: number;
  action: ProjectActionName;
  schema: JsonSchemaObject;
  defaults: Record<string, unknown>;
}

export interface JobLogsResponse {
  jobId: number;
  logs?: string;
  steps?: { number: number; logs?: string }[];
  error?: string;
}

type PRDiffAPIResponse = GitDiffPayload & { error?: string };

export async function responseError(response: Response, fallback: string): Promise<Error> {
  const text = await response.text().catch(() => '');
  const trimmed = text.trim();
  if (trimmed) {
    try {
      const parsed = JSON.parse(trimmed) as { error?: unknown };
      if (typeof parsed.error === 'string' && parsed.error.trim()) {
        return new Error(`${fallback}: ${parsed.error.trim()}`);
      }
    } catch {
      return new Error(`${fallback}: ${trimmed}`);
    }
    return new Error(`${fallback}: ${trimmed}`);
  }
  return new Error(`${fallback} (${response.status})`);
}

async function fetchJSON<T>(url: string, signal: AbortSignal, context: string): Promise<T> {
  let response: Response;
  try {
    response = await fetch(url, { signal });
  } catch (cause) {
    if (signal.aborted) throw cause;
    throw new Error(`${context}: ${cause instanceof Error ? cause.message : 'request failed'}`, { cause });
  }
  if (!response.ok) throw await responseError(response, context);
  try {
    return await response.json() as T;
  } catch (cause) {
    throw new Error(`${context}: invalid JSON response`, { cause });
  }
}

export function orgsQuery() {
  return queryOptions({
    queryKey: ['orgs', { includeIgnored: true }] as const,
    queryFn: ({ signal }) => fetchJSON<Org[]>('/api/orgs?include-ignored=1', signal, 'Failed to load organizations'),
    staleTime: oneShotStaleTime,
  });
}

export function projectActionSchemaQuery(projectName: string, action: ProjectActionName | null) {
  return queryOptions({
    queryKey: ['projects', projectName, 'actions', 'schema', action] as const,
    queryFn: ({ signal }) => {
      if (!action) throw new Error(`Cannot load action options for ${projectName} without an action`);
      const url = `/api/projects/${encodeURIComponent(projectName)}/actions/schema?action=${action}`;
      return fetchJSON<ProjectActionSchemaResponse>(url, signal, `Failed to load ${action} options for ${projectName}`);
    },
    staleTime: oneShotStaleTime,
  });
}

export function jobLogsQuery(repo: string, runId: number, jobId: number, tail = 100) {
  return queryOptions({
    queryKey: ['prs', 'job-logs', repo, runId, jobId, tail] as const,
    queryFn: async ({ signal }) => {
      const url = `/api/prs/job-logs?repo=${encodeURIComponent(repo)}&runId=${runId}&jobId=${jobId}&tail=${tail}`;
      const result = await fetchJSON<JobLogsResponse>(url, signal, `Failed to load logs for job ${jobId}`);
      if (result.error) throw new Error(`Failed to load logs for job ${jobId}: ${result.error}`);
      return result;
    },
    staleTime: oneShotStaleTime,
  });
}

export function promptDetailQuery(id: string, scopeQuery: string) {
  return queryOptions({
    queryKey: ['settings', 'prompt-detail', id, scopeQuery] as const,
    queryFn: ({ signal }) => fetchJSON<PromptSpecDetail>(
      `/api/settings/prompts/${encodeURIComponent(id)}?${scopeQuery}`,
      signal,
      `Failed to load prompt ${id}`,
    ),
    staleTime: oneShotStaleTime,
  });
}

export function prCommitDiffQuery(repo: string, sha: string) {
  return diffQuery(
    ['prs', 'commit-diff', repo, sha] as const,
    `/api/prs/commits/diff?repo=${encodeURIComponent(repo)}&sha=${encodeURIComponent(sha)}`,
    `Failed to load diff for commit ${sha}`,
  );
}

export function prFileDiffQuery(repo: string, number: number, path: string) {
  return diffQuery(
    ['prs', 'file-diff', repo, number, path] as const,
    `/api/prs/files/diff?repo=${encodeURIComponent(repo)}&number=${number}&path=${encodeURIComponent(path)}`,
    `Failed to load diff for ${path}`,
  );
}

function diffQuery(queryKey: readonly unknown[], url: string, context: string) {
  return queryOptions({
    queryKey,
    queryFn: async ({ signal }) => {
      const payload = await fetchJSON<PRDiffAPIResponse>(url, signal, context);
      if (payload.error) throw new Error(`${context}: ${payload.error}`);
      return {
        diff: payload.diff || '',
        truncated: !!payload.truncated,
        binary: !!payload.binary,
      } satisfies GitDiffPayload;
    },
    staleTime: oneShotStaleTime,
  });
}
