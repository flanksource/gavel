import { useCallback, useEffect, useState } from 'react';
import { apiUrl } from '../config';

export type ProcessFacet = 'env' | 'files' | 'sockets' | 'ports';

export interface ProcessFacetState {
  data?: unknown;
  error?: string;
  loading: boolean;
  loadedAt?: Date;
  reload: () => void;
}

// useProcessFacet reads one facet of a live process from the test UI server,
// once per pid and on reload; nothing is fetched while facet is null.
export function useProcessFacet(facet: ProcessFacet | null, pid: number): ProcessFacetState {
  const [state, setState] = useState<Omit<ProcessFacetState, 'reload'>>({ loading: false });
  const [generation, setGeneration] = useState(0);
  const reload = useCallback(() => setGeneration(g => g + 1), []);

  useEffect(() => {
    if (!facet) return;
    const controller = new AbortController();
    setState({ loading: true });
    fetch(apiUrl(`/api/process/${facet}?pid=${pid}`), { signal: controller.signal })
      .then(async res => {
        if (!res.ok) throw new Error((await res.text()).trim() || `HTTP ${res.status}`);
        return res.json() as Promise<unknown>;
      })
      .then(data => setState({ data, loading: false, loadedAt: new Date() }))
      .catch((e: unknown) => {
        if (controller.signal.aborted) return;
        setState({ error: e instanceof Error ? e.message : String(e), loading: false });
      });
    return () => controller.abort();
  }, [facet, pid, generation]);

  return { ...state, reload };
}
