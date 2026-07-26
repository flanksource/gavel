import { useEffect, useState } from 'react';
import { TestRunResults } from './TestRunResults';
import type { RunSnapshot } from './types';

export function TestRunDetail({ project, projectDir, runId, onTodoCreated }: {
  project: string;
  projectDir: string;
  runId: string;
  onTodoCreated?: () => void;
}) {
  const [snap, setSnap] = useState<RunSnapshot | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    setSnap(null);
    setError('');
    const url = `/api/tests/run?project=${encodeURIComponent(project)}&runId=${encodeURIComponent(runId)}`;
    fetch(url)
      .then(r => (r.ok ? r.json() : r.json().then(e => Promise.reject(e.error || 'failed to load run'))))
      .then((s: RunSnapshot) => setSnap(s))
      .catch(e => setError(typeof e === 'string' ? e : 'Failed to load run'));
  }, [project, runId]);

  if (error) return <Centered>{error}</Centered>;
  if (!snap) return <Centered>Loading…</Centered>;
  return (
    <TestRunResults
      snapshot={snap}
      done
      runKey={runId}
      projectName={project}
      projectDir={projectDir}
      onTodoCreated={onTodoCreated}
    />
  );
}

function Centered({ children }: { children: React.ReactNode }) {
  return <div className="flex h-full items-center justify-center p-6 text-sm text-muted-foreground">{children}</div>;
}
