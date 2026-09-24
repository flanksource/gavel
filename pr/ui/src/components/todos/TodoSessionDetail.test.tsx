import type React from 'react';
import { fireEvent, render, renderHook, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { TodoSessionAttempt } from '../../types';
import { attemptCollection, captainSessionUrl, CopyAllDetailsButton, selectAttempt, useTodoSessionDetail } from './TodoSessionDetail';
import { queryTestWrapper } from './queryTestWrapper';

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, variant: _variant, size: _size, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement> & { variant?: string; size?: string }) => (
    // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky Button itself.
    <button type="button" {...props}>
      {children}
    </button>
  ),
}));

afterEach(() => {
  vi.unstubAllGlobals();
});

function attempt(ordinal: number, overrides: Partial<TodoSessionAttempt> = {}): TodoSessionAttempt {
  return {
    promptRunId: `run-${ordinal}`,
    ordinal,
    step: 'run',
    requested: {},
    resolved: {},
    status: 'completed',
    processActive: false,
    state: 'succeeded',
    phase: 'finished',
    queuedAt: `2026-07-14T1${ordinal}:00:00Z`,
    admissionSessionId: `admission-${ordinal}`,
    createdAt: `2026-07-14T1${ordinal}:00:00Z`,
    updatedAt: `2026-07-14T1${ordinal}:01:00Z`,
    verification: null,
    ...overrides,
  };
}

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });
}

describe('useTodoSessionDetail', () => {
  it('asks for the todo attempt list and nothing else', async () => {
    const fetchMock = vi.fn(async (_input: RequestInfo | URL) => jsonResponse({ attempts: [attempt(1)] }));
    vi.stubGlobal('fetch', fetchMock);

    const { result, unmount } = renderHook(() => useTodoSessionDetail('/repo', 'todo-1', true), { wrapper: queryTestWrapper() });

    await waitFor(() => expect(result.current.detail?.attempts).toHaveLength(1));
    expect(String(fetchMock.mock.calls[0]?.[0])).toBe('/api/todos/session/detail?dir=%2Frepo&ref=todo-1');
    unmount();
  });

  it('surfaces the server error instead of an empty attempt list', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => jsonResponse({ error: 'native TODO storage is unavailable' }, 501)));

    const { result, unmount } = renderHook(() => useTodoSessionDetail('/repo', 'todo-1', true), { wrapper: queryTestWrapper() });

    await waitFor(() => expect(result.current.error).toContain('native TODO storage is unavailable'));
    expect(result.current.detail).toBeNull();
    unmount();
  });

  it('keeps the loaded attempts when only the poll period changes', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => jsonResponse({ attempts: [attempt(1, { step: 'verify' })] })));

    // Opening the Verification tab drops the badge's slow poll to a fast one; the
    // already-listed attempts must survive that switch.
    const { result, rerender, unmount } = renderHook(
      ({ intervalMs }: { intervalMs: number }) => useTodoSessionDetail('/repo', 'todo-1', true, { intervalMs }),
      { initialProps: { intervalMs: 15000 }, wrapper: queryTestWrapper() }
    );
    await waitFor(() => expect(result.current.detail?.attempts).toHaveLength(1));

    rerender({ intervalMs: 1500 });
    expect(result.current.detail?.attempts).toHaveLength(1);
    unmount();
  });
});

describe('attemptCollection', () => {
  const current = attempt(2, {
    mode: 'plan',
    provider: 'anthropic',
    runtimeMode: 'agent',
    model: 'claude-opus-4-8',
    effort: 'high',
    executionSessionId: 'execution/2',
    providerSessionId: 'provider-2',
    durationMs: 60_000,
  });
  const providerOnly = attempt(1, { providerSessionId: 'provider-1', pid: 4242, stopping: true });
  const unstarted = attempt(3, { status: 'queued', phase: 'queued' });

  it('loads each attempt from its Captain session, preferring the execution session', () => {
    const collection = attemptCollection({ todoRef: 'todo-1', attempts: [unstarted, current, providerOnly], currentId: 'run-2', launch: null });

    expect(collection).toEqual({
      kind: 'session-collection',
      id: 'todo-attempts:todo-1',
      currentSessionId: 'run-2',
      sessions: [
        {
          id: 'run-3', label: 'Attempt #3', mode: 'run', status: 'queued',
          summary: { mode: 'run', status: 'queued', updatedAt: unstarted.updatedAt },
          session: { id: 'run-3', messages: [] },
        },
        {
          id: 'run-2', label: 'Attempt #2', mode: 'plan', status: 'completed',
          summary: {
            provider: 'anthropic', modelMode: 'agent', model: 'claude-opus-4-8', effort: 'high',
            mode: 'plan', status: 'completed', durationMs: 60_000, updatedAt: current.updatedAt,
          },
          src: '/api/captain/sessions/execution%2F2',
        },
        {
          id: 'run-1', label: 'Attempt #1', mode: 'run', status: 'stopping',
          summary: { mode: 'run', status: 'stopping', pid: 4242, updatedAt: providerOnly.updatedAt },
          src: '/api/captain/sessions/provider-1',
        },
      ],
    });
  });

  it('leads with the launching attempt until the server lists it', () => {
    const launching = attemptCollection({
      todoRef: 'todo-1', attempts: [current], currentId: 'run-new',
      launch: { status: 'admitted', step: 'run', promptRunId: 'run-new' },
    });
    expect(launching.currentSessionId).toBe('run-new');
    expect(launching.sessions.map(item => [item.id, item.label, item.status])).toEqual([
      ['run-new', 'Attempt #2', 'running'],
      ['run-2', 'Attempt #2', 'completed'],
    ]);
    expect(launching.sessions[0]?.session).toEqual({ id: 'run-new', messages: [] });

    const listed = attemptCollection({
      todoRef: 'todo-1', attempts: [attempt(3, { promptRunId: 'run-new' }), current], currentId: 'run-new',
      launch: { status: 'admitted', step: 'run', promptRunId: 'run-new' },
    });
    expect(listed.sessions.map(item => item.id)).toEqual(['run-new', 'run-2']);
  });
});

describe('selectAttempt', () => {
  const newest = attempt(2, { executionSessionId: 'execution-2', providerSessionId: 'provider-2' });
  const older = attempt(1, { executionSessionId: 'execution-1', providerSessionId: 'provider-1' });

  it.each([
    ['prompt run', 'run-1'],
    ['admission session', 'admission-1'],
    ['execution session', 'execution-1'],
    ['provider session', 'provider-1'],
  ])('picks the attempt its %s names', (_label, sessionId) => {
    expect(selectAttempt([newest, older], sessionId)).toBe(older);
  });

  it('falls back to the newest attempt for an unknown or missing session id', () => {
    expect(selectAttempt([newest, older], 'someone-else')).toBe(newest);
    expect(selectAttempt([newest, older], undefined)).toBe(newest);
    expect(selectAttempt([], undefined)).toBeUndefined();
  });
});

describe('CopyAllDetailsButton', () => {
  it('copies the attempts with the selected attempt session read from Captain', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', { configurable: true, value: { writeText } });
    const session = { id: 'execution-2', messages: [{ id: 'm1', role: 'assistant', parts: [{ type: 'text', text: 'plan ready' }] }] };
    const fetchMock = vi.fn(async (_input: RequestInfo | URL) => jsonResponse(session));
    vi.stubGlobal('fetch', fetchMock);
    const selected = attempt(2, { executionSessionId: 'execution-2' });

    render(<CopyAllDetailsButton detail={{ attempts: [selected] }} attempt={selected} />);
    fireEvent.click(screen.getByRole('button', { name: 'Copy all session details' }));

    await waitFor(() => expect(writeText).toHaveBeenCalledTimes(1));
    expect(String(fetchMock.mock.calls[0]?.[0])).toBe(captainSessionUrl('execution-2'));
    expect(JSON.parse(writeText.mock.calls[0]?.[0] as string)).toEqual({ attempts: [selected], session });
  });

  it('reports a failed session read instead of copying a partial payload', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', { configurable: true, value: { writeText } });
    vi.stubGlobal('fetch', vi.fn(async () => jsonResponse({ error: 'session execution-2 not found' }, 404)));
    const selected = attempt(2, { executionSessionId: 'execution-2' });

    render(<CopyAllDetailsButton detail={{ attempts: [selected] }} attempt={selected} />);
    fireEvent.click(screen.getByRole('button', { name: 'Copy all session details' }));

    await waitFor(() => expect(screen.getByRole('button', { name: 'Copy all session details' }).getAttribute('title')).toContain('session execution-2 not found'));
    expect(writeText).not.toHaveBeenCalled();
  });
});
