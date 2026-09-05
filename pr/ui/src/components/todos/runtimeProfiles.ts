import type { TodoRunOptions } from '../../types';
import type { AISpecRuntimeValue } from '@flanksource/clicky-ui/ai';
import type { RunContext } from './providers';

export function effectiveTodoRuntime(options: TodoRunOptions, context: RunContext): Pick<AISpecRuntimeValue, 'model' | 'mode'> {
  const defaults = context.promptDefaults?.[options.step ?? 'run'];
  const reference = options.runtimeProfile ?? defaults?.runtimeProfile;
  const ref = reference?.toLowerCase();
  const profile = context.runtimeProfiles?.find(entry => entry.id === reference || entry.name.toLowerCase() === ref);
  const stepProfile = !options.runtimeProfile || reference === defaults?.runtimeProfile || profile?.id === defaults?.runtimeProfile;
  return {
    model: options.spec?.model ?? (stepProfile ? defaults?.model : undefined),
    mode: options.spec?.mode ?? (stepProfile ? defaults?.mode ?? context.defaultMode : undefined),
  };
}

export function unresolvedTodoRuntimeProfile(options: TodoRunOptions, context: RunContext): string | undefined {
  if (!options.runtimeProfile) return undefined;
  const runtime = effectiveTodoRuntime(options, context);
  if (runtime.model && runtime.mode) return undefined;
  const reference = options.runtimeProfile;
  return context.runtimeProfiles?.find(profile => profile.id === reference || profile.name.toLowerCase() === reference.toLowerCase())?.name ?? reference;
}

export function selectTodoRuntimeProfile({ options, defaults, runtimeProfile }: {
  options: TodoRunOptions;
  defaults: TodoRunOptions;
  runtimeProfile: string | undefined;
}): TodoRunOptions {
  const next = { ...options };
  if (runtimeProfile) {
    next.runtimeProfile = runtimeProfile;
    if (!options.runtimeProfile) next.spec = (changedValue(options.spec ?? {}, defaults.spec ?? {}) ?? {}) as AISpecRuntimeValue;
  } else delete next.runtimeProfile;
  return next;
}

function changedValue(value: unknown, inherited: unknown): unknown {
  if (!isRecord(value) || !isRecord(inherited)) return JSON.stringify(value) === JSON.stringify(inherited) ? undefined : value;
  const entries = Object.entries(value)
    .map(([key, item]) => [key, changedValue(item, inherited[key])])
    .filter(([, item]) => item !== undefined);
  return entries.length ? Object.fromEntries(entries) : undefined;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}
