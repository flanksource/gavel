import { UiFolderGit } from '@flanksource/clicky-ui/icons';
import type { Project, ProcStatus } from '../types';
import type { ProjectCatalog } from '../useProjectCatalog';
import { ErrorBoundary } from './ErrorBoundary';
import { ProjectsBar } from './ProjectsBar';
import { ProjectStatusView } from './ProjectStatusView';
import { TestRunDetail } from './tests/TestRunDetail';

// The projects tab renders as two AppShell slots — ProjectsSidebar into
// bodySidebar, ProjectDetailPane into children — so both take the shared
// catalog rather than owning the /api/tests state between them.

export function ProjectsSidebar({ catalog, procStatus, selectedName, selectedRunId, historyEnabled, onHistoryChange, onSelect, onSelectRun, onChanged, onAdd, onSettings }: {
  catalog: ProjectCatalog;
  procStatus: Record<string, ProcStatus>;
  selectedName: string;
  selectedRunId: string;
  historyEnabled: boolean;
  onHistoryChange: (enabled: boolean) => void;
  onSelect: (name: string) => void;
  onSelectRun: (project: string, runId: string) => void;
  onChanged: () => void;
  onAdd: () => void;
  onSettings: (project: Project) => void;
}) {
  return (
    <ProjectsBar
      projects={catalog.projects}
      runs={catalog.runs}
      procStatus={procStatus}
      selected={selectedName}
      selectedRunId={selectedRunId}
      runError={catalog.error}
      runsLoading={catalog.loading}
      historyEnabled={historyEnabled}
      onHistoryChange={onHistoryChange}
      onSelect={project => onSelect(project.name)}
      onSelectRun={onSelectRun}
      onChanged={onChanged}
      onAdd={onAdd}
      onSettings={onSettings}
    />
  );
}

export function ProjectDetailPane({ catalog, selectedName, selectedRunId, diffPath, resultsEnabled, onDiffPathChange, onChanged }: {
  catalog: ProjectCatalog;
  selectedName: string;
  selectedRunId: string;
  diffPath: string;
  resultsEnabled: boolean;
  onDiffPathChange: (path: string) => void;
  onChanged: () => void;
}) {
  const selected = catalog.selected;
  if (!selected) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
        <div className="text-center">
          <UiFolderGit className="mb-2 text-4xl" />
          <p>{selectedName ? `Project ${selectedName} was not found` : 'Select a project to view its working tree'}</p>
        </div>
      </div>
    );
  }
  if (selectedRunId) {
    return (
      <ErrorBoundary key={`${selected.name}/${selectedRunId}`}>
        <TestRunDetail project={selected.name} projectDir={selected.dir} runId={selectedRunId} onTodoCreated={onChanged} />
      </ErrorBoundary>
    );
  }
  return <ProjectStatusView key={selected.name} project={selected} diffPath={diffPath} showResults={resultsEnabled} onDiffPathChange={onDiffPathChange} onChanged={onChanged} />;
}
