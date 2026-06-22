import { useState, useEffect } from 'react';
import { Modal, Field, Combobox, Button } from '@flanksource/clicky-ui/components';
import type { ComboboxOption } from '@flanksource/clicky-ui/components';
import type { Project, TodoProvider } from '../types';

interface Props {
  open: boolean;
  onClose: () => void;
  onSaved: () => void;
  /** Repos to offer in the picker (the currently-known repos). */
  repoOptions: string[];
  /** When set, the dialog edits this project instead of adding a new one. */
  edit?: Project | null;
}

const inputClass = 'w-full rounded-md border border-input bg-background px-2.5 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring';

// AddProjectDialog drives the projects entity: a local workspace directory
// optionally bound to one or more repos. Create POSTs to /api/projects; edit
// PUTs to /api/projects/{name} (the name is the entity id, so it is locked while
// editing); delete DELETEs the same path.
export function AddProjectDialog({ open, onClose, onSaved, repoOptions, edit }: Props) {
  const [name, setName] = useState('');
  const [dir, setDir] = useState('');
  const [repos, setRepos] = useState<string[]>([]);
  // '' = auto-detect; 'grite' / 'todos' pin the workspace's todo backend.
  const [todoProvider, setTodoProvider] = useState<TodoProvider | ''>('');
  const [error, setError] = useState('');
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    if (open) {
      setName(edit?.name || '');
      setDir(edit?.dir || '');
      setRepos(edit?.repos || []);
      setTodoProvider(edit?.todoProvider || '');
      setError('');
    }
  }, [open, edit]);

  if (!open) return null;

  async function save() {
    if (!name.trim() || !dir.trim()) {
      setError('Name and directory are required');
      return;
    }
    setSaving(true);
    try {
      const url = edit ? `/api/projects/${encodeURIComponent(edit.name)}` : '/api/projects';
      const res = await fetch(url, {
        method: edit ? 'PUT' : 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: name.trim(), dir: dir.trim(), repos, todoProvider }),
      });
      if (!res.ok) {
        setError((await res.text()) || `HTTP ${res.status}`);
        setSaving(false);
        return;
      }
    } catch (e: any) {
      setError(e?.message || 'request failed');
      setSaving(false);
      return;
    }
    setSaving(false);
    onSaved();
    onClose();
  }

  async function remove() {
    if (!edit) return;
    if (!window.confirm(`Remove project "${edit.name}"? This only forgets the workspace; nothing on disk is deleted.`)) {
      return;
    }
    setDeleting(true);
    try {
      const res = await fetch(`/api/projects/${encodeURIComponent(edit.name)}`, { method: 'DELETE' });
      if (!res.ok) {
        setError((await res.text()) || `HTTP ${res.status}`);
        setDeleting(false);
        return;
      }
    } catch (e: any) {
      setError(e?.message || 'request failed');
      setDeleting(false);
      return;
    }
    setDeleting(false);
    onSaved();
    onClose();
  }

  const options: ComboboxOption[] = repoOptions.map(r => ({ value: r, label: r }));
  for (const r of repos) {
    if (!options.some(o => o.value === r)) options.push({ value: r, label: r });
  }

  return (
    <Modal
      open
      onClose={onClose}
      title={edit ? 'Edit project' : 'Add local directory'}
      size="md"
      footer={
        <div className="flex items-center justify-between gap-2">
          <div>
            {edit && (
              <Button variant="destructive" onClick={remove} loading={deleting}>Delete</Button>
            )}
          </div>
          <div className="flex gap-2">
            <Button variant="outline" onClick={onClose}>Cancel</Button>
            <Button onClick={save} loading={saving}>Save</Button>
          </div>
        </div>
      }
    >
      <div className="space-y-3">
        {error && <div className="text-sm text-destructive">{error}</div>}
        <Field label="Name" helper={edit ? 'The name identifies the project and cannot be changed' : undefined}>
          <input className={inputClass} value={name} placeholder="my-project" readOnly={!!edit}
            onChange={(e) => setName(e.currentTarget.value)} />
        </Field>
        <Field label="Directory" helper="Absolute path to the local checkout containing a Procfile">
          <input className={inputClass} value={dir} placeholder="/Users/me/src/project"
            onChange={(e) => setDir(e.currentTarget.value)} />
        </Field>
        <Field label="Repos" helper="GitHub repos this directory backs (optional)">
          <Combobox multiple options={options} value={repos}
            onChange={(v) => setRepos(v as string[])} placeholder="owner/repo" />
        </Field>
        <Field label="Todos" helper="Which todo backend this workspace uses">
          <select className={inputClass} value={todoProvider}
            onChange={(e) => setTodoProvider(e.currentTarget.value as TodoProvider | '')}>
            <option value="">Auto-detect</option>
            <option value="grite">Grite</option>
            <option value="todos">.todos files</option>
          </select>
        </Field>
      </div>
    </Modal>
  );
}
