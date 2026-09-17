import type { ComponentProps, ReactNode } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import type { AIPromptRunValue } from '@flanksource/clicky-ui/ai';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { TodoRunOptions } from '../../types';
import { TodoRunAdvancedDialog } from './TodoRunAdvancedDialog';
import { requestStepFor } from './runChoiceStorage';
import { rememberTodoRunOptions } from './run';
import type { RunContext } from './providers';
const preview = vi.hoisted(() => vi.fn());
const recentOptions = vi.hoisted(() => vi.fn((): TodoRunOptions[] => []));
const lastUsedOptions = vi.hoisted(() => vi.fn((): TodoRunOptions | undefined => undefined));
// contextVersion lets a test force a new `context` object identity on a
// later render — simulating a background run-context query refetch — while
// FIXTURE_CONTEXT's actual data stays put, without changing any props.
const contextVersion = vi.hoisted(() => ({ current: 0 }));
// One context object per version, like react-query's structurally shared data:
// a fresh identity on every render would re-run the dialog's preview effect on
// every render and never settle.
const contextsByVersion = vi.hoisted(() => new Map<number, unknown>());

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
  runtimePresets: [{ id: 'review-preset', name: 'Review preset', scope: 'context', spec: { model: 'preset-model' } }],
  lifecycle: { steps: STEPS },
};

vi.mock('./run', async importOriginal => ({
  ...(await importOriginal<typeof import('./run')>()),
  useTodoRunContext: () => {
    const version = contextVersion.current;
    if (!contextsByVersion.has(version)) contextsByVersion.set(version, { ...FIXTURE_CONTEXT, __v: version });
    return { context: contextsByVersion.get(version) as RunContext, loading: false, error: '' };
  },
  useTodoRunPreview: () => ({ isPending: false, mutate: preview }),
  lastTodoRunOptions: lastUsedOptions,
  loadRecentAdvancedTodoRunOptions: recentOptions,
  // A reconcile that visibly rewrites effort: the dialog must never route a
  // restore or a submit through it, so any 'reconciled-effort' in a request
  // body is a regression.
  reconcileTodoRunOptions: (_action: string, options: TodoRunOptions) => ({ ...options, spec: { ...options.spec, effort: 'reconciled-effort' } }),
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
  PromptRunEditor: ({ value, onChange, presets, children }: { value: AIPromptRunValue & { presets?: string[] }; onChange: (value: AIPromptRunValue & { presets?: string[] }) => void; presets?: Array<{ id: string; name: string }>; children?: ReactNode }) => (
    <div>
      <Button onClick={() => onChange({ ...value, spec: { ...value.spec, model: 'operator-model' } })}>Change model</Button>
      {presets?.map(preset => <Button key={preset.id} onClick={() => onChange({ ...value, presets: [preset.id] })}>{preset.name}</Button>)}
      {children}
    </div>
  ),
  promptRuntimeValueToPayload: (value: Record<string, unknown>) => ({ spec: value }),
}));

beforeEach(() => {
  preview.mockReset();
  recentOptions.mockReturnValue([]);
  lastUsedOptions.mockReturnValue(undefined);
  contextVersion.current = 0;
  contextsByVersion.clear();
  localStorage.clear();
});

afterEach(() => {
  vi.restoreAllMocks();
});

function setup(props: Partial<ComponentProps<typeof TodoRunAdvancedDialog>> = {}) {
  const onRun = vi.fn();
  const onClose = vi.fn();
  const utils = render(
    <TodoRunAdvancedDialog
      open
      onClose={onClose}
      onRun={onRun}
      dir="/workspace"
      refID="todo-1"
      {...props}
    />,
  );
  return { onRun, onClose, ...utils };
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

  it('dispatches a built-in step by name without legacy bookkeeping', () => {
    const { onRun } = setup({ initialMode: 'run' });

    fireEvent.click(picker().getByRole('button', { name: 'Plan' }));
    fireEvent.click(footer().getByRole('button', { name: 'Plan' }));

    expect(onRun).toHaveBeenCalledTimes(1);
    const options = onRun.mock.calls[0]![0] as TodoRunOptions;
    expect(requestStepFor(options)).toBe('plan');
    expect(options).not.toHaveProperty('driver');
    expect(options).not.toHaveProperty('runMode');
  });
});

describe('TodoRunAdvancedDialog runtime presets', () => {
  it('previews and submits a sparse step without promoting catalog metadata into a model override', () => {
    const { onRun } = setup({ initialMode: 'plan' });
    fireEvent.click(footer().getByRole('button', { name: 'Plan' }));
    expect(JSON.parse(JSON.stringify(preview.mock.lastCall?.[0].body))).toEqual({ ref: 'todo-1', step: 'plan', spec: {} });
    expect(JSON.parse(JSON.stringify(onRun.mock.lastCall?.[0]))).toEqual({ step: 'plan', spec: {} });
  });

  it('shows preflight warnings, replaces them on preview changes and clears them on success or error', () => {
    const consoleError = vi.spyOn(console, 'error');
    preview.mockImplementation(({ body }, callbacks) => callbacks.onSuccess({
      prompt: '', specYaml: '', count: 1,
      warnings: body.presets ? [] : ['permissions.plugins is unsupported', 'permissions.plugins is unsupported'],
    }));
    setup({ initialMode: 'plan' });
    expect(screen.getAllByText('permissions.plugins is unsupported')).toHaveLength(2);
    expect(consoleError).not.toHaveBeenCalled();
    expect((footer().getByRole('button', { name: 'Plan' }) as HTMLButtonElement).disabled).toBe(false);
    fireEvent.click(screen.getByRole('button', { name: 'Review preset' }));
    expect(screen.queryByText('permissions.plugins is unsupported')).toBeNull();
    preview.mockImplementation((_request, callbacks) => callbacks.onSuccess({ prompt: '', specYaml: '', count: 1, warnings: ['permissions.skills is unsupported'] }));
    fireEvent.click(screen.getByRole('button', { name: 'Change model' }));
    expect(screen.getByText('permissions.skills is unsupported')).toBeTruthy();
    preview.mockImplementation((_request, callbacks) => callbacks.onError(new Error('judge prompt is missing')));
    fireEvent.click(picker().getByRole('button', { name: 'Run' }));
    expect(screen.queryByText('permissions.skills is unsupported')).toBeNull();
    expect(screen.getByText('judge prompt is missing')).toBeTruthy();
  });

  it('restores the saved presets and resume choice from recent advanced options', () => {
    recentOptions.mockReturnValue([{ step: 'plan', presets: ['saved-preset'], spec: { mode: 'cmux' }, resume: true }]);
    const { onRun } = setup({ initialMode: 'plan' });
    fireEvent.click(screen.getByRole('button', { name: 'Review preset' }));
    fireEvent.click(screen.getByRole('button', { name: /^1\./ }));
    fireEvent.click(footer().getByRole('button', { name: 'Plan' }));
    expect(JSON.parse(JSON.stringify(onRun.mock.lastCall?.[0]))).toEqual({ step: 'plan', presets: ['saved-preset'], spec: { mode: 'cmux' }, resume: true });
  });

  it('offers resume for cmux presets resolved by preview without sending inherited mode', () => {
    preview.mockImplementation(({ body }, callbacks) => callbacks.onSuccess({
      prompt: '', specYaml: '', model: 'resolved-model', runtimeMode: body.presets?.includes('review-preset') ? 'cmux' : 'agent', count: 1,
    }));
    const { onRun } = setup({ initialMode: 'plan' });
    fireEvent.click(screen.getByRole('button', { name: 'Review preset' }));
    expect(screen.getByText('resolved-model · cmux')).toBeTruthy();
    fireEvent.click(screen.getByRole('checkbox', { name: 'Resume session' }));
    fireEvent.click(footer().getByRole('button', { name: 'Plan' }));
    expect(JSON.parse(JSON.stringify(onRun.mock.lastCall?.[0]))).toEqual({ step: 'plan', presets: ['review-preset'], spec: {}, resume: true });
  });

  it.each(['plan', 'verify', 'security-review'])('previews and submits %s with presets and no seeded model override', step => {
    const { onRun } = setup({ initialMode: step });
    fireEvent.click(screen.getByRole('button', { name: 'Review preset' }));
    fireEvent.click(footer().getByRole('button', { name: STEPS.find(entry => entry.name === step)!.label }));
    expect(JSON.parse(JSON.stringify(preview.mock.lastCall?.[0].body))).toEqual({ ref: 'todo-1', step, presets: ['review-preset'], spec: {} });
    expect(JSON.parse(JSON.stringify(onRun.mock.lastCall?.[0]))).toEqual({ step, presets: ['review-preset'], spec: {} });
  });

  it.each(['before', 'after'])('retains a model edited %s selecting presets', when => {
    const { onRun } = setup({ initialMode: 'plan' });
    if (when === 'before') fireEvent.click(screen.getByRole('button', { name: 'Change model' }));
    fireEvent.click(screen.getByRole('button', { name: 'Review preset' }));
    if (when === 'after') fireEvent.click(screen.getByRole('button', { name: 'Change model' }));
    fireEvent.click(footer().getByRole('button', { name: 'Plan' }));
    expect(JSON.parse(JSON.stringify(onRun.mock.lastCall?.[0]))).toEqual({ step: 'plan', presets: ['review-preset'], spec: { model: 'operator-model' } });
  });
});

describe('TodoRunAdvancedDialog last used', () => {
  it('opens with an empty spec even when a last-used entry exists', () => {
    // Real storage, read by the quick-run loader the dialog used to seed from.
    rememberTodoRunOptions('plan', { step: 'plan', presets: ['review-preset'], spec: { mode: 'cmux', model: 'claude-opus-4-8', effort: 'high' } });
    lastUsedOptions.mockReturnValue({ step: 'plan', presets: ['review-preset'], spec: { mode: 'cmux', model: 'claude-opus-4-8' }, resume: true });
    const { onRun } = setup({ initialMode: 'plan' });
    fireEvent.click(footer().getByRole('button', { name: 'Plan' }));
    expect(JSON.parse(JSON.stringify(preview.mock.lastCall?.[0].body))).toEqual({ ref: 'todo-1', step: 'plan', spec: {} });
    expect(JSON.parse(JSON.stringify(onRun.mock.lastCall?.[0]))).toEqual({ step: 'plan', spec: {} });
  });

  it('restores the remembered spec, presets and resume from "Last used"', () => {
    lastUsedOptions.mockReturnValue({ step: 'plan', presets: ['review-preset'], spec: { mode: 'cmux', model: 'claude-opus-4-8' }, resume: true });
    const { onRun } = setup({ initialMode: 'plan' });
    fireEvent.click(screen.getByRole('button', { name: /^Last used/ }));
    fireEvent.click(footer().getByRole('button', { name: 'Plan' }));
    expect(JSON.parse(JSON.stringify(onRun.mock.lastCall?.[0]))).toEqual({
      step: 'plan', presets: ['review-preset'], spec: { mode: 'cmux', model: 'claude-opus-4-8' }, resume: true,
    });
  });

  it('does not offer "Last used" when nothing has been run for the step yet', () => {
    setup({ initialMode: 'plan' });
    expect(screen.queryByRole('button', { name: /^Last used/ })).toBeNull();
  });

  it('hides a recent-advanced entry equal to the last-used entry', () => {
    const duplicate: TodoRunOptions = { step: 'plan', spec: { model: 'claude-opus-4-8' } };
    lastUsedOptions.mockReturnValue(duplicate);
    recentOptions.mockReturnValue([duplicate, { step: 'plan', spec: { model: 'other-model' } }]);
    setup({ initialMode: 'plan' });
    expect(screen.getAllByRole('button', { name: /^\d+\./ })).toHaveLength(1);
    expect(screen.getByRole('button', { name: /^1\./ }).textContent).toContain('other-model');
  });

  it('offers the verification prompt history as "Last used" for the verify step', () => {
    // Written directly: remembering goes through this file's effort-rewriting reconcile mock.
    localStorage.setItem('gavel.pr-ui.promptRunChoices.v2', JSON.stringify({ verification: { last: { step: 'verify', spec: { model: 'claude-opus-4-8', effort: 'high' } } } }));
    const { onRun } = setup({ initialMode: 'verify' });
    fireEvent.click(screen.getByRole('button', { name: /^Last used/ }));
    fireEvent.click(footer().getByRole('button', { name: 'Verify' }));
    expect(JSON.parse(JSON.stringify(onRun.mock.lastCall?.[0]))).toEqual({
      step: 'verify', spec: { model: 'claude-opus-4-8', effort: 'high' },
    });
  });

  it('keeps an explicit edit when the run context refetches', () => {
    const { onRun, rerender } = setup({ initialMode: 'plan' });
    fireEvent.click(screen.getByRole('button', { name: 'Change model' }));
    contextVersion.current += 1;
    rerender(
      <TodoRunAdvancedDialog open onClose={() => {}} onRun={onRun} dir="/workspace" refID="todo-1" initialMode="plan" />,
    );
    fireEvent.click(footer().getByRole('button', { name: 'Plan' }));
    expect(JSON.parse(JSON.stringify(onRun.mock.lastCall?.[0]))).toEqual({ step: 'plan', spec: { model: 'operator-model' } });
  });
});
