import { useCallback, useState } from 'react';
import { Modal, Field, Button, Select, useModal, useToast, type HostedModalProps, type ModalFormDefinition } from '@flanksource/clicky-ui/components';
import type { Project, TodoItem, TodoPriority, TodoStatus, TodoVisibility } from '../../types';
import { currentAuthMode } from '../../query';
import { todoVisibilityVisible } from '../../hostedUI';
import { ScreenshotPicker, todoFormData, useAttachments } from './attachments';
import { inputClass, priorities, statuses, statusLabel } from './format';
import { TodoBodyField } from './TodoBodyField';
import { createTodoSubmissionResult, useCreateTodoMutation, type CreateTodoResponse } from './todoMutations';
import { RunAfterCreation, useRunAfterCreation, type RunAfterCreationState } from './RunAfterCreation';
import { TodoPRField, todoPRLink } from './TodoPRLink';
import { TodoVisibilitySelect } from './TodoVisibility';
import { prTodoHref, prTodoRunRecoveryFormKey } from './prTodoDraft';

export const newTodoFormKey = 'todo:create';

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

type CreateTodoTarget = {
  // defaultDir preselects the workspace (the current todo's) when the dialog opens.
  defaultDir?: string;
} & (
  | { workspaces: Project[]; parent?: undefined }
  | { parent: CreateTodoParent; workspaces?: undefined }
);

// CreateTodoDraft is what an opener hands the hosted form. onOpen backs the
// success toast's "Open todo" so the host can select the todo in place.
export type CreateTodoDraft = CreateTodoTarget & { onOpen?: (dir: string, todo: TodoItem) => void };

type CreateTodoDialogProps = CreateTodoTarget & {
  hosted: HostedModalProps;
  onResult: (dir: string, result: CreateTodoResponse, run: RunAfterCreationState['submitted']) => void;
};

// CreateTodoDialog is the form for adding a todo to a chosen workspace, or as a
// child of `parent`. It submits in the background: the hosted modal minimizes
// into a "Submitting…" card while the request runs, and reopens with the draft
// and the error if the server refuses it.
export function CreateTodoDialog({ hosted, onResult, workspaces, parent, defaultDir }: CreateTodoDialogProps) {
  const { modalProps, submit: submitHosted } = hosted;
  const [chosenDir, setDir] = useState(() => initialDir(defaultDir, workspaces ?? []));
  const dir = parent ? parent.dir : chosenDir;
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [prURL, setPRURL] = useState('');
  const [priority, setPriority] = useState<TodoPriority>('medium');
  const [status, setStatus] = useState<TodoStatus>('pending');
  const [error, setError] = useState('');
  const showVisibility = todoVisibilityVisible(currentAuthMode());
  const [visibility, setVisibility] = useState<TodoVisibility>('private');
  const runAfter = useRunAfterCreation({ dir, status });
  // The form stays mounted while minimized, so pasting elsewhere must not attach.
  const { attachments, previews, add, remove } = useAttachments({ pasteAnywhere: !modalProps.minimized });
  const createTodo = useCreateTodoMutation(dir);
  const busy = createTodo.isPending;

  async function submit() {
    if (!title.trim() || !dir || busy || runAfter.blocked) return;
    setError('');
    try {
      const runFields = runAfter.capture();
      // /api/todos/new accepts both JSON and multipart; post the image bytes as
      // multipart when screenshots are attached, otherwise the lighter JSON path.
      const fields = {
        title, body, priority, status, ...runFields, pr: todoPRLink(prURL),
        ...(parent ? { parent: parent.ref } : {}),
        ...(showVisibility ? { visibility } : {}),
      };
      const mutation = attachments.length
        ? { body: todoFormData(fields, attachments) }
        : { headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(fields) };
      await submitHosted({
        draft: {},
        request: () => createTodoSubmissionResult(() => createTodo.mutateAsync(mutation)),
        onSuccess: result => onResult(dir, result, runFields.step ? { dir, ...runFields, step: runFields.step } : null),
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create todo');
    }
  }

  return (
    <Modal
      {...modalProps}
      title={parent ? 'New child todo' : modalProps.title}
      error={error || modalProps.error}
      size="xl"
      footer={
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={modalProps.onClose}>Cancel</Button>
          <Button onClick={submit} loading={busy} disabled={!title.trim() || !dir || runAfter.blocked}>Add todo</Button>
        </div>
      }
    >
      <div className="space-y-3">
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
            aria-label="Title"
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
        {showVisibility && (
          <Field label="Visibility">
            <TodoVisibilitySelect value={visibility} onChange={setVisibility} disabled={busy} />
          </Field>
        )}
        <RunAfterCreation state={runAfter} disabled={busy} />
        <TodoPRField value={prURL} onChange={setPRURL} disabled={busy} />
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

function HostedCreateTodoDialog({ initialDraft, modalProps, submit }: HostedModalProps) {
  const draft = initialDraft as CreateTodoDraft;
  const { onOpen } = draft;
  const { openModal } = useModal();
  const { toast } = useToast();

  function onResult(dir: string, result: CreateTodoResponse, run: RunAfterCreationState['submitted']) {
    if (result.run?.status === 'failed' && run) {
      openModal(prTodoRunRecoveryFormKey, { result, run });
      return;
    }
    const { todo } = result;
    toast({
      tone: 'success',
      message: <>Todo created. {onOpen
        ? <Button type="button" variant="link" size="sm" className="h-auto p-0 text-inherit underline" onClick={() => onOpen(dir, todo)}>Open todo</Button>
        : <a className="underline" href={prTodoHref(todo.ref)}>Open todo</a>}</>,
    });
  }

  return <CreateTodoDialog {...draft} hosted={{ initialDraft, modalProps, submit }} onResult={onResult} />;
}

export const todoForms = {
  [newTodoFormKey]: { title: 'New todo', Form: HostedCreateTodoDialog },
} satisfies Record<string, ModalFormDefinition>;

/** Opens the new-todo form in the app's modal host. */
export function useOpenCreateTodo() {
  const { openModal } = useModal();
  return useCallback((draft: CreateTodoDraft) => openModal(newTodoFormKey, draft), [openModal]);
}
