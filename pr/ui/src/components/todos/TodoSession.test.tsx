import type React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, fireEvent, render, renderHook, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { SessionToolDecision } from '@flanksource/clicky-ui/ai';
import type { TodoItem, TodoSessionAttempt } from '../../types';
import { TodoSession, useSessionStatus } from './TodoSession';
import { SessionErrorDetails } from './SessionErrorDetails';
import { queryTestWrapper } from './queryTestWrapper';
import { todoQueryKeys } from './todoQueries';

const sessionDetailMock = vi.hoisted(() => ({
  attempt: {
    promptRunId: 'run-1',
    ordinal: 1,
    step: 'implement',
    requested: {},
    resolved: {},
    status: 'running',
    processActive: true,
    state: 'running',
    phase: 'working',
    queuedAt: '2026-08-02T10:00:00Z',
    admissionSessionId: 'admission-1',
    createdAt: '2026-08-02T10:00:00Z',
    updatedAt: '2026-08-02T10:01:00Z',
  } satisfies TodoSessionAttempt,
}));

vi.mock('@flanksource/clicky-ui/ai', () => ({
  SessionInspector: () => <div data-testid="session-inspector" />,
  SessionViewer: () => <div data-testid="session-viewer" />,
}));

vi.mock('./TodoSessionDetail', () => ({
  CopyAllDetailsButton: () => null,
  SessionDiagnostics: () => null,
  ThreadInspector: ({
    onPendingToolDecision,
    onStop,
  }: {
    onPendingToolDecision: (decision: SessionToolDecision) => Promise<void>;
    onStop: (attempt: TodoSessionAttempt) => Promise<void>;
  }) => (
    <div>
      {/* oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test control for the mocked ThreadInspector callback. */}
      <button type="button" onClick={() => void onPendingToolDecision({
        event: { id: 'question-1', kind: 'tool', tool: 'AskUserQuestion' },
        allow: true,
        message: 'Proceed',
      })}>Answer question</button>
      {/* oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test control for the mocked ThreadInspector callback. */}
      <button type="button" onClick={() => void onStop(sessionDetailMock.attempt)}>Stop attempt</button>
    </div>
  ),
  useTodoSessionDetail: () => ({
    detail: {
      attempts: [sessionDetailMock.attempt],
      diagnostics: [],
      thread: { providerSessionId: 'session-1', status: 'ask' },
    },
    error: '',
  }),
}));

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({
    children,
    variant: _variant,
    size: _size,
    ...props
  }: React.ButtonHTMLAttributes<HTMLButtonElement> & { variant?: string; size?: string }) => (
    // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky Button itself.
    <button type="button" {...props}>
      {children}
    </button>
  ),
  DropdownMenu: ({ trigger, children }: {
    trigger: React.ReactNode;
    children: (close: () => void) => React.ReactNode;
  }) => <div>{trigger}{children(() => {})}</div>,
}));

vi.mock('@flanksource/clicky-ui/icons', async (importOriginal) => {
  const Icon = (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />;
  return {
    ...(await importOriginal<object>()),
    UiCancel: Icon,
    UiCheck: Icon,
    UiChevronDown: Icon,
    UiChevronRight: Icon,
    UiCircleFilled: Icon,
    UiComment: Icon,
    UiCopy: Icon,
    UiError: Icon,
    UiLightbulb: Icon,
    UiPass: Icon,
    UiShield: Icon,
    UiWarningTriangle: Icon,
  };
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('session errors', () => {
  it('keeps the complete backend error from a failed stats request', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
      statusText: 'Internal Server Error',
      text: async () => JSON.stringify({
        error: 'captain session conflict: provider session ID "session-1" is ambiguous',
        matches: ['captain-1', 'captain-2'],
      }),
    });
    vi.stubGlobal('fetch', fetchMock);

    const { result, unmount } = renderHook(() => useSessionStatus('/repo', 'session-1', true), {
      wrapper: queryTestWrapper(),
    });

    await waitFor(() => expect(result.current.error).toContain('HTTP 500 Internal Server Error'));
    expect(result.current.error).toContain('provider session ID "session-1" is ambiguous');
    expect(result.current.error).toContain('"captain-1"');
    expect(result.current.error).toContain('"captain-2"');
    unmount();
  });

  it('updates the exact session stats cache after approving a pending tool', async () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const statsKey = todoQueryKeys.sessionStats('/repo', 'session-1');
    vi.stubGlobal('fetch', vi.fn()
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          found: true,
          state: 'approval',
          inProgress: true,
          approval: { sessionId: 'session-1', tool: 'Bash', input: {}, toolUseId: 'tool-1' },
        }),
      })
      .mockResolvedValueOnce({ ok: true, json: async () => ({ resolved: true, allow: true }) }));

    const { result } = renderHook(() => useSessionStatus('/repo', 'session-1', true), {
      wrapper: ({ children }) => <QueryClientProvider client={client}>{children}</QueryClientProvider>,
    });
    await waitFor(() => expect(result.current.approval?.toolUseId).toBe('tool-1'));

    await act(() => result.current.approve(true));

    expect(client.getQueryData<{ approval?: unknown }>(statsKey)?.approval).toBeUndefined();
  });

  it('prefixes approval failures with the session action context', async () => {
    vi.stubGlobal('fetch', vi.fn()
      .mockResolvedValueOnce({ ok: true, json: async () => ({ found: true, state: 'approval', inProgress: true }) })
      .mockResolvedValueOnce({
        ok: false,
        status: 409,
        json: async () => ({ error: 'approval is no longer pending' }),
        text: async () => JSON.stringify({ error: 'approval is no longer pending' }),
      }));
    const { result } = renderHook(() => useSessionStatus('/repo', 'session-1', true), {
      wrapper: queryTestWrapper(),
    });
    await waitFor(() => expect(result.current.state).toBe('approval'));

    await expect(result.current.approve(false)).rejects.toThrow(/session approval update failed.*approval is no longer pending/i);
  });

  it('reveals and copies every labeled session error', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText },
    });
    render(<SessionErrorDetails errors={[
      { source: 'Session stream', message: 'stream failed\nprovider session ID is ambiguous' },
      { source: 'Session status', message: 'HTTP 500 Internal Server Error\nsecond backend detail' },
    ]} />);

    expect(screen.queryByText('provider session ID is ambiguous')).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Show details' }));
    expect(screen.getByText(/provider session ID is ambiguous/)).toBeTruthy();
    expect(screen.getByText(/second backend detail/)).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'Copy error details' }));
    await waitFor(() => expect(writeText).toHaveBeenCalledWith([
      'Session stream',
      'stream failed\nprovider session ID is ambiguous',
      '',
      'Session status',
      'HTTP 500 Internal Server Error\nsecond backend detail',
    ].join('\n')));
  });

  it('uses the DOM copy path when clipboard permission is denied', async () => {
    const writeText = vi.fn().mockRejectedValue(new Error('clipboard permission denied'));
    const execCommand = vi.fn().mockReturnValue(true);
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText },
    });
    Object.defineProperty(document, 'execCommand', {
      configurable: true,
      value: execCommand,
    });
    render(<SessionErrorDetails errors={[
      { source: 'Session status', message: 'HTTP 500 Internal Server Error' },
    ]} />);

    fireEvent.click(screen.getByRole('button', { name: 'Show details' }));
    fireEvent.click(screen.getByRole('button', { name: 'Copy error details' }));

    await waitFor(() => expect(execCommand).toHaveBeenCalledWith('copy'));
    expect(screen.getByRole('button', { name: 'Copy error details' }).textContent).toContain('Copied');
  });
});

describe('TodoSession mutations', () => {
  const todo: TodoItem = {
    ref: 'todo-1',
    title: 'Answer the agent',
    status: 'ask',
    priority: 'medium',
    sessionId: 'session-1',
    questions: [{ text: 'Proceed?' }],
  };

  function renderSession(fetchMock: ReturnType<typeof vi.fn>) {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const invalidateQueries = vi.spyOn(client, 'invalidateQueries');
    vi.stubGlobal('EventSource', class {
      addEventListener() {}
      close() {}
    });
    vi.stubGlobal('fetch', fetchMock);
    const onChanged = vi.fn();
    render(<TodoSession dir="/repo" sessionId="session-1" active todo={todo} onChanged={onChanged} />, {
      wrapper: ({ children }) => <QueryClientProvider client={client}>{children}</QueryClientProvider>,
    });
    return { client, invalidateQueries, onChanged };
  }

  const sessionStats = {
    found: true,
    state: 'ask',
    inProgress: false,
    durationMs: 0,
    inputTokens: 0,
    outputTokens: 0,
    cacheReadTokens: 0,
    cacheCreationTokens: 0,
    totalTokens: 0,
    contextTokens: 0,
    contextWindow: 0,
    turns: 0,
    compactions: 0,
    costUsd: 0,
  };

  it('projects an answered todo into caches and invalidates only its current session reads', async () => {
    const answered = { ...todo, status: 'in_progress' as const, questions: [] };
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes('/session/stats')) return { ok: true, json: async () => sessionStats };
      if (url === '/api/todos/answer') return { ok: true, json: async () => ({ todo: answered, status: 'resumed' }) };
      throw new Error(`unexpected fetch ${url}`);
    });
    const { invalidateQueries, onChanged } = renderSession(fetchMock);

    fireEvent.click(screen.getByRole('button', { name: 'Answer question' }));

    await waitFor(() => expect(onChanged).toHaveBeenCalledWith(answered));
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: todoQueryKeys.sessionStats('/repo', 'session-1') });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: todoQueryKeys.sessionDetail('/repo', 'todo-1', 'session-1', false) });
  });

  it('invalidates the stopped attempt and session caches after the server accepts cancellation', async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes('/session/stats')) return { ok: true, json: async () => ({ ...sessionStats, state: 'working', inProgress: true }) };
      if (url.includes('/session/stop')) return { ok: true, json: async () => ({ status: 'stopping', promptRunId: 'run-1' }) };
      throw new Error(`unexpected fetch ${url}`);
    });
    const { invalidateQueries } = renderSession(fetchMock);

    fireEvent.click(screen.getByRole('button', { name: 'Stop attempt' }));

    await waitFor(() => expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: todoQueryKeys.sessionDetail('/repo', 'todo-1', 'session-1', false),
    }));
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: todoQueryKeys.sessionDetail('/repo', 'todo-1', undefined, true) });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: todoQueryKeys.sessionStats('/repo', 'session-1') });
  });
});
