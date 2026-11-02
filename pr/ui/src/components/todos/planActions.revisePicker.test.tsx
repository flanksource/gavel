import React from 'react';
import { fireEvent, render as rtlRender, renderHook, screen, waitFor, type RenderOptions } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { TodoItem } from '../../types';
import { PLAN_ACTIONS_CONTEXT as context } from './planActionsTestContext';
import { TodoReviewBanner } from './planActions';
import { PlanReviewBar, type ReviewMode } from './PlanReview';
import { REVISE_NEW_SESSION_HINT, REVISE_SAME_AGENT_LABEL } from './ReviseRuntimePicker';
import { useWorkspaceTodos } from './useWorkspaceTodos';
import { queryTestWrapper } from './queryTestWrapper';

// The revise box's optional agent picker: "same agent" (no options on the wire)
// unless the reviewer picks a runtime, and back to "same agent" when the
// selected todo changes. The wire-level revise body is covered in
// planActions.requestBody.test.tsx; this file covers the picker's UI contract.

function render(ui: React.ReactElement, options?: RenderOptions) {
  return rtlRender(ui, { wrapper: queryTestWrapper(), ...options });
}

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({
    children,
    variant: _variant,
    size: _size,
    loading: _loading,
    loadingLabel: _loadingLabel,
    asChild: _asChild,
    ...props
  }: React.ButtonHTMLAttributes<HTMLButtonElement> & { variant?: string; size?: string; loading?: boolean; loadingLabel?: React.ReactNode; asChild?: boolean }) => (
    // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky Button itself.
    <button type="button" {...props}>
      {children}
    </button>
  ),
  DropdownMenu: ({ trigger, children }: { trigger: React.ReactNode; children: (close: () => void) => React.ReactNode }) => (
    <div>
      {trigger}
      {children(() => {})}
    </div>
  ),
  Modal: ({ children, open }: { children?: React.ReactNode; open?: boolean }) => (open === false ? null : <div>{children}</div>),
}));

vi.mock('@flanksource/clicky-ui/ai', async importOriginal => ({
  ...(await importOriginal<typeof import('@flanksource/clicky-ui/ai')>()),
  effortOptionsForModel: (_model: unknown, fallback: string[]) => fallback,
  RuntimeBar: ({ value, onChange, ariaLabel }: {
    value: Record<string, unknown>;
    onChange: (value: Record<string, unknown>) => void;
    ariaLabel?: string;
  }) => (
    // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for RuntimeBar itself.
    <button type="button" aria-label={ariaLabel} onClick={() => onChange({ ...value, mode: 'agent', model: 'claude-opus-4-8', effort: 'high' })}>
      Select Claude runtime
    </button>
  ),
  SpecRuntimeEditor: () => <div>Advanced runtime editor</div>,
  promptRuntimeValueToPayload: (value: unknown) => value,
  reconcileModelCapabilities: (value: unknown) => value,
}));

const reviewTodo = (ref: string): TodoItem => ({ ref, title: `Plan ${ref}`, status: 'review', priority: 'medium', sessionId: `session-${ref}` });
const FEEDBACK = 'Split the migration into two steps.';
const PICKED_SPEC = { mode: 'agent', model: 'claude-opus-4-8', effort: 'high' };

function stubFetch() {
  const fetchMock = vi.fn(async (url: string, _init?: RequestInit) => {
    if (url.includes('/api/todos/run/context')) return { ok: true, json: async () => context } as Response;
    if (url.includes('/api/todos/session/plan')) return { ok: true, json: async () => ({ found: false }) } as Response;
    return new Response(
      `event: admitted\ndata: ${JSON.stringify({ todo: { ...reviewTodo('todo-1'), status: 'in_progress' }, status: 'revising', promptRunId: 'run-1' })}\n\n`,
      { headers: { 'Content-Type': 'text/event-stream' } },
    );
  });
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

function reviseBody(fetchMock: ReturnType<typeof stubFetch>): Record<string, unknown> {
  const call = fetchMock.mock.calls.find(([url]) => url.includes('/api/todos/plan/revise'));
  expect(call).toBeDefined();
  return JSON.parse(String(call![1]!.body));
}

function openChangesBox() {
  fireEvent.click(screen.getByRole('button', { name: /Request changes/ }));
  fireEvent.change(screen.getByPlaceholderText(/Describe what to change/), { target: { value: FEEDBACK } });
}

beforeEach(() => {
  const store: Record<string, string> = {};
  vi.stubGlobal('localStorage', {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => { store[key] = String(value); }),
    removeItem: vi.fn((key: string) => { delete store[key]; }),
    clear: vi.fn(() => { for (const key of Object.keys(store)) delete store[key]; }),
  });
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('revise box agent picker', () => {
  it('defaults to the plan\'s own agent and posts no options', async () => {
    const fetchMock = stubFetch();
    const onChanged = vi.fn();
    render(<TodoReviewBanner todo={reviewTodo('todo-1')} dir="/workspace" onChanged={onChanged} />);

    openChangesBox();
    expect(screen.getByText(new RegExp(REVISE_SAME_AGENT_LABEL))).toBeTruthy();
    expect(screen.queryByText(REVISE_NEW_SESSION_HINT)).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Send changes & re-plan' }));

    await waitFor(() => expect(onChanged).toHaveBeenCalledTimes(1));
    expect(reviseBody(fetchMock)).toEqual({ ref: 'todo-1', feedback: FEEDBACK });
  });

  it('opening the picker without choosing a runtime still posts no options', async () => {
    const fetchMock = stubFetch();
    const onChanged = vi.fn();
    render(<TodoReviewBanner todo={reviewTodo('todo-1')} dir="/workspace" onChanged={onChanged} />);

    openChangesBox();
    fireEvent.click(screen.getByRole('button', { name: 'Change' }));
    await screen.findByRole('button', { name: 'Plan runtime' });
    expect(screen.queryByText(REVISE_NEW_SESSION_HINT)).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Send changes & re-plan' }));

    await waitFor(() => expect(onChanged).toHaveBeenCalledTimes(1));
    expect(reviseBody(fetchMock)).toEqual({ ref: 'todo-1', feedback: FEEDBACK });
  });

  it('posts the picked runtime as plan options and explains the new session', async () => {
    const fetchMock = stubFetch();
    const onChanged = vi.fn();
    render(<TodoReviewBanner todo={reviewTodo('todo-1')} dir="/workspace" onChanged={onChanged} />);

    openChangesBox();
    fireEvent.click(screen.getByRole('button', { name: 'Change' }));
    fireEvent.click(await screen.findByRole('button', { name: 'Plan runtime' }));
    expect(await screen.findByText(REVISE_NEW_SESSION_HINT)).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Send changes & re-plan' }));

    await waitFor(() => expect(onChanged).toHaveBeenCalledTimes(1));
    expect(reviseBody(fetchMock)).toEqual({ ref: 'todo-1', feedback: FEEDBACK, options: { step: 'plan', spec: PICKED_SPEC } });
  });

  it('goes back to the plan\'s own agent when the reviewer clears the pick', async () => {
    const fetchMock = stubFetch();
    const onChanged = vi.fn();
    render(<TodoReviewBanner todo={reviewTodo('todo-1')} dir="/workspace" onChanged={onChanged} />);

    openChangesBox();
    fireEvent.click(screen.getByRole('button', { name: 'Change' }));
    fireEvent.click(await screen.findByRole('button', { name: 'Plan runtime' }));
    fireEvent.click(screen.getByRole('button', { name: /Use same agent as the plan/ }));
    expect(screen.queryByText(REVISE_NEW_SESSION_HINT)).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Send changes & re-plan' }));

    await waitFor(() => expect(onChanged).toHaveBeenCalledTimes(1));
    expect(reviseBody(fetchMock)).toEqual({ ref: 'todo-1', feedback: FEEDBACK });
  });

  it('resets the picked agent when the selected todo changes', async () => {
    const fetchMock = stubFetch();
    const onChanged = vi.fn();
    const { rerender } = render(<TodoReviewBanner todo={reviewTodo('todo-1')} dir="/workspace" onChanged={onChanged} />);

    openChangesBox();
    fireEvent.click(screen.getByRole('button', { name: 'Change' }));
    fireEvent.click(await screen.findByRole('button', { name: 'Plan runtime' }));
    expect(await screen.findByText(REVISE_NEW_SESSION_HINT)).toBeTruthy();

    rerender(<TodoReviewBanner todo={reviewTodo('todo-2')} dir="/workspace" onChanged={onChanged} />);
    openChangesBox();
    expect(screen.getByText(new RegExp(REVISE_SAME_AGENT_LABEL))).toBeTruthy();
    expect(screen.queryByText(REVISE_NEW_SESSION_HINT)).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Send changes & re-plan' }));

    await waitFor(() => expect(onChanged).toHaveBeenCalledTimes(1));
    expect(reviseBody(fetchMock)).toEqual({ ref: 'todo-2', feedback: FEEDBACK });
  });
});

describe('PlanReviewBar revise box agent picker', () => {
  function renderBar(ref: string) {
    const todo = reviewTodo(ref);
    const { result } = renderHook(() => useWorkspaceTodos([]), { wrapper: queryTestWrapper() });
    const review: ReviewMode = {
      active: true, count: 1, index: 0,
      queue: [{ todo, workspace: { name: 'Workspace', dir: '/workspace', repos: [] } }],
      enter: vi.fn(), exit: vi.fn(), goNext: vi.fn(), goPrev: vi.fn(), advanceAfterAction: vi.fn(),
    };
    const bar = (selectedRef: string) => (
      <PlanReviewBar review={review} todos={{
        ...result.current,
        selected: { dir: '/workspace', ref: selectedRef },
        detail: { ...reviewTodo(selectedRef) },
        updateItem: vi.fn(), refresh: vi.fn(),
      }} />
    );
    return { review, bar, ...render(bar(ref)) };
  }

  it('opens with c, keeps Cmd+Enter submit, and posts no options by default', async () => {
    const fetchMock = stubFetch();
    const { review } = renderBar('todo-1');

    fireEvent.keyDown(document, { key: 'c' });
    const box = await screen.findByPlaceholderText(/Describe what to change/);
    fireEvent.change(box, { target: { value: FEEDBACK } });
    fireEvent.keyDown(box, { key: 'Enter', metaKey: true });

    await waitFor(() => expect(review.advanceAfterAction).toHaveBeenCalledTimes(1));
    expect(reviseBody(fetchMock)).toEqual({ ref: 'todo-1', feedback: FEEDBACK });
  });

  it('posts the picked runtime and resets it when the selected todo changes', async () => {
    const fetchMock = stubFetch();
    const { review, bar, rerender } = renderBar('todo-1');

    fireEvent.keyDown(document, { key: 'c' });
    await screen.findByPlaceholderText(/Describe what to change/);
    fireEvent.click(screen.getByRole('button', { name: 'Change' }));
    fireEvent.click(await screen.findByRole('button', { name: 'Plan runtime' }));
    expect(await screen.findByText(REVISE_NEW_SESSION_HINT)).toBeTruthy();

    rerender(bar('todo-2'));
    fireEvent.keyDown(document, { key: 'c' });
    const box = await screen.findByPlaceholderText(/Describe what to change/);
    expect(screen.getByText(new RegExp(REVISE_SAME_AGENT_LABEL))).toBeTruthy();
    expect(screen.queryByText(REVISE_NEW_SESSION_HINT)).toBeNull();
    fireEvent.change(box, { target: { value: FEEDBACK } });
    fireEvent.keyDown(box, { key: 'Enter', ctrlKey: true });

    await waitFor(() => expect(review.advanceAfterAction).toHaveBeenCalledTimes(1));
    expect(reviseBody(fetchMock)).toEqual({ ref: 'todo-2', feedback: FEEDBACK });
  });

  it('posts the picked runtime as plan options', async () => {
    const fetchMock = stubFetch();
    const { review } = renderBar('todo-1');

    fireEvent.keyDown(document, { key: 'c' });
    const box = await screen.findByPlaceholderText(/Describe what to change/);
    fireEvent.click(screen.getByRole('button', { name: 'Change' }));
    fireEvent.click(await screen.findByRole('button', { name: 'Plan runtime' }));
    fireEvent.change(box, { target: { value: FEEDBACK } });
    fireEvent.keyDown(box, { key: 'Enter', metaKey: true });

    await waitFor(() => expect(review.advanceAfterAction).toHaveBeenCalledTimes(1));
    expect(reviseBody(fetchMock)).toEqual({ ref: 'todo-1', feedback: FEEDBACK, options: { step: 'plan', spec: PICKED_SPEC } });
  });
});
