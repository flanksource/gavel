import type { AISpecRuntimeValue } from '@flanksource/clicky-ui/ai';
import type { TodoRunOptions } from '../../types';
import type { RunContext } from './providers';

export function effectiveTodoRuntime(options: TodoRunOptions, context: RunContext): Pick<AISpecRuntimeValue, 'model' | 'mode'> {
  const defaults = context.promptDefaults?.[options.step ?? 'run'];
  const references = options.presets ?? [];
  const selected = references
    .map(reference => context.runtimePresets?.find(entry => entry.id === reference || entry.name.toLowerCase() === reference.toLowerCase()))
    .filter(entry => entry !== undefined);
  const presetRuntime = selected.reduce<Pick<AISpecRuntimeValue, 'model' | 'mode'>>((runtime, preset) => ({
    model: preset.spec.model ?? runtime.model,
    mode: preset.spec.mode ?? runtime.mode,
  }), {});
  const stepDefaults = options.presets === undefined || JSON.stringify(references) === JSON.stringify(defaults?.presets ?? []);
  return {
    model: options.spec?.model ?? presetRuntime.model ?? (stepDefaults ? defaults?.spec?.model ?? defaults?.model : undefined),
    mode: options.spec?.mode ?? presetRuntime.mode ?? (stepDefaults ? defaults?.spec?.mode ?? defaults?.mode : undefined),
  };
}

export function unresolvedTodoRuntimePreset(options: TodoRunOptions, context: RunContext): string | undefined {
  if (options.presets === undefined) return undefined;
  const missing = options.presets.filter(reference => !context.runtimePresets?.some(
    preset => preset.id === reference || preset.name.toLowerCase() === reference.toLowerCase(),
  ));
  return missing.length > 0 ? missing.join(', ') : undefined;
}
