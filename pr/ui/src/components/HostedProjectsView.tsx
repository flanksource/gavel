// The Projects tab of a hosted organization: its repositories, and on the selected one test or
// lint started as a sandbox job on a picked branch, with the repository's recent test and lint runs.
import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Button, Select } from '@flanksource/clicky-ui/components';
import { UiFolderGit } from '@flanksource/clicky-ui/icons';
import { fetchJSON, mutationJSON, queryKeys } from '../query';
import type { Project } from '../types';
import { kindLabels, runStatus, useSandboxRuns } from './RunsView';

interface Branches {
  default: string;
  names: string[];
}

interface StartedAction {
  runId: string;
}

type Action = 'test' | 'lint';

const actions: { action: Action; label: string }[] = [{ action: 'test', label: 'Test' }, { action: 'lint', label: 'Lint' }];

// recentRunsShown is how many of the repository's test and lint runs the pane lists.
const recentRunsShown = 10;

export function HostedProjectsSidebar({ projects, selectedName, onSelect }: {
  projects: Project[];
  selectedName: string;
  onSelect: (name: string) => void;
}) {
  return (
    <nav aria-label="Projects" className="flex h-full flex-col gap-0.5 overflow-y-auto p-2 text-sm">
      {projects.length === 0 && <p className="p-2 text-muted-foreground">No repositories yet. Connect GitHub or push a repository.</p>}
      {projects.map(project => (
        <button
          key={project.name}
          type="button"
          aria-current={project.name === selectedName}
          onClick={() => onSelect(project.name)}
          className={`rounded px-2 py-1.5 text-left hover:bg-muted ${project.name === selectedName ? 'bg-primary/10 font-medium' : ''}`}
        >
          {project.name}
        </button>
      ))}
    </nav>
  );
}

export function HostedProjectPane({ project, onSelectRun }: { project: Project | undefined; onSelectRun: (id: string) => void }) {
  if (!project) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
        <div className="text-center">
          <UiFolderGit className="mb-2 text-4xl" />
          <p>Select a repository to run its tests or linters</p>
        </div>
      </div>
    );
  }
  return <ProjectActions key={project.name} project={project} onSelectRun={onSelectRun} />;
}

function ProjectActions({ project, onSelectRun }: { project: Project; onSelectRun: (id: string) => void }) {
  const queryClient = useQueryClient();
  const name = encodeURIComponent(project.name);
  const branches = useQuery({
    queryKey: queryKeys.projectBranches(project.name),
    queryFn: ({ signal }) => fetchJSON<Branches>({ url: `/api/projects/${name}/branches`, signal, context: `load the branches of ${project.name}` }),
  });
  const [picked, setPicked] = useState('');
  const branch = picked || branches.data?.default || '';
  const start = useMutation({
    mutationFn: (action: Action) => mutationJSON<StartedAction>({
      url: `/api/projects/${name}/actions`, method: 'POST', body: { action, branch }, context: `start ${action} on ${branch}`,
    }),
    onSuccess: started => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.sandboxRuns() });
      onSelectRun(started.runId);
    },
  });
  const runs = useSandboxRuns();
  const recent = (runs.data ?? []).filter(run => run.repo === project.name && (run.kind === 'test' || run.kind === 'lint')).slice(0, recentRunsShown);

  return (
    <div className="flex h-full min-h-0 flex-col gap-4 overflow-y-auto p-4 text-sm">
      <h2 className="text-base font-semibold">{project.name}</h2>
      <div className="flex flex-wrap items-center gap-2">
        <label htmlFor="project-branch" className="text-muted-foreground">Branch</label>
        <Select
          id="project-branch"
          value={branch}
          disabled={!branches.data}
          onChange={event => setPicked(event.target.value)}
          className="rounded border border-border bg-background px-2 py-1 font-mono text-xs"
        >
          {(branches.data?.names ?? []).map(name => <option key={name} value={name}>{name}</option>)}
        </Select>
        {actions.map(({ action, label }) => (
          <Button key={action} type="button" disabled={!branch || start.isPending} onClick={() => start.mutate(action)}>
            {label}
          </Button>
        ))}
      </div>
      {branches.error && <p role="alert" className="text-red-600">{branches.error.message}</p>}
      {start.error && <p role="alert" className="text-red-600">{start.error.message}</p>}
      <section className="flex flex-col gap-1">
        <h3 className="text-xs font-semibold uppercase text-muted-foreground">Recent runs</h3>
        {runs.error && <p className="text-muted-foreground">{runs.error.message}</p>}
        {runs.data && recent.length === 0 && <p className="text-muted-foreground">No test or lint runs yet.</p>}
        <ul aria-label="Recent runs" className="flex flex-col">
          {recent.map(run => (
            <li key={run.id}>
              <button type="button" onClick={() => onSelectRun(run.id)} className="w-full rounded px-2 py-1 text-left text-xs hover:bg-muted">
                {kindLabels[run.kind] ?? run.kind} · {run.ref?.replace(/^refs\/heads\//, '')} · {runStatus(run)}
                <span className="ml-2 text-muted-foreground">{run.user} · {new Date(run.startedAt ?? run.createdAt).toLocaleString()}</span>
              </button>
            </li>
          ))}
        </ul>
      </section>
    </div>
  );
}
