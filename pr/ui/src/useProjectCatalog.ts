import { useEffect, useMemo, useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import type { ProjectRuns, TestRunsResponse } from './components/tests/types';
import { fetchJSON, queryKeys } from './query';
import type { Project } from './types';
import { useDocumentVisible } from './useDocumentVisible';

export interface ProjectCatalog {
  // Configured projects (/api/projects) merged with history-only ones that only
  // appear in the run history, so a project with runs but no config still lists.
  projects: Project[];
  runs: ProjectRuns[];
  selected: Project | null;
  loading: boolean;
  error: string;
}

const emptyRuns: TestRunsResponse = { projects: [] };

// useProjectCatalog loads the latest test and lint runs and merges them with the
// configured project catalog. It lives above the projects UI because the AppShell
// body sidebar and the detail pane are rendered into different AppShell slots and
// so cannot share component state — one fetch and one SSE stream feed both.
//
// `enabled` keeps that stream off while another tab is showing: the sidebar is
// only mounted on the projects tab, and an idle EventSource per tab would hold a
// connection open for a list nobody is looking at.
export function useProjectCatalog({ configured, selectedName, enabled }: {
  configured: Project[];
  selectedName: string;
  enabled: boolean;
}): ProjectCatalog {
  const [streamError, setStreamError] = useState('');
  const visible = useDocumentVisible();
  const queryClient = useQueryClient();
  const queryEnabled = enabled && visible;
  const runsQuery = useQuery({
    queryKey: queryKeys.testRuns(),
    queryFn: async ({ signal }) => parseLatestRuns(
      await fetchJSON<unknown>({ url: '/api/tests', signal, context: 'Load test and lint runs' }),
    ),
    enabled: queryEnabled,
    staleTime: 30_000,
  });
  const latestRuns = enabled ? (runsQuery.data ?? emptyRuns) : emptyRuns;

  const projects = useMemo(() => {
    const merged = new Map(configured.map(project => [project.name, project]));
    for (const project of latestRuns.projects) {
      if (merged.has(project.name)) continue;
      merged.set(project.name, {
        name: project.name,
        dir: project.dir,
        repos: [...new Set(project.runs.flatMap(run => run.repo ? [run.repo] : []))],
      });
    }
    return [...merged.values()];
  }, [latestRuns.projects, configured]);

  useEffect(() => {
    if (queryEnabled) return;
    void queryClient.cancelQueries({ queryKey: queryKeys.testRuns(), exact: true });
  }, [queryClient, queryEnabled]);

  useEffect(() => {
    if (!queryEnabled) return;
    const apply = (payload: unknown) => {
      const next = parseLatestRuns(payload);
      queryClient.setQueryData(queryKeys.testRuns(), next);
      setStreamError('');
    };

    const stream = new EventSource('/api/tests/stream');
    stream.addEventListener('message', event => {
      try {
        apply(JSON.parse((event as MessageEvent<string>).data));
      } catch {
        setStreamError('Latest test and lint runs received an invalid update.');
      }
    });
    stream.onerror = () => {
      setStreamError(current => current || 'Latest test and lint runs disconnected; reconnecting…');
    };
    return () => stream.close();
  }, [queryClient, queryEnabled]);

  const queryError = runsQuery.failureReason ?? runsQuery.error;

  return {
    projects,
    runs: latestRuns.projects,
    selected: projects.find(project => project.name === selectedName) ?? null,
    loading: enabled && runsQuery.data === undefined && runsQuery.isPending,
    error: enabled ? streamError || (queryError instanceof Error ? queryError.message : '') : '',
  };
}

function parseLatestRuns(payload: unknown): TestRunsResponse {
  if (!payload || typeof payload !== 'object' || !('projects' in payload) || !Array.isArray(payload.projects)) {
    throw new Error('Latest test and lint runs returned an invalid response');
  }
  for (const project of payload.projects) {
    if (!project || typeof project !== 'object' || !('name' in project) || typeof project.name !== 'string' || !('dir' in project) || typeof project.dir !== 'string' || !('runs' in project) || !Array.isArray(project.runs)) {
      throw new Error('Latest test and lint runs returned an invalid project');
    }
    for (const run of project.runs) {
      if (!run || typeof run !== 'object' || !('runId' in run) || typeof run.runId !== 'string' || !('kind' in run) || (run.kind !== 'test' && run.kind !== 'lint' && run.kind !== 'test+lint')) {
        throw new Error('Latest test and lint runs returned an invalid run');
      }
    }
  }
  return payload as TestRunsResponse;
}
