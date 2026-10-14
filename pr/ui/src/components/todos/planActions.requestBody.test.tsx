import { act, fireEvent, render, renderHook, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { TodoItem, TodoRunOptions } from '../../types';
import { TodoReviewBanner, usePlanActions } from './planActions';
import { PlanReviewBar, type ReviewMode } from './PlanReview';
import { useWorkspaceTodos } from './useWorkspaceTodos';
import { queryTestWrapper } from './queryTestWrapper';

// Split out of planActions.test.tsx (which was already past the 500-line
// limit) purely to keep that file's size in check — see planActions.test.tsx
// for the component-level coverage of the same hook's UI.
//
// /api/todos/plan/approve, /plan/revise and /answer all continue a run
// through the same seam as /api/todos/run (see todo_review.go's
// continuation), which decodes just as strictly: driver/runMode/prompt sent
// inside `options` are rejected, or worse — a mismatched `step` inside it
// fails the request with "this action runs the %s step; options.step %q
// cannot change it". These tests assert the exact wire body usePlanActions
// sends, so a raw TodoRunOptions (still carrying driver/runMode for the
// dialog's own bookkeeping) can never leak back onto the wire.

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

function fetchLaunch(payload: unknown) {
  const fetchMock = vi.fn().mockResolvedValue(new Response(
    `event: resolved\ndata: ${JSON.stringify({ spec: {}, specYaml: '{}\n', step: 'run' })}\n\nevent: admitted\ndata: ${JSON.stringify({ ...payload as object, promptRunId: 'run-1', run: { promptRunId: 'run-1' } })}\n\n`,
    { headers: { 'Content-Type': 'text/event-stream' } },
  ));
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

describe('usePlanActions request bodies', () => {
  it('sends only step/spec through approve, stripping the driver/runMode the caller carried', async () => {
    const fetchMock = fetchLaunch({ todo: { ref: 'todo-1', title: 'x', status: 'in_progress', priority: 'medium' } });
    const { result } = renderHook(() => usePlanActions('/workspace'), { wrapper: queryTestWrapper() });
    const options: TodoRunOptions = { driver: 'agent', runMode: 'plan', spec: { mode: 'agent', model: 'claude-opus-4-8', effort: 'high' } };

    await act(async () => {
      await result.current.approve('todo-1', { run: true, options });
    });

    const request = fetchMock.mock.calls[0]![1] as RequestInit;
    expect(JSON.parse(String(request.body))).toEqual({
      ref: 'todo-1',
      run: true,
      options: { step: 'run', spec: { mode: 'agent', model: 'claude-opus-4-8', effort: 'high' } },
    });
  });

  it('sends the plan step for revise, built from the last remembered plan options', async () => {
    localStorage.setItem(
      'gavel.pr-ui.todoRunChoices.v3',
      JSON.stringify({
        last: { plan: { driver: 'cli', runMode: 'plan', plan: true, spec: { mode: 'agent', model: 'claude-opus-4-8', effort: 'medium' } } },
        recentAdvanced: {},
      }),
    );
    const fetchMock = fetchLaunch({ todo: { ref: 'todo-1', title: 'x', status: 'review', priority: 'medium' }, status: 'revising' });
    const { result } = renderHook(() => usePlanActions('/workspace'), { wrapper: queryTestWrapper() });

    await act(async () => {
      await result.current.revise('todo-1', 'Please rework the migration.');
    });

    const request = fetchMock.mock.calls[0]![1] as RequestInit;
    expect(JSON.parse(String(request.body))).toEqual({
      ref: 'todo-1',
      feedback: 'Please rework the migration.',
      options: { step: 'plan', spec: { mode: 'agent', model: 'claude-opus-4-8', effort: 'medium' } },
    });
  });

  it('strips driver/runMode from answer options without asserting a step of its own', async () => {
    const fetchMock = fetchLaunch({ todo: { ref: 'todo-1', title: 'x', status: 'in_progress', priority: 'medium' }, status: 'resumed' });
    const { result } = renderHook(() => usePlanActions('/workspace'), { wrapper: queryTestWrapper() });
    const options: TodoRunOptions = { driver: 'cmux', runMode: 'run', spec: { mode: 'cmux', model: 'gpt-5.5', effort: 'high' }, resume: true };

    await act(async () => {
      await result.current.answer('todo-1', { sessionId: ' session-1 ', answer: 'Use Postgres.', options });
    });

    const request = fetchMock.mock.calls[0]![1] as RequestInit;
    expect(JSON.parse(String(request.body))).toEqual({
      ref: 'todo-1',
      sessionId: 'session-1',
      answer: 'Use Postgres.',
      options: { spec: { mode: 'cmux', model: 'gpt-5.5', effort: 'high' }, resume: true },
    });
  });

  it('reports a missing answering session without sending a resume request', async () => {
    const fetchMock = fetchLaunch({ todo: { ref: 'todo-1', title: 'x', status: 'in_progress', priority: 'medium' } });
    const { result } = renderHook(() => usePlanActions('/workspace'), { wrapper: queryTestWrapper() });

    await act(async () => {
      await result.current.answer('todo-1', { sessionId: ' ', answer: 'Use Postgres.' });
    });

    expect(fetchMock).not.toHaveBeenCalled();
    expect(result.current.error).toMatch(/session.*required/i);
  });

  it('keeps an answer draft visible without opening a launch when its session is missing', async () => {
    const fetchMock = fetchLaunch({});
    const onLaunch = vi.fn();
    render(<TodoReviewBanner todo={{ ref: 'todo-1', title: 'Question', status: 'ask', priority: 'medium' }} dir="/workspace" onChanged={vi.fn()} onLaunch={onLaunch} />, { wrapper: queryTestWrapper() });

    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'Use Postgres.' } });
    fireEvent.click(screen.getByRole('button', { name: 'Send & resume' }));

    expect(await screen.findByText(/session.*required/i)).toBeTruthy();
    expect(fetchMock).not.toHaveBeenCalled();
    expect(onLaunch).not.toHaveBeenCalled();
    expect((screen.getByRole('textbox') as HTMLTextAreaElement).value).toBe('Use Postgres.');
  });
});

describe('PlanReviewBar answers', () => {
  it.each([false, true])('resumes the asking session from the current todo (detail loaded: %s)', async detailLoaded => {
    const todo: TodoItem = { ref: 'todo-1', title: 'Question', status: 'ask', priority: 'medium', sessionId: 'queue-session' };
    const fetchMock = fetchLaunch({ todo: { ...todo, status: 'in_progress' }, status: 'resumed' });
    const { result } = renderHook(() => useWorkspaceTodos([]), { wrapper: queryTestWrapper() });
    const review: ReviewMode = {
      active: true, count: 1, index: 0,
      queue: [{ todo, workspace: { name: 'Workspace', dir: '/workspace', repos: [] } }],
      enter: vi.fn(), exit: vi.fn(), goNext: vi.fn(), goPrev: vi.fn(), advanceAfterAction: vi.fn(),
    };
    render(<PlanReviewBar review={review} todos={{
      ...result.current,
      selected: { dir: '/workspace', ref: todo.ref },
      detail: detailLoaded ? { ...todo, sessionId: 'detail-session' } : null,
      updateItem: vi.fn(), refresh: vi.fn(),
    }} />, { wrapper: queryTestWrapper() });

    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'Use Postgres.' } });
    fireEvent.click(screen.getByRole('button', { name: 'Send & resume' }));

    await waitFor(() => expect(review.advanceAfterAction).toHaveBeenCalledTimes(1));
    const answerCall = fetchMock.mock.calls.find(call => typeof call[0] === 'string' && call[0].includes('/api/todos/answer'));
    expect(JSON.parse(String(answerCall?.[1].body))).toEqual({
      ref: todo.ref, sessionId: detailLoaded ? 'detail-session' : 'queue-session', answer: 'Use Postgres.',
    });
  });
});
