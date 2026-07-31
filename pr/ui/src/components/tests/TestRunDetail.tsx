import { useEffect, useState } from 'react';
import { TestRunResults } from './TestRunResults';
import { fetchRunSnapshot, type RunSnapshot } from './types';

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
    fetchRunSnapshot({ project, runId })
      .then((s: RunSnapshot) => setSnap(s))
      .catch(e => setError(e instanceof Error ? e.message : 'Failed to load run'));
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
