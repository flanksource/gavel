import { useEffect, useMemo, useState, type ComponentType } from 'react';
import { Markdown } from '@flanksource/clicky-ui/data';
import { UiHubot, UiRobotAi, type IconProps } from '@flanksource/clicky-ui/icons';
import type { TodoItem, TodoRunOptions } from '../../types';
import { Spinner } from '../../icons/Spinner';
import {
  TodoRunActionButton,
  TodoRunContextError,
  TodoRunEffortBadge,
  loadLastTodoRunOptions,
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
  const previewMutation = useTodoRunPreview(dir);

  useEffect(() => {
    if (!ref || !options) return;
    let cancelled = false;
    const controller = new AbortController();
    setError('');
    previewMutation.mutate({
      body: {
        ref,
        driver: options.driver,
        runMode: 'run',
        spec: { backend: options.spec?.backend, model: options.spec?.model, effort: options.spec?.effort },
      },
      signal: controller.signal,
    }, {
      onSuccess: data => {
        if (!cancelled) setPrompt(data.prompt ?? '');
      },
      onError: err => {
        if (!cancelled && !(err instanceof DOMException && err.name === 'AbortError')) setError(err.message);
      },
    });
    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [dir, ref, options?.driver, options?.spec?.backend, options?.spec?.model, options?.spec?.effort, previewMutation.mutate]);

  return { prompt, loading: previewMutation.isPending, error };
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
}: {
  dir: string;
  todo: TodoItem;
  onRun?: (options?: TodoRunOptions) => void;
  onAdvanced?: (action: TodoRunAction) => void;
  runBusy?: boolean;
  runDisabled?: boolean;
}) {
  const { context, loading: contextLoading, error: contextError } = useTodoRunContext();
  const options = useMemo(() => context ? loadLastTodoRunOptions('run', context) : null, [context]);
  const presentation = context && options ? todoRunButtonPresentation(options, context) : null;
  const runtime = context && options ? todoRunModeLabel(options, context) : '';
  const { prompt, loading, error } = useRunPromptPreview(dir, todo.ref, options);
  // provider.icon may be a runtime icon name (string) rather than a component;
  // only render it directly when it is a component, otherwise fall back
  // (mirrors iconForRunBackend in run.tsx).
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
          <TodoRunActionButton action="run" disabled={runDisabled || !context || !!contextError} loading={runBusy} onRun={onRun} onAdvanced={onAdvanced ?? (() => {})} />
          <TodoRunActionButton action="plan" disabled={runDisabled || !context || !!contextError} loading={runBusy} onRun={onRun} onAdvanced={onAdvanced ?? (() => {})} />
        </div>
      )}
    </div>
  );
}
