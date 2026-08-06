import { describe, expect, it } from 'vitest';
import { buildRunFamilies, type RunContext } from './providers';
import {
	defaultRunOptionsForAction,
	runButtonLabelForOptions,
	reconcileTodoRunOptions,
	runSpec,
	shortTodoRunModelName,
	todoRunOptionsForRuntimeChange,
	todoRunButtonPresentation,
	todoRunEffortPresentation,
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

describe('todo run runtime adapter', () => {
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

    expect(buildRunFamilies(captainContext)).toEqual([]);
  });

  it('maps RuntimeBar edits back to nested run options without losing advanced fields', () => {
    const initial = {
      ...defaultRunOptionsForAction('run', context),
      spec: {
        ...runSpec(defaultRunOptionsForAction('run', context)),
        budget: { timeout: '45m', maxTurns: 12 },
        setup: { cwd: '/workspace' },
        permissions: { mode: 'dontAsk' as const },
        prompt: { appendSystem: 'Keep the adapter seam.' },
        temperature: 0.7,
      },
    };

    const options = todoRunOptionsForRuntimeChange({
      action: 'run',
      context,
      options: initial,
      runtime: {
        ...runSpec(initial),
        backend: 'claude-agent',
        model: 'claude-opus-4-8',
        effort: 'high',
      },
    });

    expect(options).toMatchObject({
      driver: 'claude-headless',
      runMode: 'run',
      spec: {
        backend: 'claude-agent',
        model: 'claude-opus-4-8',
        effort: 'high',
        budget: { timeout: '45m', maxTurns: 12 },
        setup: { cwd: '/workspace' },
        permissions: { mode: 'dontAsk' },
        prompt: { appendSystem: 'Keep the adapter seam.' },
        workflow: { commits: [{ on: 'run', gates: 'full' }] },
      },
    });
    expect(options.spec).not.toHaveProperty('temperature');
  });

  it('keeps plan lifecycle fields when RuntimeBar changes backend, model, and effort', () => {
    const initial = defaultRunOptionsForAction('plan', context);
    expect(todoRunOptionsForRuntimeChange({
      action: 'plan',
      context,
      options: initial,
      runtime: {
        ...runSpec(initial),
        backend: 'claude-agent',
        model: 'claude-opus-4-8',
        effort: 'high',
      },
    })).toMatchObject({
      driver: 'claude-headless',
      runMode: 'plan',
      plan: true,
      spec: { backend: 'claude-agent', model: 'claude-opus-4-8', effort: 'high' },
    });
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
  });

  it('gives every effort tier a distinct tone and makes ultra rainbow', () => {
    const classes = (['low', 'medium', 'high', 'xhigh', 'max'] as const).map(effort => todoRunEffortPresentation(effort).className);
    expect(new Set(classes).size).toBe(classes.length);
    expect(todoRunEffortPresentation('ultra').className).toContain('bg-gradient-to-r');
    expect(todoRunEffortPresentation('ultra').className).toContain('from-fuchsia-500');
  });
});
