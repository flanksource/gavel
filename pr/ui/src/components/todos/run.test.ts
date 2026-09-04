import { beforeEach, describe, expect, it } from 'vitest';
import { buildRunFamilies, type RunContext } from './providers';
import {
  buildTodoRunPayload,
  defaultRunOptionsForAction,
	loadLastTodoRunOptions,
	requestStepFor,
	rememberTodoRunOptions,
	runButtonLabelForOptions,
	reconcileTodoRunOptions,
	runSpec,
	shortTodoRunModelName,
	todoRunOptionsForRuntimeChange,
	todoRunButtonPresentation,
	todoRunEffortPresentation,
} from './run';

const RUN_CHOICE_STORAGE_KEY_V2 = 'gavel.pr-ui.todoRunChoices.v2';
const RUN_CHOICE_STORAGE_KEY_V3 = 'gavel.pr-ui.todoRunChoices.v3';

const context: RunContext = {
  defaultMode: 'cmux',
  defaultProvider: 'openai',
  efforts: ['low', 'medium', 'high', 'xhigh'],
  tools: [],
  runtimes: [],
  models: [],
  lifecycle: { steps: [
    { name: 'plan', label: 'Plan', prompt: 'plan', readOnly: false },
    { name: 'run', label: 'Run', prompt: 'run', readOnly: false },
    { name: 'triage', label: 'Triage', prompt: 'triage', readOnly: false },
  ] },
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
  beforeEach(() => {
    localStorage.clear();
  });

  // POST /api/todos/run decodes strictly and rejects runMode/driver/prompt at
  // the top level, so the payload builder must not leak them — only `step`
  // (the lifecycle step name) plus spec/resume/ref survive onto the wire.
  it('uses one complete payload for preview and submit without resending a generated prompt', () => {
    const runtime = {
      mode: 'cmux',
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
      runMode: 'cmux',
      runtime,
      mode: 'run',
      resume: true,
      promptDraft: 'Generated prompt with effort directive',
      promptDirty: false,
    })).toEqual({
      ref: 'todo-123',
      step: 'run',
      resume: true,
      spec: {
        ...runtime,
        mode: 'cmux',
      },
    });

    expect(buildTodoRunPayload({
      ref: 'todo-123',
      driver: 'cmux',
      runMode: 'cmux',
      runtime,
      mode: 'plan',
      resume: false,
      promptDraft: 'Use this edited prompt',
      promptDirty: true,
    })).toMatchObject({
      step: 'plan',
      spec: {
        budget: { timeout: '45m', maxTurns: 12 },
        permissions: { mode: 'dontAsk' },
        setup: { cwd: '/workspace' },
        prompt: { user: 'Use this edited prompt' },
      },
    });
    // Neither payload carries the fields the endpoint now rejects.
    expect(buildTodoRunPayload({
      ref: 'todo-123',
      driver: 'cmux',
      runMode: 'cmux',
      runtime,
      mode: 'triage',
      resume: false,
      promptDraft: '',
      promptDirty: false,
    })).not.toHaveProperty('driver');
  });

  it('names triage by its own prompt rather than collapsing it into plan/run', () => {
    expect(buildTodoRunPayload({
      ref: 'todo-123',
      driver: 'cmux',
      runMode: 'cmux',
      runtime: { mode: 'cmux' },
      mode: 'triage',
      resume: false,
      promptDraft: '',
      promptDirty: false,
    }).step).toBe('triage');
  });

  describe('requestStepFor', () => {
    it('prefers an explicit step over the legacy runMode/plan/prompt fields', () => {
      expect(requestStepFor({ step: 'custom-review', runMode: 'plan', plan: true })).toBe('custom-review');
    });

    it('falls back to a custom prompt name when no step is set', () => {
      expect(requestStepFor({ prompt: 'triage' })).toBe('triage');
    });

    it('derives plan/run from the legacy flags as the last resort', () => {
      expect(requestStepFor({ runMode: 'plan' })).toBe('plan');
      expect(requestStepFor({ plan: true })).toBe('plan');
      expect(requestStepFor({})).toBe('run');
    });
  });

  describe('run-choice storage', () => {
    // v3: the request body now sends `step`, not runMode/driver/prompt — a v2
    // entry's shape is stale under the new contract, so it must never be read
    // back, even though the value is still sitting in localStorage.
    it('ignores a v2 entry instead of migrating it', () => {
      localStorage.setItem(RUN_CHOICE_STORAGE_KEY_V2, JSON.stringify({
        last: { run: { driver: 'cmux', runMode: 'run', spec: { mode: 'cmux', model: 'stale-v2-model' } } },
        recentAdvanced: {},
      }));

      expect(runSpec(loadLastTodoRunOptions('run', context)).model).not.toBe('stale-v2-model');
      expect(localStorage.getItem(RUN_CHOICE_STORAGE_KEY_V3)).toBeNull();
    });

    it('remembers a run under the v3 key, keyed by step', () => {
      rememberTodoRunOptions('run', { driver: 'cmux', spec: { mode: 'cmux', model: 'gpt-5.5' } });
      const stored = JSON.parse(localStorage.getItem(RUN_CHOICE_STORAGE_KEY_V3) ?? '{}');
      expect(stored.last).toHaveProperty('run');
      expect(stored.last).not.toHaveProperty('plan');
    });
  });

  it('rejects a missing Captain runtime catalog instead of inventing one', () => {
    const captainContext: RunContext = {
      defaultMode: 'agent',
      defaultProvider: 'openai',
      efforts: ['medium'],
      tools: [],
      runtimes: [],
      models: [],
      lifecycle: { steps: [] },
      modes: [{
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
        mode: 'agent',
        model: 'claude-opus-4-8',
        effort: 'high',
      },
    });

    expect(options).toMatchObject({
      driver: 'agent',
      runMode: 'run',
      spec: {
        mode: 'agent',
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

  // Seeding every action from defaultMode sent that mode as if the
  // operator had chosen it, which outranks the frontmatter the prompt pins.
  // todos-triage.prompt pins `model: claude` and allows only Read/Glob/Grep, so
  // under a codex default Captain refused the run outright: "mode
  // codex-agent cannot enforce a per-tool policy (Glob, Grep, Read)".
  // promptDefaults is keyed by lifecycle step name (run/plan/triage/...), the
  // same identifier a request's `step` field carries — not by a separate
  // "prompt name" axis.
  it('seeds a prompt from its own resolved runtime, not the account default', () => {
    const withPromptDefaults: RunContext = {
      ...context,
      promptDefaults: {
        triage: { mode: 'agent', model: 'claude-opus-4-8' },
      },
    };

    expect(runSpec(defaultRunOptionsForAction('triage', withPromptDefaults))).toMatchObject({
      mode: 'agent',
      model: 'claude-opus-4-8',
    });
    // An action the server reported no default for still falls back to it.
    expect(runSpec(defaultRunOptionsForAction('run', withPromptDefaults))).toMatchObject({
      mode: 'cmux',
    });
  });

  it("keeps a remembered mode that disagrees with the prompt default", () => {
    const withPromptDefaults: RunContext = {
      ...context,
      promptDefaults: { triage: { mode: 'agent', model: 'claude-opus-4-8' } },
    };

    expect(runSpec(reconcileTodoRunOptions('triage', {
      prompt: 'triage',
      driver: 'cmux',
      spec: { mode: 'cmux', model: 'gpt-5.5' },
    }, withPromptDefaults))).toMatchObject({ mode: 'cmux' });
  });

  it("keeps plan lifecycle fields when RuntimeBar changes mode, model, and effort", () => {
    const initial = defaultRunOptionsForAction('plan', context);
    expect(todoRunOptionsForRuntimeChange({
      action: 'plan',
      context,
      options: initial,
      runtime: {
        ...runSpec(initial),
        mode: 'agent',
        model: 'claude-opus-4-8',
        effort: 'high',
      },
    })).toMatchObject({
      driver: 'agent',
      runMode: 'plan',
      plan: true,
      spec: { mode: 'agent', model: 'claude-opus-4-8', effort: 'high' },
    });
  });

	it('labels primary buttons with mechanism and short model', () => {
    expect(runButtonLabelForOptions('plan', { driver: 'cmux', runMode: 'plan', spec: { mode: 'cmux', model: 'gpt-5.5' } }, context)).toBe('Plan (cmux:gpt-5.5)');
    expect(runButtonLabelForOptions('run', { driver: 'agent', runMode: 'run', spec: { mode: 'agent', model: 'claude-opus-4-8' } }, context)).toBe('Run (Agent:opus-4.8)');
	});

  it('builds compact primary-button model and effort presentation', () => {
    expect(todoRunButtonPresentation({
      driver: 'agent',
      runMode: 'run',
      spec: { mode: 'agent', model: 'claude-opus-4-8', effort: 'high' },
    }, context)).toMatchObject({
      model: 'opus-4.8',
      effort: 'high',
      provider: { id: 'claude', iconColor: '#D97757' },
    });

    const fixedEffortContext: RunContext = {
      ...context,
      modes: context.modes.map(runtime => runtime.agent !== "claude" ? runtime : {
        ...runtime,
        models: runtime.models.map(model => ({
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
      spec: { mode: 'agent', model: 'claude-opus-4-8', effort: 'high' },
    }, fixedEffortContext).effort).toBeUndefined();

    const openAI = todoRunButtonPresentation({
      driver: 'cmux',
      runMode: 'run',
      spec: { mode: 'cmux', model: 'gpt-5.5', effort: 'medium' },
    }, context);
    expect(openAI).toMatchObject({ model: 'gpt-5.5', provider: { id: 'codex' } });
    expect(openAI.provider?.iconColor).toBe('#10A37F');
  });

	it("reconciles stale remembered runtime models while preserving advanced options", () => {
		expect(reconcileTodoRunOptions('run', {
			driver: 'cmux',
			runMode: 'run',
			spec: {
				mode: 'cmux',
				model: 'gpt-5.3-removed',
				effort: 'high',
				budget: { timeout: '45m', maxTurns: 12 },
			},
		}, context)).toMatchObject({
			driver: 'cmux',
			runMode: 'run',
			spec: {
				mode: 'cmux',
				model: 'gpt-5.5',
				effort: 'high',
				budget: { timeout: '45m', maxTurns: 12 },
			},
		});

		expect(reconcileTodoRunOptions('plan', {
			driver: 'agent',
			runMode: 'plan',
			spec: { mode: 'agent', model: 'claude-opus-4-8', effort: 'xhigh', temperature: 0.7 },
		}, context)).toEqual(expect.objectContaining({
			runMode: 'plan',
			spec: expect.objectContaining({
				mode: 'agent',
				model: 'claude-opus-4-8',
				effort: 'high',
			}),
		}));
		expect(reconcileTodoRunOptions('plan', {
			driver: 'agent',
			spec: { mode: 'agent', model: 'claude-opus-4-8', effort: 'xhigh', temperature: 0.7 },
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
