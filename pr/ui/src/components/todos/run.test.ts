import { describe, expect, it } from 'vitest';
import { buildRunFamilies, type RunContext } from './providers';
import {
  buildTodoRunPayload,
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
  defaultBackend: 'cmux',
  defaultProvider: 'openai',
  efforts: ['low', 'medium', 'high', 'xhigh'],
  tools: [],
  runtimes: [],
  models: [],
  backends: [
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
        { id: 'gpt-5.4', provider: 'openai', label: 'GPT-5.4', reasoning: true, configured: true },
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
  it('uses one complete payload for preview and submit without resending a generated prompt', () => {
    const runtime = {
      backend: 'cmux',
      model: 'gpt-5.5',
      effort: 'high' as const,
      budget: { timeout: '45m', maxTurns: 12 },
      permissions: { mode: 'dontAsk' as const },
      setup: { cwd: '/workspace' },
      workflow: { commits: [{ on: 'run' as const, gates: 'full' as const }] },
    };

    expect(buildTodoRunPayload({
      ref: 'todo-123',
      driver: 'cmux',
      runBackend: 'cmux',
      runtime,
      mode: 'run',
      resume: true,
      promptDraft: 'Generated prompt with effort directive',
      promptDirty: false,
    })).toEqual({
      ref: 'todo-123',
      driver: 'cmux',
      runMode: 'run',
      resume: true,
      spec: {
        ...runtime,
        backend: 'cmux',
      },
    });

    expect(buildTodoRunPayload({
      ref: 'todo-123',
      driver: 'cmux',
      runBackend: 'cmux',
      runtime,
      mode: 'plan',
      resume: false,
      promptDraft: 'Use this edited prompt',
      promptDirty: true,
    })).toMatchObject({
      runMode: 'plan',
      plan: true,
      spec: {
        budget: { timeout: '45m', maxTurns: 12 },
        permissions: { mode: 'dontAsk' },
        setup: { cwd: '/workspace' },
        prompt: { user: 'Use this edited prompt' },
      },
    });
  });

  it('rejects a missing Captain runtime catalog instead of inventing one', () => {
    const captainContext: RunContext = {
      defaultBackend: 'agent',
      defaultProvider: 'openai',
      efforts: ['medium'],
      tools: [],
      runtimes: [],
      models: [],
      backends: [{
        id: 'agent',
        label: 'Codex Agent',
        provider: 'openai',
        agent: 'codex',
        defaultModel: '',
        driver: 'agent',
        mechanisms: [{ value: 'agent', label: 'Agent', driver: 'agent' }],
        models: [],
        configured: false,
        modelError: 'Captain returned no models',
      }],
    };

    expect(() => buildRunFamilies(captainContext)).toThrow('Captain returned no runtime catalog');
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
        backend: 'agent',
        model: 'claude-opus-4-8',
        effort: 'high',
      },
    });

    expect(options).toMatchObject({
      driver: 'agent',
      runMode: 'run',
      spec: {
        backend: 'agent',
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

  // Seeding every action from defaultBackend sent that backend as if the
  // operator had chosen it, which outranks the frontmatter the prompt pins.
  // todos-triage.prompt pins `model: claude` and allows only Read/Glob/Grep, so
  // under a codex default Captain refused the run outright: "backend
  // codex-agent cannot enforce a per-tool policy (Glob, Grep, Read)".
  it('seeds a prompt from its own resolved runtime, not the account default', () => {
    const withPromptDefaults: RunContext = {
      ...context,
      promptDefaults: {
        triage: { backend: 'agent', model: 'claude-opus-4-8' },
      },
    };

    expect(runSpec(defaultRunOptionsForAction('triage', withPromptDefaults))).toMatchObject({
      backend: 'agent',
      model: 'claude-opus-4-8',
    });
    // An action the server reported no default for still falls back to it.
    expect(runSpec(defaultRunOptionsForAction('run', withPromptDefaults))).toMatchObject({
      backend: 'cmux',
    });
  });

  it('keeps a remembered backend that disagrees with the prompt default', () => {
    const withPromptDefaults: RunContext = {
      ...context,
      promptDefaults: { triage: { backend: 'agent', model: 'claude-opus-4-8' } },
    };

    expect(runSpec(reconcileTodoRunOptions('triage', {
      prompt: 'triage',
      driver: 'cmux',
      spec: { backend: 'cmux', model: 'gpt-5.5' },
    }, withPromptDefaults))).toMatchObject({ backend: 'cmux' });
  });

  it('keeps plan lifecycle fields when RuntimeBar changes backend, model, and effort', () => {
    const initial = defaultRunOptionsForAction('plan', context);
    expect(todoRunOptionsForRuntimeChange({
      action: 'plan',
      context,
      options: initial,
      runtime: {
        ...runSpec(initial),
        backend: 'agent',
        model: 'claude-opus-4-8',
        effort: 'high',
      },
    })).toMatchObject({
      driver: 'agent',
      runMode: 'plan',
      plan: true,
      spec: { backend: 'agent', model: 'claude-opus-4-8', effort: 'high' },
    });
  });

	it('labels primary buttons with mechanism and short model', () => {
    expect(runButtonLabelForOptions('plan', { driver: 'cmux', runMode: 'plan', spec: { backend: 'cmux', model: 'gpt-5.5' } }, context)).toBe('Plan (cmux:gpt-5.5)');
    expect(runButtonLabelForOptions('run', { driver: 'agent', runMode: 'run', spec: { backend: 'agent', model: 'claude-opus-4-8' } }, context)).toBe('Run (Agent:opus-4.8)');
	});

  it('builds compact primary-button model and effort presentation', () => {
    expect(todoRunButtonPresentation({
      driver: 'agent',
      runMode: 'run',
      spec: { backend: 'agent', model: 'claude-opus-4-8', effort: 'high' },
    }, context)).toMatchObject({
      model: 'opus-4.8',
      effort: 'high',
      provider: { id: 'claude', iconColor: '#D97757' },
    });

    const fixedEffortContext: RunContext = {
      ...context,
      backends: context.backends.map(backend => backend.agent !== 'claude' ? backend : {
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
      driver: 'agent',
      runMode: 'run',
      spec: { backend: 'agent', model: 'claude-opus-4-8', effort: 'high' },
    }, fixedEffortContext).effort).toBeUndefined();

    const openAI = todoRunButtonPresentation({
      driver: 'cmux',
      runMode: 'run',
      spec: { backend: 'cmux', model: 'gpt-5.5', effort: 'medium' },
    }, context);
    expect(openAI).toMatchObject({ model: 'gpt-5.5', provider: { id: 'codex' } });
    expect(openAI.provider?.iconColor).toBe('#10A37F');
  });

	it('reconciles stale remembered backend models while preserving advanced options', () => {
		expect(reconcileTodoRunOptions('run', {
			driver: 'cmux',
			runMode: 'run',
			spec: {
				backend: 'cmux',
				model: 'gpt-5.3-removed',
				effort: 'high',
				budget: { timeout: '45m', maxTurns: 12 },
			},
		}, context)).toMatchObject({
			driver: 'cmux',
			runMode: 'run',
			spec: {
				backend: 'cmux',
				model: 'gpt-5.5',
				effort: 'high',
				budget: { timeout: '45m', maxTurns: 12 },
			},
		});

		expect(reconcileTodoRunOptions('plan', {
			driver: 'agent',
			runMode: 'plan',
			spec: { backend: 'agent', model: 'claude-opus-4-8', effort: 'xhigh', temperature: 0.7 },
		}, context)).toEqual(expect.objectContaining({
			runMode: 'plan',
			spec: expect.objectContaining({
				backend: 'agent',
				model: 'claude-opus-4-8',
				effort: 'high',
			}),
		}));
		expect(reconcileTodoRunOptions('plan', {
			driver: 'agent',
			spec: { backend: 'agent', model: 'claude-opus-4-8', effort: 'xhigh', temperature: 0.7 },
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
