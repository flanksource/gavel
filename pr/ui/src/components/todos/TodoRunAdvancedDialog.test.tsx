import type { ComponentProps, ReactNode } from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { TodoRunOptions } from '../../types';
import { TodoRunAdvancedDialog } from './TodoRunAdvancedDialog';
import { requestStepFor } from './runChoiceStorage';
import type { RunContext } from './providers';

// The run dialog's step picker now lists whatever the project's lifecycle
// declares — including a custom, non-built-in step — rather than a hardcoded
// run/plan/triage SegmentedControl. `useTodoRunContext`/`useTodoRunPreview`
// and the localStorage-backed "last used" helpers are stubbed so these tests
// exercise only the picker's own behaviour: which steps it lists, which one it
// preselects, and that the chosen step (not a driver/runMode guess) is what
// reaches the wire.
const STEPS = [
  { name: 'plan', label: 'Plan', prompt: 'plan', readOnly: false },
  { name: 'run', label: 'Run', prompt: 'run', readOnly: false },
  { name: 'verify', label: 'Verify', prompt: 'verify', readOnly: true },
  { name: 'security-review', label: 'Security review', prompt: 'security-review', readOnly: false },
];

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
      models: [{ id: 'claude-opus-4-8', provider: 'anthropic', label: 'Claude Opus 4.8', reasoning: true, configured: true }],
      configured: true,
    },
  ],
  promptDefaults: { 'security-review': { mode: 'agent', model: 'claude-opus-4-8' } },
  lifecycle: { steps: STEPS },
};

vi.mock('./run', async importOriginal => ({
  ...(await importOriginal<typeof import('./run')>()),
  useTodoRunContext: () => ({ context: FIXTURE_CONTEXT, loading: false, error: '' }),
  useTodoRunPreview: () => ({ isPending: false, mutate: () => {} }),
  loadLastTodoRunOptions: () => ({ spec: { mode: 'agent', model: 'claude-opus-4-8', effort: 'medium' } }),
  loadRecentAdvancedTodoRunOptions: () => [],
  reconcileTodoRunOptions: (_action: string, options: TodoRunOptions) => options,
}));

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, variant: _variant, size: _size, loading: _loading, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement> & { variant?: string; size?: string; loading?: boolean }) => (
    // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky Button itself.
    <button type="button" {...props}>{children}</button>
  ),
  Field: ({ children }: { children?: ReactNode }) => <div>{children}</div>,
  Modal: ({ children, footer, open }: { children?: ReactNode; footer?: ReactNode; open?: boolean }) => (
    open === false ? null : <div><div>{children}</div><div data-testid="dialog-footer">{footer}</div></div>
  ),
  SegmentedControl: <T extends string>({
    options,
    value,
    onChange,
    'aria-label': ariaLabel,
  }: {
    options: Array<{ id: T; label: string; disabled?: boolean }>;
    value: T;
    onChange: (value: T) => void;
    'aria-label'?: string;
  }) => (
    <div role="group" aria-label={ariaLabel}>
      {options.map(option => (
        // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for SegmentedControl itself.
        <button
          key={option.id}
          type="button"
          disabled={option.disabled}
          aria-pressed={option.id === value}
          onClick={() => onChange(option.id)}
        >
          {option.label}
        </button>
      ))}
    </div>
  ),
  Tabs: <T extends string>({ tabs, value, onChange }: { tabs: Array<{ id: T; label: string }>; value: T; onChange: (value: T) => void }) => (
    <div>
      {tabs.map(tab => (
        // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for Tabs itself.
        <button key={tab.id} type="button" aria-pressed={tab.id === value} onClick={() => onChange(tab.id)}>{tab.label}</button>
      ))}
    </div>
  ),
}));

vi.mock('@flanksource/clicky-ui/data', () => ({
  CodeBlock: () => null,
}));

vi.mock('@flanksource/clicky-ui/ai', async importOriginal => ({
  ...(await importOriginal<typeof import('@flanksource/clicky-ui/ai')>()),
  PromptRunEditor: () => null,
  promptRuntimeValueToPayload: (value: Record<string, unknown>) => ({ spec: value }),
}));

afterEach(() => {
  vi.restoreAllMocks();
});

function setup(props: Partial<ComponentProps<typeof TodoRunAdvancedDialog>> = {}) {
  const onRun = vi.fn();
  const onClose = vi.fn();
  render(
    <TodoRunAdvancedDialog
      open
      onClose={onClose}
      onRun={onRun}
      dir="/workspace"
      refID="todo-1"
      {...props}
    />,
  );
  return { onRun, onClose };
}

// The picker's currently-selected option always shares its label with the
// modal footer's submit button (see submitLabel in the component), so picker
// assertions scope to the "Prompt" group and submit-button assertions scope
// to the footer, rather than querying the whole document by button name.
function picker() {
  return within(screen.getByRole('group', { name: 'Prompt' }));
}

function footer() {
  return within(screen.getByTestId('dialog-footer'));
}

describe('TodoRunAdvancedDialog step picker', () => {
  it('lists every lifecycle step the run context declares, including a project-defined custom one', () => {
    setup({ initialMode: 'run' });
    for (const step of STEPS) {
      expect(picker().getByRole('button', { name: step.label })).toBeTruthy();
    }
  });

  it('disables a read-only step so the picker cannot dispatch it', () => {
    setup({ initialMode: 'run' });
    expect((picker().getByRole('button', { name: 'Verify' }) as HTMLButtonElement).disabled).toBe(true);
    expect((picker().getByRole('button', { name: 'Run' }) as HTMLButtonElement).disabled).toBe(false);
  });

  it('preselects the todo-suggested next step when the caller forces none', () => {
    setup({ initialMode: undefined, nextStep: 'security-review' });
    expect(picker().getByRole('button', { name: 'Security review' }).getAttribute('aria-pressed')).toBe('true');
    expect(picker().getByRole('button', { name: 'Run' }).getAttribute('aria-pressed')).toBe('false');
  });

  it('honors an explicit initialMode over the suggested next step', () => {
    setup({ initialMode: 'plan', nextStep: 'security-review' });
    expect(picker().getByRole('button', { name: 'Plan' }).getAttribute('aria-pressed')).toBe('true');
    expect(picker().getByRole('button', { name: 'Security review' }).getAttribute('aria-pressed')).toBe('false');
  });

  it('dispatches the exact custom step the operator picked, not a driver/runMode guess', () => {
    const { onRun } = setup({ initialMode: 'run' });

    fireEvent.click(picker().getByRole('button', { name: 'Security review' }));
    fireEvent.click(footer().getByRole('button', { name: 'Security review' }));

    expect(onRun).toHaveBeenCalledTimes(1);
    const options = onRun.mock.calls[0]![0] as TodoRunOptions;
    expect(requestStepFor(options)).toBe('security-review');
    expect(options).not.toHaveProperty('driver');
    expect(options).not.toHaveProperty('runMode');
  });

  it('dispatches a built-in step (plan) with its bookkeeping fields for storage, but no driver', () => {
    const { onRun } = setup({ initialMode: 'run' });

    fireEvent.click(picker().getByRole('button', { name: 'Plan' }));
    fireEvent.click(footer().getByRole('button', { name: 'Plan' }));

    expect(onRun).toHaveBeenCalledTimes(1);
    const options = onRun.mock.calls[0]![0] as TodoRunOptions;
    expect(requestStepFor(options)).toBe('plan');
    expect(options).not.toHaveProperty('driver');
  });
});
