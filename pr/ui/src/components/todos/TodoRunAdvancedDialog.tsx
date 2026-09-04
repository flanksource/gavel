import { lazy, Suspense, useEffect, useRef, useState } from "react";
import { Button, Field, Modal, SegmentedControl, Tabs } from "@flanksource/clicky-ui/components";
import { CodeBlock } from "@flanksource/clicky-ui/data";
import { PromptRunEditor, promptRuntimeValueToPayload, type AIPromptRunValue, type AISpecRuntimeValue } from "@flanksource/clicky-ui/ai";
import type { TodoRunOptions } from "../../types";
import { Spinner } from "../../icons/Spinner";
import { inputClass } from "./format";
import {
  loadLastTodoRunOptions,
  loadRecentAdvancedTodoRunOptions,
  normalizeRunOptions,
  reconcileTodoRunOptions,
  runChoiceDetail,
  runOptionsKey,
  runSpec,
  TODO_RUN_ACTIONS,
  TodoRunContextError,
  type TodoRunAction,
  type TodoRunRequestPayload,
  useTodoRunContext,
  useTodoRunPreview,
} from "./run";
import {
  agentForRuntime,
  buildRunFamilies,
  defaultModelForSelection,
  isCmuxMode,
  type RunContext,
} from "./providers";

const RUN_SPEC_SECTIONS = ["model", "prompt", "permissions", "workspace", "verify", "commit"] as const;
const INITIAL_RUNTIME_VALUE: AISpecRuntimeValue = { effort: "medium" };
const MdxEditorField = lazy(() => import("@flanksource/clicky-ui/mdx-editor").then((module) => ({ default: module.MdxEditorField })));

// isKnownRunAction is true for the three built-in behaviour classes, which
// still carry client-side bookkeeping (storage, recent-config buttons) keyed
// by TodoRunAction. A project-declared custom lifecycle step has no such
// bucket — it dispatches by name only (its `.step` reaches requestStepFor
// directly) and starts from a fresh runtime seeded off promptDefaults[step].
function isKnownRunAction(step: string): step is TodoRunAction {
  return (TODO_RUN_ACTIONS as readonly string[]).includes(step);
}

function seedRuntimeForStep(step: string, context: RunContext): AISpecRuntimeValue {
  if (isKnownRunAction(step)) {
    return runSpec(reconcileTodoRunOptions(step, loadLastTodoRunOptions(step, context), context));
  }
  const promptDefault = context.promptDefaults?.[step] ?? {};
  const agent = agentForRuntime(context, promptDefault.mode, promptDefault.model);
  const mode = promptDefault.mode || context.defaultMode || "";
  const model = promptDefault.model || defaultModelForSelection(context, agent, mode);
  return { effort: "medium", mode, model };
}

// initialStepFor picks the step the picker opens on: a step the caller forced
// (one of the Run/Plan buttons) outranks everything, then the todo's own
// server-computed suggestion (TodoLifecycle.next), then the first step the
// project's lifecycle declares.
function initialStepFor(initialMode: string | undefined, nextStep: string | null | undefined, context: RunContext | null): string {
  if (initialMode) return initialMode;
  if (nextStep) return nextStep;
  return context?.lifecycle.steps[0]?.name ?? "run";
}

export function TodoRunAdvancedDialog({
  open,
  onClose,
  onRun,
  loading,
  title = "Run todo",
  initialMode,
  nextStep = null,
  dir,
  refID,
}: {
  open: boolean;
  onClose: () => void;
  onRun: (options: TodoRunOptions) => void;
  loading?: boolean;
  title?: string;
  // A specific lifecycle step name the caller wants preselected (e.g. the
  // Run/Plan buttons' own advanced-options cog). Omit to let the todo's own
  // suggestion (nextStep) or the project's first declared step decide.
  initialMode?: string;
  nextStep?: string | null;
  dir: string;
  refID: string;
}) {
  const [runRequest, setRunRequest] = useState<AIPromptRunValue>({ spec: INITIAL_RUNTIME_VALUE });
  const runtimeValue = runRequest.spec ?? {};
  const [step, setStep] = useState<string>(initialMode ?? nextStep ?? "run");
  const [resume, setResume] = useState(false);
  const [promptDraft, setPromptDraft] = useState("");
  const [promptDirty, setPromptDirty] = useState(false);
  const [previewError, setPreviewError] = useState("");
  const [specYAML, setSpecYAML] = useState("");
  const [view, setView] = useState<"form" | "yaml">("form");
  const [regenNonce, setRegenNonce] = useState(0);
  const { context, loading: contextLoading, error: contextError } = useTodoRunContext(open);
  const previewMutation = useTodoRunPreview(dir);
  const promptDirtyRef = useRef(false);

  const steps = context?.lifecycle.steps ?? [];
  const selectedStep = steps.find((item) => item.name === step);
  const submitLabel = selectedStep?.label ?? step;
  const families = context ? buildRunFamilies(context) : [];
  const agent = context ? agentForRuntime(context, runtimeValue.mode, runtimeValue.model) : undefined;
  const isCmux = isCmuxMode(runtimeValue.mode);
  const activeModels = context?.models ?? [];
  const modelFallback = context && agent ? defaultModelForSelection(context, agent, runtimeValue.mode) : "";
  const recentAdvanced = context && isKnownRunAction(step) ? loadRecentAdvancedTodoRunOptions(step, context) : [];

  function changeStep(next: string) {
    setStep(next);
    setPromptDirty(false);
    promptDirtyRef.current = false;
  }

  function editPrompt(value: string) {
    setPromptDraft(value);
    setPromptDirty(true);
    promptDirtyRef.current = true;
  }

  function regeneratePrompt() {
    setPromptDirty(false);
    promptDirtyRef.current = false;
    setRegenNonce((value) => value + 1);
  }

  useEffect(() => {
    if (!open) return;
    setResume(false);
    setPromptDraft("");
    setPromptDirty(false);
    setSpecYAML("");
    setView("form");
    promptDirtyRef.current = false;
  }, [open, initialMode]);

  useEffect(() => {
    if (!open || !context) return;
    const resolved = initialStepFor(initialMode, nextStep, context);
    setStep(resolved);
    setRunRequest({ spec: seedRuntimeForStep(resolved, context) });
  }, [open, initialMode, nextStep, context]);

  const previewModel = runtimeValue.model?.trim() || modelFallback;

  function buildRequestPayload(): TodoRunRequestPayload {
    const { spec } = promptRuntimeValueToPayload({ ...runtimeValue, model: previewModel });
    const prompt = promptDirty ? { ...spec.prompt, user: promptDraft } : spec.prompt;
    return { ref: refID, step, spec: { ...spec, prompt }, resume: (isCmux && resume) || undefined };
  }

  useEffect(() => {
    if (!open) {
      setPreviewError("");
      return;
    }
    if (!context || contextError || !refID) {
      setPreviewError("");
      return;
    }
    const body = buildRequestPayload();
    let cancelled = false;
    const controller = new AbortController();
    setPreviewError("");
    previewMutation.mutate({ body, signal: controller.signal }, {
      onSuccess: data => {
        if (cancelled) return;
        setSpecYAML(data.specYaml ?? "");
        if (!promptDirtyRef.current) setPromptDraft(data.prompt ?? "");
      },
      onError: error => {
        if (!cancelled && !(error instanceof DOMException && error.name === "AbortError")) setPreviewError(error.message);
      },
    });
    return () => {
      cancelled = true;
      controller.abort();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, context, contextError, refID, step, previewModel, runtimeValue, resume, isCmux, promptDraft, promptDirty, regenNonce, previewMutation.mutate]);

  if (!open) return null;

  function submit() {
    if (!context) return;
    const { ref: _ref, ...payload } = buildRequestPayload();
    // Bookkeeping fields (runMode/plan/prompt) stay on the returned options for
    // callers that still bucket "last used"/"recent" choices by TodoRunAction
    // (see TodoDetail.tsx's rememberTodoRunOptionsForMode) — they never reach
    // the wire (see run.tsx's useTodoRun, which rebuilds the body from `step`).
    // A custom lifecycle step has no such bucket, so it carries none.
    const bookkeeping = isKnownRunAction(step) ? normalizeRunOptions(step, {}) : {};
    onRun({ ...bookkeeping, ...payload });
  }

  const promptEditorNode = (
    <div className="space-y-1">
      <div className="flex items-center justify-end gap-2">
        {previewMutation.isPending && <Spinner className="text-xs text-muted-foreground" />}
        <Button variant="ghost" type="button" onClick={regeneratePrompt} disabled={previewMutation.isPending} title="Discard edits and regenerate from the options above" className="h-auto rounded px-1.5 py-0.5 text-[11px] text-muted-foreground hover:bg-muted hover:text-foreground">Regenerate</Button>
      </div>
      {previewError && <div className="text-xs text-red-600">{previewError}</div>}
      <Suspense fallback={<textarea className={`${inputClass} h-auto min-h-[16rem] resize-y font-mono`} value={promptDraft} onChange={(event) => editPrompt(event.currentTarget.value)} placeholder={previewMutation.isPending ? "Loading prompt…" : "Prompt"} />}>
        <MdxEditorField value={promptDraft} onChange={editPrompt} placeholder={previewMutation.isPending ? "Loading prompt…" : "Prompt"} className="min-h-[16rem]" />
      </Suspense>
      {promptDirty && <div className="text-[11px] text-muted-foreground">Edited — sent verbatim as the prompt.</div>}
    </div>
  );

  return (
    <Modal open onClose={onClose} title={title} size="2xl" footer={<div className="flex justify-end gap-2"><Button variant="outline" onClick={onClose}>Cancel</Button><Button onClick={submit} loading={loading} disabled={contextLoading || !context || !!contextError}>{submitLabel}</Button></div>}>
      <div className="space-y-3">
        <TodoRunContextError error={contextError} />
        {contextLoading && <div className="text-xs text-muted-foreground">Loading Captain run providers…</div>}
        {context && !contextError && (
          <>
            <Tabs tabs={[{ id: "form", label: "Form" }, { id: "yaml", label: "YAML" }]} value={view} onChange={value => setView(value as "form" | "yaml")} />
            {view === "form" ? (
              <div role="tabpanel" aria-label="Form" className="space-y-3">
                <Field label="Prompt">
                  <SegmentedControl
                    aria-label="Prompt"
                    value={step}
                    onChange={(value) => changeStep(value as string)}
                    options={steps.map((item) => ({ id: item.name, label: item.label, disabled: item.readOnly }))}
                  />
                </Field>
                <PromptRunEditor value={runRequest} onChange={setRunRequest} models={activeModels} families={families} tools={context.tools} specSections={RUN_SPEC_SECTIONS} promptEditor={promptEditorNode} promptLabel="Prompt">
                  <>
                    {isCmux && <label className="inline-flex items-center gap-2 text-xs"><input type="checkbox" checked={resume} onChange={(event) => setResume(event.currentTarget.checked)} /><span>Resume session</span></label>}
                    {recentAdvanced.length > 0 && (
                      <div className="space-y-1 border-t border-border pt-3">
                        <div className="text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">Recent advanced</div>
                        <div className="flex flex-wrap gap-1.5">
                          {recentAdvanced.map((options, index) => (
                            <Button
                              key={runOptionsKey(options)}
                              variant="outline"
                              size="sm"
                              type="button"
                              // Only rendered when isKnownRunAction(step) — see recentAdvanced above.
                              onClick={() => setRunRequest((current) => ({ ...current, spec: runSpec(reconcileTodoRunOptions(step as TodoRunAction, options, context)) }))}
                            >
                              {index + 1}. {runChoiceDetail(options, "advanced", context)}
                            </Button>
                          ))}
                        </div>
                      </div>
                    )}
                  </>
                </PromptRunEditor>
              </div>
            ) : (
              <div role="tabpanel" aria-label="YAML" className="space-y-2">
                <div className="text-xs text-muted-foreground">Rendered Captain prompt spec sent when this {submitLabel.toLowerCase()} starts.</div>
                {previewMutation.isPending && !specYAML ? <Spinner className="text-sm text-muted-foreground" /> : null}
                {previewError ? <div className="text-xs text-red-600">{previewError}</div> : null}
                <CodeBlock language="yaml" source={specYAML} copyable className="max-h-[60vh] overflow-auto" />
              </div>
            )}
          </>
        )}
      </div>
    </Modal>
  );
}
