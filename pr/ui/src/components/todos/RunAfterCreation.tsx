import { useState } from 'react';
import { Button, Select } from '@flanksource/clicky-ui/components';
import { useNewTodoSteps } from './newTodoSteps';
import { readCreateStepChoice, writeCreateStepChoice } from './runChoiceStorage';
import { todoMutationJSON, type CreateTodoResponse } from './todoMutations';
import type { TodoRunOptions } from '../../types';
import { loadLastTodoRunOptions, reconcileTodoRunOptions, rememberTodoRunOptions, useTodoRunContext, type TodoRunContextState } from './run';
import { TodoRunRuntimeBar } from './TodoRunActionButton';

const NONE = '';

export type CreateTodoRunFields = Pick<TodoRunOptions, 'step' | 'spec' | 'presets'>;
type SubmittedTodoRun = CreateTodoRunFields & { dir: string; step: string };

export interface RunAfterCreationState {
  /** The step to send with the create request, or '' for None. A remembered
   *  choice the current status does not offer reads as None. */
  step: string;
  choose: (step: string) => void;
  steps: { name: string; label: string }[];
  loading: boolean;
  error: string;
  runtime: TodoRunContextState;
  options: TodoRunOptions;
  selectOptions: (options: TodoRunOptions) => void;
  fields: CreateTodoRunFields;
  blocked: boolean;
  submitted: SubmittedTodoRun | null;
  capture: () => CreateTodoRunFields;
}

/** Owns the "run after creation" choice for a create form: the steps the server
 *  offers for `status`, refetched when the status changes, and the last choice
 *  remembered across forms. */
export function useRunAfterCreation({ dir, status, enabled = true }: { dir: string; status: string; enabled?: boolean }): RunAfterCreationState {
  const [chosen, setChosen] = useState(readCreateStepChoice);
  const query = useNewTodoSteps({ dir, status, enabled });
  const steps = query.data ?? [];
  const step = steps.some(step => step.name === chosen) ? chosen : NONE;
  const runtime = useTodoRunContext({ dir, enabled: enabled && !!step });
  const context = runtime.error ? null : runtime.context;
  const [selection, setSelection] = useState<{ dir: string; step: string; options: TodoRunOptions } | null>(null);
  const [submitted, setSubmitted] = useState<SubmittedTodoRun | null>(null);
  const candidate = !step ? {} : selection?.dir === dir && selection.step === step ? selection.options : loadLastTodoRunOptions(step, context);
  const options = step && context ? reconcileTodoRunOptions(step, candidate, context) : candidate;
  const fields = step ? { step, spec: options.spec ?? {}, presets: options.presets } : {};
  const blocked = !!step && (runtime.loading || !runtime.context || !!runtime.error);
  return {
    step,
    choose: step => { setChosen(step); writeCreateStepChoice(step); },
    steps,
    loading: query.isPending && enabled,
    error: query.error instanceof Error ? query.error.message : step ? runtime.error : '',
    runtime, options, fields, blocked, submitted,
    selectOptions: options => setSelection({ dir, step, options: rememberTodoRunOptions(step, options) }),
    capture: () => {
      if (blocked) throw new Error(runtime.error || 'Runtime is not ready');
      setSubmitted(step ? { dir, ...fields, step } : null);
      return fields;
    },
  };
}

export function RunAfterCreation({ state, disabled }: { state: RunAfterCreationState; disabled: boolean }) {
  return (
    <div className="space-y-1">
      <div className="flex flex-wrap items-center gap-2">
        <label className="flex items-center gap-2 whitespace-nowrap text-sm [&>div]:w-auto">
          Run after creation
          <Select
            value={state.step}
            onChange={event => state.choose(event.currentTarget.value)}
            disabled={disabled || state.loading}
            className="h-8 w-auto rounded-md border border-input bg-background pl-2.5 pr-8 text-sm"
            aria-label="Run after creation"
          >
            <option value={NONE}>{state.loading ? 'Loading steps...' : 'None'}</option>
            {state.steps.map(step => <option key={step.name} value={step.name}>{step.label}</option>)}
          </Select>
        </label>
        {state.step && (state.runtime.context && !state.runtime.error ? (
          <TodoRunRuntimeBar action={state.step} context={state.runtime.context} options={state.options} onChange={state.selectOptions} disabled={disabled || state.blocked} />
        ) : (
          <Button type="button" variant="outline" size="sm" disabled>{state.runtime.loading ? 'Loading runtime...' : 'Runtime unavailable'}</Button>
        ))}
      </div>
      {state.error && <p role="alert" className="text-xs text-destructive">{state.error}</p>}
      <p className="text-xs text-muted-foreground">Start the chosen lifecycle step on the new todo as soon as it is saved.</p>
    </div>
  );
}

export function RunCreationRecovery({ result, request, onDone }: { result: CreateTodoResponse; request: SubmittedTodoRun | null; onDone: () => void }) {
  if (!request) throw new Error('Missing submitted workflow run for recovery');
  const { dir, step, spec, presets } = request;
  const [error, setError] = useState(result.run?.error ?? `${step} could not start`);
  const [busy, setBusy] = useState(false);
  async function retry() {
    setBusy(true);
    try {
      await todoMutationJSON('/api/todos/run', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dir, ref: result.todo.ref, step, spec, presets }),
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
