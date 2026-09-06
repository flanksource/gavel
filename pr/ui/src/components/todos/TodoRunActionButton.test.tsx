import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { TodoRunOptions } from '../../types';
import { TodoRunActionButton } from './TodoRunActionButton';
import type { RunContext } from './providers';

// TodoRunActionButton renders the primary Run/Plan button, its runtime chip,
// and the advanced-options cog. It reasons only in terms of the fixed
// TodoRunAction it is given (`action`) — never a driver/runMode pair — so
// these tests stay focused on the three things a caller depends on: the
// button dispatches the current options, the runtime bar surfaces a change,
// and the cog opens the advanced dialog for the same action.
//
// Only useTodoRunContext is mocked here (to avoid a network round trip);
// run.tsx's storage/reconciliation helpers run for real against the fixture
// below, because several of them call each other internally — partially
// mocking just one export leaves those internal calls resolving against the
// *real* sibling function, which produced confusing, inconsistent values in
// an earlier version of this test.
const FIXTURE_CONTEXT: RunContext = {
  defaultMode: 'agent',
  defaultProvider: 'anthropic',
  efforts: ['low', 'medium', 'high'],
  tools: [],
  runtimes: [
    { family: 'claude', provider: 'anthropic', catalogPrefix: 'anthropic', modes: [{ mode: 'agent', schema: { type: 'object' } }] },
  ],
  models: [
    { id: 'claude-opus-4-8', provider: 'anthropic', label: 'Claude Opus 4.8', reasoning: true, configured: true },
    { id: 'claude-opus-4-9', provider: 'anthropic', label: 'Claude Opus 4.9', reasoning: true, configured: true },
  ],
  modes: [
    {
      id: 'agent',
      label: 'Claude Agent',
      provider: 'anthropic',
      agent: 'claude',
      defaultModel: 'claude-opus-4-8',
      driver: 'agent',
      mechanisms: [{ value: 'agent', label: 'Agent', driver: 'agent' }],
      models: [
        { id: 'claude-opus-4-8', provider: 'anthropic', label: 'Claude Opus 4.8', reasoning: true, configured: true },
        { id: 'claude-opus-4-9', provider: 'anthropic', label: 'Claude Opus 4.9', reasoning: true, configured: true },
      ],
      configured: true,
    },
  ],
  lifecycle: { steps: [{ name: 'run', label: 'Run', prompt: 'run', readOnly: false }] },
  runtimeProfiles: [{ id: 'review-profile', name: 'Review profile', presets: ['cmux-preset'] }],
};

vi.mock('./run', async importOriginal => ({
  ...(await importOriginal<typeof import('./run')>()),
  useTodoRunContext: () => ({ context: FIXTURE_CONTEXT, loading: false, error: '' }),
}));

vi.mock('@flanksource/clicky-ui/ai', async importOriginal => ({
  ...(await importOriginal<typeof import('@flanksource/clicky-ui/ai')>()),
  RuntimeBar: ({ ariaLabel, onChange, value }: { ariaLabel?: string; onChange: (value: Record<string, unknown>) => void; value: Record<string, unknown> }) => (
    // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for RuntimeBar itself.
    <button type="button" aria-label={ariaLabel} onClick={() => onChange({ ...value, model: 'claude-opus-4-9' })}>runtime</button>
  ),
}));

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
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('TodoRunActionButton', () => {
  it('shows a sparse profile by identity without claiming the step default runtime', () => {
    const onRun = vi.fn();
    const options = { step: 'run', runtimeProfile: 'review-profile', spec: {} };
    render(<TodoRunActionButton dir="/repo" action="run" options={options} onRun={onRun} onAdvanced={vi.fn()} />);
    expect(screen.getByText('Profile: Review profile')).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Run runtime' })).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Run' }));
    expect(onRun).toHaveBeenCalledWith(options);
  });

  it('dispatches a sparse request when the operator has not changed the runtime', () => {
    const onRun = vi.fn();
    render(<TodoRunActionButton dir="/repo" action="run" onRun={onRun} onAdvanced={vi.fn()} />);

    fireEvent.click(screen.getByRole('button', { name: 'Run' }));

    expect(onRun).toHaveBeenCalledWith(expect.objectContaining({
      step: 'run',
      spec: {},
    }) as TodoRunOptions);
  });

  it('opens the advanced dialog for the same action', () => {
    const onAdvanced = vi.fn();
    render(<TodoRunActionButton dir="/repo" action="plan" onRun={vi.fn()} onAdvanced={onAdvanced} />);

    fireEvent.click(screen.getByRole('button', { name: 'Advanced plan options' }));

    expect(onAdvanced).toHaveBeenCalledWith('plan');
  });

  it('surfaces a runtime change through onOptionsChange without running', () => {
    const onRun = vi.fn();
    const onOptionsChange = vi.fn();
    render(<TodoRunActionButton dir="/repo" action="run" onRun={onRun} onAdvanced={vi.fn()} onOptionsChange={onOptionsChange} />);

    fireEvent.click(screen.getByRole('button', { name: 'Run runtime' }));

    expect(onRun).not.toHaveBeenCalled();
    expect(onOptionsChange).toHaveBeenCalledWith(expect.objectContaining({
      spec: expect.objectContaining({ model: 'claude-opus-4-9' }),
    }) as TodoRunOptions);
  });
});
