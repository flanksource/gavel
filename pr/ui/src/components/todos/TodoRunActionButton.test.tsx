import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { TodoRunOptions } from '../../types';
import { TodoRunActionButton, TodoRunRuntimeBar } from './TodoRunActionButton';
import type { RunContext } from './providers';
import { setAuthMode } from '../../query';

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
  runtimePresets: [{ id: 'review-preset', name: 'Review preset', scope: 'context', spec: {} }],
};

vi.mock('./run', async importOriginal => ({
  ...(await importOriginal<typeof import('./run')>()),
  useTodoRunContext: () => ({ context: FIXTURE_CONTEXT, loading: false, error: '' }),
}));

const { runtimeBarProps } = vi.hoisted(() => ({ runtimeBarProps: { current: null as null | Record<string, unknown> } }));

vi.mock('@flanksource/clicky-ui/ai', async importOriginal => ({
  ...(await importOriginal<typeof import('@flanksource/clicky-ui/ai')>()),
  RuntimeBar: (props: { ariaLabel?: string; onChange: (value: Record<string, unknown>) => void; value: Record<string, unknown> }) => {
    runtimeBarProps.current = props;
    const { ariaLabel, onChange, value } = props;
    return (
      // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for RuntimeBar itself.
      <button type="button" aria-label={ariaLabel} onClick={() => onChange({ ...value, model: 'claude-opus-4-9' })}>runtime</button>
    );
  },
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
  it('shows a missing preset by identity without claiming the step default runtime', () => {
    const onRun = vi.fn();
    const options = { step: 'run', presets: ['missing-preset'], spec: {} };
    render(<TodoRunActionButton dir="/repo" action="run" options={options} onRun={onRun} onAdvanced={vi.fn()} />);
    expect(screen.getByText('Preset: missing-preset')).toBeTruthy();
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

  it('offers a hosted org no runtime or advanced options: its runs use the repository\'s runtime', () => {
    setAuthMode('clerk', vi.fn());
    try {
      const onRun = vi.fn();
      render(<TodoRunActionButton dir="repo:1" action="run" onRun={onRun} onAdvanced={vi.fn()} />);
      expect(screen.queryByRole('button', { name: 'Run runtime' })).toBeNull();
      expect(screen.queryByRole('button', { name: 'Advanced run options' })).toBeNull();
      fireEvent.click(screen.getByRole('button', { name: 'Run' }));
      expect(onRun).toHaveBeenCalledOnce();
    } finally {
      setAuthMode('none', vi.fn());
    }
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

describe('TodoRunRuntimeBar', () => {
  it('offers the full runtime options: timeout, cost, permission mode, worktree source and commits', () => {
    render(<TodoRunRuntimeBar action="run" context={FIXTURE_CONTEXT} options={{ step: 'run' }} onChange={vi.fn()} />);

    const props = runtimeBarProps.current as { showTimeout?: boolean; showCost?: boolean; actions?: { fields?: { id: string }[] } };
    expect({
      showTimeout: props.showTimeout,
      showCost: props.showCost,
      fields: props.actions?.fields?.map(field => field.id),
    }).toEqual({
      showTimeout: true,
      showCost: true,
      fields: expect.arrayContaining(['setup.checkout.worktree.mode', 'workflow.commits']),
    });
  });

  it('applies a commit cadence picked in the options menu to the run options', () => {
    const onChange = vi.fn();
    render(<TodoRunRuntimeBar action="run" context={FIXTURE_CONTEXT} options={{ step: 'run' }} onChange={onChange} />);

    type Item = { label: { props: { text?: string } }; onSelect: () => void };
    const fields = (runtimeBarProps.current as { actions: { fields: { id: string; items: Item[] }[] } }).actions.fields;
    const everyTurn = fields.find(field => field.id === 'workflow.commits')!.items.find(item => item.label.props.text === 'Every turn')!;
    everyTurn.onSelect();

    expect(onChange.mock.calls.at(-1)?.[0]?.spec?.workflow).toEqual({ commits: [{ on: 'turn' }] });
  });
});

describe('TodoRunActionButton branch choice', () => {
  const branchChoice = { branch: 'shell/1', unresolved: 2 };
  const branchSelect = () => screen.getByRole('combobox', { name: 'Branch for this run' }) as HTMLSelectElement;

  it('continues on the branch by default and says how many comments are unresolved', () => {
    const onRun = vi.fn();
    render(<TodoRunActionButton dir="/repo" action="run" branchChoice={branchChoice} onRun={onRun} onAdvanced={vi.fn()} />);

    expect(Array.from(branchSelect().options).map(option => option.textContent)).toEqual(['Continue on shell/1 · 2 unresolved comments', 'New branch']);
    fireEvent.click(screen.getByRole('button', { name: 'Run' }));

    expect(onRun).toHaveBeenCalledWith(expect.objectContaining({ step: 'run', reuseBranch: true }) as TodoRunOptions);
  });

  it('dispatches reuseBranch false after "New branch" is chosen, without remembering the choice', () => {
    const onRun = vi.fn();
    render(<TodoRunActionButton dir="/repo" action="run" branchChoice={branchChoice} onRun={onRun} onAdvanced={vi.fn()} />);

    fireEvent.change(branchSelect(), { target: { value: 'new' } });
    fireEvent.click(screen.getByRole('button', { name: 'Run' }));

    expect(onRun).toHaveBeenCalledWith(expect.objectContaining({ reuseBranch: false }) as TodoRunOptions);
    expect(vi.mocked(localStorage.setItem).mock.calls.length).toBeGreaterThan(0);
    expect(String(vi.mocked(localStorage.setItem).mock.lastCall?.[1])).not.toContain('reuseBranch');
  });

  it('offers no choice and sends no reuseBranch without a reusable branch', () => {
    const onRun = vi.fn();
    render(<TodoRunActionButton dir="/repo" action="run" branchChoice={null} onRun={onRun} onAdvanced={vi.fn()} />);

    expect(screen.queryByRole('combobox', { name: 'Branch for this run' })).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Run' }));

    expect(onRun.mock.lastCall?.[0]).not.toHaveProperty('reuseBranch');
  });

  it('offers the choice on Run only, never on Plan', () => {
    render(<TodoRunActionButton dir="/repo" action="plan" branchChoice={branchChoice} onRun={vi.fn()} onAdvanced={vi.fn()} />);

    expect(screen.queryByRole('combobox', { name: 'Branch for this run' })).toBeNull();
  });
});
