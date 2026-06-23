import { useEffect, useState } from 'react';
import { Modal, Field, Button } from '@flanksource/clicky-ui/components';
import type { Project, TodoItem, TodoPriority, TodoStatus } from '../../types';
import { inputClass, priorities, statuses, statusLabel, todoQuery } from './format';

// CreateTodoDialog is a modal form for adding a todo to a chosen workspace.
export function CreateTodoDialog({
  open,
  onClose,
  workspaces,
  onCreated,
}: {
  open: boolean;
  onClose: () => void;
  workspaces: Project[];
  onCreated: (dir: string, todo: TodoItem) => void;
}) {
  const [dir, setDir] = useState(workspaces[0]?.dir ?? '');
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [priority, setPriority] = useState<TodoPriority>('medium');
  const [status, setStatus] = useState<TodoStatus>('pending');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (open) {
      setDir(workspaces[0]?.dir ?? '');
      setTitle('');
      setBody('');
      setPriority('medium');
      setStatus('pending');
      setError('');
      setBusy(false);
    }
  }, [open, workspaces]);

  if (!open) return null;

  async function submit() {
    if (!title.trim() || !dir || busy) return;
    setBusy(true);
    setError('');
    try {
      const provider = workspaces.find(w => w.dir === dir)?.todoProvider || 'auto';
      const response = await fetch(`/api/todos?${todoQuery(dir, provider)}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title, body, priority, status }),
      });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || 'Create failed');
      onCreated(dir, data as TodoItem);
    } catch (err: any) {
      setError(err?.message || 'Create failed');
      setBusy(false);
    }
  }

  return (
    <Modal
      open
      onClose={onClose}
      title="New todo"
      size="md"
      footer={
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={onClose}>Cancel</Button>
          <Button onClick={submit} loading={busy} disabled={!title.trim() || !dir}>Add todo</Button>
        </div>
      }
    >
      <div className="space-y-3">
        {error && <div className="text-sm text-destructive">{error}</div>}
        <Field label="Workspace">
          <select value={dir} onChange={e => setDir(e.currentTarget.value)} className={inputClass} aria-label="Workspace">
            {workspaces.map(w => <option key={w.dir} value={w.dir}>{w.name}</option>)}
          </select>
        </Field>
        <Field label="Title">
          <input
            className={inputClass}
            value={title}
            placeholder="What needs doing?"
            onChange={e => setTitle(e.currentTarget.value)}
            autoFocus
          />
        </Field>
        <div className="flex gap-3">
          <div className="flex-1">
            <Field label="Priority">
              <select value={priority} onChange={e => setPriority(e.currentTarget.value as TodoPriority)} className={inputClass} aria-label="Priority">
                {priorities.map(p => <option key={p} value={p}>{p}</option>)}
              </select>
            </Field>
          </div>
          <div className="flex-1">
            <Field label="Status">
              <select value={status} onChange={e => setStatus(e.currentTarget.value as TodoStatus)} className={inputClass} aria-label="Status">
                {statuses.map(s => <option key={s} value={s}>{statusLabel(s)}</option>)}
              </select>
            </Field>
          </div>
        </div>
        <Field label="Body">
          <textarea
            className={`${inputClass} h-28 resize-none`}
            value={body}
            placeholder="Details (optional)"
            onChange={e => setBody(e.currentTarget.value)}
          />
        </Field>
      </div>
    </Modal>
  );
}
