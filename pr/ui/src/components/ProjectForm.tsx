import { useState, useEffect } from 'react';
import { Field, Combobox } from '@flanksource/clicky-ui/components';
import type { ComboboxOption } from '@flanksource/clicky-ui/components';
import type { Project, TodoProvider } from '../types';

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
  todoProvider: TodoProvider | '';
  setTodoProvider: (v: TodoProvider | '') => void;
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

// useProjectRegistration drives the projects entity: create POSTs to
// /api/projects, edit PUTs to /api/projects/{name} (name is the id, locked while
// editing), delete DELETEs the same path. Pass project=null for the add flow.
export function useProjectRegistration(open: boolean, project: Project | null): ProjectRegistration {
  const [name, setName] = useState('');
  const [dir, setDir] = useState('');
  const [repos, setRepos] = useState<string[]>([]);
  // Retained until M7 removes the legacy provider field from the project type.
  // Runtime persistence is PostgreSQL-only and this state is never saved.
  const [todoProvider, setTodoProvider] = useState<TodoProvider | ''>('');
  const [error, setError] = useState('');
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const editing = !!project;

  useEffect(() => {
    if (!open) return;
    setName(project?.name || '');
    setDir(project?.dir || '');
    setRepos(project?.repos || []);
    setTodoProvider(project?.todoProvider || '');
    setError('');
  }, [open, project?.name]);

  async function save(): Promise<boolean> {
    if (!name.trim() || !dir.trim()) {
      setError('Name and directory are required');
      return false;
    }
    setSaving(true);
    try {
      const url = project ? `/api/projects/${encodeURIComponent(project.name)}` : '/api/projects';
      const res = await fetch(url, {
        method: project ? 'PUT' : 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: name.trim(), dir: dir.trim(), repos }),
      });
      if (!res.ok) {
        setError((await res.text()) || `HTTP ${res.status}`);
        setSaving(false);
        return false;
      }
    } catch (e: any) {
      setError(e?.message || 'request failed');
      setSaving(false);
      return false;
    }
    setSaving(false);
    return true;
  }

  async function remove(): Promise<boolean> {
    if (!project) return false;
    if (!window.confirm(`Remove project "${project.name}"? This only forgets the workspace; nothing on disk is deleted.`)) {
      return false;
    }
    setDeleting(true);
    try {
      const res = await fetch(`/api/projects/${encodeURIComponent(project.name)}`, { method: 'DELETE' });
      if (!res.ok) {
        setError((await res.text()) || `HTTP ${res.status}`);
        setDeleting(false);
        return false;
      }
    } catch (e: any) {
      setError(e?.message || 'request failed');
      setDeleting(false);
      return false;
    }
    setDeleting(false);
    return true;
  }

  return {
    name, setName, dir, setDir, repos, setRepos, todoProvider, setTodoProvider,
    error, saving, deleting, editing, save, remove,
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
      {reg.error && <div className="text-sm text-destructive">{reg.error}</div>}
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
