import { lazy, Suspense, useCallback, useEffect, useRef, useState } from 'react';
import { Button, Combobox, DropdownMenu, Field, Modal, SegmentedControl } from '@flanksource/clicky-ui/components';
import type { TodoRunAgent, TodoRunEffort, TodoRunOptions, TodoRunPreviewResponse, TodoRunResponse } from '../../types';
import { GavelIcon } from '../GavelIcon';
import { inputClass, todoQuery } from './format';
import { PROVIDERS, providerCatalog, driverFor, type RunMechanism } from './providers';

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
  const [model, setModel] = useState('claude');
  const [effort, setEffort] = useState<TodoRunEffort>('medium');
  const [mode, setMode] = useState<RunMode>('run');
  const [resume, setResume] = useState(false);
  const [timeout, setTimeoutValue] = useState('30m');
  const [maxCost, setMaxCost] = useState('');
  const [maxTurns, setMaxTurns] = useState('');
  const [dirty, setDirty] = useState(false);
  const [dryRun, setDryRun] = useState(false);
  const [commit, setCommit] = useState(true);
  const [check, setCheck] = useState(false);
  // promptDraft is the editable prompt body sent as the verbatim override;
  // promptDirty stops the live preview from clobbering the user's edits.
  const [promptDraft, setPromptDraft] = useState('');
  const [promptDirty, setPromptDirty] = useState(false);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [previewError, setPreviewError] = useState('');
  const [verifyBusy, setVerifyBusy] = useState(false);
  const [verifyError, setVerifyError] = useState('');
  const [regenNonce, setRegenNonce] = useState(0);
  // Ref mirror of promptDirty so the preview effect can read it without
  // refetching on every keystroke.
  const promptDirtyRef = useRef(false);

  const catalog = providerCatalog(agent);
  const driver = driverFor(agent, mechanism);
  const isCmux = mechanism === 'cmux';
  const plan = mode === 'plan';
  const isVerify = mode === 'verify';
  const canVerify = refs.length === 1; // verify scores one issue's commits

  // Switching mode re-seeds the editor from the matching preview (Run/Plan share
  // the run body; Verify uses the verify prompt).
  function changeMode(next: RunMode) {
    setMode(next);
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
    setAgent(next);
    setMechanism(nextMechanism);
    setModel(nextCatalog.defaultModel);
    if (!nextCatalog.efforts.includes(effort)) setEffort(nextCatalog.efforts[0]);
    if (nextMechanism !== 'cmux' && mode === 'plan') setMode('run');
  }

  function changeMechanism(next: RunMechanism) {
    setMechanism(next);
    if (next !== 'cmux' && mode === 'plan') setMode('run');
  }

  useEffect(() => {
    if (!open) return;
    setAgent('claude');
    setMechanism('cmux');
    setModel('claude');
    setEffort('medium');
    setMode('run');
    setResume(false);
    setTimeoutValue('30m');
    setMaxCost('');
    setMaxTurns('');
    setDirty(false);
    setDryRun(false);
    setCommit(true);
    setCheck(false);
    setPromptDraft('');
    setPromptDirty(false);
    promptDirtyRef.current = false;
    setVerifyError('');
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
      ? { provider, dir, ref: list[0], model: model.trim() || agent }
      : { refs: list, driver, model: model.trim() || agent, effort, plan: isCmux ? plan : undefined, resume: isCmux ? resume : undefined };

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
  }, [open, dir, provider, refsKey, driver, agent, model, effort, plan, resume, isCmux, isVerify, regenNonce]);

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
        body: JSON.stringify({ provider, dir, ref: list[0], model: model.trim() || agent, prompt: promptDraft }),
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
    const cost = Number(maxCost);
    const turns = Number.parseInt(maxTurns, 10);
    onRun({
      driver,
      model: model.trim() || agent,
      effort,
      plan: isCmux ? plan : undefined,
      resume: isCmux ? resume : undefined,
      timeout: timeout.trim() || '30m',
      maxCost: !isCmux && maxCost.trim() && Number.isFinite(cost) ? cost : undefined,
      maxTurns: !isCmux && maxTurns.trim() && Number.isFinite(turns) ? turns : undefined,
      dirty: !isCmux ? dirty : undefined,
      dryRun,
      // Plan-only runs make no changes, so there is nothing to auto-commit.
      commit: plan ? false : commit,
      // Plan-only runs make no changes, so the post-completion test/lint check
      // loop has nothing to verify.
      check: plan ? false : check,
      // The edited prompt body is sent verbatim as the override.
      prompt: promptDraft.trim() ? promptDraft : undefined,
    });
  }

  const modeOptions: { id: RunMode; label: string }[] = [{ id: 'run', label: 'Run' }];
  if (isCmux) modeOptions.push({ id: 'plan', label: 'Plan' });
  if (canVerify) modeOptions.push({ id: 'verify', label: 'Verify' });

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
        <Field label="Provider">
          <SegmentedControl
            aria-label="Provider"
            value={agent}
            onChange={changeProvider}
            options={PROVIDERS.map(p => ({ id: p.id, label: p.label, icon: p.icon }))}
          />
        </Field>
        {isVerify ? (
          <Field label="Model">
            <Combobox
              ariaLabel="Model"
              value={model}
              onChange={setModel}
              options={catalog.models}
              placeholder={catalog.defaultModel}
            />
          </Field>
        ) : (
          <>
            <div className="grid grid-cols-2 gap-3">
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
              <Field label="Model">
                <Combobox
                  ariaLabel="Model"
                  value={model}
                  onChange={setModel}
                  options={catalog.models}
                  placeholder={catalog.defaultModel}
                />
              </Field>
            </div>
            <div className="grid grid-cols-3 gap-3">
              <Field label="Effort">
                <Combobox
                  ariaLabel="Effort"
                  value={effort}
                  onChange={v => setEffort(v as TodoRunEffort)}
                  options={catalog.efforts.map(e => ({ value: e, label: e }))}
                  allowCustomValue={false}
                  required
                />
              </Field>
              <Field label="Timeout">
                <input className={inputClass} value={timeout} onChange={e => setTimeoutValue(e.currentTarget.value)} />
              </Field>
              <Field label="Max turns">
                <input className={inputClass} type="number" min="0" value={maxTurns} onChange={e => setMaxTurns(e.currentTarget.value)} disabled={isCmux} />
              </Field>
            </div>
            <Field label="Max cost">
              <input className={inputClass} type="number" min="0" step="0.01" value={maxCost} onChange={e => setMaxCost(e.currentTarget.value)} disabled={isCmux} />
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
