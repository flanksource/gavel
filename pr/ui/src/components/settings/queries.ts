import { queryOptions } from '@tanstack/react-query';
import type { JsonSchemaObject } from '@flanksource/clicky-ui/components';
import type { RunContext } from '../todos/providers';
import type { GavelTrace } from './provenance';
import type { PromptDescriptor } from './schema';

const settingsStaleTime = 5 * 60 * 1000;

export interface SettingsConfigResponse {
  config?: Record<string, unknown>;
  path?: string;
}

async function fetchSettingsJSON<T>(url: string, signal: AbortSignal, context: string): Promise<T> {
  let response: Response;
  try {
    response = await fetch(url, { signal });
  } catch (cause) {
    if (signal.aborted) throw cause;
    throw new Error(`${context}: ${cause instanceof Error ? cause.message : 'request failed'}`, { cause });
  }
  if (!response.ok) {
    const detail = (await response.text().catch(() => '')).trim();
    throw new Error(detail ? `${context}: ${detail}` : `${context} (${response.status})`);
  }
  try {
    return await response.json() as T;
  } catch (cause) {
    throw new Error(`${context}: invalid JSON response`, { cause });
  }
}

export function settingsSchemaQuery() {
  return queryOptions({
    queryKey: ['settings', 'schema'] as const,
    queryFn: ({ signal }) => fetchSettingsJSON<JsonSchemaObject>(
      '/api/settings/schema',
      signal,
      'Failed to load settings schema',
    ),
    staleTime: settingsStaleTime,
  });
}

export function settingsPromptsQuery() {
  return queryOptions({
    queryKey: ['settings', 'prompts'] as const,
    queryFn: async ({ signal }) => {
      const prompts = await fetchSettingsJSON<PromptDescriptor[]>(
        '/api/settings/prompts',
        signal,
        'Failed to load prompt registry',
      );
      return Object.fromEntries((prompts ?? []).map(prompt => [prompt.id, prompt]));
    },
    staleTime: settingsStaleTime,
  });
}

export function settingsRunContextQuery() {
  return queryOptions({
    queryKey: ['todos', 'run-context'] as const,
    queryFn: ({ signal }) => fetchSettingsJSON<RunContext>(
      '/api/todos/run/context',
      signal,
      'Failed to load Captain run providers',
    ),
    staleTime: settingsStaleTime,
  });
}

export function settingsTraceQuery(scopeQuery: string) {
  return queryOptions({
    queryKey: ['settings', 'gavel-trace', scopeQuery] as const,
    queryFn: ({ signal }) => fetchSettingsJSON<GavelTrace>(
      `/api/settings/gavel/trace?${scopeQuery}`,
      signal,
      'Failed to load settings provenance',
    ),
    staleTime: settingsStaleTime,
  });
}

export function settingsConfigQuery(scopeQuery: string) {
  const scopeLabel = scopeQuery.startsWith('project=') ? 'project settings' : 'global settings';
  return queryOptions({
    queryKey: ['settings', 'gavel', scopeQuery] as const,
    queryFn: ({ signal }) => fetchSettingsJSON<SettingsConfigResponse>(
      `/api/settings/gavel?${scopeQuery}`,
      signal,
      `Failed to load ${scopeLabel}`,
    ),
    staleTime: settingsStaleTime,
  });
}
