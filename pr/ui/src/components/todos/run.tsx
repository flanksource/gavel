import { lazy, Suspense, useCallback, useEffect, useRef, useState, type ComponentType } from "react";
import { Button, Combobox, DropdownMenu, Field, Modal, SegmentedControl } from "@flanksource/clicky-ui/components";
import { ModelSelector, ProviderSelector } from "@flanksource/clicky-ui/chat";
import { PromptRunEditor, promptRuntimeValueToPayload, type AISpecRuntimeValue } from "@flanksource/clicky-ui/ai";
import { UiChevronDown, UiCog, UiPlay, UiSparkles, UiTerminal, type IconProps } from "@flanksource/clicky-ui/icons";
import type { TodoRunAgent, TodoRunOptions, TodoRunPreviewResponse, TodoRunResponse } from "../../types";
import { Spinner } from "../../icons/Spinner";
import { inputClass, todoQuery } from "./format";
import {
  PROVIDERS,
  agentForBackend,
  backendCatalog as findBackendCatalog,
  backendsForAgent,
  buildRunFamilies,
  defaultBackendForAgent,
  defaultModelForSelection,
  driverForSelection,
  isCmuxBackend,
  modelsForSelection,
  runContextWithFallback,
  type RunContext,
} from "./providers";

// RunMode is the prompt the dialog runs: Run (implement), Plan (propose only),
// or Verify (score the committed work against acceptance criteria).
type RunMode = "run" | "plan" | "verify";

// The spec editor exposes exactly what gavel's dispatch reads: model/effort/budget,
// the prompt override, tool/permission posture, plus the run's Workspace (dirty
// worktree), Verify (checks), and Commit (auto-commit / dry-run). The last three
// replace the old loose checkboxes now that those options live on the api.Spec.
const RUN_SPEC_SECTIONS = ["model", "prompt", "permissions", "workspace", "verify", "commit"] as const;

// Runs auto-commit by default — the old commit=true default, now expressed as the
// spec's Workflow.PostRun.commit. The advanced dialog's Commit section can turn it
// off; a plan-only run never commits (the server suppresses commit in plan mode).
const AUTO_COMMIT: Pick<TodoRunOptions, "workflow"> = { workflow: { postRun: { commit: true } } };

// MdxEditorField is the same markdown editor field JsonSchemaForm uses for its
// markdown fields. It lazily pulls in the heavy @mdxeditor/editor, so it is
// code-split and rendered under Suspense with a plain-textarea fallback.
const MdxEditorField = lazy(() => import("@flanksource/clicky-ui/mdx-editor").then((m) => ({ default: m.MdxEditorField })));

export const defaultRunOptions: TodoRunOptions = { driver: "claude-cmux", model: "claude", effort: "medium", ...AUTO_COMMIT };

type RunPreset = { label: string; icon: ComponentType<IconProps>; options: TodoRunOptions };

// The split-button menu offers two actions — Run (implement) and Plan (propose
// a plan without changing code) — each with a Claude and a Codex option, plus
// Advanced for the full dialog.
export const runActionGroups: Array<{ action: "Run" | "Plan"; detail: string; presets: RunPreset[] }> = [
  {
    action: "Run",
    detail: "implement",
    presets: [
      { label: "Claude", icon: UiSparkles, options: { driver: "claude-cmux", model: "claude", effort: "medium", ...AUTO_COMMIT } },
      { label: "Codex", icon: UiTerminal, options: { driver: "codex-cmux", model: "codex", effort: "medium", ...AUTO_COMMIT } },
    ],
  },
  {
    action: "Plan",
    detail: "plan only · no changes",
    presets: [
      { label: "Claude", icon: UiSparkles, options: { driver: "claude-cmux", model: "claude", effort: "medium", plan: true } },
      { label: "Codex", icon: UiTerminal, options: { driver: "codex-cmux", model: "codex", effort: "medium", plan: true } },
    ],
  },
];

// useTodoRun POSTs a run for one or more todo refs in a workspace. A single ref
// runs on its own; multiple refs run together in one agent session (the server
// dispatches them as a combined group). Both the single-todo detail pane and the
// list's multi-select bar drive runs through this one hook.
export function useTodoRun(dir: string, provider: string) {
  const [runBusy, setRunBusy] = useState(false);
  const [runMessage, setRunMessage] = useState("");
  const [runError, setRunError] = useState("");

  const reset = useCallback(() => {
    setRunMessage("");
    setRunError("");
  }, []);

  const run = useCallback(
    async (refs: string[], options: TodoRunOptions = defaultRunOptions): Promise<TodoRunResponse | null> => {
      const cleaned = refs.map((r) => r.trim()).filter(Boolean);
      if (cleaned.length === 0 || runBusy) return null;
      setRunBusy(true);
      setRunError("");
      setRunMessage("");
      try {
        // Send `ref` for a single todo (matching the original payload) and `refs`
        // for a multi-select group run.
        const body = cleaned.length === 1 ? { ref: cleaned[0], ...options } : { refs: cleaned, ...options };
        const response = await fetch(`/api/todos/run?${todoQuery(dir, provider)}`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(body),
        });
        const data = await response.json();
        if (!response.ok) throw new Error(data.error || "Run failed");
        const result = data as TodoRunResponse;
        setRunMessage(result.message || (result.status === "dry_run" ? "Todo run validated" : "Todo run started"));
        return result;
      } catch (err: any) {
        setRunError(err?.message || "Run failed");
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
  label = "Run",
  icon = UiPlay,
  tone = "default",
  title = "Run todo",
  onRun,
  onAdvanced,
}: {
  disabled?: boolean;
  loading?: boolean;
  label?: string;
  icon?: ComponentType<IconProps>;
  tone?: "default" | "danger";
  title?: string;
  onRun: (options?: TodoRunOptions) => void;
  onAdvanced: () => void;
}) {
  const primaryTone = tone === "danger" ? "text-red-600 hover:bg-red-500/10 hover:text-red-700" : "text-foreground hover:bg-muted";
  const PrimaryIcon = loading ? Spinner : icon;
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
        <PrimaryIcon className="text-xs" />
        <span>{label}</span>
      </Button>
      <DropdownMenu
        align="right"
        menuLabel="Run todo"
        menuClassName="w-[280px] max-w-[calc(100vw-24px)]"
        trigger={
          <Button variant="ghost" size="icon" type="button" disabled={disabled} title="Run options" aria-label="Run options" className="h-8 w-7 rounded-none text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50">
            <UiChevronDown className="text-xs" />
          </Button>
        }
      >
        {() => (
          <div className="p-1 text-xs">
            {runActionGroups.map((group) => (
              <div key={group.action}>
                <div className="px-2 pb-0.5 pt-1.5 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">{group.action}</div>
                {group.presets.map((preset) => {
                  const Icon = preset.icon;
                  return (
                    <Button key={`${group.action}:${preset.label}`} variant="ghost" type="button" onClick={() => onRun(preset.options)} className="flex h-auto w-full items-center justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted">
                      <Icon className="shrink-0 text-sm text-muted-foreground" />
                      <span className="min-w-0 flex-1">
                        <span className="block truncate font-medium text-foreground">{preset.label}</span>
                        <span className="block truncate text-[11px] text-muted-foreground">cmux · {group.detail}</span>
                      </span>
                    </Button>
                  );
                })}
              </div>
            ))}
            <div className="my-1 border-t border-border" />
            <Button variant="ghost" type="button" onClick={onAdvanced} className="flex h-auto w-full items-center justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted">
              <UiCog className="shrink-0 text-sm text-muted-foreground" />
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

const INITIAL_RUNTIME_VALUE: AISpecRuntimeValue = { backend: "claude-cmux", model: "claude", effort: "medium", ...AUTO_COMMIT };
const INITIAL_VERIFY_BACKEND = "claude-agent";
const INITIAL_VERIFY_MODEL = "claude-agent-sonnet";

export function TodoRunAdvancedDialog({
  open,
  onClose,
  onRun,
  loading,
  title = "Run todo",
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
  // Run/Plan share one AISpecRuntimeValue (model/backend/effort/budget/
  // permissions/prompt), edited via clicky's PromptRunEditor. Verify only ever
  // needs a captain backend + model (its wire body has no effort/budget/
  // permissions), so it keeps its own small, independent selection instead of
  // sharing — and instead of showing Effort/Budget/Permissions controls that
  // verify's request would silently ignore.
  const [runtimeValue, setRuntimeValue] = useState<AISpecRuntimeValue>(INITIAL_RUNTIME_VALUE);
  const [mode, setMode] = useState<RunMode>("run");
  // Resume stays a discrete toggle (session-identity decision, cmux only); dirty,
  // auto-commit, dry-run, and checks now live on runtimeValue's spec (Workspace/
  // Commit/Verify sections), not as parallel booleans.
  const [resume, setResume] = useState(false);
  const [verifyAgent, setVerifyAgent] = useState<TodoRunAgent>("claude");
  const [verifyBackend, setVerifyBackend] = useState(INITIAL_VERIFY_BACKEND);
  const [verifyModel, setVerifyModel] = useState(INITIAL_VERIFY_MODEL);
  // promptDraft is the editable prompt body sent as the verbatim override;
  // promptDirty stops the live preview from clobbering the user's edits.
  const [promptDraft, setPromptDraft] = useState("");
  const [promptDirty, setPromptDirty] = useState(false);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [previewError, setPreviewError] = useState("");
  const [verifyBusy, setVerifyBusy] = useState(false);
  const [verifyError, setVerifyError] = useState("");
  const [regenNonce, setRegenNonce] = useState(0);
  const [runContext, setRunContext] = useState<RunContext | null>(null);
  // Ref mirror of promptDirty so the preview effect can read it without
  // refetching on every keystroke.
  const promptDirtyRef = useRef(false);

  const context = runContextWithFallback(runContext);
  const families = buildRunFamilies(context);
  const agent = agentForBackend(context, runtimeValue.backend);
  const isCmux = isCmuxBackend(agent, runtimeValue.backend);
  const activeModels = modelsForSelection(context, agent, runtimeValue.backend);
  const modelFallback = defaultModelForSelection(context, agent, runtimeValue.backend);
  const { driver, runBackend } = driverForSelection(context, agent, runtimeValue.backend);
  const plan = mode === "plan";
  const isVerify = mode === "verify";
  const canVerify = refs.length === 1; // verify scores one issue's commits
  const verifyBackendCatalog = findBackendCatalog(context, verifyBackend, verifyAgent);
  const verifyModelFallback = verifyBackendCatalog.defaultModel;

  // A backend switch (cmux <-> a captain backend, or a family switch) can leave
  // a model id that no longer belongs to the new mode (cmux uses bare ids like
  // "opus"; captain backends use prefixed ids like "claude-agent-opus") — reset
  // to the new mode's default in that case, mirroring the old changeMechanism/
  // changeProvider/changeBackend resets.
  function changeRuntime(next: AISpecRuntimeValue) {
    const prevBackend = runtimeValue.backend ?? "";
    const nextBackend = next.backend ?? "";
    if (nextBackend === prevBackend) {
      setRuntimeValue(next);
      return;
    }
    const nextAgent = agentForBackend(context, nextBackend);
    const candidates = modelsForSelection(context, nextAgent, nextBackend);
    const modelStillValid = !!next.model && candidates.some((m) => m.id === next.model);
    setRuntimeValue({
      ...next,
      model: modelStillValid ? next.model : defaultModelForSelection(context, nextAgent, nextBackend),
    });
    // Plan only runs in cmux; leaving cmux drops back to Run, same as before.
    if (!isCmuxBackend(nextAgent, nextBackend) && mode === "plan") setMode("run");
  }

  function changeMode(next: RunMode) {
    setMode(next);
    setPromptDirty(false);
    promptDirtyRef.current = false;
    setVerifyError("");
  }

  function editPrompt(v: string) {
    setPromptDraft(v);
    setPromptDirty(true);
    promptDirtyRef.current = true;
  }

  function regeneratePrompt() {
    setPromptDirty(false);
    promptDirtyRef.current = false;
    setRegenNonce((n) => n + 1);
  }

  function changeVerifyAgent(next: TodoRunAgent) {
    const nextBackend = defaultBackendForAgent(context, next);
    setVerifyAgent(next);
    setVerifyBackend(nextBackend.id);
    setVerifyModel(nextBackend.defaultModel);
  }

  function changeVerifyBackend(next: string) {
    const nextBackend = findBackendCatalog(context, next, verifyAgent);
    setVerifyBackend(nextBackend.id);
    setVerifyModel(nextBackend.defaultModel);
  }

  useEffect(() => {
    if (!open) return;
    setRuntimeValue(INITIAL_RUNTIME_VALUE);
    setMode("run");
    setResume(false);
    setVerifyAgent("claude");
    setVerifyBackend(INITIAL_VERIFY_BACKEND);
    setVerifyModel(INITIAL_VERIFY_MODEL);
    setPromptDraft("");
    setPromptDirty(false);
    promptDirtyRef.current = false;
    setVerifyError("");
  }, [open]);

  useEffect(() => {
    if (!open) return;
    let cancelled = false;
    fetch("/api/todos/run/context")
      .then(async (res) => {
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || "Failed to load run context");
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
  const refsKey = refs.join("\n");
  const previewModel = isVerify ? verifyModel.trim() || verifyModelFallback : runtimeValue.model?.trim() || modelFallback;
  const previewBackend = isVerify ? verifyBackend : runBackend;

  // Fetch the prompt that will be sent whenever the dialog is open and a
  // prompt-affecting option changes (driver/model/effort/plan/resume). The
  // server builds it from the same code path the run uses, so it matches exactly.
  // Fetch the generated prompt body (Run/Plan) or verify prompt (Verify) and seed
  // the editor with it unless the user has edited it. The server builds it from
  // the same code path the run/verify uses, so it matches what would be sent.
  useEffect(() => {
    if (!open) {
      setPreviewError("");
      return;
    }
    const list = refsKey.split("\n").filter(Boolean);
    if (list.length === 0) {
      setPreviewError("");
      return;
    }
    const url = isVerify ? `/api/todos/verify/preview?${todoQuery(dir, provider)}` : `/api/todos/run/preview?${todoQuery(dir, provider)}`;
    const body = isVerify
      ? { provider, dir, ref: list[0], backend: previewBackend, model: previewModel }
      : {
          refs: list,
          driver,
          backend: previewBackend,
          model: previewModel,
          effort: runtimeValue.effort,
          plan: isCmux ? plan : undefined,
          resume: isCmux ? resume : undefined,
        };

    let cancelled = false;
    const controller = new AbortController();
    setPreviewLoading(true);
    setPreviewError("");
    fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
      signal: controller.signal,
    })
      .then(async (res) => {
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || "Preview failed");
        if (!cancelled && !promptDirtyRef.current) setPromptDraft((data as TodoRunPreviewResponse).prompt ?? "");
      })
      .catch((err: any) => {
        if (!cancelled && err?.name !== "AbortError") setPreviewError(err?.message || "Preview failed");
      })
      .finally(() => {
        if (!cancelled) setPreviewLoading(false);
      });
    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [open, dir, provider, refsKey, driver, previewBackend, previewModel, runtimeValue.effort, plan, resume, isCmux, isVerify, regenNonce]);

  if (!open) return null;

  // runVerify POSTs the (edited) verify prompt to the verification endpoint and
  // closes on success; the parent's todo polling reflects the new status.
  async function runVerify() {
    const list = refsKey.split("\n").filter(Boolean);
    if (list.length === 0) return;
    setVerifyBusy(true);
    setVerifyError("");
    try {
      const res = await fetch(`/api/todos/verify?${todoQuery(dir, provider)}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ provider, dir, ref: list[0], backend: verifyBackend, model: previewModel, prompt: promptDraft }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || "Verify failed");
      onClose();
    } catch (err: any) {
      setVerifyError(err?.message || "Verify failed");
    } finally {
      setVerifyBusy(false);
    }
  }

  function submit() {
    if (isVerify) {
      void runVerify();
      return;
    }
    // spec carries the model/effort/permissions plus the run's setup (dirty),
    // workflow.verify (checks) and workflow.postRun (auto-commit / dry-run) — all
    // edited via the spec editor's Workspace/Verify/Commit sections. Plan-only runs
    // never commit or verify; the server suppresses both for run mode plan.
    const { spec } = promptRuntimeValueToPayload(runtimeValue);
    onRun({
      ...spec,
      driver,
      backend: runBackend,
      plan: isCmux ? plan : undefined,
      resume: isCmux ? resume : undefined,
      prompt: {
        // The edited prompt body is sent verbatim as the override.
        user: promptDraft.trim() ? promptDraft : undefined,
      },
    });
  }

  const modeOptions: { id: RunMode; label: string }[] = [{ id: "run", label: "Run" }];
  if (isCmux) modeOptions.push({ id: "plan", label: "Plan" });
  if (canVerify) modeOptions.push({ id: "verify", label: "Verify" });
  const verifyBackendOptions = backendsForAgent(context, verifyAgent).map((item) => ({
    value: item.id,
    label: item.configured === false ? `${item.label} (not ready)` : item.label,
  }));

  // Shared regenerate/error/editor/dirty-note block: passed to PromptRunEditor's
  // `promptEditor` slot for Run/Plan (it supplies the "Prompt" block title), and
  // rendered under a matching manual title for Verify (which doesn't use
  // PromptRunEditor, since its wire body has no effort/budget/permissions to edit).
  const promptEditorNode = (
    <div className="space-y-1">
      <div className="flex items-center justify-end gap-2">
        {previewLoading && <Spinner className="text-xs text-muted-foreground" />}
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
      {previewError && <div className="text-xs text-red-600">{previewError}</div>}
      <Suspense fallback={<textarea className={`${inputClass} h-auto min-h-[16rem] resize-y font-mono`} value={promptDraft} onChange={(e) => editPrompt(e.currentTarget.value)} placeholder={previewLoading ? "Loading prompt…" : "Prompt"} />}>
        <MdxEditorField value={promptDraft} onChange={editPrompt} placeholder={previewLoading ? "Loading prompt…" : "Prompt"} className="min-h-[16rem]" />
      </Suspense>
      {promptDirty && <div className="text-[11px] text-muted-foreground">Edited — sent verbatim as the prompt.</div>}
    </div>
  );

  return (
    <Modal
      open
      onClose={onClose}
      title={title}
      size="2xl"
      footer={
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={submit} loading={isVerify ? verifyBusy : loading}>
            {isVerify ? "Verify" : plan ? "Plan" : "Run"}
          </Button>
        </div>
      }
    >
      <div className="space-y-3">
        <Field label="Mode">
          <SegmentedControl aria-label="Mode" value={mode} onChange={(v) => changeMode(v as RunMode)} options={modeOptions} />
        </Field>
        {isVerify ? (
          <>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
              <Field label="Agent">
                <ProviderSelector ariaLabel="Agent" value={verifyAgent} onChange={changeVerifyAgent} providers={PROVIDERS.map((p) => ({ id: p.id, label: p.label, icon: p.icon }))} />
              </Field>
              <Field label="Backend">
                <Combobox ariaLabel="Captain backend" value={verifyBackendCatalog.id} onChange={changeVerifyBackend} options={verifyBackendOptions} allowCustomValue={false} required />
              </Field>
              <Field label="Model">
                <ModelSelector models={verifyBackendCatalog.models} value={verifyModel} onChange={setVerifyModel} className="w-full" />
              </Field>
            </div>
            <div className="space-y-1">
              <div className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Verify prompt</div>
              {promptEditorNode}
              {verifyError && <div className="text-xs text-red-600">{verifyError}</div>}
            </div>
          </>
        ) : (
          <PromptRunEditor value={runtimeValue} onChange={changeRuntime} models={activeModels} families={families} tools={context.tools} specSections={RUN_SPEC_SECTIONS} promptEditor={promptEditorNode} promptLabel="Prompt">
            {isCmux && (
              // Resume is the one run-orchestration toggle without a spec home: it
              // continues the todo's prior claude session (--resume) rather than
              // minting a fresh one, and only applies to cmux runs.
              <label className="inline-flex items-center gap-2 text-xs">
                <input type="checkbox" checked={resume} onChange={(e) => setResume(e.currentTarget.checked)} />
                <span>Resume session</span>
              </label>
            )}
          </PromptRunEditor>
        )}
      </div>
    </Modal>
  );
}
