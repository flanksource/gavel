import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { RunContext } from './providers';
import { defaultRunOptionsForAction, loadLastTodoRunOptions, reconcileTodoRunOptions, rememberTodoRunOptions } from './run';
import { loadPromptRunOptions, rememberPromptRunOptions } from './PromptRunButton';

const context: RunContext = {
  defaultMode: 'agent', defaultProvider: 'openai', models: [], runtimes: [], tools: [], efforts: ['medium'], lifecycle: { steps: [] },
  modes: [{ id: 'agent', label: 'Agent', provider: 'openai', agent: 'codex', defaultModel: 'example-default', driver: 'agent', mechanisms: [], models: [{ id: 'example-default', label: 'Default', provider: 'openai', reasoning: true }] }],
  runtimePresets: [{ id: 'review-preset', name: 'Review', scope: 'context', spec: { model: 'example-preset' } }],
  promptDefaults: { plan: { mode: 'agent', model: 'example-default', presets: ['review-preset'] }, verify: { presets: ['review-preset'] } },
};

beforeEach(() => {
  const values = new Map<string, string>();
  vi.stubGlobal('localStorage', { getItem: (key: string) => values.get(key) ?? null, setItem: (key: string, value: string) => values.set(key, value) });
});
afterEach(() => vi.unstubAllGlobals());

describe('preset-aware run history', () => {
  it('seeds ordered presets without promoting the resolved model to a request override', () => {
    expect(defaultRunOptionsForAction('plan', context)).toEqual({ step: 'plan', presets: ['review-preset'], spec: {} });
  });

  it('does not add mode, model, or effort when reconciling preset references', () => {
    const options = { step: 'plan', presets: ['review-preset'], spec: { budget: { maxTurns: 8 } } };
    expect(reconcileTodoRunOptions('plan', options, context)).toEqual(options);
  });

  it('round-trips presets and an explicit model override without replacing either', () => {
    const options = { step: 'plan', presets: ['review-preset'], spec: { model: 'example-override' } };
    rememberTodoRunOptions('plan', options, true);
    expect(loadLastTodoRunOptions('plan', context)).toEqual(options);
  });

  it('preserves verification presets through its separate history', () => {
    expect(loadPromptRunOptions('verification', context)).toEqual({ step: 'verify', presets: ['review-preset'], spec: {} });
    rememberPromptRunOptions('verification', { presets: ['review-preset'], spec: { budget: { maxTurns: 8 } } }, context);
    expect(loadPromptRunOptions('verification', context)).toEqual({ step: 'verify', presets: ['review-preset'], spec: expect.objectContaining({ budget: { maxTurns: 8 } }) });
  });
});
