import { useMutation, useQueryClient, type QueryClient } from '@tanstack/react-query';
import type { PromptSpecDetail, PromptSpecSavePayload } from '@flanksource/clicky-ui/ai';
import { promptDetailQuery, responseError } from '../oneShotQueries';
import { settingsConfigQuery, settingsTraceQuery, type SettingsConfigResponse } from './queries';

async function putJSON<T>(url: string, body: unknown, context: string): Promise<T> {
  let response: Response;
  try {
    response = await fetch(url, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
  } catch (cause) {
    throw new Error(`${context}: ${cause instanceof Error ? cause.message : 'request failed'}`, { cause });
  }
  if (!response.ok) throw await responseError(response, context);
  try {
    return await response.json() as T;
  } catch (cause) {
    throw new Error(`${context}: invalid JSON response`, { cause });
  }
}

function invalidateTrace(client: QueryClient, scopeQuery: string) {
  if (scopeQuery === 'scope=global') {
    return client.invalidateQueries({ queryKey: ['settings', 'gavel-trace'] });
  }
  return client.invalidateQueries({ queryKey: settingsTraceQuery(scopeQuery).queryKey, exact: true });
}

function invalidatePromptDetails(client: QueryClient, scopeQuery: string) {
  return client.invalidateQueries({
    predicate: ({ queryKey }) => (
      queryKey[0] === 'settings'
      && queryKey[1] === 'prompt-detail'
      && queryKey[3] === scopeQuery
    ),
  });
}

export function useSaveSettingsMutation(scopeQuery: string) {
  const client = useQueryClient();
  const scopeLabel = scopeQuery.startsWith('project=') ? 'project settings' : 'global settings';
  return useMutation({
    mutationKey: ['settings', 'gavel', 'save', scopeQuery],
    mutationFn: (config: Record<string, unknown>) => putJSON<SettingsConfigResponse>(
      `/api/settings/gavel?${scopeQuery}`,
      config,
      `Failed to save ${scopeLabel}`,
    ),
    onSuccess: async (saved) => {
      client.setQueryData(settingsConfigQuery(scopeQuery).queryKey, saved);
      await Promise.all([
        invalidateTrace(client, scopeQuery),
        invalidatePromptDetails(client, scopeQuery),
      ]);
    },
  });
}

export function useSavePromptMutation(id: string, scopeQuery: string) {
  const client = useQueryClient();
  return useMutation({
    mutationKey: ['settings', 'prompt-detail', 'save', id, scopeQuery],
    mutationFn: (payload: PromptSpecSavePayload) => putJSON<PromptSpecDetail>(
      `/api/settings/prompts/${encodeURIComponent(id)}?${scopeQuery}`,
      payload,
      `Failed to save prompt ${id}`,
    ),
    onSuccess: async (saved) => {
      client.setQueryData(promptDetailQuery(id, scopeQuery).queryKey, saved);
      await Promise.all([
        client.invalidateQueries({ queryKey: settingsConfigQuery(scopeQuery).queryKey, exact: true }),
        invalidateTrace(client, scopeQuery),
      ]);
    },
  });
}
