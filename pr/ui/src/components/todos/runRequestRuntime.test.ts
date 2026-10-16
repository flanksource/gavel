// Tests what a run request may choose of its runtime: a hosted org's runs use the repository's
// runtime for each step, so they send no model, mode or presets; a local run sends its choice.
import { describe, expect, it } from 'vitest';
import { runRequestRuntime } from './run';

const chosen = {
  ref: 'todo-1',
  step: 'plan',
  presets: ['fast'],
  spec: { model: 'claude-opus-5-5', mode: 'agent', provider: 'anthropic', id: 'x', fallbacks: ['y'], effort: 'high', prompt: { user: 'go' } },
};

describe('runRequestRuntime', () => {
  it('drops a hosted run\'s runtime choice and keeps the rest of its spec', () => {
    expect(runRequestRuntime(chosen, 'clerk')).toEqual({
      ref: 'todo-1',
      step: 'plan',
      spec: { effort: 'high', prompt: { user: 'go' } },
    });
  });

  it('keeps a local run\'s runtime choice', () => {
    expect(runRequestRuntime(chosen, 'none')).toEqual(chosen);
  });
});
