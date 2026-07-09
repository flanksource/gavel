import { describe, expect, it } from 'vitest';
import type { RunContext } from './providers';
import {
  runButtonLabelForOptions,
  runChoicesForAction,
  runChoicesForRuntimeMode,
  shortTodoRunModelName,
} from './run';

const context: RunContext = {
  defaultBackend: 'codex-cmux',
  efforts: ['low', 'medium', 'high', 'xhigh'],
  tools: [],
  backends: [
    {
      id: 'codex-cmux',
      label: 'Codex cmux',
      provider: 'openai',
      agent: 'codex',
      defaultModel: 'gpt-5.5',
      driver: 'codex-cmux',
      mechanisms: [{ value: 'cmux', label: 'cmux (TUI)', driver: 'codex-cmux' }],
      models: [
        { id: 'gpt-5.5', provider: 'openai', label: 'GPT-5.5', reasoning: true, configured: true },
        { id: 'gpt-5.4', provider: 'openai', label: 'GPT-5.4', reasoning: true, configured: true },
      ],
      configured: true,
    },
    {
      id: 'claude-agent',
      label: 'Claude Agent',
      provider: 'anthropic',
      agent: 'claude',
      defaultModel: 'claude-opus-4-8',
      driver: 'claude-headless',
      mechanisms: [{ value: 'agent', label: 'agent', driver: 'claude-headless' }],
      models: [
        { id: 'claude-opus-4-8', provider: 'anthropic', label: 'Claude Opus 4.8', reasoning: true, configured: true },
      ],
      configured: true,
    },
  ],
};

describe('todo run model choices', () => {
  it('builds plan choices for every backend from run context', () => {
    const choices = runChoicesForAction(context, 'plan');

    expect(choices.map(choice => `${choice.backend.id}:${choice.modelID}`)).toEqual([
      'codex-cmux:gpt-5.5',
      'codex-cmux:gpt-5.4',
      'claude-agent:claude-opus-4-8',
    ]);
    expect(choices.find(choice => choice.backend.id === 'claude-agent')?.options).toMatchObject({
      driver: 'claude-headless',
      backend: 'claude-agent',
      model: 'claude-opus-4-8',
      runMode: 'plan',
      plan: true,
    });

    expect(runChoicesForAction(context, 'run', 'high')[0]?.options).toMatchObject({
      effort: 'high',
      runMode: 'run',
    });

    expect(runChoicesForRuntimeMode(context, 'run', 'cmux').map(choice => choice.backend.id)).toEqual([
      'codex-cmux',
      'codex-cmux',
    ]);
    expect(runChoicesForRuntimeMode(context, 'run', 'agent').map(choice => choice.backend.id)).toEqual([
      'claude-agent',
    ]);
  });

  it('labels primary buttons with mechanism and short model', () => {
    expect(runButtonLabelForOptions('plan', { driver: 'codex-cmux', backend: 'codex-cmux', model: 'gpt-5.5', runMode: 'plan' }, context)).toBe('Plan (cmux:gpt-5.5)');
    expect(runButtonLabelForOptions('run', { driver: 'claude-headless', backend: 'claude-agent', model: 'claude-opus-4-8', runMode: 'run' }, context)).toBe('Run (agent:opus-4.8)');
  });

  it('shortens provider and Claude version prefixes', () => {
    expect(shortTodoRunModelName('anthropic/claude-opus-4-8')).toBe('opus-4.8');
    expect(shortTodoRunModelName('Claude Opus 4.8')).toBe('opus-4.8');
    expect(shortTodoRunModelName('claude-sonnet-5')).toBe('sonnet-5');
    expect(shortTodoRunModelName('gpt-5.5')).toBe('gpt-5.5');
  });
});
