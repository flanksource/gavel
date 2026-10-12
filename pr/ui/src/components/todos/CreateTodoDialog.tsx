import { useState } from 'react';
import { Modal, Field, Button, Select } from '@flanksource/clicky-ui/components';
import type { Project, TodoItem, TodoPriority, TodoStatus } from '../../types';
import { ScreenshotPicker, todoFormData, useAttachments } from './attachments';
import { inputClass, priorities, statuses, statusLabel } from './format';
import { TodoBodyField } from './TodoBodyField';
import { useCreateTodoMutation, type CreateTodoResponse } from './todoMutations';
import { TriageAfterCreation, TriageCreationRecovery } from './TriageAfterCreation';

// initialDir picks the workspace to preselect: the current todo's workspace when
// it's a configured one, otherwise the first workspace.
function initialDir(defaultDir: string | undefined, workspaces: Project[]): string {
  if (defaultDir && workspaces.some(w => w.dir === defaultDir)) return defaultDir;
  return workspaces[0]?.dir ?? '';
}

// The todo a new child hangs under. Its workspace is the child's workspace, so
// a child form has no workspace to choose.
export interface CreateTodoParent {
  ref: string;
  title: string;
  dir: string;
}

type CreateTodoFormProps = {
  onClose: () => void;
  onCreated: (dir: string, todo: TodoItem) => void;
  // defaultDir preselects the workspace (the current todo's) when the dialog opens.
  defaultDir?: string;
} & (
  | { workspaces: Project[]; parent?: undefined }
  | { parent: CreateTodoParent; workspaces?: undefined }
);

// CreateTodoDialog is a modal form for adding a todo to a chosen workspace, or
// as a child of `parent`. The form mounts on open, so each opening starts from a
// fresh draft while prop changes during editing (a re-filtered workspaces list,
// a new selection) leave the draft alone.
export function CreateTodoDialog({ open, ...props }: CreateTodoFormProps & { open: boolean }) {
  if (!open) return null;
  return <CreateTodoForm {...props} />;
}

function CreateTodoForm({ onClose, workspaces, parent, onCreated, defaultDir }: CreateTodoFormProps) {
  const [chosenDir, setDir] = useState(() => initialDir(defaultDir, workspaces ?? []));
  const dir = parent ? parent.dir : chosenDir;
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [priority, setPriority] = useState<TodoPriority>('medium');
  const [status, setStatus] = useState<TodoStatus>('pending');
  const [error, setError] = useState('');
  const [triage, setTriage] = useState(true);
  const [created, setCreated] = useState<CreateTodoResponse | null>(null);
  // The paste listener lives only while the form is mounted, i.e. while open.
  const { attachments, previews, add, remove } = useAttachments({ pasteAnywhere: true });
  const createTodo = useCreateTodoMutation(dir);
  const busy = createTodo.isPending;

  async function submit() {
    if (!title.trim() || !dir || busy || created) return;
    setError('');
    try {
      // /api/todos/new accepts both JSON and multipart; post the image bytes as
      // multipart when screenshots are attached, otherwise the lighter JSON path.
      const fields = { title, body, priority, status, triage, ...(parent ? { parent: parent.ref } : {}) };
      const result = await createTodo.mutateAsync(attachments.length
        ? { body: todoFormData(fields, attachments) }
        : {
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(fields),
          });
      if (result.triage?.status === 'failed') setCreated(result);
      else onCreated(dir, result.todo);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create todo');
    }
  }

  return (
    <Modal
      open
      onClose={onClose}
      title={parent ? 'New child todo' : 'New todo'}
      size="xl"
      footer={created ? undefined :
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={onClose}>Cancel</Button>
          <Button onClick={submit} loading={busy} disabled={!title.trim() || !dir}>Add todo</Button>
        </div>
      }
    >
      {created ? <TriageCreationRecovery result={created} dir={dir} onDone={() => onCreated(dir, created.todo)} /> : <div className="space-y-3">
        {error && <div className="text-sm text-destructive">{error}</div>}
        {parent ? (
          <Field label="Child of">
            <div className={`${inputClass} truncate`} title={parent.title}>{parent.title}</div>
          </Field>
        ) : (
          <Field label="Workspace">
            <Select value={dir} onChange={e => setDir(e.currentTarget.value)} className={inputClass} aria-label="Workspace">
              {workspaces.map(w => <option key={w.dir} value={w.dir}>{w.name}</option>)}
            </Select>
          </Field>
        )}
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
        <TriageAfterCreation checked={triage} onChange={setTriage} disabled={busy} />
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
      </div>}
    </Modal>
  );
}
