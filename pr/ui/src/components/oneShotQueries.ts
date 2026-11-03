import { queryOptions } from '@tanstack/react-query';
import { familiesFromRuntimeCatalog, type PromptSpecDetail, type RuntimeCatalogFamily } from '@flanksource/clicky-ui/ai';
import type { ChatModel } from '@flanksource/clicky-ui/chat';
import type { GitDiffPayload } from '@flanksource/clicky-ui/data';
import type { JsonSchemaObject } from '@flanksource/clicky-ui/components';
import type { Org } from '../types';
import { queryKeys } from '../query';

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

export async function fetchJSON<T>(url: string, signal: AbortSignal, context: string): Promise<T> {
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

// MeResponse is /api/me: the caller (null when nobody is logged in locally), their
// current organization, and every organization they belong to.
export interface MeOrg {
  id: string;
  slug: string;
  name: string;
  role: string;
}

export interface MeResponse {
  user: { id: string; name: string; email: string; githubLogin: string } | null;
  org: MeOrg;
  orgs: MeOrg[];
}

export function meQuery() {
  return queryOptions({
    queryKey: queryKeys.me(),
    queryFn: ({ signal }) => fetchJSON<MeResponse>('/api/me', signal, 'Failed to load the signed-in user'),
    staleTime: oneShotStaleTime,
  });
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
    queryKey: ['projects', projectName, 'actions', 'schema', action, 2] as const,
    queryFn: async ({ signal }) => {
      if (!action) throw new Error(`Cannot load action options for ${projectName} without an action`);
      const url = `/api/projects/${encodeURIComponent(projectName)}/actions/schema?action=${action}`;
      const definition = await fetchJSON<ProjectActionSchemaResponse>(url, signal, `Failed to load ${action} options for ${projectName}`);
      if (definition.schemaVersion !== 2) throw new Error(`Unsupported ${action} options schema version ${definition.schemaVersion}; reload the updated Gavel UI and server`);
      return definition;
    },
    staleTime: oneShotStaleTime,
  });
}

export function chatModelsQuery() {
  return queryOptions({
    queryKey: ['chat', 'models'] as const,
    queryFn: async ({ signal }) => {
      const models = await fetchJSON<ChatModel[]>('/api/chat/models', signal, 'Failed to load AI models');
      if (!Array.isArray(models) || !models.length) throw new Error('AI model catalog is empty or invalid');
      return models;
    },
    staleTime: oneShotStaleTime,
  });
}

export function chatRuntimesQuery() {
  return queryOptions({
    queryKey: ['chat', 'runtimes'] as const,
    queryFn: async ({ signal }) => {
      const catalog = await fetchJSON<RuntimeCatalogFamily[]>('/api/chat/runtimes', signal, 'Failed to load AI runtimes');
      if (!Array.isArray(catalog) || !catalog.length || catalog.some(family => !family?.modes?.length)) {
        throw new Error('AI runtime catalog is empty or invalid');
      }
      return familiesFromRuntimeCatalog(catalog);
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
