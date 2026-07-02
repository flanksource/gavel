import { useEffect, useMemo, useState } from 'react';
import { SplitPane } from '@flanksource/clicky-ui/components';
import { useDocumentVisible } from '../useDocumentVisible';
import { ErrorBoundary } from './ErrorBoundary';
import { GavelIcon } from './GavelIcon';
import { TestRunList } from './tests/TestRunList';
import { TestRunDetail } from './tests/TestRunDetail';
import type { TestRunsResponse } from './tests/types';

export function TestsView({
  selectedPath,
  onSelect,
  query = '',
}: {
  selectedPath: string;
  onSelect: (path: string) => void;
  query?: string;
}) {
  const [data, setData] = useState<TestRunsResponse>({ projects: [] });
  const visible = useDocumentVisible();

  // Stream the run list while visible; the syncer pushes a fresh snapshot
  // whenever a scan finds new runs (with a slow ticker fallback server-side).
  useEffect(() => {
    if (!visible) return;
    fetch('/api/tests')
      .then(r => r.json())
      .then((d: TestRunsResponse) => setData(d))
      .catch(() => {});
    const es = new EventSource('/api/tests/stream');
    es.addEventListener('message', (e: MessageEvent) => {
      try {
        setData(JSON.parse(e.data));
      } catch {
        /* ignore malformed frame */
      }
    });
    es.onerror = () => {
      /* EventSource auto-reconnects */
    };
    return () => es.close();
  }, [visible]);

  // selectedPath is "{project}/{runId}"; runId (run-<timestamp>) never contains
  // a slash, so split on the first one.
  const [project, runId] = useMemo(() => {
    const i = selectedPath.indexOf('/');
    return i < 0 ? [selectedPath, ''] : [selectedPath.slice(0, i), selectedPath.slice(i + 1)];
  }, [selectedPath]);

  // The global search filters the run list by project name or run kind/id; a
  // project whose name matches keeps all its runs.
  const projects = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return data.projects;
    return data.projects
      .map(p =>
        p.name.toLowerCase().includes(q)
          ? p
          : { ...p, runs: p.runs.filter(r => r.kind.toLowerCase().includes(q) || r.runId.toLowerCase().includes(q)) },
      )
      .filter(p => p.runs.length > 0);
  }, [data.projects, query]);

  return (
    <SplitPane
      left={<TestRunList projects={projects} selectedPath={selectedPath} onSelect={onSelect} />}
      right={
        runId ? (
          // Keyed on the path so navigating to another run resets a tripped boundary.
          <ErrorBoundary key={selectedPath}>
            <TestRunDetail project={project} runId={runId} />
          </ErrorBoundary>
        ) : (
          <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
            <div className="text-center">
              <GavelIcon name="codicon:beaker" className="mb-2 text-4xl" />
              <p>Select a run to view its results</p>
            </div>
          </div>
        )
      }
    />
  );
}
