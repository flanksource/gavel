import { runtimeModelError, runtimeModeFromModel, type RuntimeBarValue } from '@flanksource/clicky-ui/ai';

export const runtimeKeys = ['model', 'mode', 'effort'] as const;

export function commitRuntimeValue(options: Record<string, unknown>): RuntimeBarValue {
  const runtime: RuntimeBarValue = {};
  for (const key of runtimeKeys) {
    const value = options[key];
    if (value === undefined || value === '') continue;
    if (typeof value !== 'string') throw new Error(`Commit ${key} must be a string`);
    runtime[key] = value;
  }
  const error = runtimeModelError(runtime.model);
  if (error) throw new Error(error);
  if (runtime.model && !runtime.model.includes(',') && runtimeModeFromModel(runtime.model)) {
    const [mode, model, effort] = runtime.model.split(':').map(part => part.trim());
    runtime.mode = mode;
    runtime.model = model;
    if (effort) runtime.effort = effort;
  }
  return runtime;
}

export function withCommitRuntime(options: Record<string, unknown>, runtime: RuntimeBarValue): Record<string, unknown> {
  const next = { ...options };
  for (const key of runtimeKeys) {
    delete next[key];
    if (runtime[key] !== undefined && runtime[key] !== '') next[key] = runtime[key];
  }
  return next;
}

export function withoutRemovedAIOptions(options: Record<string, unknown>): Record<string, unknown> {
  const next = { ...options };
  for (const key of ['group-model', 'fallback', 'temperature', 'no-cache']) delete next[key];
  return next;
}
