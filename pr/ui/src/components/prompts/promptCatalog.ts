import { queryOptions, type QueryClient } from '@tanstack/react-query';
import type {
  PromptCatalogEntry,
  PromptCatalogLayer,
  PromptPageAdapter,
  PromptRenderInput,
  PromptRenderResult,
  PromptSpecDetail,
  PromptSpecSavePayload,
} from '@flanksource/clicky-ui/ai';
import { fetchJSON, promptDetailQuery, responseError } from '../oneShotQueries';
import type { PromptDescriptor } from '../settings/schema';

const catalogStaleTime = 60 * 1000;

// scopeQueryFor maps the prompts tab's scope project to the settings query the
// prompt endpoints take: the global chain, or a registered project's directory.
export function scopeQueryFor(project: string): string {
  return project ? `project=${encodeURIComponent(project)}` : 'scope=global';
}

export function promptCatalogQuery(scopeQuery: string) {
  return queryOptions({
    queryKey: ['settings', 'prompt-catalog', scopeQuery] as const,
    queryFn: ({ signal }) => fetchJSON<PromptCatalogEntry[]>(
      `/api/settings/prompts/catalog?${scopeQuery}`,
      signal,
      'Failed to load prompt catalog',
    ),
    staleTime: catalogStaleTime,
  });
}

// withDefaults attaches each registered prompt's built-in document so the page
// can diff an override against it; named todos prompts have no default.
export function withDefaults(
  entries: PromptCatalogEntry[],
  descriptors: Record<string, PromptDescriptor> | undefined,
): PromptCatalogEntry[] {
  if (!descriptors) return entries;
  return entries.map(entry => {
    const descriptor = descriptors[entry.id];
    return descriptor?.default ? { ...entry, defaultRaw: descriptor.default } : entry;
  });
}

async function sendJSON<T>(method: 'PUT' | 'POST', url: string, body: unknown, context: string): Promise<T> {
  let response: Response;
  try {
    response = await fetch(url, { method, headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
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

function layerScope(layer: PromptCatalogLayer): string {
  if (!layer.scope) throw new Error(`${layer.path} is not editable from the dashboard`);
  return layer.scope;
}

// invalidateAfterPromptSave drops every settings view a prompt write changes:
// the catalog for every scope (a home-layer edit changes every project's
// effective prompt), the layer's config and provenance, and the detail itself.
async function invalidateAfterPromptSave(client: QueryClient, id: string, scopeQuery: string, saved: PromptSpecDetail) {
  client.setQueryData(promptDetailQuery(id, scopeQuery).queryKey, saved);
  await Promise.all([
    client.invalidateQueries({ queryKey: ['settings', 'prompt-catalog'] }),
    client.invalidateQueries({ queryKey: ['settings', 'gavel', scopeQuery], exact: true }),
    client.invalidateQueries({ queryKey: ['settings', 'gavel-trace'] }),
    client.invalidateQueries({
      predicate: ({ queryKey }) => queryKey[0] === 'settings' && queryKey[1] === 'prompt-detail' && queryKey[2] === id,
    }),
  ]);
}

// promptPageAdapter binds the shared PromptPage to Gavel's settings prompt
// endpoints: each layer loads and saves through its own scope query, and
// preview renders against the scope the table was opened with.
export function promptPageAdapter(client: QueryClient, scopeQuery: string): PromptPageAdapter {
  return {
    loadDetail: (entry, layer) => client.fetchQuery(promptDetailQuery(entry.id, layerScope(layer))),
    saveDetail: async (entry, layer, payload: PromptSpecSavePayload) => {
      const scope = layerScope(layer);
      const saved = await sendJSON<PromptSpecDetail>(
        'PUT',
        `/api/settings/prompts/${encodeURIComponent(entry.id)}?${scope}`,
        payload,
        `Failed to save prompt ${entry.id}`,
      );
      await invalidateAfterPromptSave(client, entry.id, scope, saved);
      return saved;
    },
    render: (entry, input: PromptRenderInput) => sendJSON<PromptRenderResult>(
      'POST',
      `/api/settings/prompts/${encodeURIComponent(entry.id)}/render?${scopeQuery}`,
      input,
      `Failed to render prompt ${entry.id}`,
    ),
  };
}
