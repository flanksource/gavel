import { lazy, Suspense, useEffect, useRef, useState } from "react";
import { Button, Field, Modal, SegmentedControl, Tabs } from "@flanksource/clicky-ui/components";
import { CodeBlock } from "@flanksource/clicky-ui/data";
import { PromptRunEditor, type AIPromptRunValue, type AISpecRuntimeValue } from "@flanksource/clicky-ui/ai";
import type { TodoRunOptions } from "../../types";
import { Spinner } from "../../icons/Spinner";
import { inputClass } from "./format";
import {
  buildTodoRunPayload,
  loadLastTodoRunOptions,
  loadRecentAdvancedTodoRunOptions,
  reconcileTodoRunOptions,
  runChoiceDetail,
  runOptionsKey,
  runSpec,
  TodoRunContextError,
  type RunMode,
  type TodoRunAction,
  useTodoRunContext,
  useTodoRunPreview,
} from "./run";
import {
  agentForRuntime,
  buildRunFamilies,
  defaultModelForSelection,
  driverForSelection,
  isCmuxBackend,
} from "./providers";

const RUN_SPEC_SECTIONS = ["model", "prompt", "permissions", "workspace", "verify", "commit"] as const;
const AUTO_COMMIT: Pick<AISpecRuntimeValue, "workflow"> = { workflow: { commits: [{ on: "run", gates: "full" }] } };
const INITIAL_RUNTIME_VALUE: AISpecRuntimeValue = { effort: "medium", ...AUTO_COMMIT };
const MdxEditorField = lazy(() => import("@flanksource/clicky-ui/mdx-editor").then((module) => ({ default: module.MdxEditorField })));

export function TodoRunAdvancedDialog({
  open,
  onClose,
  onRun,
  loading,
  title = "Run todo",
  initialMode = "run",
  dir,
  refID,
}: {
  open: boolean;
  onClose: () => void;
  onRun: (options: TodoRunOptions) => void;
  loading?: boolean;
  title?: string;
  initialMode?: RunMode;
  dir: string;
  refID: string;
}) {
  const [runRequest, setRunRequest] = useState<AIPromptRunValue>({ spec: INITIAL_RUNTIME_VALUE });
  const runtimeValue = runRequest.spec ?? {};
  const [mode, setMode] = useState<RunMode>("run");
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

  const families = context ? buildRunFamilies(context) : [];
  const agent = context ? agentForRuntime(context, runtimeValue.backend, runtimeValue.model) : undefined;
  const isCmux = isCmuxBackend(runtimeValue.backend);
  const activeModels = context?.models ?? [];
  const modelFallback = context && agent ? defaultModelForSelection(context, agent, runtimeValue.backend) : "";
  const selection = context && agent ? driverForSelection(context, agent, runtimeValue.backend) : null;
  const driver = selection?.driver;
  const runBackend = selection?.runBackend;
  const plan = mode === "plan";
  const advancedAction: TodoRunAction = plan ? "plan" : "run";
  const recentAdvanced = context ? loadRecentAdvancedTodoRunOptions(advancedAction, context) : [];

  function changeMode(next: RunMode) {
    setMode(next);
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
    setRunRequest({ spec: INITIAL_RUNTIME_VALUE });
    setMode(initialMode);
    setResume(false);
    setPromptDraft("");
    setPromptDirty(false);
    setSpecYAML("");
    setView("form");
    promptDirtyRef.current = false;
  }, [open, initialMode]);

  useEffect(() => {
    if (!open || !context) return;
    const action: TodoRunAction = initialMode === "plan" ? "plan" : "run";
    setRunRequest({ spec: runSpec(reconcileTodoRunOptions(action, loadLastTodoRunOptions(action, context), context)) });
  }, [open, initialMode, context]);

  const previewModel = runtimeValue.model?.trim() || modelFallback;

  useEffect(() => {
    if (!open) {
      setPreviewError("");
      return;
    }
    if (!context || contextError || !refID || !driver) {
      setPreviewError("");
      return;
    }
    const body = buildTodoRunPayload({
      ref: refID,
      driver,
      runBackend,
      runtime: { ...runtimeValue, model: previewModel },
      mode,
      resume: isCmux && resume,
      promptDraft,
      promptDirty,
    });
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
  }, [open, context, contextError, refID, driver, runBackend, previewModel, runtimeValue, mode, resume, isCmux, promptDraft, promptDirty, regenNonce, previewMutation.mutate]);

  if (!open) return null;

  function submit() {
    if (!context || !driver) return;
    const { ref: _, ...options } = buildTodoRunPayload({
      ref: refID,
      driver,
      runBackend,
      runtime: { ...runtimeValue, model: previewModel },
      mode,
      resume: isCmux && resume,
      promptDraft,
      promptDirty,
    });
    onRun(options);
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
    <Modal open onClose={onClose} title={title} size="2xl" footer={<div className="flex justify-end gap-2"><Button variant="outline" onClick={onClose}>Cancel</Button><Button onClick={submit} loading={loading} disabled={contextLoading || !context || !!contextError}>{plan ? "Plan" : "Run"}</Button></div>}>
      <div className="space-y-3">
        <TodoRunContextError error={contextError} />
        {contextLoading && <div className="text-xs text-muted-foreground">Loading Captain run providers…</div>}
        {context && !contextError && (
          <>
            <Tabs tabs={[{ id: "form", label: "Form" }, { id: "yaml", label: "YAML" }]} value={view} onChange={value => setView(value as "form" | "yaml")} />
            {view === "form" ? (
              <div role="tabpanel" aria-label="Form" className="space-y-3">
                <Field label="Mode"><SegmentedControl aria-label="Mode" value={mode} onChange={(value) => changeMode(value as RunMode)} options={[{ id: "run", label: "Run" }, { id: "plan", label: "Plan" }]} /></Field>
                <PromptRunEditor value={runRequest} onChange={setRunRequest} models={activeModels} families={families} tools={context.tools} specSections={RUN_SPEC_SECTIONS} promptEditor={promptEditorNode} promptLabel="Prompt">
                  <>
                    {isCmux && <label className="inline-flex items-center gap-2 text-xs"><input type="checkbox" checked={resume} onChange={(event) => setResume(event.currentTarget.checked)} /><span>Resume session</span></label>}
                    {recentAdvanced.length > 0 && (
                      <div className="space-y-1 border-t border-border pt-3">
                        <div className="text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">Recent advanced</div>
                        <div className="flex flex-wrap gap-1.5">
                          {recentAdvanced.map((options, index) => (
                            <Button key={runOptionsKey(options)} variant="outline" size="sm" type="button" onClick={() => setRunRequest((current) => ({ ...current, spec: runSpec(reconcileTodoRunOptions(advancedAction, options, context)) }))}>
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
                <div className="text-xs text-muted-foreground">Rendered Captain prompt spec sent when this {plan ? "plan" : "run"} starts.</div>
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
