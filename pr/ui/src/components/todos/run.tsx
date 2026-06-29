import { lazy, Suspense, useCallback, useEffect, useRef, useState } from 'react';
import { Button, Combobox, DropdownMenu, Field, Modal, SegmentedControl } from '@flanksource/clicky-ui/components';
import { BudgetSelector, EffortSelector, ModelSelector, ProviderSelector, type ChatBudgetConfig, type ClaudePermissionMode, type ToolMode } from '@flanksource/clicky-ui/chat';
import { ToolPreferences } from '@flanksource/clicky-ui/ai';
import type { TodoRunAgent, TodoRunEffort, TodoRunOptions, TodoRunPreviewResponse, TodoRunResponse } from '../../types';
import { GavelIcon } from '../GavelIcon';
import { inputClass, todoQuery } from './format';
import {
  PROVIDERS,
  backendCatalog as findBackendCatalog,
  backendsForAgent,
  defaultBackendForAgent,
  driverFor,
  providerCatalog,
  runContextWithFallback,
  type RunContext,
  type RunMechanism,
} from './providers';

// RunMode is the prompt the dialog runs: Run (implement), Plan (propose only),
// or Verify (score the committed work against acceptance criteria).
type RunMode = 'run' | 'plan' | 'verify';

// MdxEditorField is the same markdown editor field JsonSchemaForm uses for its
// markdown fields. It lazily pulls in the heavy @mdxeditor/editor, so it is
// code-split and rendered under Suspense with a plain-textarea fallback.
const MdxEditorField = lazy(() =>
  import('@flanksource/clicky-ui/mdx-editor').then(m => ({ default: m.MdxEditorField })),
);

export const defaultRunOptions: TodoRunOptions = { driver: 'claude-cmux', model: 'claude', effort: 'medium' };

type RunPreset = { label: string; icon: string; options: TodoRunOptions };

// The split-button menu offers two actions — Run (implement) and Plan (propose
// a plan without changing code) — each with a Claude and a Codex option, plus
// Advanced for the full dialog.
export const runActionGroups: Array<{ action: 'Run' | 'Plan'; detail: string; presets: RunPreset[] }> = [
  {
    action: 'Run',
    detail: 'implement',
    presets: [
      { label: 'Claude', icon: 'codicon:sparkle', options: { driver: 'claude-cmux', model: 'claude', effort: 'medium' } },
      { label: 'Codex', icon: 'codicon:terminal', options: { driver: 'codex-cmux', model: 'codex', effort: 'medium' } },
    ],
  },
  {
    action: 'Plan',
    detail: 'plan only · no changes',
    presets: [
      { label: 'Claude', icon: 'codicon:sparkle', options: { driver: 'claude-cmux', model: 'claude', effort: 'medium', plan: true } },
      { label: 'Codex', icon: 'codicon:terminal', options: { driver: 'codex-cmux', model: 'codex', effort: 'medium', plan: true } },
    ],
  },
];

// useTodoRun POSTs a run for one or more todo refs in a workspace. A single ref
// runs on its own; multiple refs run together in one agent session (the server
// dispatches them as a combined group). Both the single-todo detail pane and the
// list's multi-select bar drive runs through this one hook.
export function useTodoRun(dir: string, provider: string) {
  const [runBusy, setRunBusy] = useState(false);
  const [runMessage, setRunMessage] = useState('');
  const [runError, setRunError] = useState('');

  const reset = useCallback(() => {
    setRunMessage('');
    setRunError('');
  }, []);

  const run = useCallback(
    async (refs: string[], options: TodoRunOptions = defaultRunOptions): Promise<TodoRunResponse | null> => {
      const cleaned = refs.map(r => r.trim()).filter(Boolean);
      if (cleaned.length === 0 || runBusy) return null;
      setRunBusy(true);
      setRunError('');
      setRunMessage('');
      try {
        // Send `ref` for a single todo (matching the original payload) and `refs`
        // for a multi-select group run.
        const body = cleaned.length === 1 ? { ref: cleaned[0], ...options } : { refs: cleaned, ...options };
        const response = await fetch(`/api/todos/run?${todoQuery(dir, provider)}`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(body),
        });
        const data = await response.json();
        if (!response.ok) throw new Error(data.error || 'Run failed');
        const result = data as TodoRunResponse;
        setRunMessage(result.message || (result.status === 'dry_run' ? 'Todo run validated' : 'Todo run started'));
        return result;
      } catch (err: any) {
        setRunError(err?.message || 'Run failed');
        return null;
      } finally {
        setRunBusy(false);
      }
    },
    [dir, provider, runBusy],
  );

  return { runBusy, runMessage, runError, reset, run };
}

export function TodoRunSplitButton({
  disabled,
  loading,
  label = 'Run',
  icon = 'codicon:play',
  tone = 'default',
  title = 'Run todo',
  onRun,
  onAdvanced,
}: {
  disabled?: boolean;
  loading?: boolean;
  label?: string;
  icon?: string;
  tone?: 'default' | 'danger';
  title?: string;
  onRun: (options?: TodoRunOptions) => void;
  onAdvanced: () => void;
}) {
  const primaryTone = tone === 'danger'
    ? 'text-red-600 hover:bg-red-500/10 hover:text-red-700'
    : 'text-foreground hover:bg-muted';
  return (
    <div className="inline-flex h-8 shrink-0 items-stretch rounded-md border border-border bg-background">
      <Button
        variant="ghost"
        type="button"
        onClick={() => onRun(defaultRunOptions)}
        disabled={disabled}
        title={title}
        className={`inline-flex h-8 items-center gap-1 rounded-none border-r border-border px-2 text-xs font-medium disabled:opacity-50 ${primaryTone}`}
      >
        <GavelIcon name={loading ? 'svg-spinners:ring-resize' : icon} className="text-xs" />
        <span>{label}</span>
      </Button>
      <DropdownMenu
        align="right"
        menuLabel="Run todo"
        menuClassName="w-[280px] max-w-[calc(100vw-24px)]"
        trigger={
          <Button
            variant="ghost"
            size="icon"
            type="button"
            disabled={disabled}
            title="Run options"
            aria-label="Run options"
            className="h-8 w-7 rounded-none text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
          >
            <GavelIcon name="codicon:chevron-down" className="text-xs" />
          </Button>
        }
      >
        {() => (
          <div className="p-1 text-xs">
            {runActionGroups.map(group => (
              <div key={group.action}>
                <div className="px-2 pb-0.5 pt-1.5 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">
                  {group.action}
                </div>
                {group.presets.map(preset => (
                  <Button
                    key={`${group.action}:${preset.label}`}
                    variant="ghost"
                    type="button"
                    onClick={() => onRun(preset.options)}
                    className="flex h-auto w-full items-center justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted"
                  >
                    <GavelIcon name={preset.icon} className="shrink-0 text-sm text-muted-foreground" />
                    <span className="min-w-0 flex-1">
                      <span className="block truncate font-medium text-foreground">{preset.label}</span>
                      <span className="block truncate text-[11px] text-muted-foreground">cmux · {group.detail}</span>
                    </span>
                  </Button>
                ))}
              </div>
            ))}
            <div className="my-1 border-t border-border" />
            <Button
              variant="ghost"
              type="button"
              onClick={onAdvanced}
              className="flex h-auto w-full items-center justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted"
            >
              <GavelIcon name="codicon:settings-gear" className="shrink-0 text-sm text-muted-foreground" />
              <span className="min-w-0 flex-1">
                <span className="block truncate font-medium text-foreground">Advanced</span>
                <span className="block truncate text-[11px] text-muted-foreground">model, effort, timeout, limits</span>
              </span>
            </Button>
          </div>
        )}
      </DropdownMenu>
    </div>
  );
}

export function TodoRunAdvancedDialog({
  open,
  onClose,
  onRun,
  loading,
  title = 'Run todo',
  dir,
  provider,
  refs,
}: {
  open: boolean;
  onClose: () => void;
  onRun: (options: TodoRunOptions) => void;
  loading?: boolean;
  title?: string;
  // dir/provider/refs identify the todo(s) this dialog will run, so it can fetch
  // a live preview of the prompt that will be sent as the options change.
  dir: string;
  provider: string;
  refs: string[];
}) {
  // The driver splits into two picker axes: the provider (claude/codex, the
  // segmented control) and the mechanism (cmux/headless/…). agent is the provider
  // and isCmux gates the cmux-only (plan/resume) vs structured-only (max
  // cost/turns, dirty worktree) fields.
  const [agent, setAgent] = useState<TodoRunAgent>('claude');
  const [mechanism, setMechanism] = useState<RunMechanism>('cmux');
  const [backend, setBackend] = useState('claude-agent');
  const [model, setModel] = useState('claude');
  const [effort, setEffort] = useState<TodoRunEffort>('medium');
  const [mode, setMode] = useState<RunMode>('run');
  const [resume, setResume] = useState(false);
  const [timeout, setTimeoutValue] = useState('30m');
  const [budget, setBudget] = useState<ChatBudgetConfig>({});
  const [dirty, setDirty] = useState(false);
  const [dryRun, setDryRun] = useState(false);
  const [commit, setCommit] = useState(true);
  const [check, setCheck] = useState(false);
  // toolModes scopes each agent tool for the run (enabled/ask/disabled);
  // permissionMode is the base posture. Both feed the run's api.Permissions.
  const [toolModes, setToolModes] = useState<Record<string, ToolMode>>({});
  const [permissionMode, setPermissionMode] = useState<ClaudePermissionMode>('default');
  // promptDraft is the editable prompt body sent as the verbatim override;
  // promptDirty stops the live preview from clobbering the user's edits.
  const [promptDraft, setPromptDraft] = useState('');
  const [promptDirty, setPromptDirty] = useState(false);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [previewError, setPreviewError] = useState('');
  const [verifyBusy, setVerifyBusy] = useState(false);
  const [verifyError, setVerifyError] = useState('');
  const [regenNonce, setRegenNonce] = useState(0);
  const [runContext, setRunContext] = useState<RunContext | null>(null);
  // Ref mirror of promptDirty so the preview effect can read it without
  // refetching on every keystroke.
  const promptDirtyRef = useRef(false);

  const context = runContextWithFallback(runContext);
  const catalog = providerCatalog(agent);
  const isCmux = mechanism === 'cmux';
  const plan = mode === 'plan';
  const isVerify = mode === 'verify';
  const usesCaptainBackend = isVerify || mechanism === 'headless';
  const selectedBackend = findBackendCatalog(context, backend, agent);
  const driver = usesCaptainBackend && !isVerify ? selectedBackend.driver : driverFor(agent, mechanism);
  const models = usesCaptainBackend ? selectedBackend.models : catalog.models;
  const efforts = context.efforts.length > 0 ? context.efforts : catalog.efforts;
  const modelFallback = usesCaptainBackend ? selectedBackend.defaultModel : agent;
  const runBackend = usesCaptainBackend ? selectedBackend.id : undefined;
  const canVerify = refs.length === 1; // verify scores one issue's commits

  // Switching mode re-seeds the editor from the matching preview (Run/Plan share
  // the run body; Verify uses the verify prompt).
  function changeMode(next: RunMode) {
    setMode(next);
    if (next === 'verify') {
      const nextBackend = findBackendCatalog(context, backend, agent);
      setModel(nextBackend.defaultModel);
    } else if (mode === 'verify') {
      setModel(mechanism === 'headless' ? selectedBackend.defaultModel : catalog.defaultModel);
    }
    setPromptDirty(false);
    promptDirtyRef.current = false;
    setVerifyError('');
  }

  function editPrompt(v: string) {
    setPromptDraft(v);
    setPromptDirty(true);
    promptDirtyRef.current = true;
  }

  function regeneratePrompt() {
    setPromptDirty(false);
    promptDirtyRef.current = false;
    setRegenNonce(n => n + 1);
  }

  // Switching provider re-scopes the mechanism/model/effort to what the new
  // provider offers, keeping the current mechanism when it is still valid.
  function changeProvider(next: TodoRunAgent) {
    const nextCatalog = providerCatalog(next);
    const nextMechanism = nextCatalog.mechanisms.some(m => m.value === mechanism) ? mechanism : 'cmux';
    const nextBackend = defaultBackendForAgent(context, next);
    const nextUsesCaptainBackend = mode === 'verify' || nextMechanism === 'headless';
    setAgent(next);
    setMechanism(nextMechanism);
    setBackend(nextBackend.id);
    setModel(nextUsesCaptainBackend ? nextBackend.defaultModel : nextCatalog.defaultModel);
    if (!efforts.includes(effort)) setEffort((efforts[0] || 'medium') as TodoRunEffort);
    if (nextMechanism !== 'cmux' && mode === 'plan') setMode('run');
  }

  function changeMechanism(next: RunMechanism) {
    setMechanism(next);
    if (next === 'headless') {
      const nextBackend = findBackendCatalog(context, backend, agent);
      setModel(nextBackend.defaultModel);
    } else {
      setModel(catalog.defaultModel);
    }
    if (next !== 'cmux' && mode === 'plan') setMode('run');
  }

  function changeBackend(next: string) {
    const nextBackend = findBackendCatalog(context, next, agent);
    setBackend(nextBackend.id);
    setModel(nextBackend.defaultModel);
  }

  function changeEffort(next: string) {
    setEffort((next || 'medium') as TodoRunEffort);
  }

  useEffect(() => {
    if (!open) return;
    setAgent('claude');
    setMechanism('cmux');
    setBackend('claude-agent');
    setModel('claude');
    setEffort('medium');
    setMode('run');
    setResume(false);
    setTimeoutValue('30m');
    setBudget({});
    setDirty(false);
    setDryRun(false);
    setCommit(true);
    setCheck(false);
    setToolModes({});
    setPermissionMode('default');
    setPromptDraft('');
    setPromptDirty(false);
    promptDirtyRef.current = false;
    setVerifyError('');
  }, [open]);

  useEffect(() => {
    if (!open) return;
    let cancelled = false;
    fetch('/api/todos/run/context')
      .then(async res => {
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'Failed to load run context');
        if (!cancelled) setRunContext(data as RunContext);
      })
      .catch(() => {
        if (!cancelled) setRunContext(null);
      });
    return () => {
      cancelled = true;
    };
  }, [open]);

  // refs is a fresh array each render at the call sites, so key the preview fetch
  // on its contents rather than its identity to avoid an endless refetch loop.
  const refsKey = refs.join('\n');

  // Fetch the prompt that will be sent whenever the dialog is open and a
  // prompt-affecting option changes (driver/model/effort/plan/resume). The
  // server builds it from the same code path the run uses, so it matches exactly.
  // Fetch the generated prompt body (Run/Plan) or verify prompt (Verify) and seed
  // the editor with it unless the user has edited it. The server builds it from
  // the same code path the run/verify uses, so it matches what would be sent.
  useEffect(() => {
    if (!open) {
      setPreviewError('');
      return;
    }
    const list = refsKey.split('\n').filter(Boolean);
    if (list.length === 0) {
      setPreviewError('');
      return;
    }
    const url = isVerify
      ? `/api/todos/verify/preview?${todoQuery(dir, provider)}`
      : `/api/todos/run/preview?${todoQuery(dir, provider)}`;
    const body = isVerify
      ? { provider, dir, ref: list[0], backend: runBackend, model: model.trim() || modelFallback }
      : {
          refs: list,
          driver,
          backend: runBackend,
          model: model.trim() || modelFallback,
          effort,
          plan: isCmux ? plan : undefined,
          resume: isCmux ? resume : undefined,
        };

    let cancelled = false;
    const controller = new AbortController();
    setPreviewLoading(true);
    setPreviewError('');
    fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
      signal: controller.signal,
    })
      .then(async res => {
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'Preview failed');
        if (!cancelled && !promptDirtyRef.current) setPromptDraft((data as TodoRunPreviewResponse).prompt ?? '');
      })
      .catch((err: any) => {
        if (!cancelled && err?.name !== 'AbortError') setPreviewError(err?.message || 'Preview failed');
      })
      .finally(() => {
        if (!cancelled) setPreviewLoading(false);
      });
    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [open, dir, provider, refsKey, driver, runBackend, model, modelFallback, effort, plan, resume, isCmux, isVerify, regenNonce]);

  if (!open) return null;

  // runVerify POSTs the (edited) verify prompt to the verification endpoint and
  // closes on success; the parent's todo polling reflects the new status.
  async function runVerify() {
    const list = refsKey.split('\n').filter(Boolean);
    if (list.length === 0) return;
    setVerifyBusy(true);
    setVerifyError('');
    try {
      const res = await fetch(`/api/todos/verify?${todoQuery(dir, provider)}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ provider, dir, ref: list[0], backend: runBackend, model: model.trim() || modelFallback, prompt: promptDraft }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Verify failed');
      onClose();
    } catch (err: any) {
      setVerifyError(err?.message || 'Verify failed');
    } finally {
      setVerifyBusy(false);
    }
  }

  function submit() {
    if (isVerify) {
      void runVerify();
      return;
    }
    // Send the effective per-tool posture (catalog default unless the user
    // overrode it) so the backend sees a complete picture — otherwise a single
    // override would leave the untouched tools unspecified (and brokered). Sent
    // only when the user actually changed something; an untouched run keeps the
    // backend's default posture.
    const toolPreferences =
      Object.keys(toolModes).length > 0
        ? Object.fromEntries([
            ...context.tools.map(t => [t.name, toolModes[t.name] ?? t.defaultMode ?? 'enabled'] as const),
            ...Object.entries(toolModes),
          ])
        : undefined;
    onRun({
      driver,
      backend: runBackend,
      model: model.trim() || modelFallback,
      effort,
      plan: isCmux ? plan : undefined,
      resume: isCmux ? resume : undefined,
      timeout: timeout.trim() || '30m',
      maxCost: !isCmux && budget.cost !== undefined ? budget.cost : undefined,
      maxTurns: !isCmux && budget.maxTokens !== undefined ? budget.maxTokens : undefined,
      dirty: !isCmux ? dirty : undefined,
      dryRun,
      // Plan-only runs make no changes, so there is nothing to auto-commit.
      commit: plan ? false : commit,
      // Plan-only runs make no changes, so the post-completion test/lint check
      // loop has nothing to verify.
      check: plan ? false : check,
      // The edited prompt body is sent verbatim as the override.
      prompt: promptDraft.trim() ? promptDraft : undefined,
      // Per-tool modes and permission posture, sent only when the user changed
      // them from the defaults so a plain run keeps the backend's default posture.
      toolPreferences,
      permissionMode: permissionMode !== 'default' ? permissionMode : undefined,
    });
  }

  const modeOptions: { id: RunMode; label: string }[] = [{ id: 'run', label: 'Run' }];
  if (isCmux) modeOptions.push({ id: 'plan', label: 'Plan' });
  if (canVerify) modeOptions.push({ id: 'verify', label: 'Verify' });
  const backendOptions = backendsForAgent(context, agent).map(item => ({
    value: item.id,
    label: item.configured === false ? `${item.label} (not ready)` : item.label,
  }));

  return (
    <Modal
      open
      onClose={onClose}
      title={title}
      size="md"
      footer={
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={onClose}>Cancel</Button>
          <Button onClick={submit} loading={isVerify ? verifyBusy : loading}>
            {isVerify ? 'Verify' : plan ? 'Plan' : 'Run'}
          </Button>
        </div>
      }
    >
      <div className="space-y-3">
        <Field label="Mode">
          <SegmentedControl
            aria-label="Mode"
            value={mode}
            onChange={v => changeMode(v as RunMode)}
            options={modeOptions}
          />
        </Field>
        <Field label="Agent">
          <ProviderSelector
            ariaLabel="Agent"
            value={agent}
            onChange={changeProvider}
            providers={PROVIDERS.map(p => ({ id: p.id, label: p.label, icon: p.icon }))}
          />
        </Field>
        {isVerify ? (
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <Field label="Backend">
              <Combobox
                ariaLabel="Captain backend"
                value={selectedBackend.id}
                onChange={changeBackend}
                options={backendOptions}
                allowCustomValue={false}
                required
              />
            </Field>
            <Field label="Model">
              <ModelSelector
                models={models}
                value={model}
                onChange={setModel}
                className="w-full"
              />
            </Field>
          </div>
        ) : (
          <>
            <div className={`grid gap-3 ${usesCaptainBackend ? 'grid-cols-1 sm:grid-cols-3' : 'grid-cols-2'}`}>
              <Field label="Type">
                <Combobox
                  ariaLabel="Driver type"
                  value={mechanism}
                  onChange={v => changeMechanism(v as RunMechanism)}
                  options={catalog.mechanisms.map(m => ({ value: m.value, label: m.label }))}
                  allowCustomValue={false}
                  required
                />
              </Field>
              {usesCaptainBackend && (
                <Field label="Backend">
                  <Combobox
                    ariaLabel="Captain backend"
                    value={selectedBackend.id}
                    onChange={changeBackend}
                    options={backendOptions}
                    allowCustomValue={false}
                    required
                  />
                </Field>
              )}
              <Field label="Model">
                <ModelSelector
                  models={models}
                  value={model}
                  onChange={setModel}
                  className="w-full"
                />
              </Field>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <Field label="Effort">
                <EffortSelector
                  efforts={efforts}
                  value={effort}
                  onChange={changeEffort}
                  className="w-full"
                />
              </Field>
              <Field label="Timeout">
                <input className={inputClass} value={timeout} onChange={e => setTimeoutValue(e.currentTarget.value)} />
              </Field>
            </div>
            <Field label="Budget">
              <BudgetSelector
                budget={budget}
                onBudgetChange={setBudget}
                maxTokensLabel="Max turns"
                disabled={isCmux}
              />
            </Field>
            <div className="flex flex-wrap gap-3 text-xs">
              <label className="inline-flex items-center gap-2">
                <input type="checkbox" checked={resume} onChange={e => setResume(e.currentTarget.checked)} disabled={!isCmux} />
                <span>Resume session</span>
              </label>
              <label className="inline-flex items-center gap-2">
                <input type="checkbox" checked={dirty} onChange={e => setDirty(e.currentTarget.checked)} disabled={isCmux} />
                <span>Dirty worktree</span>
              </label>
              <label className="inline-flex items-center gap-2">
                <input type="checkbox" checked={commit && !plan} onChange={e => setCommit(e.currentTarget.checked)} disabled={plan} />
                <span>Auto-commit</span>
              </label>
              <label className="inline-flex items-center gap-2" title="Run the configured test/lint checks after the agent finishes and feed failures back to it">
                <input type="checkbox" checked={check && !plan} onChange={e => setCheck(e.currentTarget.checked)} disabled={plan} />
                <span>Run checks (test/lint)</span>
              </label>
              <label className="inline-flex items-center gap-2">
                <input type="checkbox" checked={dryRun} onChange={e => setDryRun(e.currentTarget.checked)} />
                <span>Dry run</span>
              </label>
            </div>
            <Field label="Tools">
              <ToolPreferences
                tools={context.tools}
                value={toolModes}
                onChange={setToolModes}
                permissionMode={permissionMode}
                onPermissionModeChange={setPermissionMode}
              />
            </Field>
          </>
        )}
        <div className="space-y-1">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-muted-foreground">{isVerify ? 'Verify prompt' : 'Prompt'}</span>
            <div className="flex items-center gap-2">
              {previewLoading && <GavelIcon name="svg-spinners:ring-resize" className="text-xs text-muted-foreground" />}
              <Button
                variant="ghost"
                type="button"
                onClick={regeneratePrompt}
                disabled={previewLoading}
                title="Discard edits and regenerate from the options above"
                className="h-auto rounded px-1.5 py-0.5 text-[11px] text-muted-foreground hover:bg-muted hover:text-foreground"
              >
                Regenerate
              </Button>
            </div>
          </div>
          {(previewError || verifyError) && <div className="text-xs text-red-600">{previewError || verifyError}</div>}
          <Suspense
            fallback={
              <textarea
                className={`${inputClass} h-auto min-h-[16rem] resize-y font-mono`}
                value={promptDraft}
                onChange={e => editPrompt(e.currentTarget.value)}
                placeholder={previewLoading ? 'Loading prompt…' : 'Prompt'}
              />
            }
          >
            <MdxEditorField
              value={promptDraft}
              onChange={editPrompt}
              placeholder={previewLoading ? 'Loading prompt…' : 'Prompt'}
              className="min-h-[16rem]"
            />
          </Suspense>
          {promptDirty && <div className="text-[11px] text-muted-foreground">Edited — sent verbatim as the prompt.</div>}
        </div>
      </div>
    </Modal>
  );
}
