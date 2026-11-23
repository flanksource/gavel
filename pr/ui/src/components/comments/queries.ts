import { useEffect, useMemo, useState } from 'react';
import { queryOptions, useQueries, useQueryClient } from '@tanstack/react-query';
import { createGavelCommentsClient, projectScope } from '@flanksource/clicky-ui/comments/gavel';
import { openEventStream } from '../../eventHub';
import { dashboardFetch, fetchJSON } from '../../query';
import type { Project } from '../../types';

export const commentListKey = (scope: string) => ['comments', 'list', scope] as const;
const eventsPath = '/api/v1/comments/events';

export function commentClient(scope: string, signal?: AbortSignal) {
  return createGavelCommentsClient({ baseUrl: '', scope, fetch: (url, init) => dashboardFetch(url, signal ? { ...init, signal } : init) });
}

export function capabilitiesQuery(id: string, version: number) {
  return queryOptions({
    queryKey: ['comments', 'capabilities', id, version] as const,
    queryFn: async ({ signal }) => {
      const data = await fetchJSON<{ actions: string[] }>({ url: `/api/v1/comments/${encodeURIComponent(id)}/capabilities`, signal, context: 'Load comment actions' });
      if (!Array.isArray(data?.actions)) throw new Error('The comment response omitted its actions');
      return data.actions;
    },
  });
}

export function commentSessionsQuery(scope: string) {
  return queryOptions({
    queryKey: ['comments', 'agent-sessions', scope] as const,
    queryFn: ({ signal }) => commentClient(scope, signal).agentSessions(),
    staleTime: 15_000,
  });
}

export function useCommentListings(projects: Project[], enabled: boolean, visible: boolean) {
  const query = useQueryClient();
  const [eventErrors, setEventErrors] = useState<Record<string, string>>({});
  const sources = useMemo(() => projects.map(({ name }) => ({ scope: projectScope(name), project: name })), [projects]);
  const results = useQueries({ queries: sources.map(source => ({
    queryKey: commentListKey(source.scope),
    queryFn: async ({ signal }) => ({ ...source, threads: await commentClient(source.scope, signal).list() }),
    enabled,
    staleTime: 15_000,
  })) });

  useEffect(() => {
    if (!enabled || !visible) return;
    const cursors = new Map<string, number>();
    const stops = sources.map(({ scope }) => {
      const stream = openEventStream(`${eventsPath}?${new URLSearchParams({ scope })}`);
      let timer: ReturnType<typeof setTimeout> | undefined;
      const invalidate = () => {
        clearTimeout(timer);
        timer = setTimeout(() => void query.invalidateQueries({ queryKey: commentListKey(scope) }), 200);
      };
      const onError = (event: Event) => setEventErrors(previous => ({ ...previous, [scope]: event instanceof ErrorEvent ? event.message : 'Event connection lost' }));
      stream.addEventListener('error', onError);
      stream.addEventListener('open', () => {
        setEventErrors(previous => ({ ...previous, [scope]: '' }));
        invalidate();
      });
      stream.addEventListener('comment', (event) => {
        const sequence = Number((event as MessageEvent).lastEventId);
        if (sequence > (cursors.get(scope) ?? 0)) {
          cursors.set(scope, sequence);
          invalidate();
        }
      });
      return () => { clearTimeout(timer); stream.close(); };
    });
    return () => stops.forEach(stop => stop());
  }, [enabled, query, sources, visible]);

  const listings = results.flatMap(result => result.data ? [result.data] : []);
  const errors = results.flatMap((result, index) => {
    const source = sources[index];
    const message = result.error?.message || (source && eventErrors[source.scope]);
    return source && message ? [`${source.project}: ${message}`] : [];
  });
  return { listings: enabled && (listings.length > 0 || sources.length === 0) ? listings : undefined, errors };
}
