import type { RunContext } from './providers';

export const PLAN_ACTIONS_CONTEXT: RunContext = {
  defaultMode: 'agent',
  defaultProvider: 'openai',
  efforts: ['low', 'medium', 'high', 'xhigh'],
  tools: [],
  lifecycle: { steps: [
    { name: 'plan', label: 'Plan', prompt: 'plan', readOnly: false },
    { name: 'run', label: 'Run', prompt: 'run', readOnly: false },
  ] },
  runtimes: [
    { family: 'codex', provider: 'openai', catalogPrefix: 'openai', modes: [{ mode: 'agent', schema: { type: 'object' } }, { mode: 'cmux', schema: { type: 'object' } }] },
    { family: 'claude', provider: 'anthropic', catalogPrefix: 'anthropic', modes: [{ mode: 'agent', schema: { type: 'object' } }] },
  ],
  models: [
    { id: 'gpt-5.5', provider: 'openai', label: 'GPT-5.5', reasoning: true, configured: true, runtime: { model: 'gpt-5.5' } },
    { id: 'claude-opus-4-8', provider: 'anthropic', label: 'Claude Opus 4.8', reasoning: true, configured: true, runtime: { model: 'claude-opus-4-8' } },
  ],
  modes: [
    {
      id: 'cmux',
      label: 'Codex cmux',
      provider: 'openai',
      agent: 'codex',
      defaultModel: 'gpt-5.5',
      driver: 'cmux',
      mechanisms: [{ value: 'cmux', label: 'cmux (TUI)', driver: 'cmux' }],
      models: [
        { id: 'gpt-5.5', provider: 'openai', label: 'GPT-5.5', reasoning: true, configured: true },
      ],
      configured: true,
    },
    {
      id: 'agent',
      label: 'Codex Agent',
      provider: 'openai',
      agent: 'codex',
      defaultModel: 'gpt-5.5',
      driver: 'agent',
      mechanisms: [{ value: 'agent', label: 'Agent', driver: 'agent' }],
      models: [
        { id: 'gpt-5.5', provider: 'openai', label: 'GPT-5.5', reasoning: true, configured: true },
      ],
      configured: true,
    },
    {
      id: 'agent',
      label: 'Claude Agent',
      provider: 'anthropic',
      agent: 'claude',
      defaultModel: 'claude-opus-4-8',
      driver: 'agent',
      mechanisms: [{ value: 'agent', label: 'agent', driver: 'agent' }],
      models: [
        { id: 'claude-opus-4-8', provider: 'anthropic', label: 'Claude Opus 4.8', reasoning: true, configured: true },
      ],
      configured: true,
    },
  ],
};
