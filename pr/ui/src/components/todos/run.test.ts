import { describe, expect, it } from 'vitest';
import { buildRunFamilies, type RunContext } from './providers';
import {
	runButtonLabelForOptions,
	runChoicesForAction,
	runChoicesForRuntimeMode,
	reconcileTodoRunOptions,
	shortTodoRunModelName,
	todoRunButtonPresentation,
	todoRunEffortPresentation,
	todoRunModelFamilyName,
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
        {
          id: 'claude-opus-4-8',
          provider: 'anthropic',
          label: 'Claude Opus 4.8',
          capabilitiesKnown: true,
          reasoning: true,
          temperature: false,
          supportedEfforts: ['low', 'high'],
          defaultEffort: 'high',
          configured: true,
        },
      ],
      configured: true,
    },
  ],
};

describe('todo run model choices', () => {
  it('does not invent providers or models missing from the Captain context', () => {
    const captainContext: RunContext = {
      defaultBackend: 'codex-agent',
      efforts: ['medium'],
      tools: [],
      backends: [{
        id: 'codex-agent',
        label: 'Codex Agent',
        provider: 'openai',
        agent: 'codex',
        defaultModel: '',
        driver: 'codex-headless',
        mechanisms: [{ value: 'agent', label: 'Agent', driver: 'codex-headless' }],
        models: [],
        configured: false,
        modelError: 'Captain returned no models',
      }],
    };

    expect(runChoicesForAction(captainContext, 'run')).toEqual([]);
    expect(buildRunFamilies(captainContext).map(family => family.id)).toEqual(['codex']);
  });

  it('builds plan choices for every backend from run context', () => {
    const choices = runChoicesForAction(context, 'plan');

    expect(choices.map(choice => `${choice.backend.id}:${choice.modelID}`)).toEqual([
      'codex-cmux:gpt-5.5',
      'codex-cmux:gpt-5.4',
      'claude-agent:claude-opus-4-8',
    ]);
    expect(choices.find(choice => choice.backend.id === 'claude-agent')?.options).toMatchObject({
      driver: 'claude-headless',
      runMode: 'plan',
      plan: true,
      spec: { backend: 'claude-agent', model: 'claude-opus-4-8' },
    });

    expect(runChoicesForAction(context, 'run', 'high')[0]?.options).toMatchObject({
      runMode: 'run',
      spec: { effort: 'high', workflow: { commits: [{ on: 'run', gates: 'full' }] } },
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
    expect(runButtonLabelForOptions('plan', { driver: 'codex-cmux', runMode: 'plan', spec: { backend: 'codex-cmux', model: 'gpt-5.5' } }, context)).toBe('Plan (cmux:gpt-5.5)');
    expect(runButtonLabelForOptions('run', { driver: 'claude-headless', runMode: 'run', spec: { backend: 'claude-agent', model: 'claude-opus-4-8' } }, context)).toBe('Run (Agent:opus-4.8)');
	});

  it('builds compact primary-button model and effort presentation', () => {
    expect(todoRunButtonPresentation({
      driver: 'claude-headless',
      runMode: 'run',
      spec: { backend: 'claude-agent', model: 'claude-opus-4-8', effort: 'high' },
    }, context)).toMatchObject({
      model: 'opus-4.8',
      effort: 'high',
      provider: { id: 'claude', iconColor: '#D97757' },
    });

    const fixedEffortContext: RunContext = {
      ...context,
      backends: context.backends.map(backend => backend.id !== 'claude-agent' ? backend : {
        ...backend,
        models: backend.models.map(model => ({
          ...model,
          capabilitiesKnown: true,
          reasoning: false,
          supportedEfforts: [],
        })),
      }),
    };
    expect(todoRunButtonPresentation({
      driver: 'claude-headless',
      runMode: 'run',
      spec: { backend: 'claude-agent', model: 'claude-opus-4-8', effort: 'high' },
    }, fixedEffortContext).effort).toBeUndefined();

    const openAI = todoRunButtonPresentation({
      driver: 'codex-cmux',
      runMode: 'run',
      spec: { backend: 'codex-cmux', model: 'gpt-5.5', effort: 'medium' },
    }, context);
    expect(openAI).toMatchObject({ model: 'gpt-5.5', provider: { id: 'codex' } });
    expect(openAI.provider?.iconColor).toBe('#10A37F');
  });

	it('reconciles stale remembered backend models while preserving advanced options', () => {
		expect(reconcileTodoRunOptions('run', {
			driver: 'codex-cmux',
			runMode: 'run',
			spec: {
				backend: 'codex-cmux',
				model: 'gpt-5.3-removed',
				effort: 'high',
				budget: { timeout: '45m', maxTurns: 12 },
			},
		}, context)).toMatchObject({
			driver: 'codex-cmux',
			runMode: 'run',
			spec: {
				backend: 'codex-cmux',
				model: 'gpt-5.5',
				effort: 'high',
				budget: { timeout: '45m', maxTurns: 12 },
			},
		});

		expect(reconcileTodoRunOptions('plan', {
			driver: 'claude-headless',
			runMode: 'plan',
			spec: { backend: 'claude-agent', model: 'claude-opus-4-8', effort: 'xhigh', temperature: 0.7 },
		}, context)).toEqual(expect.objectContaining({
			runMode: 'plan',
			spec: expect.objectContaining({
				backend: 'claude-agent',
				model: 'claude-opus-4-8',
				effort: 'high',
			}),
		}));
		expect(reconcileTodoRunOptions('plan', {
			driver: 'claude-headless',
			spec: { backend: 'claude-agent', model: 'claude-opus-4-8', effort: 'xhigh', temperature: 0.7 },
		}, context)).not.toHaveProperty('spec.temperature');
	});

  it('shortens provider and Claude version prefixes', () => {
    expect(shortTodoRunModelName('anthropic/claude-opus-4-8')).toBe('opus-4.8');
    expect(shortTodoRunModelName('Claude Opus 4.8')).toBe('opus-4.8');
    expect(shortTodoRunModelName('claude-sonnet-5')).toBe('sonnet-5');
    expect(shortTodoRunModelName('gpt-5.5')).toBe('gpt-5.5');
    expect(todoRunModelFamilyName('Claude Opus 4.8')).toBe('Opus');
    expect(todoRunModelFamilyName('claude-fable-5')).toBe('Fable');
    expect(todoRunModelFamilyName('gpt-5.6-sol')).toBe('Sol');
    expect(todoRunModelFamilyName('gpt-5.5')).toBe('GPT');
  });

  it('gives every effort tier a distinct tone and makes ultra rainbow', () => {
    const classes = (['low', 'medium', 'high', 'xhigh', 'max'] as const).map(effort => todoRunEffortPresentation(effort).className);
    expect(new Set(classes).size).toBe(classes.length);
    expect(todoRunEffortPresentation('ultra').className).toContain('bg-gradient-to-r');
    expect(todoRunEffortPresentation('ultra').className).toContain('from-fuchsia-500');
  });
});
