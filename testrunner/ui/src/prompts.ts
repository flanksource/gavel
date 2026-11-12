import { resolveLookupScope, type FieldOption, type LookupFetcher } from '@flanksource/clicky-ui/components';
import { lookupNodePresentation, type OperationLookupNode } from '@flanksource/clicky-ui/rpc';
import { apiUrl } from './config';

/** The prompt API's base under a run's server: GET {base}/prompts, POST {base}/prompts/{id}/answer. */
export function promptsBase(scope: { runId?: string } = {}): string {
  return apiUrl('/api', scope);
}

/**
 * The connection API's base under a run's server, the forms' `connectionApiBase`:
 * a picker's inline "New connection" form reads its schema from and saves to
 * {base}/connection.
 */
export function connectionsBase(scope: { runId?: string } = {}): string {
  return apiUrl('/api/v1', scope);
}

interface LookupResponse {
  filters?: Record<string, { options?: Record<string, OperationLookupNode> }>;
}

/**
 * Resolves a form's `x-clicky-lookup` fields against the run's server: a
 * fixture `format: connection` picker and an `x-query` field alike. It
 * requests the descriptor's URL under the page's base, keeping its own query
 * (`?types=`), adding the scope the field's linked param holds (`connection`),
 * the filter and the search text, and keeps each option's icon, description,
 * group and badges.
 */
export function connectionLookupFetcher(scope: { runId?: string } = {}): LookupFetcher {
  return async ({ descriptor, query, rootValue }): Promise<FieldOption[]> => {
    if (!descriptor.url) throw new Error(`lookup ${descriptor.filter} has no url`);
    const url = new URL(apiUrl(descriptor.url, scope), window.location.origin);
    for (const [param, value] of Object.entries(resolveLookupScope(descriptor, rootValue))) url.searchParams.set(param, value);
    url.searchParams.set('__lookup_filter', descriptor.filter);
    url.searchParams.set(descriptor.searchParam ?? '__lookup_q', query);
    const res = await fetch(`${url.pathname}${url.search}`, { headers: { Accept: 'application/json' } });
    if (!res.ok) throw new Error(`lookup ${descriptor.filter}: ${res.status} ${(await res.text()).trim()}`);
    const filter = ((await res.json()) as LookupResponse).filters?.[descriptor.filter];
    if (!filter) throw new Error(`lookup ${descriptor.filter}: the response has no ${descriptor.filter} filter`);
    return Object.entries(filter.options ?? {}).map(([value, node]) => ({
      value,
      label: node.plain || node.text || value,
      ...lookupNodePresentation(value, node),
    }));
  };
}
