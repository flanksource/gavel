import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { TodoRunOptions } from '../../types';
import { usePlanActions } from './planActions';
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

function fetchJSON(payload: unknown) {
  const fetchMock = vi.fn().mockResolvedValue({ ok: true, json: async () => payload });
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

describe('usePlanActions request bodies', () => {
  it('sends only step/spec through approve, stripping the driver/runMode the caller carried', async () => {
    const fetchMock = fetchJSON({ todo: { ref: 'todo-1', title: 'x', status: 'in_progress', priority: 'medium' } });
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
    const fetchMock = fetchJSON({ todo: { ref: 'todo-1', title: 'x', status: 'review', priority: 'medium' }, status: 'revising' });
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
    const fetchMock = fetchJSON({ todo: { ref: 'todo-1', title: 'x', status: 'in_progress', priority: 'medium' }, status: 'resumed' });
    const { result } = renderHook(() => usePlanActions('/workspace'), { wrapper: queryTestWrapper() });
    const options: TodoRunOptions = { driver: 'cmux', runMode: 'run', spec: { mode: 'cmux', model: 'gpt-5.5', effort: 'high' }, resume: true };

    await act(async () => {
      await result.current.answer('todo-1', { answer: 'Use Postgres.', options });
    });

    const request = fetchMock.mock.calls[0]![1] as RequestInit;
    expect(JSON.parse(String(request.body))).toEqual({
      ref: 'todo-1',
      answer: 'Use Postgres.',
      options: { spec: { mode: 'cmux', model: 'gpt-5.5', effort: 'high' }, resume: true },
    });
  });
});
