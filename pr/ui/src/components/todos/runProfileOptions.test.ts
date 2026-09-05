import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { RunContext } from './providers';
import { defaultRunOptionsForAction, loadLastTodoRunOptions, reconcileTodoRunOptions, rememberTodoRunOptions } from './run';
import { loadPromptRunOptions, rememberPromptRunOptions } from './PromptRunButton';

const context: RunContext = {
  defaultMode: 'agent', defaultProvider: 'openai', models: [], runtimes: [], tools: [], efforts: ['medium'], lifecycle: { steps: [] },
  modes: [{ id: 'agent', label: 'Agent', provider: 'openai', agent: 'codex', defaultModel: 'example-default', driver: 'agent', mechanisms: [], models: [{ id: 'example-default', label: 'Default', provider: 'openai', reasoning: true }] }],
  runtimeProfiles: [{ id: 'review-profile', name: 'Review', model: 'example-profile', presets: [] }],
  promptDefaults: { plan: { mode: 'agent', model: 'example-default', runtimeProfile: 'review-profile' }, verify: { runtimeProfile: 'review-profile' } },
};

beforeEach(() => {
  const values = new Map<string, string>();
  vi.stubGlobal('localStorage', { getItem: (key: string) => values.get(key) ?? null, setItem: (key: string, value: string) => values.set(key, value) });
});
afterEach(() => vi.unstubAllGlobals());

describe('profile-aware run history', () => {
  it('seeds a named profile without promoting the resolved model to a request override', () => {
    expect(defaultRunOptionsForAction('plan', context)).toEqual({ step: 'plan', runtimeProfile: 'review-profile', spec: {} });
  });

  it('does not add mode, model, or effort when reconciling a profile reference', () => {
    const options = { step: 'plan', runtimeProfile: 'review-profile', spec: { budget: { maxTurns: 8 } } };
    expect(reconcileTodoRunOptions('plan', options, context)).toEqual(options);
  });

  it('round-trips a profile and an explicit model override without replacing either', () => {
    const options = { step: 'plan', runtimeProfile: 'review-profile', spec: { model: 'example-override' } };
    rememberTodoRunOptions('plan', options, true);
    expect(loadLastTodoRunOptions('plan', context)).toEqual(options);
  });

  it('preserves a verification profile through its separate history', () => {
    expect(loadPromptRunOptions('verification', context)).toEqual({ step: 'verify', runtimeProfile: 'review-profile', spec: {} });
    rememberPromptRunOptions('verification', { runtimeProfile: 'review-profile', spec: { budget: { maxTurns: 8 } } }, context);
    expect(loadPromptRunOptions('verification', context)).toEqual({ step: 'verify', runtimeProfile: 'review-profile', spec: expect.objectContaining({ budget: { maxTurns: 8 } }) });
  });
});
