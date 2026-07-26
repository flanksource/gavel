import { useEffect, useMemo, useState } from 'react';
import { SplitPane } from '@flanksource/clicky-ui/components';
import { UiFolderGit } from '@flanksource/clicky-ui/icons';
import type { Project, ProcStatus } from '../types';
import { useDocumentVisible } from '../useDocumentVisible';
import { ErrorBoundary } from './ErrorBoundary';
import { ProjectsBar } from './ProjectsBar';
import { ProjectStatusView } from './ProjectStatusView';
import { TestRunDetail } from './tests/TestRunDetail';
import type { TestRunsResponse } from './tests/types';

interface Props {
  projects: Project[];
  procStatus: Record<string, ProcStatus>;
  selectedName: string;
  selectedRunId: string;
  diffPath: string;
  onSelect: (name: string) => void;
  onSelectRun: (project: string, runId: string) => void;
  onDiffPathChange: (path: string) => void;
  onChanged: () => void;
  onAdd: () => void;
  onSettings: (project: Project) => void;
}

export function ProjectsView(props: Props) {
  const [latestRuns, setLatestRuns] = useState<TestRunsResponse>({ projects: [] });
  const [runsLoading, setRunsLoading] = useState(true);
  const [runError, setRunError] = useState('');
  const visible = useDocumentVisible();
  const projects = useMemo(() => {
    const merged = new Map(props.projects.map(project => [project.name, project]));
    for (const project of latestRuns.projects) {
      if (merged.has(project.name)) continue;
      merged.set(project.name, {
        name: project.name,
        dir: project.dir,
        repos: [...new Set(project.runs.flatMap(run => run.repo ? [run.repo] : []))],
      });
    }
    return [...merged.values()];
  }, [latestRuns.projects, props.projects]);
  const selected = projects.find(project => project.name === props.selectedName) ?? null;

  useEffect(() => {
    if (!visible) return;
    let active = true;
    const apply = (payload: unknown) => {
      const next = parseLatestRuns(payload);
      if (!active) return;
      setLatestRuns(next);
      setRunsLoading(false);
      setRunError('');
    };

    fetch('/api/tests')
      .then(async response => {
        if (!response.ok) throw new Error(`Load test and lint runs failed (HTTP ${response.status})`);
        apply(await response.json());
      })
      .catch(cause => {
        if (!active) return;
        setRunsLoading(false);
        setRunError(cause instanceof Error ? cause.message : 'Load test and lint runs failed');
      });

    const stream = new EventSource('/api/tests/stream');
    stream.addEventListener('message', event => {
      try {
        apply(JSON.parse((event as MessageEvent<string>).data));
      } catch {
        if (active) setRunError('Latest test and lint runs received an invalid update.');
      }
    });
    stream.onerror = () => {
      if (active) setRunError(current => current || 'Latest test and lint runs disconnected; reconnecting…');
    };
    return () => {
      active = false;
      stream.close();
    };
  }, [visible]);

  return (
    <SplitPane
      defaultSplit={22}
      minLeft={16}
      minRight={55}
      left={
        <ProjectsBar
          projects={projects}
          runs={latestRuns.projects}
          procStatus={props.procStatus}
          selected={props.selectedName}
          selectedRunId={props.selectedRunId}
          runError={runError}
          runsLoading={runsLoading}
          onSelect={project => props.onSelect(project.name)}
          onSelectRun={props.onSelectRun}
          onChanged={props.onChanged}
          onAdd={props.onAdd}
          onSettings={props.onSettings}
        />
      }
      right={selected ? (
        props.selectedRunId ? (
          <ErrorBoundary key={`${selected.name}/${props.selectedRunId}`}>
            <TestRunDetail project={selected.name} projectDir={selected.dir} runId={props.selectedRunId} onTodoCreated={props.onChanged} />
          </ErrorBoundary>
        ) : (
          <ProjectStatusView key={selected.name} project={selected} diffPath={props.diffPath} onDiffPathChange={props.onDiffPathChange} onChanged={props.onChanged} />
        )
      ) : (
        <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
          <div className="text-center">
            <UiFolderGit className="mb-2 text-4xl" />
            <p>{props.selectedName ? `Project ${props.selectedName} was not found` : 'Select a project to view its working tree'}</p>
          </div>
        </div>
      )}
    />
  );
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
