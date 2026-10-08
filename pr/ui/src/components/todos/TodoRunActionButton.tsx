import { useEffect, useState, type ComponentType } from "react";
import { Button, Select } from "@flanksource/clicky-ui/components";
import { RuntimeBar, type AISpecRuntimeValue } from "@flanksource/clicky-ui/ai";
import { UiCog, type IconProps } from "@flanksource/clicky-ui/icons";
import type { TodoRunOptions } from "../../types";
import { Spinner } from "../../icons/Spinner";
import {
  loadLastTodoRunOptions,
  reconcileTodoRunOptions,
  rememberTodoRunOptions,
  runActionConfig,
  runSpec,
  TodoRunContextError,
  todoRunOptionsForRuntimeChange,
  type TodoRunAction,
  useTodoRunContext,
} from "./run";
import { buildRunFamilies, type RunContext } from "./providers";
import { effectiveTodoRuntime, unresolvedTodoRuntimePreset } from './runtimePresets';
import { continueBranchLabel, NEW_BRANCH_LABEL, type RunBranchChoice } from './runBranchChoice';

const CONTINUE = "continue";
const NEW = "new";

export function TodoRunRuntimeBar({
  action,
  context,
  options,
  disabled,
  onChange,
}: {
  action: TodoRunAction;
  context: RunContext;
  options: TodoRunOptions;
  disabled?: boolean;
  onChange: (options: TodoRunOptions) => void;
}) {
  const spec = runSpec(options);
  const inherited = effectiveTodoRuntime(options, context);
  const preset = unresolvedTodoRuntimePreset(options, context);
  if (preset) return <span className="inline-flex items-center rounded-md border border-border px-2 text-xs text-muted-foreground">Preset: {preset}</span>;
  return (
    <fieldset disabled={disabled} className="min-w-0 border-0 p-0 disabled:opacity-50">
      <RuntimeBar<AISpecRuntimeValue>
        value={spec}
        effectiveModel={inherited.model}
        effectiveMode={inherited.mode}
        families={buildRunFamilies(context)}
        models={context.models ?? []}
        reasoningEfforts={context.efforts}
        ariaLabel={`${runActionConfig[action].label} runtime`}
        className="min-w-0"
        onChange={runtime => onChange(todoRunOptionsForRuntimeChange({ action, context, options, runtime }))}
      />
    </fieldset>
  );
}

export function TodoRunActionButton({
  dir,
  action,
  disabled,
  loading,
  label,
  icon,
  tone = "default",
  title,
  options: controlledOptions,
  onOptionsChange,
  onRun,
  onAdvanced,
  branchChoice,
}: {
  dir: string;
  action: TodoRunAction;
  disabled?: boolean;
  loading?: boolean;
  label?: string;
  icon?: ComponentType<IconProps>;
  tone?: "default" | "danger";
  title?: string;
  options?: TodoRunOptions;
  onOptionsChange?: (options: TodoRunOptions) => void;
  onRun: (options?: TodoRunOptions) => void;
  onAdvanced: (action: TodoRunAction) => void;
  // The branch a run may continue on; when set, a Run offers continuing on it or
  // starting a new branch, and the choice rides on the dispatched options.
  branchChoice?: RunBranchChoice | null;
}) {
  const config = runActionConfig[action];
  const { context, loading: contextLoading, error: contextError } = useTodoRunContext({ dir });
  const [selectedOptions, setSelectedOptions] = useState<TodoRunOptions | null>(null);
  useEffect(() => setSelectedOptions(null), [action, context]);
  const [reuseBranch, setReuseBranch] = useState(true);
  useEffect(() => setReuseBranch(true), [branchChoice?.branch]);
  const offersBranch = action === "run" && !!branchChoice;
  const unavailable = contextLoading || !context || !!contextError;
  const candidateOptions = controlledOptions ?? selectedOptions ?? loadLastTodoRunOptions(action, context);
  const currentOptions = context ? reconcileTodoRunOptions(action, candidateOptions, context) : candidateOptions;
  const primaryTone = tone === "danger" ? "text-red-600 hover:bg-red-500/10 hover:text-red-700" : "text-foreground hover:bg-muted";
  const PrimaryIcon = loading ? Spinner : icon ?? config.icon;

  function selectOptions(options: TodoRunOptions) {
    const remembered = rememberTodoRunOptions(action, options);
    setSelectedOptions(remembered);
    onOptionsChange?.(remembered);
  }

  function runWith(options: TodoRunOptions) {
    const remembered = rememberTodoRunOptions(action, options);
    onRun(offersBranch ? { ...remembered, reuseBranch } : remembered);
  }

  return (
    <div className="flex flex-col items-start gap-1">
      <div className="inline-flex min-h-8 shrink-0 items-stretch gap-1">
        <Button variant="ghost" type="button" onClick={() => runWith(currentOptions)} disabled={disabled || unavailable} title={title ?? config.title} className={`inline-flex h-8 items-center gap-1 rounded-md border border-border px-2 text-xs font-medium disabled:opacity-50 ${primaryTone}`}>
          <PrimaryIcon className="text-xs" />
          <span>{label ?? config.label}</span>
        </Button>
        {context && !contextError ? (
          <TodoRunRuntimeBar action={action} context={context} options={currentOptions} disabled={disabled || unavailable} onChange={selectOptions} />
        ) : (
          <Button variant="outline" size="sm" type="button" disabled title={`${config.label} runtime unavailable`} aria-label={`${config.label} runtime unavailable`} className="h-8 px-2 text-xs">Runtime</Button>
        )}
        <Button variant="outline" size="icon" type="button" disabled={disabled || unavailable} title={`Advanced ${config.label.toLowerCase()} options`} aria-label={`Advanced ${config.label.toLowerCase()} options`} onClick={() => onAdvanced(action)} className="h-8 w-8 shrink-0 text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50">
          <UiCog className="text-sm" />
        </Button>
      </div>
      {offersBranch && branchChoice && (
        <Select
          aria-label="Branch for this run"
          value={reuseBranch ? CONTINUE : NEW}
          disabled={disabled || unavailable}
          onChange={event => setReuseBranch(event.currentTarget.value === CONTINUE)}
          options={[
            { value: CONTINUE, label: continueBranchLabel(branchChoice) },
            { value: NEW, label: NEW_BRANCH_LABEL },
          ]}
          className="h-8 max-w-full text-xs"
        />
      )}
      <TodoRunContextError error={contextError} />
    </div>
  );
}
