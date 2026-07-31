import { useEffect, useMemo, useState } from 'react';
import type { ProjectRuns, TestRunsResponse } from './components/tests/types';
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
  const [latestRuns, setLatestRuns] = useState<TestRunsResponse>({ projects: [] });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const visible = useDocumentVisible();

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
    if (!enabled || !visible) return;
    let active = true;
    const apply = (payload: unknown) => {
      const next = parseLatestRuns(payload);
      if (!active) return;
      setLatestRuns(next);
      setLoading(false);
      setError('');
    };

    fetch('/api/tests')
      .then(async response => {
        if (!response.ok) throw new Error(`Load test and lint runs failed (HTTP ${response.status})`);
        apply(await response.json());
      })
      .catch(cause => {
        if (!active) return;
        setLoading(false);
        setError(cause instanceof Error ? cause.message : 'Load test and lint runs failed');
      });

    const stream = new EventSource('/api/tests/stream');
    stream.addEventListener('message', event => {
      try {
        apply(JSON.parse((event as MessageEvent<string>).data));
      } catch {
        if (active) setError('Latest test and lint runs received an invalid update.');
      }
    });
    stream.onerror = () => {
      if (active) setError(current => current || 'Latest test and lint runs disconnected; reconnecting…');
    };
    return () => {
      active = false;
      stream.close();
    };
  }, [enabled, visible]);

  return {
    projects,
    runs: latestRuns.projects,
    selected: projects.find(project => project.name === selectedName) ?? null,
    loading,
    error,
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
