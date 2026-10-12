import { useState } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { todoMutationJSON, type CreateTodoResponse } from './todoMutations';

export function TriageAfterCreation({ checked, onChange, disabled }: { checked: boolean; onChange: (checked: boolean) => void; disabled: boolean }) {
  return (
    <div className="space-y-1">
      <label className="flex items-center gap-2 text-sm">
        <input type="checkbox" checked={checked} onChange={event => onChange(event.currentTarget.checked)} disabled={disabled} />
        Triage after creation
      </label>
      <p className="text-xs text-muted-foreground">Choose a title and labels and check for related work. Relationships require review.</p>
    </div>
  );
}

export function TriageCreationRecovery({ result, dir, onDone }: { result: CreateTodoResponse; dir: string; onDone: () => void }) {
  const [error, setError] = useState(result.triage?.error ?? 'Triage could not start');
  const [busy, setBusy] = useState(false);
  async function retry() {
    setBusy(true);
    try {
      await todoMutationJSON('/api/todos/run', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dir, ref: result.todo.ref, step: 'triage.new' }),
      }, 'Failed to start triage');
      onDone();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Failed to start triage');
    } finally {
      setBusy(false);
    }
  }
  return (
    <div className="space-y-3" role="status">
      <p>Todo created. Triage could not start: {error}</p>
      <a href={`/todos/${encodeURIComponent(result.todo.ref)}`}>Open todo</a>
      <div className="flex gap-2">
        <Button onClick={retry} loading={busy}>Retry triage</Button>
        <Button variant="outline" onClick={onDone} disabled={busy}>Continue without triage</Button>
      </div>
    </div>
  );
}
