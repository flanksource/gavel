import { useState } from 'react';
import { Modal, Field, Button, Select } from '@flanksource/clicky-ui/components';
import type { Project, TodoItem, TodoPriority, TodoStatus } from '../../types';
import { ScreenshotPicker, todoFormData, useAttachments } from './attachments';
import { inputClass, priorities, statuses, statusLabel } from './format';
import { TodoBodyField } from './TodoBodyField';
import { useCreateTodoMutation } from './todoMutations';

// initialDir picks the workspace to preselect: the current todo's workspace when
// it's a configured one, otherwise the first workspace.
function initialDir(defaultDir: string | undefined, workspaces: Project[]): string {
  if (defaultDir && workspaces.some(w => w.dir === defaultDir)) return defaultDir;
  return workspaces[0]?.dir ?? '';
}

interface CreateTodoFormProps {
  onClose: () => void;
  workspaces: Project[];
  onCreated: (dir: string, todo: TodoItem) => void;
  // defaultDir preselects the workspace (the current todo's) when the dialog opens.
  defaultDir?: string;
}

// CreateTodoDialog is a modal form for adding a todo to a chosen workspace. The
// form mounts on open, so each opening starts from a fresh draft while prop
// changes during editing (a re-filtered workspaces list, a new selection) leave
// the draft alone.
export function CreateTodoDialog({ open, ...props }: CreateTodoFormProps & { open: boolean }) {
  if (!open) return null;
  return <CreateTodoForm {...props} />;
}

function CreateTodoForm({ onClose, workspaces, onCreated, defaultDir }: CreateTodoFormProps) {
  const [dir, setDir] = useState(() => initialDir(defaultDir, workspaces));
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [priority, setPriority] = useState<TodoPriority>('medium');
  const [status, setStatus] = useState<TodoStatus>('pending');
  const [error, setError] = useState('');
  // The paste listener lives only while the form is mounted, i.e. while open.
  const { attachments, previews, add, remove } = useAttachments({ pasteAnywhere: true });
  const createTodo = useCreateTodoMutation(dir);
  const busy = createTodo.isPending;

  async function submit() {
    if (!title.trim() || !dir || busy) return;
    setError('');
    try {
      // /api/todos/new accepts both JSON and multipart; post the image bytes as
      // multipart when screenshots are attached, otherwise the lighter JSON path.
      const result = await createTodo.mutateAsync(attachments.length
        ? { body: todoFormData({ title, body, priority, status }, attachments) }
        : {
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ title, body, priority, status }),
          });
      onCreated(dir, result.todo);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create todo');
    }
  }

  return (
    <Modal
      open
      onClose={onClose}
      title="New todo"
      size="xl"
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
        <TodoBodyField
          label="Body"
          value={body}
          onChange={setBody}
          placeholder="Details (optional)"
          disabled={busy}
        />
        <Field label="Screenshot">
          <ScreenshotPicker previews={previews} onAdd={add} onRemove={remove} disabled={busy} />
        </Field>
      </div>
    </Modal>
  );
}
