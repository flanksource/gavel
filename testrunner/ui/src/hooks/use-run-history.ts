import { useCallback, useEffect, useRef, useState } from 'react';
import { fetchRunsPage, type RunInfo, type RunsPage } from '../runHistory';

export interface RunHistory {
  /** The host serves a run history; false until the first page answers, and for hosts without one. */
  available: boolean;
  /** The run being viewed; undefined for the default (latest) view. */
  current?: string;
  latest: string[];
  latestHref: string;
  /** Every page fetched so far, newest first. */
  runs: RunInfo[];
  hasMore: boolean;
  loading: boolean;
  error?: string;
  loadMore: () => void;
}

/** Loads the run history page by page from `/api/runs` under the page's base path. */
export function useRunHistory(): RunHistory {
  const [pages, setPages] = useState<RunsPage[]>([]);
  const [available, setAvailable] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>();
  const inFlight = useRef(false);

  const load = useCallback(async (before?: string) => {
    if (inFlight.current) return;
    inFlight.current = true;
    setLoading(true);
    try {
      const page = await fetchRunsPage(before ? { before } : {});
      if (page) {
        setAvailable(true);
        setPages(prev => (before ? [...prev, page] : [page]));
      }
      setError(undefined);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      inFlight.current = false;
      setLoading(false);
    }
  }, []);

  useEffect(() => { void load(); }, [load]);

  const first = pages[0];
  const next = pages.at(-1)?.next;
  const loadMore = useCallback(() => { if (next) void load(next); }, [next, load]);

  return {
    available,
    ...(first?.current ? { current: first.current } : {}),
    latest: first?.latest ?? [],
    latestHref: first?.latestHref ?? '/',
    runs: pages.flatMap(page => page.runs),
    hasMore: !!next,
    loading,
    ...(error ? { error } : {}),
    loadMore,
  };
}
