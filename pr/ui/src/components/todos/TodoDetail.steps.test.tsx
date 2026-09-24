import { Button } from '@flanksource/clicky-ui/components';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { TodoItem, TodoRunOptions } from '../../types';
import type { RunContext } from './providers';
import { StatefulTodoDetail } from './todoDetailTestHarness';
import type { TodoDetailProps } from './TodoDetail';
import { queryTestWrapper } from './queryTestWrapper';
import type { TodoDetailView } from '../../routes';
import { readRunChoiceState } from './runChoiceStorage';

const execution = vi.hoisted(() => ({ run: vi.fn(), verify: vi.fn(), reset: vi.fn(), step: 'verify', spec: undefined as Record<string, unknown> | undefined }));
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
vi.mock('./TodoSession', () => ({
  TodoSession: ({ sessionTab, sessionIds, onSessionTabChange, onSessionIdsChange }: {
    sessionTab?: string;
    sessionIds?: string[];
    onSessionTabChange: (tab: string) => void;
    onSessionIdsChange: (ids: string[]) => void;
  }) => (
    <div>
      <div data-testid="session-view">{`${sessionTab ?? 'default'}|${(sessionIds ?? []).join(',')}`}</div>
      <Button onClick={() => onSessionTabChange('metadata')}>Inspector metadata</Button>
      <Button onClick={() => onSessionIdsChange(['run-b'])}>Select run-b</Button>
    </div>
  ),
}));
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
    <div data-testid="advanced-step">{initialMode}<Button onClick={() => onRun({ step: execution.step, spec: execution.spec ?? spec })}>Dispatch selected step</Button></div>
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
  execution.spec = undefined;
  const values = new Map<string, string>();
  vi.stubGlobal('localStorage', { getItem: (key: string) => values.get(key) ?? null, setItem: (key: string, value: string) => values.set(key, value) });
});
afterEach(() => vi.unstubAllGlobals());

async function openDetail(initialView: TodoDetailView = {}, onViewChange?: TodoDetailProps['onViewChange']) {
  render(<StatefulTodoDetail todo={todo} loading={false} dir="/repo" onChanged={() => {}} onDeleted={() => {}} initialView={initialView} onViewChange={onViewChange} />, { wrapper: queryTestWrapper() });
  await act(async () => {});
}

describe('TodoDetail routed view', () => {
  it('reports a detail tab click as a pushed view change', async () => {
    const onViewChange = vi.fn();
    await openDetail({}, onViewChange);

    fireEvent.click(screen.getByRole('button', { name: 'Session' }));

    expect(onViewChange).toHaveBeenCalledWith({ tab: 'session' }, undefined);
    expect(screen.getByTestId('session-view').textContent).toBe('default|');
  });

  it('hands the routed inspector tab and attempts to the session, pushing tab changes and replacing selections', async () => {
    const onViewChange = vi.fn();
    const view: TodoDetailView = { tab: 'session', sessionTab: 'costs', sessionIds: ['run-a'] };
    await openDetail(view, onViewChange);
    expect(screen.getByTestId('session-view').textContent).toBe('costs|run-a');

    fireEvent.click(screen.getByRole('button', { name: 'Inspector metadata' }));
    expect(onViewChange).toHaveBeenLastCalledWith({ ...view, sessionTab: 'metadata' }, undefined);

    fireEvent.click(screen.getByRole('button', { name: 'Select run-b' }));
    expect(onViewChange).toHaveBeenLastCalledWith({ ...view, sessionTab: 'metadata', sessionIds: ['run-b'] }, 'replace');
  });

  it('opens the Session tab for a new run and drops earlier attempt picks', async () => {
    const onViewChange = vi.fn();
    await openDetail({ tab: 'plan', sessionTab: 'raw', sessionIds: ['run-a'] }, onViewChange);

    fireEvent.click(screen.getByRole('button', { name: 'Start plan' }));

    await waitFor(() => expect(onViewChange).toHaveBeenCalledWith({ tab: 'session', sessionTab: 'raw' }, undefined));
  });

  it('lands on the Verification tab after dispatching a verification', async () => {
    const onViewChange = vi.fn();
    execution.verify.mockResolvedValue({});
    await openDetail({ tab: 'plan' }, onViewChange);

    fireEvent.click(screen.getByRole('button', { name: 'Start verify' }));

    await waitFor(() => expect(onViewChange).toHaveBeenCalledWith({ tab: 'verification' }, undefined));
  });
});

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

  // The catalog only offers `medium` for this model, so reconciling the request
  // would rewrite `high`; the advanced dialog's choice must reach the server as set.
  it('dispatches an advanced verification runtime exactly as set, without reconciling its effort', async () => {
    const chosen = { mode: 'agent', model: 'example-model', effort: 'high' };
    execution.spec = chosen;
    await openDetail();
    fireEvent.click(screen.getByRole('button', { name: 'Advanced verify' }));
    fireEvent.click(screen.getByRole('button', { name: 'Dispatch selected step' }));
    await waitFor(() => expect(execution.verify).toHaveBeenCalledWith({ ref: todo.ref, spec: chosen }));
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

  // The edited prompt was rendered for this todo. The advanced dispatch sends it
  // once; a later quick run of the same step, on this or any other todo, must
  // render its own prompt instead of replaying the remembered one.
  it('dispatches an edited prompt once without replaying it on the next quick run', async () => {
    const editedPrompt = { user: '## Review lifecycle routing\n\nEdited prompt body' };
    execution.step = 'plan';
    execution.spec = { ...spec, prompt: editedPrompt };
    await openDetail();
    fireEvent.click(screen.getByRole('button', { name: 'Advanced plan' }));
    fireEvent.click(screen.getByRole('button', { name: 'Dispatch selected step' }));
    await waitFor(() => expect(execution.run).toHaveBeenCalledWith(todo.ref, { step: 'plan', spec: { ...spec, prompt: editedPrompt } }));
    expect(readRunChoiceState().last.plan).toEqual({ step: 'plan', spec });
    fireEvent.click(screen.getByRole('button', { name: 'Start plan' }));
    await waitFor(() => expect(execution.run).toHaveBeenLastCalledWith(todo.ref, { step: 'plan', spec }));
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
