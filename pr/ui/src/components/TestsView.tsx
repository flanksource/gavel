import { useEffect, useMemo, useState } from 'react';
import { SplitPane } from '@flanksource/clicky-ui/components';
import { useDocumentVisible } from '../useDocumentVisible';
import { GavelIcon } from './GavelIcon';
import { TestRunList } from './tests/TestRunList';
import { TestRunDetail } from './tests/TestRunDetail';
import type { TestRunsResponse } from './tests/types';

export function TestsView({
  selectedPath,
  onSelect,
}: {
  selectedPath: string;
  onSelect: (path: string) => void;
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

  return (
    <SplitPane
      left={<TestRunList projects={data.projects} selectedPath={selectedPath} onSelect={onSelect} />}
      right={
        runId ? (
          <TestRunDetail project={project} runId={runId} />
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
