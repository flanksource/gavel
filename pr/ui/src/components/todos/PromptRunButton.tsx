import { useEffect, useState, type ComponentType } from 'react';
import { SpecRuntimeEditor, promptRuntimeValueToPayload, type AISpecRuntimeValue } from '@flanksource/clicky-ui/ai';
import { Button, DropdownMenu, Modal } from '@flanksource/clicky-ui/components';
import { UiChevronDown, UiCog, UiHistory, UiPlay, type IconProps } from '@flanksource/clicky-ui/icons';
import type { TodoRunOptions } from '../../types';
import { Spinner } from '../../icons/Spinner';
import {
  TodoRunDropdownContent,
  defaultRunOptionsForAction,
  loadLastTodoRunOptions,
  loadRecentAdvancedTodoRunOptions,
  reconcileTodoRunOptions,
  runButtonQualifierForOptions,
  useTodoRunContext,
} from './run';
import { agentForBackend, buildRunFamilies, driverForSelection, type RunContext } from './providers';

export type PromptRunScope = 'approval' | 'verification';

type PromptRunHistory = {
  last?: TodoRunOptions;
  recent?: TodoRunOptions[];
};

type PromptRunHistoryState = Partial<Record<PromptRunScope, PromptRunHistory>>;

const PROMPT_RUN_STORAGE_KEY = 'gavel.pr-ui.promptRunChoices.v1';
const LEGACY_RUN_STORAGE_KEY = 'gavel.pr-ui.todoRunChoices.v1';

function cloneOptions(options: TodoRunOptions): TodoRunOptions {
  return JSON.parse(JSON.stringify(options)) as TodoRunOptions;
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

export function verificationSpecFromOptions(options: TodoRunOptions): TodoRunOptions {
  const workflow = options.workflow?.verify
    ? { verify: cloneOptions({ workflow: { verify: options.workflow.verify } }).workflow?.verify }
    : undefined;
  return {
    model: options.model,
    id: options.id,
    backend: options.backend,
    temperature: options.temperature,
    effort: options.effort,
    noCache: options.noCache,
    fallbacks: options.fallbacks,
    budget: options.budget,
    memory: options.memory,
    permissions: options.permissions,
    workflow,
    cliArgs: options.cliArgs,
  };
}

function normalizePromptRunOptions(scope: PromptRunScope, options: TodoRunOptions, context: RunContext): TodoRunOptions {
  const reconciled = reconcileTodoRunOptions('run', options, context);
  return scope === 'verification' ? verificationSpecFromOptions(reconciled) : reconciled;
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
  const context = useTodoRunContext(!disabled);
  const [revision, setRevision] = useState(0);
  const options = loadPromptRunOptions(scope, context);
  const recent = loadRecentPromptRunOptions(scope, context);
  const PrimaryIcon = loading ? Spinner : icon;

  function run(optionsToRemember: TodoRunOptions, close?: () => void) {
    close?.();
    const remembered = rememberPromptRunOptions(scope, optionsToRemember, context);
    setRevision(value => value + 1);
    onRun(remembered);
  }

  return (
    <div className="inline-flex h-8 shrink-0 items-stretch rounded-md border border-border bg-background" data-history-revision={revision}>
      <Button
        variant="ghost"
        type="button"
        disabled={disabled}
        onClick={() => run(options)}
        className="inline-flex h-8 items-center gap-1 rounded-none border-r border-border px-2 text-xs font-medium text-foreground hover:bg-muted disabled:opacity-50"
        title={title}
      >
        <PrimaryIcon className="text-xs" />
        <span>{label} {runButtonQualifierForOptions(options, context)}</span>
      </Button>
      <DropdownMenu
        align="right"
        menuLabel={`${label} options`}
        menuClassName="max-h-[70vh] w-[320px] max-w-[calc(100vw-24px)] overflow-y-auto"
        trigger={
          <Button
            variant="ghost"
            size="icon"
            type="button"
            disabled={disabled}
            title={`${label} options`}
            aria-label={`${label} options`}
            className="h-8 w-7 rounded-none text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
          >
            <UiChevronDown className="text-xs" />
          </Button>
        }
      >
        {close => (
          <div>
            <TodoRunDropdownContent
              context={context}
              initialAction="run"
              closeParent={close}
              onSelect={(_action, selected) => run(selected)}
              showAdvanced={false}
            />
            {recent.length > 0 && (
              <div className="border-t border-border p-1 text-xs">
                <div className="flex items-center gap-1 px-2 py-1 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">
                  <UiHistory className="text-xs" />
                  Recent configs
                </div>
                {recent.map((item, index) => (
                  <Button
                    key={optionsKey(item)}
                    variant="ghost"
                    type="button"
                    onClick={() => run(item, close)}
                    className="flex h-8 w-full items-center justify-start rounded px-2 text-left text-xs hover:bg-muted"
                  >
                    {index + 1}. {runButtonQualifierForOptions(item, context)}
                    {item.effort ? ` · ${item.effort}` : ''}
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
  const context = useTodoRunContext(open);
  const initialValue = scope === 'verification' ? verificationSpecFromOptions(initial) : cloneOptions(initial);
  const [value, setValue] = useState<AISpecRuntimeValue>(() => initialValue);

  useEffect(() => {
    if (open) setValue(scope === 'verification' ? verificationSpecFromOptions(initial) : cloneOptions(initial));
  }, [open, initial, scope]);

  if (!open) return null;
  const models = context.backends.flatMap(backend => backend.models);
  const verification = scope === 'verification';

  return (
    <Modal open onClose={onClose} title={verification ? 'Verification run options' : 'Approve and run options'} size="2xl">
      <SpecRuntimeEditor
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
            onRun(rememberPromptRunOptions(scope, verificationSpecFromOptions(runtimeSpec), context));
            return;
          }
          const agent = agentForBackend(context, runtimeSpec.backend);
          const selection = driverForSelection(context, agent, runtimeSpec.backend);
          onRun(rememberPromptRunOptions(scope, {
            ...runtimeSpec,
            driver: selection.driver,
            backend: selection.runBackend,
            runMode: 'run',
          }, context));
        }}
        saveLabel={loading ? 'Running…' : verification ? 'Run verification' : 'Approve & run'}
        footerStatus={verification ? 'The persisted fixture supplies the prompt.' : 'Ready to implement the approved plan.'}
      />
    </Modal>
  );
}
