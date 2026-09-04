import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { TodoRunActionButton } from './TodoRunActionButton';
import type { RunContext } from './providers';
import { queryTestWrapper } from './queryTestWrapper';

const context: RunContext = {
  defaultMode: 'agent',
  defaultProvider: 'openai',
  efforts: ['low', 'medium', 'high', 'xhigh'],
  tools: [],
  lifecycle: { steps: [
    { name: 'plan', label: 'Plan', prompt: 'plan', readOnly: false },
    { name: 'run', label: 'Run', prompt: 'run', readOnly: false },
  ] },
  runtimes: [
    { family: 'codex', provider: 'openai', catalogPrefix: 'openai', modes: [{ mode: 'agent', schema: { type: 'object' } }] },
    { family: 'claude', provider: 'anthropic', catalogPrefix: 'anthropic', modes: [{ mode: 'agent', schema: { type: 'object' } }, { mode: 'cli', schema: { type: 'object' } }] },
  ],
  models: [
    { id: 'gpt-5.5', provider: 'openai', label: 'GPT-5.5', reasoning: true, configured: true, runtime: { model: 'gpt-5.5' } },
    { id: 'claude-opus-4-8', provider: 'anthropic', label: 'Claude Opus 4.8', capabilitiesKnown: true, reasoning: true, supportedEfforts: ['low', 'high'], defaultEffort: 'high', configured: true, runtime: { model: 'claude-opus-4-8' } },
  ],
  modes: [
    {
      id: 'agent',
      label: 'Codex Agent',
      provider: 'openai',
      agent: 'codex',
      defaultModel: 'gpt-5.5',
      driver: 'agent',
      mechanisms: [{ value: 'agent', label: 'Agent', driver: 'agent' }],
      models: [{ id: 'gpt-5.5', provider: 'openai', label: 'GPT-5.5', reasoning: true, configured: true }],
      configured: true,
    },
    {
      id: 'agent',
      label: 'Claude Agent',
      provider: 'anthropic',
      agent: 'claude',
      defaultModel: 'claude-opus-4-8',
      driver: 'agent',
      mechanisms: [{ value: 'agent', label: 'Agent', driver: 'agent' }],
      models: [{
        id: 'claude-opus-4-8',
        provider: 'anthropic',
        label: 'Claude Opus 4.8',
        capabilitiesKnown: true,
        reasoning: true,
        supportedEfforts: ['low', 'high'],
        defaultEffort: 'high',
        configured: true,
      }],
      configured: true,
    },
    {
      id: 'cli',
      label: 'Claude CLI',
      provider: 'anthropic',
      agent: 'claude',
      defaultModel: 'claude-opus-4-8',
      driver: 'cli',
      mechanisms: [{ value: 'cli', label: 'CLI', driver: 'cli' }],
      models: [{
        id: 'claude-opus-4-8',
        provider: 'anthropic',
        label: 'Claude Opus 4.8',
        capabilitiesKnown: true,
        reasoning: true,
        supportedEfforts: ['low', 'high'],
        defaultEffort: 'high',
        configured: true,
      }],
      configured: true,
    },
  ],
};

beforeEach(() => {
  const store: Record<string, string> = {};
  vi.stubGlobal('localStorage', {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = String(value);
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key];
    }),
    clear: vi.fn(() => {
      for (const key of Object.keys(store)) delete store[key];
    }),
  });
  vi.stubGlobal('fetch', vi.fn(async () => ({ ok: true, json: async () => context }) as Response));
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('TodoRunActionButton RuntimeBar', () => {
  it('selects and remembers family, mode, model, and effort before the primary action executes', async () => {
    const onRun = vi.fn();
    render(
      <TodoRunActionButton action="run" onRun={onRun} onAdvanced={vi.fn()} />,
      { wrapper: queryTestWrapper() },
    );

    const primary = screen.getByRole('button', { name: 'Run' });
    await waitFor(() => expect((primary as HTMLButtonElement).disabled).toBe(false));
    fireEvent.click(await screen.findByRole('button', {
      name: 'Run runtime: Codex, Agent, GPT-5.5, effort Medium',
    }));

    let menu = screen.getByRole('menu');
    fireEvent.click(within(menu).getByRole('radio', { name: 'Claude' }));
    menu = screen.getByRole('menu');
    fireEvent.click(within(menu).getByRole('radio', { name: 'CLI' }));
    menu = screen.getByRole('menu');
    fireEvent.click(within(menu).getByRole('button', { name: 'Claude Opus 4.8' }));
    menu = screen.getByRole('menu');
    expect(within(menu).getByRole('slider', { name: 'Reasoning effort' }).getAttribute('aria-valuetext')).toBe('High');

    expect(onRun).not.toHaveBeenCalled();
    await waitFor(() => expect(JSON.parse(localStorage.getItem('gavel.pr-ui.todoRunChoices.v2') ?? '{}')).toMatchObject({
      last: {
        run: {
          driver: 'cli',
          runMode: 'run',
          spec: {
            mode: 'cli',
            model: 'claude-opus-4-8',
            effort: 'high',
            workflow: { commits: [{ on: 'run', gates: 'full' }] },
          },
        },
      },
    }));

    fireEvent.click(primary);
    expect(onRun).toHaveBeenCalledWith(expect.objectContaining({
      driver: 'cli',
      runMode: 'run',
      spec: expect.objectContaining({ mode: 'cli', model: 'claude-opus-4-8', effort: 'high' }),
    }));
  });

  it('disables runtime selection and preserves the Advanced entry point', async () => {
    const onAdvanced = vi.fn();
    const { rerender } = render(
      <TodoRunActionButton action="plan" onRun={vi.fn()} onAdvanced={onAdvanced} />,
      { wrapper: queryTestWrapper() },
    );
    const advanced = await screen.findByRole('button', { name: 'Advanced plan options' });
    await waitFor(() => expect((advanced as HTMLButtonElement).disabled).toBe(false));
    fireEvent.click(advanced);
    expect(onAdvanced).toHaveBeenCalledWith('plan');

    rerender(<TodoRunActionButton action="plan" disabled onRun={vi.fn()} onAdvanced={onAdvanced} />);
    const runtime = screen.getByRole('button', { name: /^Plan runtime:/ });
    expect((runtime.closest('fieldset') as HTMLFieldSetElement).disabled).toBe(true);
    expect((screen.getByRole('button', { name: 'Advanced plan options' }) as HTMLButtonElement).disabled).toBe(true);
  });
});
