import { describe, expect, it } from 'vitest';
import type { RunContext } from './providers';
import { defaultRunOptionsForAction, reconcileTodoRunOptions, runButtonQualifierForOptions, todoRunButtonPresentation } from './run';

const context: RunContext = {
  modes: [{ id: 'agent', label: 'Agent', provider: 'openai', agent: 'codex', defaultModel: 'catalog-first', driver: 'agent', mechanisms: [], models: [{ id: 'catalog-first', label: 'Catalog first', provider: 'openai', reasoning: true }] }],
  models: [], runtimes: [], tools: [], efforts: [], lifecycle: { steps: [] },
};

describe('run defaults remain descriptive', () => {
  it('leaves an unconfigured run sparse even when the catalog offers a default model', () => {
    expect(defaultRunOptionsForAction('run', context)).toEqual({ step: 'run', spec: {} });
    expect(runButtonQualifierForOptions({ step: 'run', spec: {} }, context)).toBe('(Choose model)');
    expect(todoRunButtonPresentation({ step: 'run', spec: {} }, context)).toEqual({ provider: undefined, model: 'Choose model', effort: undefined });
  });

  it('displays configured step defaults without copying their fields into a request', () => {
    const configured = { ...context, promptDefaults: { run: { mode: 'agent', model: 'catalog-first' } } };
    expect(defaultRunOptionsForAction('run', configured)).toEqual({ step: 'run', spec: {} });
    expect(runButtonQualifierForOptions({ step: 'run', spec: {} }, configured)).toBe('(Agent:Catalog first)');
  });

  it('preserves an unavailable explicit model for the shared preflight to diagnose', () => {
    const options = { step: 'run', spec: { model: 'operator-model', budget: { maxTurns: 7 } } };
    expect(reconcileTodoRunOptions('run', options, context)).toEqual(options);
  });

  it('describes a model-free verification as a fixture', () => {
    expect(todoRunButtonPresentation({ step: 'verify', spec: {} }, context)).toEqual({ provider: undefined, model: 'Fixture', effort: undefined });
  });
});
