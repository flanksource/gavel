import { mergeRuntimeSpec, type AISpecRuntimeValue, type RuntimeCatalogFamily, type RuntimePreset } from "@flanksource/clicky-ui/ai";
import type { ChatModel } from "@flanksource/clicky-ui/chat";

export type TodoRun = { step: string; spec: AISpecRuntimeValue; presets?: string[] };
export type RunResult = { status: string; sessionId?: string; promptRunId?: string; error?: string };
export type RunStep = { name: string; label: string; prompt: string; auxiliary: boolean };
export type RunContext = {
  runtimes: RuntimeCatalogFamily[];
  models: ChatModel[];
  efforts: string[];
  promptDefaults?: Record<string, { spec?: AISpecRuntimeValue; model?: string; mode?: string; presets?: string[] }>;
  runtimePresets?: RuntimePreset[];
};

export function runDefaults(context: RunContext, step: string): TodoRun {
  const presets = context.promptDefaults?.[step]?.presets;
  return { step, spec: {}, ...(presets ? { presets } : {}) };
}

export function effectiveRuntime(context: RunContext, run: TodoRun): AISpecRuntimeValue {
  const defaults = context.promptDefaults?.[run.step];
  const inherited = run.presets === undefined || JSON.stringify(run.presets) === JSON.stringify(defaults?.presets ?? []);
  let spec: AISpecRuntimeValue = inherited ? {
    ...(defaults?.model ? { model: defaults.model } : {}), ...(defaults?.mode ? { mode: defaults.mode } : {}), ...defaults?.spec,
  } : {};
  for (const reference of run.presets ?? []) {
    const preset = context.runtimePresets?.find(entry => entry.id === reference || entry.name.toLowerCase() === reference.toLowerCase());
    if (!preset) throw new Error(`Runtime preset ${reference} is not in this workspace's cached catalog`);
    spec = mergeRuntimeSpec(spec, preset.spec);
  }
  return mergeRuntimeSpec(spec, run.spec);
}

export function runCaptureError(run: TodoRun | undefined, context: RunContext | undefined, steps: RunStep[] | undefined): string {
  if (!run) return "";
  if (!context) return "The selected run needs this project's saved runtime catalog. Connect to Gavel or turn off Start a new run.";
  if (!steps?.some(step => step.name === run.step)) return `Step ${run.step} is unavailable for this project's selected status. Choose another step or turn off Start a new run.`;
  try { effectiveRuntime(context, run); }
  catch (cause) { return String(cause); }
  return "";
}
