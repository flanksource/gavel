import { useEffect, useState } from 'react';
import { Modal, Field, Button, Select } from '@flanksource/clicky-ui/components';
import type { Project, TodoItem, TodoPriority, TodoStatus } from '../../types';
import { ScreenshotPicker, todoFormData, useAttachments } from './attachments';
import { inputClass, priorities, statuses, statusLabel, todoQuery } from './format';

// initialDir picks the workspace to preselect: the current todo's workspace when
// it's a configured one, otherwise the first workspace.
function initialDir(defaultDir: string | undefined, workspaces: Project[]): string {
  if (defaultDir && workspaces.some(w => w.dir === defaultDir)) return defaultDir;
  return workspaces[0]?.dir ?? '';
}

// CreateTodoDialog is a modal form for adding a todo to a chosen workspace.
export function CreateTodoDialog({
  open,
  onClose,
  workspaces,
  onCreated,
  defaultDir,
}: {
  open: boolean;
  onClose: () => void;
  workspaces: Project[];
  onCreated: (dir: string, todo: TodoItem) => void;
  // defaultDir preselects the workspace (the current todo's) when the dialog opens.
  defaultDir?: string;
}) {
  const [dir, setDir] = useState(() => initialDir(defaultDir, workspaces));
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [priority, setPriority] = useState<TodoPriority>('medium');
  const [status, setStatus] = useState<TodoStatus>('pending');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  // Capture a pasted screenshot only while the dialog is open — it stays mounted
  // (closed) in the dashboard, so an always-on listener would hijack paste.
  const { attachments, previews, add, remove, clear } = useAttachments({ pasteAnywhere: open });

  useEffect(() => {
    if (open) {
      setDir(initialDir(defaultDir, workspaces));
      setTitle('');
      setBody('');
      setPriority('medium');
      setStatus('pending');
      setError('');
      setBusy(false);
      clear();
    }
  }, [open, workspaces, defaultDir, clear]);

  if (!open) return null;

  async function submit() {
    if (!title.trim() || !dir || busy) return;
    setBusy(true);
    setError('');
    try {
      const provider = workspaces.find(w => w.dir === dir)?.todoProvider || 'auto';
      // /api/todos/new accepts both JSON and multipart; post the image bytes as
      // multipart when screenshots are attached, otherwise the lighter JSON path.
      const url = `/api/todos/new?${todoQuery(dir, provider)}`;
      const response = attachments.length
        ? await fetch(url, { method: 'POST', body: todoFormData({ title, body, priority, status }, attachments) })
        : await fetch(url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ title, body, priority, status }),
          });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || 'Create failed');
      onCreated(dir, data.todo as TodoItem);
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
          <Select value={dir} onChange={e => setDir(e.currentTarget.value)} className={inputClass} aria-label="Workspace">
            {workspaces.map(w => <option key={w.dir} value={w.dir}>{w.name}</option>)}
          </Select>
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
              <Select value={priority} onChange={e => setPriority(e.currentTarget.value as TodoPriority)} className={inputClass} aria-label="Priority">
                {priorities.map(p => <option key={p} value={p}>{p}</option>)}
              </Select>
            </Field>
          </div>
          <div className="flex-1">
            <Field label="Status">
              <Select value={status} onChange={e => setStatus(e.currentTarget.value as TodoStatus)} className={inputClass} aria-label="Status">
                {statuses.map(s => <option key={s} value={s}>{statusLabel(s)}</option>)}
              </Select>
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
        <Field label="Screenshot">
          <ScreenshotPicker previews={previews} onAdd={add} onRemove={remove} disabled={busy} />
        </Field>
      </div>
    </Modal>
  );
}
