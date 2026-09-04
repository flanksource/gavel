import { useCallback, useEffect, useRef, useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { fetchJSON, queryKeys } from './query';
import type { Project, ProcStatus, SearchConfig, Snapshot } from './types';

const bootstrapStaleTime = 30_000;

type SnapshotUpdater = (current: Snapshot) => Snapshot;

export interface AppQueryState {
  snapshot: Snapshot;
  projects: Project[];
  projectsLoaded: boolean;
  projectError: string;
  procStatus: Record<string, ProcStatus>;
  processError: string;
  updateSnapshot: (updater: SnapshotUpdater) => void;
  refreshProjects: () => Promise<void>;
  refreshProjectsAndProcesses: () => Promise<void>;
}

export function useAppQueries({ enabled, initialConfig }: { enabled: boolean; initialConfig: SearchConfig }): AppQueryState {
  const queryClient = useQueryClient();
  const [processStreamError, setProcessStreamError] = useState('');
  const initialSnapshot = useRef(emptySnapshot(initialConfig)).current;
  const prQuery = useQuery({
    queryKey: queryKeys.prSnapshot(),
    queryFn: async ({ signal }) => mergeSnapshot(
      queryClient.getQueryData<Snapshot>(queryKeys.prSnapshot()) ?? initialSnapshot,
      parseSnapshot(await fetchJSON<unknown>({ url: '/api/prs', signal, context: 'Load pull requests' })),
    ),
    enabled,
    staleTime: bootstrapStaleTime,
  });
  const projectsQuery = useQuery({
    queryKey: queryKeys.projects(),
    queryFn: async ({ signal }) => parseProjects(
      await fetchJSON<unknown>({ url: '/api/projects', signal, context: 'Load projects' }),
    ),
    enabled,
    retry: true,
    retryDelay: attempt => Math.min(1000 * (2 ** attempt), 15_000),
    staleTime: bootstrapStaleTime,
  });
  const procQuery = useQuery({
    queryKey: queryKeys.processStatuses(),
    queryFn: async ({ signal }) => parseProcStatuses(
      await fetchJSON<unknown>({ url: '/api/proc/status', signal, context: 'Load process status' }),
    ),
    enabled,
    staleTime: bootstrapStaleTime,
  });

  useEffect(() => {
    if (enabled) return;
    void queryClient.cancelQueries({ queryKey: queryKeys.prSnapshot(), exact: true });
    void queryClient.cancelQueries({ queryKey: queryKeys.projects(), exact: true });
    void queryClient.cancelQueries({ queryKey: queryKeys.processStatuses(), exact: true });
  }, [enabled, queryClient]);

  useEffect(() => {
    if (!enabled) return;
    const stream = new EventSource('/api/prs/stream');
    stream.addEventListener('message', event => {
      try {
        const incoming = parseSnapshot(JSON.parse((event as MessageEvent<string>).data));
        queryClient.setQueryData<Snapshot>(queryKeys.prSnapshot(), current => mergeSnapshot(current ?? initialSnapshot, incoming));
      } catch {
        queryClient.setQueryData<Snapshot>(queryKeys.prSnapshot(), current => ({
          ...(current ?? initialSnapshot),
          error: 'Pull request stream received an invalid update.',
        }));
      }
    });
    stream.onerror = () => {
      queryClient.setQueryData<Snapshot>(queryKeys.prSnapshot(), current => ({
        ...(current ?? initialSnapshot),
        error: 'Connection lost — retrying...',
      }));
    };
    return () => stream.close();
  }, [enabled, initialSnapshot, queryClient]);

  useEffect(() => {
    if (!enabled) return;
    const stream = new EventSource('/api/proc/status/stream');
    stream.addEventListener('message', event => {
      try {
        queryClient.setQueryData(queryKeys.processStatuses(), parseProcStatuses(JSON.parse((event as MessageEvent<string>).data)));
        setProcessStreamError('');
      } catch {
        setProcessStreamError('Process status stream received an invalid update.');
      }
    });
    stream.onerror = () => setProcessStreamError(current => current || 'Process status stream disconnected; reconnecting…');
    return () => stream.close();
  }, [enabled, queryClient]);

  const updateSnapshot = useCallback((updater: SnapshotUpdater) => {
    queryClient.setQueryData<Snapshot>(queryKeys.prSnapshot(), current => updater(current ?? initialSnapshot));
  }, [initialSnapshot, queryClient]);
  const refreshProjects = useCallback(async () => {
    await queryClient.invalidateQueries({ queryKey: queryKeys.projects(), exact: true });
  }, [queryClient]);
  const refreshProjectsAndProcesses = useCallback(async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: queryKeys.projects(), exact: true }),
      queryClient.invalidateQueries({ queryKey: queryKeys.processStatuses(), exact: true }),
    ]);
  }, [queryClient]);

  const projectFailure = projectsQuery.failureReason ?? projectsQuery.error;
  return {
    snapshot: prQuery.data ?? (prQuery.error ? { ...initialSnapshot, error: errorMessage(prQuery.error) } : initialSnapshot),
    projects: projectsQuery.data ?? [],
    projectsLoaded: projectsQuery.data !== undefined || projectsQuery.failureCount > 0,
    projectError: projectFailure ? errorMessage(projectFailure) : '',
    procStatus: procQuery.data ?? {},
    processError: processStreamError || (procQuery.error ? errorMessage(procQuery.error) : ''),
    updateSnapshot,
    refreshProjects,
    refreshProjectsAndProcesses,
  };
}

function emptySnapshot(config: SearchConfig): Snapshot {
  return {
    prs: [],
    fetchedAt: '',
    nextFetchIn: 60,
    incremental: false,
    paused: false,
    config,
  };
}

function mergeSnapshot(current: Snapshot, incoming: Snapshot): Snapshot {
  const currentFetchedAt = Date.parse(current.fetchedAt);
  const incomingFetchedAt = Date.parse(incoming.fetchedAt);
  if (Number.isFinite(currentFetchedAt) && Number.isFinite(incomingFetchedAt) && currentFetchedAt > incomingFetchedAt) {
    return current;
  }
  return {
    ...current,
    ...incoming,
    prs: incoming.prs,
    viewer: incoming.viewer || current.viewer,
    unread: incoming.unread ?? {},
    config: {
      ...incoming.config,
      org: incoming.config.org || current.config.org,
      all: incoming.config.all || current.config.all,
    },
    rateLimit: incoming.rateLimit ?? current.rateLimit,
    syncStatus: incoming.syncStatus ?? current.syncStatus,
    gavelResults: incoming.gavelResults
      ? { ...current.gavelResults, ...incoming.gavelResults }
      : current.gavelResults,
  };
}

function parseSnapshot(payload: unknown): Snapshot {
  if (!payload || typeof payload !== 'object' || !('prs' in payload) || !Array.isArray(payload.prs) || !('config' in payload) || !payload.config || typeof payload.config !== 'object' || !('fetchedAt' in payload) || typeof payload.fetchedAt !== 'string') {
    throw new Error('Load pull requests: invalid response');
  }
  return payload as Snapshot;
}

// The projects payload is read directly by the sidebar (project.repos[0]) and by
// the process and settings dialogs, so its shape is checked here rather than
// trusted: a malformed entry fails the query loudly instead of throwing mid-render
// and unmounting the app.
function parseProjects(payload: unknown): Project[] {
  if (!Array.isArray(payload)) throw new Error('Load projects: invalid response');
  for (const project of payload) {
    if (!project || typeof project !== 'object'
      || typeof project.name !== 'string'
      || typeof project.dir !== 'string'
      || !Array.isArray(project.repos)) {
      throw new Error('Load projects: invalid project');
    }
  }
  return payload as Project[];
}

function parseProcStatuses(payload: unknown): Record<string, ProcStatus> {
  if (!payload || typeof payload !== 'object' || Array.isArray(payload)) throw new Error('Load process status: invalid response');
  return payload as Record<string, ProcStatus>;
}

function errorMessage(cause: unknown): string {
  return cause instanceof Error ? cause.message : 'Request failed';
}
