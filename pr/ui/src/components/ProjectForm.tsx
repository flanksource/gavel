import { useState, useEffect } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Field, Combobox } from '@flanksource/clicky-ui/components';
import type { ComboboxOption } from '@flanksource/clicky-ui/components';
import { mutationJSON, queryKeys } from '../query';
import type { Project } from '../types';
import { projectDiffQueryKey } from './projectMutations';

const inputClass =
  'w-full rounded-md border border-input bg-background px-2.5 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring';

// ProjectRegistration is the editable state for one workspace entity (a local
// directory optionally bound to repos). Shared by the add dialog and the settings
// dialog's Project tab so the form + fetch logic lives in one place.
export interface ProjectRegistration {
  name: string;
  setName: (v: string) => void;
  dir: string;
  setDir: (v: string) => void;
  repos: string[];
  setRepos: (v: string[]) => void;
  error: string;
  saving: boolean;
  deleting: boolean;
  /** True when editing an existing project (name locked, delete available). */
  editing: boolean;
  /** POSTs a new project (add) or PUTs the existing one (edit). Returns ok. */
  save: () => Promise<boolean>;
  /** DELETEs the project after confirmation; only meaningful when editing. */
  remove: () => Promise<boolean>;
}

export interface ProjectRegistrationOptions {
  open: boolean;
  project: Project | null;
  defaults?: Partial<Pick<Project, 'name' | 'dir' | 'repos'>>;
}

// useProjectRegistration drives the projects entity: create POSTs to
// /api/projects, edit PUTs to /api/projects/{name} (name is the id, locked while
// editing), delete DELETEs the same path. Pass project=null for the add flow.
export function useProjectRegistration(options: ProjectRegistrationOptions): ProjectRegistration {
  const { open, project, defaults } = options;
  const [name, setName] = useState('');
  const [dir, setDir] = useState('');
  const [repos, setRepos] = useState<string[]>([]);
  const [error, setError] = useState('');
  const editing = !!project;
  const projectReposKey = project?.repos.join('\0') ?? '';
  const defaultReposKey = defaults?.repos?.join('\0') ?? '';
  const queryClient = useQueryClient();
  const saveMutation = useMutation({
    mutationFn: (registration: Project) => mutationJSON<Project>({
      url: project ? `/api/projects/${encodeURIComponent(project.name)}` : '/api/projects',
      method: project ? 'PUT' : 'POST',
      body: registration,
      context: `Failed to ${project ? 'update' : 'create'} project ${project?.name ?? registration.name}`,
    }),
    onSuccess: async () => {
      const invalidations = [
        queryClient.invalidateQueries({ queryKey: queryKeys.projects(), exact: true }),
      ];
      if (project) {
        invalidations.push(
          queryClient.invalidateQueries({ queryKey: queryKeys.projectStatusScope(project.name) }),
          queryClient.invalidateQueries({ queryKey: projectDiffQueryKey(project.name) }),
        );
      }
      await Promise.all(invalidations);
    },
  });
  const removeMutation = useMutation({
    mutationFn: (projectName: string) => mutationJSON<void>({
      url: `/api/projects/${encodeURIComponent(projectName)}`,
      method: 'DELETE',
      context: `Failed to delete project ${projectName}`,
    }),
    onSuccess: async (_, projectName) => {
      queryClient.removeQueries({ queryKey: queryKeys.projectStatusScope(projectName) });
      queryClient.removeQueries({ queryKey: projectDiffQueryKey(projectName) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.projects(), exact: true });
    },
  });

  useEffect(() => {
    if (!open) return;
    setName(project ? project.name : (defaults?.name ?? ''));
    setDir(project ? project.dir : (defaults?.dir ?? ''));
    setRepos([...(project ? project.repos : (defaults?.repos ?? []))]);
    setError('');
  }, [
    open,
    project?.name,
    project?.dir,
    projectReposKey,
    defaults?.name,
    defaults?.dir,
    defaultReposKey,
  ]);

  async function save(): Promise<boolean> {
    if (!name.trim() || !dir.trim()) {
      setError('Name and directory are required');
      return false;
    }
    const registration = { name: name.trim(), dir: dir.trim(), repos };
    setError('');
    try {
      await saveMutation.mutateAsync(registration);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Project save failed');
      return false;
    }
    return true;
  }

  async function remove(): Promise<boolean> {
    if (!project) return false;
    if (!window.confirm(`Remove project "${project.name}"? This only forgets the workspace; nothing on disk is deleted.`)) {
      return false;
    }
    setError('');
    try {
      await removeMutation.mutateAsync(project.name);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Project delete failed');
      return false;
    }
    return true;
  }

  return {
    name, setName, dir, setDir, repos, setRepos,
    error, saving: saveMutation.isPending, deleting: removeMutation.isPending, editing, save, remove,
  };
}

// ProjectFields renders the registration inputs bound to a ProjectRegistration.
// The caller owns the surrounding dialog chrome (title, footer, save/delete
// buttons) so this component stays layout-agnostic.
export function ProjectFields({ reg, repoOptions }: { reg: ProjectRegistration; repoOptions: string[] }) {
  const options: ComboboxOption[] = repoOptions.map(r => ({ value: r, label: r }));
  for (const r of reg.repos) {
    if (!options.some(o => o.value === r)) options.push({ value: r, label: r });
  }

  return (
    <div className="space-y-3">
      {reg.error && <div role="alert" className="text-sm text-destructive">{reg.error}</div>}
      <Field label="Name" helper={reg.editing ? 'The name identifies the project and cannot be changed' : undefined}>
        <input className={inputClass} value={reg.name} placeholder="my-project" readOnly={reg.editing}
          onChange={(e) => reg.setName(e.currentTarget.value)} />
      </Field>
      <Field label="Directory" helper="Absolute path to the local checkout containing a Procfile">
        <input className={inputClass} value={reg.dir} placeholder="/Users/me/src/project"
          onChange={(e) => reg.setDir(e.currentTarget.value)} />
      </Field>
      <Field label="Repos" helper="GitHub repos this directory backs (optional)">
        <Combobox multiple options={options} value={reg.repos}
          onChange={(v) => reg.setRepos(v as string[])} placeholder="owner/repo" />
      </Field>
      <Field label="Todos" helper="Todo persistence is managed centrally">
        <div className={`${inputClass} cursor-default text-muted-foreground`} aria-label="Todo persistence">
          PostgreSQL
        </div>
      </Field>
    </div>
  );
}
