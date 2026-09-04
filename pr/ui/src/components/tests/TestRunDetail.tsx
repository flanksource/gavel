import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { TestRunResults } from './TestRunResults';
import { runSnapshotQuery } from './types';

export function TestRunDetail({ project, projectDir, runId, onTodoCreated }: {
  project: string;
  projectDir: string;
  runId: string;
  onTodoCreated?: () => void;
}) {
  const identity = useMemo(() => ({ project, runId }), [project, runId]);
  const query = useQuery({
    ...runSnapshotQuery(identity),
    enabled: Boolean(project && runId),
  });

  if (query.error) return <Centered>{query.error.message}</Centered>;
  if (!query.data) return <Centered>Loading…</Centered>;
  return (
    <TestRunResults
      snapshot={query.data}
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
