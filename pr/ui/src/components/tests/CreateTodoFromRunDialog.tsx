import { useEffect, useMemo, useState } from 'react';
import { Button, Field, Modal } from '@flanksource/clicky-ui/components';
import { UiAdd, UiLinkExternal, UiPass } from '@flanksource/clicky-ui/icons';
import type { TodoItem } from '../../types';
import { inputClass } from '../todos/format';
import { useCreateTodoMutation } from '../todos/todoMutations';
import { defaultRunTodoTitle, type RunFailureCandidate } from './RunFailureCandidates';

export interface CreateTodoFromRunDialogProps {
  open: boolean;
  projectName: string;
  projectDir: string;
  runId: string;
  candidates: RunFailureCandidate[];
  onClose: () => void;
  onCreated?: () => void;
}

export function CreateTodoFromRunDialog({
  open,
  projectName,
  projectDir,
  runId,
  candidates,
  onClose,
  onCreated,
}: CreateTodoFromRunDialogProps) {
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [title, setTitle] = useState('');
  const [notes, setNotes] = useState('');
  const [error, setError] = useState('');
  const [created, setCreated] = useState<TodoItem | null>(null);
  const createTodo = useCreateTodoMutation(projectDir);
  const busy = createTodo.isPending;
  const selectedCandidates = useMemo(() => candidates.filter(candidate => selected.has(candidate.key)), [candidates, selected]);

  useEffect(() => {
    if (!open) return;
    setSelected(new Set(candidates.map(candidate => candidate.key)));
    setTitle(defaultRunTodoTitle(projectName, candidates));
    setNotes('');
    createTodo.reset();
    setError('');
    setCreated(null);
  }, [candidates, open, projectName, runId, createTodo.reset]);

  async function submit() {
    if (busy || !title.trim() || selectedCandidates.length === 0) return;
    if (!projectDir.trim()) {
      setError('Project directory is required to create a todo.');
      return;
    }
    setError('');
    try {
      const runHref = `/projects/${encodeURIComponent(projectName)}/runs/${encodeURIComponent(runId)}`;
      const body = [notes.trim(), `_From [Project run \`${runId}\`](${runHref})._`].filter(Boolean).join('\n\n');
      const payload = await createTodo.mutateAsync({
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          title: title.trim(),
          body,
          priority: 'medium',
          status: 'pending',
          criteria: selectedCandidates.map(candidate => ({ text: candidate.criterion })),
        }),
      });
      if (!payload.todo?.ref) throw new Error('Create todo response did not include the created todo.');
      setCreated(payload.todo);
      onCreated?.();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Failed to create todo');
    }
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={created ? 'Todo created' : 'Create todo from run failures'}
      size="lg"
      footer={created ? (
        <Button type="button" onClick={onClose}>Done</Button>
      ) : (
        <div className="flex w-full items-center justify-between gap-3">
          <span className="text-xs text-muted-foreground">{selectedCandidates.length} of {candidates.length} failures selected</span>
          <div className="flex gap-2">
            <Button type="button" variant="outline" onClick={onClose}>Cancel</Button>
            <Button type="button" loading={busy} disabled={!title.trim() || selectedCandidates.length === 0} onClick={() => void submit()}>
              <UiAdd /> Add todo
            </Button>
          </div>
        </div>
      )}
    >
      {created ? (
        <div className="space-y-3">
          <div className="flex items-start gap-2 rounded-md border border-green-200 bg-green-50 p-3 text-sm text-green-800">
            <UiPass className="mt-0.5 shrink-0 text-green-600" />
            <div>
              <div className="font-medium">{created.title}</div>
              <div className="text-xs">{created.criteria?.length ?? selectedCandidates.length} acceptance criteria added</div>
            </div>
          </div>
          <a href={`/todos/${created.ref.split('/').map(encodeURIComponent).join('/')}`} className="inline-flex items-center gap-1 text-sm text-blue-600 hover:underline">
            <UiLinkExternal /> Open todo
          </a>
        </div>
      ) : (
        <div className="space-y-4">
          {error && <div role="alert" className="text-sm text-destructive">{error}</div>}
          <Field label="Title">
            <input aria-label="Title" className={inputClass} value={title} onChange={event => setTitle(event.currentTarget.value)} />
          </Field>
          <Field label="Notes">
            <textarea aria-label="Notes" className={`${inputClass} min-h-20`} value={notes} onChange={event => setNotes(event.currentTarget.value)} />
          </Field>
          <div className="divide-y divide-border rounded-md border border-border">
            {candidates.map(candidate => (
              <label key={candidate.key} className="flex items-start gap-2 px-3 py-2 text-sm">
                <input
                  type="checkbox"
                  className="mt-0.5"
                  aria-label={`Include ${candidate.title}`}
                  checked={selected.has(candidate.key)}
                  onChange={() => setSelected(current => toggleCandidate(current, candidate.key))}
                />
                <span className="min-w-0">
                  <span className="block font-medium text-foreground">{candidate.title}</span>
                  {candidate.location && <span className="block font-mono text-xs text-muted-foreground">{candidate.location}</span>}
                  {candidate.detail && <span className="mt-0.5 block whitespace-pre-wrap text-xs text-muted-foreground">{candidate.detail}</span>}
                </span>
              </label>
            ))}
          </div>
        </div>
      )}
    </Modal>
  );
}

function toggleCandidate(current: Set<string>, key: string): Set<string> {
  const next = new Set(current);
  if (next.has(key)) next.delete(key);
  else next.add(key);
  return next;
}
