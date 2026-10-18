import { useState } from 'react';
import { Button, Select } from '@flanksource/clicky-ui/components';
import { useNewTodoSteps } from './newTodoSteps';
import { readCreateStepChoice, writeCreateStepChoice } from './runChoiceStorage';
import { inputClass } from './format';
import { todoMutationJSON, type CreateTodoResponse } from './todoMutations';

const NONE = '';

export interface RunAfterCreationState {
  /** The step to send with the create request, or '' for None. A remembered
   *  choice the current status does not offer reads as None. */
  step: string;
  choose: (step: string) => void;
  steps: { name: string; label: string }[];
  loading: boolean;
  error: string;
}

/** Owns the "run after creation" choice for a create form: the steps the server
 *  offers for `status`, refetched when the status changes, and the last choice
 *  remembered across forms. */
export function useRunAfterCreation({ dir, status, enabled = true }: { dir: string; status: string; enabled?: boolean }): RunAfterCreationState {
  const [chosen, setChosen] = useState(readCreateStepChoice);
  const query = useNewTodoSteps({ dir, status, enabled });
  const steps = query.data ?? [];
  return {
    step: steps.some(step => step.name === chosen) ? chosen : NONE,
    choose: step => { setChosen(step); writeCreateStepChoice(step); },
    steps,
    loading: query.isPending && enabled,
    error: query.error instanceof Error ? query.error.message : '',
  };
}

export function RunAfterCreation({ state, disabled }: { state: RunAfterCreationState; disabled: boolean }) {
  return (
    <div className="space-y-1">
      <label className="flex items-center gap-2 text-sm">
        Run after creation
        <Select
          value={state.step}
          onChange={event => state.choose(event.currentTarget.value)}
          disabled={disabled || state.loading}
          className={inputClass}
          aria-label="Run after creation"
        >
          <option value={NONE}>{state.loading ? 'Loading steps...' : 'None'}</option>
          {state.steps.map(step => <option key={step.name} value={step.name}>{step.label}</option>)}
        </Select>
      </label>
      {state.error && <p role="alert" className="text-xs text-destructive">{state.error}</p>}
      <p className="text-xs text-muted-foreground">Start the chosen lifecycle step on the new todo as soon as it is saved.</p>
    </div>
  );
}

export function RunCreationRecovery({ result, dir, step, onDone }: { result: CreateTodoResponse; dir: string; step: string; onDone: () => void }) {
  const [error, setError] = useState(result.run?.error ?? `${step} could not start`);
  const [busy, setBusy] = useState(false);
  async function retry() {
    setBusy(true);
    try {
      await todoMutationJSON('/api/todos/run', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dir, ref: result.todo.ref, step }),
      }, `Failed to start ${step}`);
      onDone();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : `Failed to start ${step}`);
    } finally {
      setBusy(false);
    }
  }
  return (
    <div className="space-y-3" role="status">
      <p>Todo created. {step} could not start: {error}</p>
      <a href={`/todos/${encodeURIComponent(result.todo.ref)}`}>Open todo</a>
      <div className="flex gap-2">
        <Button onClick={retry} loading={busy}>Retry {step}</Button>
        <Button variant="outline" onClick={onDone} disabled={busy}>Continue without running</Button>
      </div>
    </div>
  );
}
