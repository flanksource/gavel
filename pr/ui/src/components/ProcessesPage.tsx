import { useMemo } from 'react';
import type { Project } from '../types';
import { UiClose } from '@flanksource/clicky-ui/icons';
import { WorkspaceGroup } from './ProcessTable';
import { ProjectsPlaceholder } from './ProjectsPlaceholder';
import { aggregateDotClass, flattenProcesses, emptyProcStatus } from '../utils';
import { useProcStatus } from '../procStatusQuery';

export function ProcessesPage({
  projects,
  projectsLoaded,
  projectError,
  onProcChanged,
}: {
  projects: Project[];
  projectsLoaded: boolean;
  projectError?: string;
  onProcChanged: () => void;
}) {
  const procStatus = useProcStatus();
  const workspaces = useMemo(
    () => projects.map(p => ({ project: p, status: procStatus[p.name] ?? emptyProcStatus })),
    [projects, procStatus],
  );
  const procs = useMemo(() => flattenProcesses(projects, procStatus), [projects, procStatus]);
  const running = procs.filter(p => p.proc.status === 'running').length;
  const dot = aggregateDotClass(procs.map(p => p.proc));

  return (
    <div className="min-h-screen bg-background text-foreground">
      <header className="flex items-center justify-between border-b border-border px-4 py-3">
        <div className="flex min-w-0 items-center gap-3">
          <img src="/brand/gavel-logo.svg" alt="gavel" className="h-7 shrink-0" />
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <span className={`inline-block h-2.5 w-2.5 rounded-full ${dot}`} />
              <span className="text-sm font-semibold">Processes</span>
            </div>
            <div className="text-xs text-muted-foreground">{running} running of {procs.length}</div>
          </div>
        </div>
        <a
          href="/prs"
          className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          title="Back to PR dashboard"
          aria-label="Back to PR dashboard"
        >
          <UiClose className="text-base" />
        </a>
      </header>

      <main className="mx-auto w-full max-w-6xl px-4 py-4">
        {workspaces.length > 0 ? (
          <div className="divide-y divide-border border-y border-border">
            {workspaces.map(w => (
              <WorkspaceGroup key={w.project.name} project={w.project} status={w.status} onChanged={onProcChanged} />
            ))}
          </div>
        ) : (
          <ProjectsPlaceholder
            loaded={projectsLoaded}
            error={projectError}
            emptyText="No projects configured"
            className="py-10 text-center text-sm text-muted-foreground"
          />
        )}
      </main>
    </div>
  );
}
