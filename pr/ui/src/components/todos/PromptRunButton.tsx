import { useEffect, useState, type ComponentType } from 'react';
import { SpecRuntimeEditor, promptRuntimeValueToPayload, type AISpecRuntimeValue } from '@flanksource/clicky-ui/ai';
import { Button, DropdownMenu, Modal } from '@flanksource/clicky-ui/components';
import { UiChevronDown, UiCog, UiHistory, UiPlay, type IconProps } from '@flanksource/clicky-ui/icons';
import type { TodoRunOptions } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { TodoRunRuntimeBar } from './TodoRunActionButton';
import {
  TodoRunContextError,
  defaultRunOptionsForAction,
  loadLastTodoRunOptions,
  loadRecentAdvancedTodoRunOptions,
  reconcileTodoRunOptions,
  runButtonQualifierForOptions,
  runSpec,
  useTodoRunContext,
} from './run';
import { agentForRuntime, buildRunFamilies, driverForSelection, type RunContext } from './providers';

export type PromptRunScope = 'approval' | 'verification';

type PromptRunHistory = {
  last?: TodoRunOptions;
  recent?: TodoRunOptions[];
};

type PromptRunHistoryState = Partial<Record<PromptRunScope, PromptRunHistory>>;

// v2 for the same reason as RUN_CHOICE_STORAGE_KEY in run.tsx: the persisted
// TodoRunOptions now nests the spec.
const PROMPT_RUN_STORAGE_KEY = 'gavel.pr-ui.promptRunChoices.v2';
const LEGACY_RUN_STORAGE_KEY = 'gavel.pr-ui.todoRunChoices.v2';

function cloneSpec(spec: AISpecRuntimeValue): AISpecRuntimeValue {
  return JSON.parse(JSON.stringify(spec)) as AISpecRuntimeValue;
}

function readPromptRunHistory(): PromptRunHistoryState {
  if (typeof localStorage === 'undefined') return {};
  const raw = localStorage.getItem(PROMPT_RUN_STORAGE_KEY);
  if (!raw) return {};
  try {
    const parsed = JSON.parse(raw) as PromptRunHistoryState;
    return parsed && typeof parsed === 'object' ? parsed : {};
  } catch {
    return {};
  }
}

function writePromptRunHistory(state: PromptRunHistoryState) {
  if (typeof localStorage === 'undefined') return;
  localStorage.setItem(PROMPT_RUN_STORAGE_KEY, JSON.stringify(state));
}

// verificationSpec keeps only what a verification run may carry: the model and
// budget knobs, the permission posture and workflow.verify. Setup, prompt,
// sessionId and workflow.commits are dropped — the persisted fixture supplies the
// prompt and the server rejects a verification spec that would commit.
export function verificationSpec(spec: AISpecRuntimeValue): AISpecRuntimeValue {
  return {
    model: spec.model,
    id: spec.id,
    backend: spec.backend,
    temperature: spec.temperature,
    effort: spec.effort,
    noCache: spec.noCache,
    fallbacks: spec.fallbacks,
    budget: spec.budget,
    memory: spec.memory,
    permissions: spec.permissions,
    workflow: spec.workflow?.verify ? { verify: cloneSpec({ workflow: { verify: spec.workflow.verify } }).workflow?.verify } : undefined,
    cliArgs: spec.cliArgs,
  };
}

function normalizePromptRunOptions(scope: PromptRunScope, options: TodoRunOptions, context: RunContext): TodoRunOptions {
  const reconciled = reconcileTodoRunOptions('run', options, context);
  // A verification run has no driver or run mode of its own — it posts a bare
  // spec to /api/todos/verification/run — so only the spec half survives.
  return scope === 'verification' ? { spec: verificationSpec(runSpec(reconciled)) } : reconciled;
}

function optionsKey(options: TodoRunOptions): string {
  return JSON.stringify(options);
}

function migrateApprovalHistory(state: PromptRunHistoryState, context: RunContext): PromptRunHistoryState {
  if (state.approval?.last) return state;
  if (typeof localStorage === 'undefined' || !localStorage.getItem(LEGACY_RUN_STORAGE_KEY)) return state;
  const next = { ...state };
  next.approval = {
    last: loadLastTodoRunOptions('run', context),
    recent: loadRecentAdvancedTodoRunOptions('run', context),
  };
  writePromptRunHistory(next);
  return next;
}

export function loadPromptRunOptions(scope: PromptRunScope, context: RunContext): TodoRunOptions {
  const state = scope === 'approval'
    ? migrateApprovalHistory(readPromptRunHistory(), context)
    : readPromptRunHistory();
  const fallback = defaultRunOptionsForAction('run', context);
  return normalizePromptRunOptions(scope, state[scope]?.last ?? fallback, context);
}

export function loadRecentPromptRunOptions(scope: PromptRunScope, context: RunContext): TodoRunOptions[] {
  const state = scope === 'approval'
    ? migrateApprovalHistory(readPromptRunHistory(), context)
    : readPromptRunHistory();
  const seen = new Set<string>();
  return (state[scope]?.recent ?? [])
    .map(options => normalizePromptRunOptions(scope, options, context))
    .filter(options => {
      const key = optionsKey(options);
      if (seen.has(key)) return false;
      seen.add(key);
      return true;
    })
    .slice(0, 3);
}

export function rememberPromptRunOptions(scope: PromptRunScope, options: TodoRunOptions, context: RunContext): TodoRunOptions {
  const remembered = normalizePromptRunOptions(scope, options, context);
  const state = scope === 'approval'
    ? migrateApprovalHistory(readPromptRunHistory(), context)
    : readPromptRunHistory();
  const recent = (state[scope]?.recent ?? [])
    .map(item => normalizePromptRunOptions(scope, item, context))
    .filter(item => optionsKey(item) !== optionsKey(remembered));
  state[scope] = { last: remembered, recent: [remembered, ...recent].slice(0, 3) };
  writePromptRunHistory(state);
  return remembered;
}

export function PromptRunButton({
  scope,
  label,
  title,
  disabled,
  loading,
  icon = UiPlay,
  onRun,
  onAdvanced,
}: {
  scope: PromptRunScope;
  label: string;
  title: string;
  disabled?: boolean;
  loading?: boolean;
  icon?: ComponentType<IconProps>;
  onRun: (options: TodoRunOptions) => void;
  onAdvanced: (options: TodoRunOptions) => void;
}) {
  const { context, loading: contextLoading, error: contextError } = useTodoRunContext(!disabled);
  const [revision, setRevision] = useState(0);
  const options = context ? loadPromptRunOptions(scope, context) : defaultRunOptionsForAction('run');
  const recent = context ? loadRecentPromptRunOptions(scope, context) : [];
  const unavailable = contextLoading || !context || !!contextError;
  const PrimaryIcon = loading ? Spinner : icon;

  function select(optionsToRemember: TodoRunOptions, close?: () => void) {
    if (!context) return;
    close?.();
    rememberPromptRunOptions(scope, optionsToRemember, context);
    setRevision(value => value + 1);
  }

  function run() {
    if (!context) return;
    onRun(rememberPromptRunOptions(scope, options, context));
  }

  return (
    <div className="flex flex-col items-start gap-1" data-history-revision={revision}>
      <div className="inline-flex min-h-8 shrink-0 items-stretch gap-1">
        <Button
          variant="ghost"
          type="button"
          disabled={disabled || unavailable}
          onClick={run}
          className="inline-flex h-8 items-center gap-1 rounded-md border border-border px-2 text-xs font-medium text-foreground hover:bg-muted disabled:opacity-50"
          title={title}
        >
          <PrimaryIcon className="text-xs" />
          <span>{label}</span>
        </Button>
        {context && !contextError ? (
          <TodoRunRuntimeBar
            action="run"
            context={context}
            options={options}
            disabled={disabled || unavailable}
            onChange={selected => select(selected)}
          />
        ) : (
          <Button variant="outline" size="sm" type="button" disabled title={`${label} runtime unavailable`} aria-label={`${label} runtime unavailable`} className="h-8 px-2 text-xs">
            Runtime
          </Button>
        )}
        {context && !contextError ? (
          <DropdownMenu
            align="right"
            menuLabel={`${label} history and advanced options`}
            menuClassName="max-h-[70vh] w-72 max-w-[calc(100vw-24px)] overflow-y-auto"
            trigger={
              <Button
                variant="ghost"
                size="icon"
                type="button"
                disabled={disabled || unavailable}
                title={`${label} history and advanced options`}
                aria-label={`${label} history and advanced options`}
                className="h-8 w-8 rounded-md text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
              >
                <UiChevronDown className="text-xs" />
              </Button>
            }
          >
            {close => (
              <div>
                {recent.length > 0 && (
                  <div className="p-1 text-xs">
                    <div className="flex items-center gap-1 px-2 py-1 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">
                      <UiHistory className="text-xs" />
                      Recent configs
                    </div>
                    {recent.map((item, index) => (
                      <Button
                        key={optionsKey(item)}
                        variant="ghost"
                        type="button"
                        onClick={() => select(item, close)}
                        className="flex h-8 w-full items-center justify-start rounded px-2 text-left text-xs hover:bg-muted"
                      >
                        {index + 1}. {runButtonQualifierForOptions(item, context)}
                        {runSpec(item).effort ? ` · ${runSpec(item).effort}` : ''}
                      </Button>
                    ))}
                  </div>
                )}
                <div className="border-t border-border p-1">
                  <Button
                    variant="ghost"
                    type="button"
                    aria-label="Advanced"
                    onClick={() => {
                      close();
                      onAdvanced(options);
                    }}
                    className="flex h-auto w-full items-center justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted"
                  >
                    <UiCog className="text-sm text-muted-foreground" />
                    <span>
                      <span className="block text-xs font-medium text-foreground">Advanced</span>
                      <span className="block text-[11px] text-muted-foreground">model, effort, timeout, permissions</span>
                    </span>
                  </Button>
                </div>
              </div>
            )}
          </DropdownMenu>
        ) : (
          <Button variant="ghost" size="icon" type="button" disabled title={`${label} history and advanced options`} aria-label={`${label} history and advanced options`} className="h-8 w-8 rounded-md">
            <UiChevronDown className="text-xs" />
          </Button>
        )}
      </div>
      <TodoRunContextError error={contextError} />
    </div>
  );
}

const APPROVAL_SPEC_SECTIONS = ['model', 'prompt', 'permissions', 'workspace', 'verify', 'commit'] as const;
const VERIFICATION_SPEC_SECTIONS = ['model', 'permissions', 'verify'] as const;

export function PromptRunAdvancedDialog({
  scope,
  open,
  initial,
  loading,
  onClose,
  onRun,
}: {
  scope: PromptRunScope;
  open: boolean;
  initial: TodoRunOptions;
  loading?: boolean;
  onClose: () => void;
  onRun: (options: TodoRunOptions) => void;
}) {
  const { context, loading: contextLoading, error: contextError } = useTodoRunContext(open);
  // The editor edits the spec half only; the driver/runMode siblings are decided
  // on save from the spec's backend.
  const initialSpec = (options: TodoRunOptions) =>
    scope === 'verification' ? verificationSpec(runSpec(options)) : cloneSpec(runSpec(options));
  const [value, setValue] = useState<AISpecRuntimeValue>(() => initialSpec(initial));

  useEffect(() => {
    if (open) setValue(initialSpec(initial));
  }, [open, initial, scope]);

  if (!open) return null;
  const models = context?.models ?? [];
  const verification = scope === 'verification';

  return (
    <Modal open onClose={onClose} title={verification ? 'Verification run options' : 'Approve and run options'} size="2xl">
      <TodoRunContextError error={contextError} />
      {contextLoading && <div className="text-xs text-muted-foreground">Loading Captain run providers…</div>}
      {context && !contextError && <SpecRuntimeEditor
        value={value}
        onChange={setValue}
        models={models}
        families={buildRunFamilies(context)}
        tools={context.tools}
        sections={verification ? VERIFICATION_SPEC_SECTIONS : APPROVAL_SPEC_SECTIONS}
        title={verification ? 'Verification runtime' : 'Implementation runtime'}
        eyebrow={verification ? 'Embedded AI prompt configuration' : 'Approved plan execution'}
        onCancel={onClose}
        onSave={() => {
          const { spec } = promptRuntimeValueToPayload(value);
          const runtimeSpec = spec ?? {};
          if (verification) {
            onRun(rememberPromptRunOptions(scope, { spec: verificationSpec(runtimeSpec) }, context));
            return;
          }
          const agent = agentForRuntime(context, runtimeSpec.backend, runtimeSpec.model);
          const selection = driverForSelection(context, agent, runtimeSpec.backend);
          onRun(rememberPromptRunOptions(scope, {
            driver: selection.driver,
            runMode: 'run',
            spec: { ...runtimeSpec, backend: selection.runBackend },
          }, context));
        }}
        saveLabel={loading ? 'Running…' : verification ? 'Run verification' : 'Approve & run'}
        footerStatus={verification ? 'The persisted fixture supplies the prompt.' : 'Ready to implement the approved plan.'}
      />}
    </Modal>
  );
}
