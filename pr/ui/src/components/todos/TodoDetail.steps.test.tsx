import { Button } from '@flanksource/clicky-ui/components';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { TodoItem, TodoRunOptions } from '../../types';
import type { RunContext } from './providers';
import { TodoDetail } from './TodoDetail';
import { readRunChoiceState } from './runChoiceStorage';

const execution = vi.hoisted(() => ({ run: vi.fn(), verify: vi.fn(), reset: vi.fn(), step: 'verify' }));
const spec = { mode: 'agent', model: 'example-model', effort: 'medium' as const };
const context: RunContext = {
  modes: [{ id: 'agent', label: 'Agent', provider: 'openai', agent: 'codex', defaultModel: spec.model, driver: 'agent', mechanisms: [], models: [{ id: spec.model, label: 'Example', provider: 'openai', reasoning: true }] }],
  defaultMode: 'agent', defaultProvider: 'openai', efforts: ['medium'], tools: [], models: [], runtimes: [],
  lifecycle: { steps: ['plan', 'run', 'verify', 'review-security'].map(name => ({ name, label: name, prompt: name, readOnly: false })) },
};

vi.mock('./run', async importOriginal => ({
  ...(await importOriginal<object>()),
  useTodoRunContext: () => ({ context, loading: false, error: '' }),
  useTodoRun: () => ({ run: execution.run, reset: execution.reset, runBusy: false, runError: '', runMessage: '' }),
}));
vi.mock('./todoMutations', async importOriginal => ({
  ...(await importOriginal<object>()),
  useUpdateTodoMutation: () => ({ isPending: false }),
  useDeleteTodoMutation: () => ({ isPending: false }),
  useTransferTodoMutation: () => ({ isPending: false }),
  useGithubPushTodoMutation: () => ({ isPending: false }),
  useTodoSessionStop: () => ({ isPending: false }),
  useTodoVerificationRun: () => ({ isPending: false, mutateAsync: execution.verify }),
}));
vi.mock('./TodoSessionTimer', () => ({ useSessionStats: () => ({ stats: null }) }));
vi.mock('./TodoSessionDetail', () => ({ useTodoSessionDetail: () => ({ detail: undefined, error: '' }) }));
vi.mock('./tagQueries', () => ({ useTodoTagIndex: () => ({}), useTodoTagCounts: () => ({}) }));
vi.mock('./TodoTag', () => ({ TodoTagField: () => null, TodoTagRow: () => null }));
vi.mock('./TodoTimeline', () => ({ TodoTimeline: () => null }));
vi.mock('./TodoCommits', () => ({ TodoCommits: () => null }));
vi.mock('./TodoSession', () => ({ TodoSession: () => null }));
vi.mock('./TodoPlan', () => ({ TodoPlan: () => null }));
vi.mock('./TodoVerification', () => ({ TodoVerification: () => null }));
vi.mock('./TodoCompose', () => ({ TodoTitleEditor: () => null, TodoBodyEditor: () => null, TodoCommentBox: () => null }));
vi.mock('./planActions', () => ({ TodoReviewBanner: () => null }));
vi.mock('./TodoPhaseButton', () => ({
  TodoPhaseTicks: () => null,
  TodoPhaseButton: ({ onAdvanced, onRunStep }: { onAdvanced: (step: string) => void; onRunStep: (step: string) => void }) => (
    <div>{context.lifecycle.steps.map(({ name }) => <div key={name}>
      <Button onClick={() => onAdvanced(name)}>Advanced {name}</Button>
      <Button onClick={() => onRunStep(name)}>Start {name}</Button>
    </div>)}</div>
  ),
}));
vi.mock('./TodoRunAdvancedDialog', () => ({
  TodoRunAdvancedDialog: ({ open, initialMode, onRun }: { open: boolean; initialMode: string; onRun: (options: TodoRunOptions) => void }) => open && (
    <div data-testid="advanced-step">{initialMode}<Button onClick={() => onRun({ step: execution.step, spec })}>Dispatch selected step</Button></div>
  ),
}));

const todo: TodoItem = {
  ref: 'example-todo', title: 'Review lifecycle routing', status: 'pending', priority: 'medium',
  lifecycle: { steps: context.lifecycle.steps.map(({ name, label }) => ({ name, label, applicable: true, suggested: name === 'verify', done: false, lastRun: null })), next: 'verify', reason: 'Verify the result' },
};

beforeEach(() => {
  execution.run.mockReset();
  execution.verify.mockReset();
  execution.step = 'verify';
  const values = new Map<string, string>();
  vi.stubGlobal('localStorage', { getItem: (key: string) => values.get(key) ?? null, setItem: (key: string, value: string) => values.set(key, value) });
});
afterEach(() => vi.unstubAllGlobals());

async function openDetail() {
  render(<TodoDetail todo={todo} loading={false} dir="/repo" onChanged={() => {}} onDeleted={() => {}} />);
  await act(async () => {});
}

describe('TodoDetail named step dispatch', () => {
  it.each(['plan', 'verify', 'review-security'])('opens Advanced %s on that exact step', async step => {
    await openDetail();
    fireEvent.click(screen.getByRole('button', { name: `Advanced ${step}` }));
    expect(screen.getByTestId('advanced-step').textContent).toBe(`${step}Dispatch selected step`);
  });

  it('dispatches verification through the verification mutation', async () => {
    await openDetail();
    fireEvent.click(screen.getByRole('button', { name: 'Advanced verify' }));
    fireEvent.click(screen.getByRole('button', { name: 'Dispatch selected step' }));
    await waitFor(() => expect(execution.verify).toHaveBeenCalledWith({ ref: todo.ref, spec: expect.objectContaining(spec) }));
    expect(execution.run).not.toHaveBeenCalled();
  });

  it('remembers and dispatches the selected step after changing the dialog from run to plan', async () => {
    execution.step = 'plan';
    await openDetail();
    fireEvent.click(screen.getByRole('button', { name: 'Advanced run' }));
    fireEvent.click(screen.getByRole('button', { name: 'Dispatch selected step' }));
    await waitFor(() => expect(execution.run).toHaveBeenCalledWith(todo.ref, { step: 'plan', spec }));
    expect(readRunChoiceState().last.plan).toEqual({ step: 'plan', spec });
    expect(readRunChoiceState().last.run).toBeUndefined();
  });

  // A step the operator has never configured dispatches by name with an empty
  // spec: the catalog's default mode/model are display defaults, and copying
  // them into the request would outrank the runtime the step's own prompt
  // frontmatter pins (see run.test.ts, "displays a prompt runtime while leaving
  // the request sparse").
  it('runs a custom lifecycle step by name without seeding a runtime it was never given', async () => {
    await openDetail();
    fireEvent.click(screen.getByRole('button', { name: 'Start review-security' }));
    await waitFor(() => expect(execution.run).toHaveBeenCalledWith(todo.ref, { step: 'review-security', spec: {} }));
  });
});
