import { describe, expect, it } from 'vitest';
import type { RunContext } from './providers';
import { effectiveTodoRuntime, unresolvedTodoRuntimePreset } from './runtimePresets';

describe('runtime preset request overrides', () => {
  it('shows the folded step default above its selected presets', () => {
    const context: RunContext = {
      modes: [], runtimes: [], models: [], efforts: [], tools: [], lifecycle: { steps: [] },
      defaultMode: 'agent',
      promptDefaults: { plan: { presets: ['preset-review'], model: 'example-step-model', mode: 'cmux' } },
      runtimePresets: [{ id: 'preset-review', name: 'Review', scope: 'context', spec: { model: 'example-preset-model' } }],
    };
    expect(effectiveTodoRuntime({ step: 'plan', presets: ['preset-review'], spec: {} }, context)).toEqual({ model: 'example-preset-model', mode: 'cmux' });
    expect(effectiveTodoRuntime({ step: 'plan', spec: {} }, context)).toEqual({ model: 'example-step-model', mode: 'cmux' });
  });

  it('does not invent fields absent from a different preset selection', () => {
    const context: RunContext = {
      modes: [], runtimes: [], models: [], efforts: [], tools: [], lifecycle: { steps: [] },
      defaultMode: 'cli',
      promptDefaults: { run: { presets: ['step-preset'], model: 'example-step-model', mode: 'cli' } },
      runtimePresets: [{ id: 'selected-preset', name: 'Selected', scope: 'context', spec: {} }],
    };
    expect(effectiveTodoRuntime({ step: 'run', presets: ['selected-preset'], spec: {} }, context)).toEqual({ model: undefined, mode: undefined });
    expect(effectiveTodoRuntime({ step: 'run', presets: ['selected-preset'], spec: { mode: 'cmux' } }, context)).toEqual({ model: undefined, mode: 'cmux' });
  });

  it('reports only preset references missing from the runtime catalog', () => {
    const context: RunContext = {
      modes: [], runtimes: [], models: [], efforts: [], tools: [], lifecycle: { steps: [] },
      runtimePresets: [{ id: 'known-id', name: 'Known preset', scope: 'context', spec: {} }],
    };
    expect(unresolvedTodoRuntimePreset({ presets: ['Known preset'], spec: {} }, context)).toBeUndefined();
    expect(unresolvedTodoRuntimePreset({ presets: ['known-id', 'missing'], spec: {} }, context)).toBe('missing');
  });
});
