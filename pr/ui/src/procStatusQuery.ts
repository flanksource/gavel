import { useQuery } from '@tanstack/react-query';
import { fetchJSON, queryKeys } from './query';
import type { ProcStatus } from './types';

const EMPTY_PROC_STATUS: Record<string, ProcStatus> = {};

// The process-status map is fetched once by useAppQueries and then kept fresh
// by /api/proc/status/stream writing straight into this cache entry. Frames
// carry live CPU/memory samples, so they change every few seconds.
export const procStatusQueryOptions = {
  queryKey: queryKeys.processStatuses(),
  queryFn: async ({ signal }: { signal: AbortSignal }) => parseProcStatuses(
    await fetchJSON<unknown>({ url: '/api/proc/status', signal, context: 'Load process status' }),
  ),
};

// useProcStatus subscribes a component to the live process-status map. Read it
// in the component that displays process state rather than passing it down from
// the app root, so each frame re-renders only those components. It never
// fetches: useAppQueries owns loading, and the stream owns freshness.
export function useProcStatus(): Record<string, ProcStatus> {
  return useQuery({ ...procStatusQueryOptions, enabled: false }).data ?? EMPTY_PROC_STATUS;
}

export function parseProcStatuses(payload: unknown): Record<string, ProcStatus> {
  if (!payload || typeof payload !== 'object' || Array.isArray(payload)) throw new Error('Load process status: invalid response');
  return payload as Record<string, ProcStatus>;
}
