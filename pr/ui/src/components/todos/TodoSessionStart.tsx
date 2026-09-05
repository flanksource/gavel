import { useEffect, useMemo, useState, type ComponentType } from 'react';
import { Markdown } from '@flanksource/clicky-ui/data';
import { UiHubot, UiRobotAi, type IconProps } from '@flanksource/clicky-ui/icons';
import type { TodoItem, TodoRunOptions, TodoRunPreviewResponse } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { TodoRunActionButton } from './TodoRunActionButton';
import {
  TodoRunContextError,
  TodoRunEffortBadge,
  loadLastTodoRunOptions,
  reconcileTodoRunOptions,
  requestStepFor,
  todoRunButtonPresentation,
  todoRunModeLabel,
  useTodoRunContext,
  useTodoRunPreview,
  type TodoRunAction,
} from './run';

// useRunPromptPreview fetches the exact prompt a Run would dispatch for this todo,
// using the same /api/todos/run/preview path the advanced dialog uses. It is a
// read-only, non-blocking preview: a failure surfaces an error line but never
// stops the user from starting the run.
function useRunPromptPreview(dir: string, ref: string, options: TodoRunOptions | null) {
  const [prompt, setPrompt] = useState('');
  const [error, setError] = useState('');
  const [resolution, setResolution] = useState<{ key: string; data: TodoRunPreviewResponse }>();
  const requestKey = JSON.stringify({ dir, ref, options });
  const previewMutation = useTodoRunPreview(dir);

  useEffect(() => {
    if (!ref || !options) return;
    let cancelled = false;
    const controller = new AbortController();
    setError('');
    previewMutation.mutate({
      body: {
        ref,
        step: requestStepFor(options),
        runtimeProfile: options.runtimeProfile,
        spec: options.spec ?? {},
      },
      signal: controller.signal,
    }, {
      onSuccess: data => {
        if (!cancelled) {
          setPrompt(data.prompt ?? '');
          setResolution({ key: requestKey, data });
        }
      },
      onError: err => {
        if (!cancelled && !(err instanceof DOMException && err.name === 'AbortError')) setError(err.message);
      },
    });
    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [requestKey, previewMutation.mutate]);

  return { prompt, loading: previewMutation.isPending, error, resolution: resolution?.key === requestKey ? resolution.data : undefined };
}

function DetailChip({
  label,
  value,
  icon: Icon,
  iconColor,
}: {
  label: string;
  value: string;
  icon?: ComponentType<IconProps>;
  iconColor?: string;
}) {
  return (
    <span className="inline-flex items-center gap-1.5 rounded-full border border-border bg-muted/40 px-2.5 py-1 text-xs text-foreground">
      {Icon && <Icon className="text-sm" style={iconColor ? { color: iconColor } : undefined} aria-hidden="true" />}
      <span className="text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">{label}</span>
      <span className="font-medium">{value}</span>
    </span>
  );
}

// TodoSessionStart is the never-run Session tab: a large "ready to run" hero that
// surfaces the model / runtime / effort a run would use and the exact prompt it
// would send, plus Run/Plan actions. Once a run starts, TodoDetail sets the
// todo's sessionId and TodoSession swaps this out for the live SessionViewer.
export function TodoSessionStart({
  dir,
  todo,
  onRun,
  onAdvanced,
  runBusy,
  runDisabled,
  runOptions,
  planOptions,
  onRunOptionsChange,
  onPlanOptionsChange,
}: {
  dir: string;
  todo: TodoItem;
  onRun?: (options?: TodoRunOptions) => void;
  onAdvanced?: (action: TodoRunAction) => void;
  runBusy?: boolean;
  runDisabled?: boolean;
  runOptions?: TodoRunOptions;
  planOptions?: TodoRunOptions;
  onRunOptionsChange?: (options: TodoRunOptions) => void;
  onPlanOptionsChange?: (options: TodoRunOptions) => void;
}) {
  const { context, loading: contextLoading, error: contextError } = useTodoRunContext({ dir });
  const options = useMemo(() => context
    ? reconcileTodoRunOptions('run', runOptions ?? loadLastTodoRunOptions('run', context), context)
    : null, [context, runOptions]);
  const selectedPlanOptions = useMemo(() => context
    ? reconcileTodoRunOptions('plan', planOptions ?? loadLastTodoRunOptions('plan', context), context)
    : planOptions, [context, planOptions]);
  const { prompt, loading, error, resolution } = useRunPromptPreview(dir, todo.ref, options);
  const displayOptions = options && resolution ? { ...options, spec: { ...options.spec, model: resolution.model ?? options.spec?.model, mode: resolution.runtimeMode ?? options.spec?.mode } } : options;
  const presentation = context && displayOptions ? todoRunButtonPresentation(displayOptions, context) : null;
  const runtime = context && displayOptions ? todoRunModeLabel(displayOptions, context) : '';
  // provider.icon may be a runtime icon name rather than a component, so use the
  // generic agent icon unless it can be rendered directly.
  const providerIcon = presentation?.provider?.icon;
  const ProviderIcon = providerIcon && typeof providerIcon !== 'string' ? (providerIcon as ComponentType<IconProps>) : UiRobotAi;

  return (
    <div className="m-3 flex min-h-0 flex-1 flex-col items-center overflow-y-auto rounded-md border border-border bg-card px-6 py-8 text-center">
      <span className="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-full border border-border bg-muted/40">
        <UiHubot className="text-3xl text-muted-foreground" aria-hidden="true" />
      </span>
      <h2 className="text-lg font-semibold text-foreground">No agent session yet</h2>
      <p className="mt-1 max-w-md text-sm text-muted-foreground">Run this todo to start an agent session. It will run as:</p>
      <TodoRunContextError error={contextError} />

      {contextLoading ? (
        <div className="mt-4 text-xs text-muted-foreground">Loading Captain run providers…</div>
      ) : presentation && (
        <div className="mt-4 flex flex-wrap items-center justify-center gap-2">
          <DetailChip label="Model" value={presentation.model} icon={ProviderIcon} iconColor={presentation.provider?.iconColor} />
          <DetailChip label="Runtime" value={runtime} />
          {presentation.effort && <TodoRunEffortBadge effort={presentation.effort} />}
        </div>
      )}

      {options && <div className="mt-5 w-full max-w-2xl text-left">
        <div className="mb-1 flex items-center gap-2 text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
          Prompt
          {loading && <Spinner className="text-xs" />}
        </div>
        {error ? (
          <div className="text-xs text-red-600">{error}</div>
        ) : (
          <div className="max-h-64 overflow-y-auto rounded-md border border-border bg-muted/20 p-3">
            <Markdown text={prompt || 'Loading prompt…'} className="text-xs" />
          </div>
        )}
      </div>}

      {onRun && (
        <div className="mt-5 flex items-center justify-center gap-2">
          <TodoRunActionButton dir={dir} action="run" disabled={runDisabled || !context || !!contextError} loading={runBusy} options={options ?? undefined} onOptionsChange={onRunOptionsChange} onRun={onRun} onAdvanced={onAdvanced ?? (() => {})} />
          <TodoRunActionButton dir={dir} action="plan" disabled={runDisabled || !context || !!contextError} loading={runBusy} options={selectedPlanOptions} onOptionsChange={onPlanOptionsChange} onRun={onRun} onAdvanced={onAdvanced ?? (() => {})} />
        </div>
      )}
    </div>
  );
}
