import type { FieldOption, LookupFetcher } from '@flanksource/clicky-ui/components';
import { apiUrl } from './config';

/** The prompt API's base under a run's server: GET {base}/prompts, POST {base}/prompts/{id}/answer. */
export function promptsBase(scope: { runId?: string } = {}): string {
  return apiUrl('/api', scope);
}

interface LookupResponse {
  filters?: Record<string, { options?: Record<string, { plain?: string; text?: string }> }>;
}

/**
 * Resolves a form's `x-clicky-lookup` pickers (a fixture `format: connection`
 * field) against the run's server: the descriptor's URL under the page's base,
 * keeping its own query (`?types=`) and adding the search text.
 */
export function connectionLookupFetcher(scope: { runId?: string } = {}): LookupFetcher {
  return async ({ descriptor, query }): Promise<FieldOption[]> => {
    if (!descriptor.url) throw new Error(`lookup ${descriptor.filter} has no url`);
    const url = new URL(apiUrl(descriptor.url, scope), window.location.origin);
    url.searchParams.set(descriptor.searchParam ?? '__lookup_q', query);
    const res = await fetch(`${url.pathname}${url.search}`, { headers: { Accept: 'application/json' } });
    if (!res.ok) throw new Error(`lookup ${descriptor.filter}: ${res.status} ${(await res.text()).trim()}`);
    const filter = ((await res.json()) as LookupResponse).filters?.[descriptor.filter];
    if (!filter) throw new Error(`lookup ${descriptor.filter}: the response has no ${descriptor.filter} filter`);
    return Object.entries(filter.options ?? {}).map(([value, node]) => ({ value, label: node.plain || node.text || value }));
  };
}
