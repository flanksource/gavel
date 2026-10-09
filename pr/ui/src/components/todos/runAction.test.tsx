import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { TodoRunActionButton } from './TodoRunActionButton';
import type { RunContext } from './providers';
import { queryTestWrapper } from './queryTestWrapper';

// RuntimeBar is a clicky-ui primitive (mode-first identity + overflow fields);
// its own interaction surface is clicky-ui's test responsibility, not
// gavel's. Mocking it here — the same pattern TodoRunActionButton.test.tsx
// uses — lets this suite verify gavel's own plumbing (remembering a runtime
// choice, keeping the primary action sparse until it fires, dispatching the
// final spec) without re-deriving clicky-ui's current DOM/ARIA structure.
vi.mock('@flanksource/clicky-ui/ai', async importOriginal => ({
  ...(await importOriginal<typeof import('@flanksource/clicky-ui/ai')>()),
  RuntimeBar: ({ value, onChange, ariaLabel }: {
    value: Record<string, unknown>;
    onChange: (value: Record<string, unknown>) => void;
    ariaLabel?: string;
  }) => (
    // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for RuntimeBar itself.
    <button
      type="button"
      aria-label={ariaLabel}
      onClick={() => onChange({ ...value, mode: 'cli', model: 'claude-opus-4-8', effort: 'high' })}
    >
      {ariaLabel}
    </button>
  ),
}));

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
      <TodoRunActionButton dir="/repo" action="run" onRun={onRun} onAdvanced={vi.fn()} />,
      { wrapper: queryTestWrapper() },
    );

    const primary = screen.getByRole('button', { name: 'Run' });
    await waitFor(() => expect((primary as HTMLButtonElement).disabled).toBe(false));

    fireEvent.click(screen.getByRole('button', { name: 'Run runtime' }));

    expect(onRun).not.toHaveBeenCalled();
    await waitFor(() => expect(JSON.parse(localStorage.getItem('gavel.pr-ui.todoRunChoices.v3') ?? '{}')).toMatchObject({
      last: {
        run: {
          step: 'run',
          spec: {
            mode: 'cli',
            model: 'claude-opus-4-8',
            effort: 'high',
          },
        },
      },
    }));

    fireEvent.click(primary);
    expect(onRun).toHaveBeenCalledWith(expect.objectContaining({
      step: 'run',
      spec: expect.objectContaining({ mode: 'cli', model: 'claude-opus-4-8', effort: 'high' }),
    }));
  });

  it('disables runtime selection and preserves the Advanced entry point', async () => {
    const onAdvanced = vi.fn();
    const { rerender } = render(
      <TodoRunActionButton dir="/repo" action="plan" onRun={vi.fn()} onAdvanced={onAdvanced} />,
      { wrapper: queryTestWrapper() },
    );
    const advanced = await screen.findByRole('button', { name: 'Advanced plan options' });
    await waitFor(() => expect((advanced as HTMLButtonElement).disabled).toBe(false));
    fireEvent.click(advanced);
    expect(onAdvanced).toHaveBeenCalledWith('plan');

    rerender(<TodoRunActionButton dir="/repo" action="plan" disabled onRun={vi.fn()} onAdvanced={onAdvanced} />);
    const runtime = screen.getByRole('button', { name: 'Plan runtime' });
    expect((runtime.closest('fieldset') as HTMLFieldSetElement).disabled).toBe(true);
    expect((screen.getByRole('button', { name: 'Advanced plan options' }) as HTMLButtonElement).disabled).toBe(true);
  });
});
