import { useEffect, useMemo, useState } from "react";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import {
  PromptRunEditor,
  familiesFromRuntimeCatalog,
  runtimeModeError,
  type AIPromptRunValue,
  type AISpecRuntimePermissionCatalog,
  type RuntimeCatalogFamily,
  type SpecRuntimeFamily,
} from "@flanksource/clicky-ui/ai";
import type { ChatModel } from "@flanksource/clicky-ui/chat";
import { Button, Field, InputField, Modal, Select } from "@flanksource/clicky-ui/components";
import { request } from "./api";
import {
  assertRunContext,
  availableMode,
  buildRunRequest,
  effectiveRuntimeOf,
  findPrompt,
  loadRememberedRun,
  modeUnavailable,
  promptLabel,
  rememberRun,
  toolMetaOf,
  variablesSchemaOf,
  type RunContext,
  type RunPreview,
  type RunRequest,
} from "./run-request";

const PREVIEW_DEBOUNCE_MS = 400;

function useDebounced<T>(value: T, delayMs: number): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delayMs);
    return () => clearTimeout(timer);
  }, [value, delayMs]);
  return debounced;
}

function useRuntimeCatalogs(fileId: string) {
  const models = useQuery({
    queryKey: ["chat", "models"],
    queryFn: async () => {
      const catalog = await request<ChatModel[]>("/api/chat/models");
      if (!Array.isArray(catalog) || catalog.length === 0) throw new Error("AI model catalog is empty or invalid");
      return catalog;
    },
  });
  const families = useQuery({
    queryKey: ["chat", "runtimes"],
    queryFn: async () => {
      const catalog = await request<RuntimeCatalogFamily[]>("/api/chat/runtimes");
      if (!Array.isArray(catalog) || catalog.length === 0 || catalog.some((family) => !family?.modes?.length))
        throw new Error("AI runtime catalog is empty or invalid");
      return familiesFromRuntimeCatalog(catalog);
    },
  });
  const context = useQuery({
    queryKey: ["editor", "agent-run-context", fileId],
    queryFn: async () => assertRunContext(await request<unknown>(`/api/file/${fileId}/agent-run/context`)),
    gcTime: 0,
    retry: false,
  });
  return { models, families, context };
}

/**
 * Starts a new managed agent session on a thread: which captain prompt runs it,
 * the presets and runtime layered over it, its variables and anything to add to
 * the thread. The preview shows what the request resolves to before it starts.
 */
export function NewAgentSessionDialog({
  thread,
  fileId,
  onPreview,
  onStart,
  onClose,
}: {
  thread: string;
  fileId: string;
  onPreview: (run: RunRequest) => Promise<RunPreview>;
  onStart: (run: RunRequest) => Promise<void>;
  onClose: () => void;
}) {
  const { models, families, context } = useRuntimeCatalogs(fileId);
  if (!models.data || !families.data || !context.data) {
    const failure = models.error ?? families.error ?? context.error;
    return (
      <Modal
        open
        onClose={onClose}
        title="New agent session"
        size="lg"
        footer={
          <div className="flex justify-end gap-2">
            <Button variant="ghost" onClick={onClose}>
              Cancel
            </Button>
          </div>
        }
      >
        {failure ? (
          <div className="flex items-center gap-2 text-sm">
            <span role="alert" className="text-destructive">
              {failure.message}
            </span>
            <Button
              variant="ghost"
              onClick={() => {
                void models.refetch();
                void families.refetch();
                void context.refetch();
              }}
            >
              Retry
            </Button>
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">Loading prompts and AI options…</p>
        )}
      </Modal>
    );
  }
  return (
    <SessionForm
      thread={thread}
      fileId={fileId}
      context={context.data}
      models={models.data}
      families={families.data}
      onPreview={onPreview}
      onStart={onStart}
      onClose={onClose}
    />
  );
}

function SessionForm({
  thread,
  fileId,
  context,
  models,
  families,
  onPreview,
  onStart,
  onClose,
}: {
  thread: string;
  fileId: string;
  context: RunContext;
  models: ChatModel[];
  families: SpecRuntimeFamily[];
  onPreview: (run: RunRequest) => Promise<RunPreview>;
  onStart: (run: RunRequest) => Promise<void>;
  onClose: () => void;
}) {
  const remembered = useMemo(loadRememberedRun, []);
  // A remembered prompt that is no longer discovered starts from the context's default instead.
  const [promptId, setPromptId] = useState(() => (findPrompt(context.prompts, remembered.prompt) ?? findPrompt(context.prompts, context.defaultPrompt))?.id);
  const [value, setValue] = useState<AIPromptRunValue>({ spec: remembered.spec ?? {}, ...(remembered.presets ? { presets: remembered.presets } : {}) });
  const [instructions, setInstructions] = useState("");
  const [error, setError] = useState<string>();
  const [starting, setStarting] = useState(false);

  const prompt = findPrompt(context.prompts, promptId);
  if (!prompt) throw new Error(`prompt ${JSON.stringify(promptId)} is not among the discovered prompts`);
  const effective = effectiveRuntimeOf({ spec: value.spec ?? {}, prompt, families, models });
  const modeError = runtimeModeError(families, value.spec?.mode) ?? modeUnavailable(families, effective.mode);
  const fallbackMode = availableMode(families);

  const built = useMemo(() => {
    try {
      return {
        schema: variablesSchemaOf(prompt.variables, context.hostVariables),
        run: buildRunRequest({ prompt: prompt.id, value, instructions, hostVariables: context.hostVariables }),
      };
    } catch (cause) {
      return { failure: cause instanceof Error ? cause.message : String(cause) };
    }
  }, [prompt, value, instructions, context.hostVariables]);

  const runKey = useDebounced(JSON.stringify(built.run ?? null), PREVIEW_DEBOUNCE_MS);
  const preview = useQuery({
    queryKey: ["editor", "agent-run-preview", fileId, runKey],
    queryFn: () => onPreview(JSON.parse(runKey) as RunRequest),
    enabled: !modeError && runKey !== "null",
    placeholderData: keepPreviousData,
    gcTime: 0,
    retry: false,
  });
  const catalog = useQuery({
    queryKey: ["editor", "permission-catalog", fileId, effective.family, effective.mode],
    queryFn: async () => {
      if (!effective.family) throw new Error(`no runtime family provides ${effective.mode} mode`);
      // Rooted server-side at the project holding the file.
      const query = new URLSearchParams({ provider: effective.family, mode: effective.mode });
      const permissions = await request<AISpecRuntimePermissionCatalog>(`/api/file/${fileId}/agent-run/permissions?${query}`);
      if (!permissions || typeof permissions !== "object") throw new Error("permission catalog is malformed");
      return permissions;
    },
    enabled: !modeError && effective.family !== undefined,
    retry: false,
  });

  const settled = runKey === JSON.stringify(built.run ?? null);
  const blocked = !!built.failure || !!modeError || (settled && (preview.isError || !!preview.data?.validationError));

  const start = async () => {
    if (!built.run) return;
    setStarting(true);
    setError(undefined);
    try {
      rememberRun({ prompt: prompt.id, ...(value.presets ? { presets: value.presets } : {}), ...(value.spec ? { spec: value.spec } : {}) });
      await onStart(built.run);
    } catch (cause) {
      setError(String(cause));
    } finally {
      setStarting(false);
    }
  };

  const problems = [built.failure, preview.error?.message, preview.data?.validationError, catalog.error && `Permission catalog unavailable: ${catalog.error.message}`].filter(
    (problem): problem is string => !!problem,
  );

  return (
    <Modal
      open
      onClose={onClose}
      title="New agent session"
      size="2xl"
      footer={
        <div className="flex justify-end gap-2">
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={() => void start()} disabled={starting || blocked}>
            {starting ? "Starting…" : "Start session"}
          </Button>
        </div>
      }
    >
      <div className="flex flex-col gap-3">
        <p className="text-sm text-muted-foreground">The session works this thread in {context.dir}, replying and resolving it as its own agent.</p>
        <blockquote className="line-clamp-3 border-l-2 border-border pl-2 text-sm">{thread}</blockquote>
        <Field label="Prompt" htmlFor="new-session-prompt" helper={prompt.description}>
          <Select
            id="new-session-prompt"
            value={prompt.id}
            onChange={(event) => {
              setPromptId(event.currentTarget.value);
              setValue((current) => ({ ...current, variables: {} }));
            }}
            options={context.prompts.map((entry) => ({ value: entry.id, label: promptLabel(entry) }))}
          />
        </Field>
        {modeError ? (
          <div className="flex flex-wrap items-center gap-2 text-sm">
            <span role="alert" className="text-destructive">{`${effective.mode} mode cannot run here. ${modeError}`}</span>
            {fallbackMode ? (
              <Button variant="outline" size="sm" onClick={() => setValue((current) => ({ ...current, spec: { mode: fallbackMode } }))}>
                {`Use ${fallbackMode} mode`}
              </Button>
            ) : (
              <span className="text-muted-foreground">No AI runtime is available; configure credentials, then retry.</span>
            )}
          </div>
        ) : (
          <PromptRunEditor
            key={prompt.id}
            value={value}
            onChange={setValue}
            presets={context.presets}
            models={models}
            families={families}
            tools={toolMetaOf(preview.data?.tools ?? [])}
            {...(catalog.data ? { permissionCatalog: catalog.data } : {})}
            {...(preview.data ? { resolution: preview.data.resolution } : {})}
            {...(built.schema ? { variablesSchema: built.schema } : {})}
            promptLabel="Additional instructions"
            promptEditor={
              <div className="flex flex-col gap-1">
                <InputField
                  id="new-session-instructions"
                  aria-label="Additional instructions"
                  as="textarea"
                  value={instructions}
                  onChange={setInstructions}
                  placeholder="e.g. Keep the existing headings"
                />
                <p className="text-xs text-muted-foreground">Appended to the thread the session is given.</p>
              </div>
            }
          />
        )}
        {preview.data?.warnings?.map((warning) => (
          <p key={warning} className="text-xs text-muted-foreground">
            {warning}
          </p>
        ))}
        {[...problems, ...(error ? [error] : [])].map((problem) => (
          <p key={problem} role="alert" className="text-sm text-destructive">
            {problem}
          </p>
        ))}
      </div>
    </Modal>
  );
}
