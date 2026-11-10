import {
  familyForMode,
  familyForModel,
  promptRuntimeValueToPayload,
  type AIPromptRunValue,
  type AISpecRuntimeSpec,
  type ResolvedRuntimeSpec,
  type RuntimePreset,
  type SpecRuntimeFamily,
  type ToolMeta,
} from "@flanksource/clicky-ui/ai";
import type { ChatModel } from "@flanksource/clicky-ui/chat";
import type { JsonSchemaObject, JsonSchemaProperty } from "@flanksource/clicky-ui/components";

/** A variable a captain prompt declares (captain PromptSummary.variables). */
export type PromptVariable = { name: string; type?: string; description?: string; required?: boolean };

/** The part of captain's PromptSummary the dialog reads. */
export type PromptSummary = {
  id: string;
  name: string;
  description?: string;
  source: string;
  model?: string;
  mode?: string;
  variables?: PromptVariable[];
};

/** GET /api/file/{id}/agent-run/context. */
export type RunContext = {
  prompts: PromptSummary[];
  presets: RuntimePreset[];
  defaultPrompt: string;
  hostVariables: string[];
  dir: string;
};

/** POST /api/file/{id}/comments/{cid}/run (and /run/preview): captain's run request plus extra instructions. */
export type RunRequest = {
  prompt?: string;
  presets?: string[];
  spec?: AISpecRuntimeSpec;
  variables?: Record<string, unknown>;
  instructions?: string;
};

/** POST .../run/preview. */
export type RunPreview = {
  model?: string;
  mode?: string;
  resolution: ResolvedRuntimeSpec;
  validationError?: string;
  warnings?: string[];
  tools: { name: string; description?: string; source: "thread" | "prompt" }[];
};

/** What the editor-comments prompt runs on when neither the request nor the prompt names a model or mode. */
export const PROMPT_RUNTIME = { model: "claude", mode: "agent" } as const;

/** Validates the context answer: a malformed one is a backend contract break, not something to render around. */
export function assertRunContext(value: unknown): RunContext {
  const context = value as Partial<RunContext> | null;
  if (
    !context ||
    !Array.isArray(context.prompts) ||
    !Array.isArray(context.presets) ||
    !Array.isArray(context.hostVariables) ||
    typeof context.defaultPrompt !== "string" ||
    typeof context.dir !== "string"
  )
    throw new Error("agent-run context is malformed: expected prompts, presets, hostVariables, defaultPrompt and dir");
  if (!findPrompt(context.prompts, context.defaultPrompt))
    throw new Error(`default prompt ${JSON.stringify(context.defaultPrompt)} is not among the ${context.prompts.length} discovered prompts`);
  return context as RunContext;
}

/** The prompt named by `idOrName`, matching an id first so a same-named prompt from another source cannot shadow it. */
export function findPrompt(prompts: readonly PromptSummary[], idOrName: string | undefined): PromptSummary | undefined {
  if (!idOrName) return undefined;
  return prompts.find((prompt) => prompt.id === idOrName) ?? prompts.find((prompt) => prompt.name === idOrName);
}

/** "name · source", the label a prompt carries in the picker. */
export function promptLabel(prompt: PromptSummary): string {
  return `${prompt.name} · ${prompt.source}`;
}

const VARIABLE_TYPES: Record<string, NonNullable<JsonSchemaProperty["type"]>> = {
  string: "string",
  number: "number",
  integer: "integer",
  int: "integer",
  float: "number",
  boolean: "boolean",
  bool: "boolean",
  array: "array",
  object: "object",
};

/**
 * The variables form for a prompt: its declared variables except the ones the
 * server fills in. An undeclared type is a string; an unrecognised one throws,
 * since guessing would render a control that sends the wrong value shape.
 */
export function variablesSchemaOf(variables: readonly PromptVariable[] | undefined, hostVariables: readonly string[]): JsonSchemaObject {
  const declared = (variables ?? []).filter((variable) => !hostVariables.includes(variable.name));
  const properties = Object.fromEntries(
    declared.map(({ name, type, description }) => {
      const resolved = type === undefined || type === "" ? "string" : VARIABLE_TYPES[type];
      if (!resolved) throw new Error(`prompt variable ${name} has unsupported type ${JSON.stringify(type)}`);
      return [name, { type: resolved, title: name, ...(description ? { description } : {}) } satisfies JsonSchemaProperty];
    }),
  );
  const required = declared.filter((variable) => variable.required).map((variable) => variable.name);
  return { type: "object", properties, ...(required.length > 0 ? { required } : {}) };
}

/**
 * The request the dialog previews and submits: the chosen prompt, ordered
 * presets, the compacted runtime spec and variables the user set (never the
 * ones the server owns), and trimmed instructions. Empty fields are omitted.
 */
export function buildRunRequest({
  prompt,
  value,
  instructions,
  hostVariables,
}: {
  prompt: string;
  value: AIPromptRunValue;
  instructions: string;
  hostVariables: readonly string[];
}): RunRequest {
  const { spec } = promptRuntimeValueToPayload(value.spec ?? {});
  const variables = Object.fromEntries(
    Object.entries(value.variables ?? {}).filter(
      ([name, entry]) => !hostVariables.includes(name) && entry !== undefined && entry !== null && entry !== "",
    ),
  );
  return {
    ...(prompt ? { prompt } : {}),
    ...(value.presets?.length ? { presets: value.presets } : {}),
    ...(Object.keys(spec).length > 0 ? { spec } : {}),
    ...(Object.keys(variables).length > 0 ? { variables } : {}),
    ...(instructions.trim() ? { instructions: instructions.trim() } : {}),
  };
}

/**
 * The model, mode and runtime family a run lands on: the request's own choice,
 * else the prompt's, else the editor-comments default. The family is what the
 * permission catalog is scoped to; undefined when no family provides the mode.
 */
export function effectiveRuntimeOf({
  spec,
  prompt,
  families,
  models,
}: {
  spec: AISpecRuntimeSpec;
  prompt: PromptSummary | undefined;
  families: SpecRuntimeFamily[];
  models: ChatModel[];
}): { model: string; mode: string; family: string | undefined } {
  const model = spec.model || spec.id || prompt?.model || PROMPT_RUNTIME.model;
  const mode = spec.mode || prompt?.mode || PROMPT_RUNTIME.mode;
  const family =
    familyForModel(families, models, model) ?? families.find((entry) => entry.id === model) ?? familyForMode(families, mode);
  return { model, mode, family: family?.id };
}

/** The preview's effective tools as the editor's tool list, grouped by where each came from. */
export function toolMetaOf(tools: RunPreview["tools"]): ToolMeta[] {
  return tools.map(({ name, description, source }) => ({
    name,
    label: name,
    group: source === "thread" ? "Thread tools" : "Prompt tools",
    ...(description ? { description } : {}),
  }));
}

type CatalogFamily = { label: string; modes: { id: string; availability?: { state: string; reason?: string; remediation?: string } }[] };

/**
 * Why `mode` cannot run here, or undefined when some family can run it. RuntimeBar
 * renders only available modes, so an unavailable inherited mode is reported
 * instead of rendered.
 */
export function modeUnavailable(families: readonly CatalogFamily[], mode: string): string | undefined {
  const candidates = families.flatMap((family) => family.modes.filter((entry) => entry.id === mode).map((entry) => ({ family, entry })));
  if (candidates.some(({ entry }) => !entry.availability || entry.availability.state === "available")) return undefined;
  if (candidates.length === 0) return `No runtime here provides ${mode} mode.`;
  return candidates
    .map(({ family, entry }) => `${family.label}: ${[entry.availability?.reason, entry.availability?.remediation].filter(Boolean).join(" ") || entry.availability?.state}`)
    .join(" ");
}

/** The first mode some family can run here, to switch to when the inherited one cannot. */
export function availableMode(families: readonly CatalogFamily[]): string | undefined {
  return families.flatMap((family) => family.modes).find((entry) => !entry.availability || entry.availability.state === "available")?.id;
}

/** What the dialog restores on the next open: a per-viewer convenience, never authoritative. */
export type RememberedRun = { prompt?: string; presets?: string[]; spec?: AISpecRuntimeSpec };

const REMEMBERED_RUN = "gavel-edit:new-session-run";

export function loadRememberedRun(): RememberedRun {
  try {
    const stored = window.localStorage.getItem(REMEMBERED_RUN);
    const parsed: unknown = stored ? JSON.parse(stored) : {};
    return parsed && typeof parsed === "object" && !Array.isArray(parsed) ? (parsed as RememberedRun) : {};
  } catch {
    // A viewer convenience only: unreadable storage starts from the prompt's own defaults.
    return {};
  }
}

export function rememberRun(run: RememberedRun) {
  try {
    window.localStorage.setItem(REMEMBERED_RUN, JSON.stringify(run));
  } catch {
    // Storage may be blocked; the choice still applies to this run.
  }
}
