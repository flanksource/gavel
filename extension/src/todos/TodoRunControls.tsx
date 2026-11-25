import { useEffect, useRef } from "react";
import { Field, MultiSelect, Select, Switch } from "@flanksource/clicky-ui";
import { familiesFromRuntimeCatalog, SpecRuntimeBar } from "@flanksource/clicky-ui/ai";
import { effectiveRuntime, runDefaults, type RunContext, type RunStep, type TodoRun } from "./runtime";

export function TodoRunControls({ run, context, steps, remembered, ready, disabled, error, onChange }: {
  run: TodoRun | undefined;
  context: RunContext | undefined;
  steps: RunStep[] | undefined;
  remembered: TodoRun | null | undefined;
  ready: boolean;
  disabled: boolean;
  error: string;
  onChange: (run: TodoRun | undefined) => void;
}) {
  const initialized = useRef(false);
  useEffect(() => {
    if (!ready || initialized.current) return;
    initialized.current = true;
    if (!run && remembered) onChange(remembered);
  }, [ready, remembered, run, onChange]);
  const chooseStep = (step: string) => {
    if (!context) throw new Error("Load this project's runtime catalog before choosing a step");
    onChange(runDefaults(context, step));
  };
  const preferred = steps?.find(step => step.name === "run") ?? steps?.[0];
  const unavailable = !context || !steps?.length;
  const inherited = run && context && !error ? effectiveRuntime(context, run) : undefined;
  return <section className="space-y-3 rounded-md border border-border p-3" aria-label="Run after creation">
    <Switch label="Start a new run" checked={!!run} disabled={disabled || (!run && unavailable)} onChange={enabled => {
      initialized.current = true;
      if (!enabled) onChange(undefined);
      else if (preferred) chooseStep(preferred.name);
    }} />
    <p className="text-xs text-muted-foreground">Start the selected lifecycle step after this todo syncs to Gavel.</p>
    {unavailable && <p role="status" className="text-xs text-muted-foreground">{ready ? "No saved run catalog for this project and status. Connect to Gavel to load its lifecycle steps and runtimes." : "Loading saved run choices…"}</p>}
    {run && <>
      <Field label="Run after creation" htmlFor="todo-run-step">
        <Select id="todo-run-step" value={run.step} disabled={disabled || !context} onChange={event => chooseStep(event.target.value)}>
          {!steps?.some(step => step.name === run.step) && <option value={run.step} disabled>{run.step} (unavailable for this status)</option>}
          {steps?.map(step => <option key={step.name} value={step.name}>{step.label}</option>)}
        </Select>
      </Field>
      {context && <Field label="Runtime presets"><MultiSelect ariaLabel="Runtime presets" value={run.presets ?? []} disabled={disabled}
        options={[...new Set([...(context.runtimePresets?.map(preset => preset.id) ?? []), ...(run.presets ?? [])])].map(value => ({ value, label: context.runtimePresets?.find(preset => preset.id === value)?.name ?? value }))}
        onChange={presets => onChange({ ...run, presets })} /></Field>}
      {context && inherited && <SpecRuntimeBar ariaLabel="Todo runtime" value={run.spec} readOnly={disabled}
        families={familiesFromRuntimeCatalog(context.runtimes)} models={context.models} reasoningEfforts={context.efforts}
        effectiveMode={inherited.mode} effectiveModel={inherited.model} inheritedCommits={inherited.workflow?.commits}
        onChange={spec => onChange({ ...run, spec })} />}
      {error && <p role="alert" className="text-xs text-destructive">{error}</p>}
    </>}
  </section>;
}
