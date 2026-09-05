import { describe, expect, it } from 'vitest';
import { effectiveTodoRuntime, selectTodoRuntimeProfile } from './runtimeProfiles';
import type { RunContext } from './providers';

const defaults = { step: 'plan', spec: { model: 'example-default', mode: 'agent', effort: 'medium' as const } };

describe('runtime profile request overrides', () => {
  it('shows the folded step default above the profile declaration', () => {
    const context: RunContext = {
      modes: [], runtimes: [], models: [], efforts: [], tools: [], lifecycle: { steps: [] },
      defaultMode: 'agent',
      promptDefaults: { plan: { runtimeProfile: 'profile-review', model: 'example-step-model', mode: 'cmux' } },
      runtimeProfiles: [{ id: 'profile-review', name: 'Review', model: 'example-profile-model' }],
    };
    expect(effectiveTodoRuntime({ step: 'plan', runtimeProfile: 'profile-review', spec: {} }, context)).toEqual({ model: 'example-step-model', mode: 'cmux' });
    expect(effectiveTodoRuntime({ step: 'plan', spec: {} }, context)).toEqual({ model: 'example-step-model', mode: 'cmux' });
  });

  it('does not invent the effective runtime of a different profile or preset-only profile', () => {
    const context: RunContext = {
      modes: [], runtimes: [], models: [], efforts: [], tools: [], lifecycle: { steps: [] },
      defaultMode: 'cli',
      promptDefaults: { run: { runtimeProfile: 'step-profile', model: 'example-step-model', mode: 'cli' } },
      runtimeProfiles: [{ id: 'selected-profile', name: 'Selected', presets: ['cmux-preset'] }],
    };
    expect(effectiveTodoRuntime({ step: 'run', runtimeProfile: 'selected-profile', spec: {} }, context)).toEqual({ model: undefined, mode: undefined });
    expect(effectiveTodoRuntime({ step: 'run', runtimeProfile: 'selected-profile', spec: { mode: 'cmux' } }, context)).toEqual({ model: undefined, mode: 'cmux' });
  });

  it('replaces preseeded runtime defaults with the selected profile reference', () => {
    expect(selectTodoRuntimeProfile({ options: defaults, defaults, runtimeProfile: 'profile-review' })).toEqual({
      step: 'plan', runtimeProfile: 'profile-review', spec: {},
    });
  });

  it('keeps operator changes and removes only unchanged defaults', () => {
    const options = { ...defaults, spec: { ...defaults.spec, effort: 'high' as const, budget: { maxTurns: 8 } } };
    expect(selectTodoRuntimeProfile({ options, defaults, runtimeProfile: 'profile-review' })).toEqual({
      step: 'plan', runtimeProfile: 'profile-review', spec: { effort: 'high', budget: { maxTurns: 8 } },
    });
  });

  it('keeps an explicitly edited model when switching profiles', () => {
    const options = { step: 'verify', runtimeProfile: 'profile-first', spec: { model: 'example-override' } };
    expect(selectTodoRuntimeProfile({ options, defaults, runtimeProfile: 'profile-second' })).toEqual({ ...options, runtimeProfile: 'profile-second' });
  });

  it('removes nested defaults without discarding edited sibling fields', () => {
    const defaultOptions = { ...defaults, spec: { ...defaults.spec, budget: { maxTurns: 8, timeout: '10m' } } };
    const options = { ...defaultOptions, spec: { ...defaultOptions.spec, budget: { maxTurns: 12, timeout: '10m' } } };
    expect(selectTodoRuntimeProfile({ options, defaults: defaultOptions, runtimeProfile: 'profile-review' }).spec).toEqual({ budget: { maxTurns: 12 } });
  });

  it('returns to step defaults while preserving explicit overrides', () => {
    const options = { step: 'plan', runtimeProfile: 'profile-review', spec: { model: 'example-override' } };
    expect(selectTodoRuntimeProfile({ options, defaults, runtimeProfile: undefined })).toEqual({ step: 'plan', spec: { model: 'example-override' } });
  });
});
