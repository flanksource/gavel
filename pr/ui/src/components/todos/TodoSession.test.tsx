import type React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, fireEvent, render, renderHook, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { SessionCollectionInput, SessionCollectionItem, SessionMetadataSummary, SessionPendingTool, SessionToolDecision } from '@flanksource/clicky-ui/ai';
import type { TodoItem, TodoSessionAttempt } from '../../types';
import { TodoSession, useSessionStatus } from './TodoSession';
import { SessionErrorDetails } from './SessionErrorDetails';
import { queryTestWrapper } from './queryTestWrapper';
import { todoQueryKeys } from './todoQueries';

interface InspectorDouble {
  session?: SessionCollectionInput;
  src?: string;
  layout?: string;
  metadata?: SessionMetadataSummary;
  toolbarActions?: React.ReactNode;
  transcriptProps?: { pendingTools?: SessionPendingTool[]; onPendingToolDecision?: (decision: SessionToolDecision) => Promise<void> };
  renderSessionActions?: (item: SessionCollectionItem) => React.ReactNode;
  selectedSessionIds?: readonly string[];
  onSelectedSessionIdsChange?: (ids: string[]) => void;
}

// The inspector itself is clicky-ui's (it fetches each item's `src`); the
// double exposes what TodoSession hands it and drives its callbacks the way the
// real attempt picker and transcript do.
vi.mock('@flanksource/clicky-ui/ai', async (importOriginal) => ({
  ...(await importOriginal<object>()),
  SessionInspector: (props: InspectorDouble) => {
    const decide = (decision: SessionToolDecision) => void props.transcriptProps?.onPendingToolDecision?.(decision);
    const items = props.session?.sessions ?? [];
    return (
      <div
        data-testid="session-inspector"
        data-src={props.src}
        data-current={props.session?.currentSessionId}
        data-items={JSON.stringify(items.map(({ id, label, status, src }) => ({ id, label, status, src })))}
        data-selected={(props.selectedSessionIds ?? []).join(',')}
        data-pending={JSON.stringify(props.transcriptProps?.pendingTools ?? [])}
        data-layout={props.layout}
        data-provider={props.metadata?.provider}
        data-mode={props.metadata?.executionMode}
        data-model={props.metadata?.model}
        data-effort={props.metadata?.reasoningEffort}
        data-context={props.metadata?.context?.freePercent}
      >
        {props.toolbarActions}
        {items.map(item => (
          <span key={item.id}>
            {/* oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test control standing in for the attempt picker. */}
            <button type="button" onClick={() => props.onSelectedSessionIdsChange?.([item.id])}>{`Pick ${item.label}`}</button>
            {props.renderSessionActions?.(item)}
          </span>
        ))}
        {/* oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test control for the transcript decision callback. */}
        <button type="button" onClick={() => decide({ event: { id: 'question-1', kind: 'tool', tool: 'AskUserQuestion' }, allow: true, message: 'Proceed' })}>Answer question</button>
        {/* A decision whose event carries an approvalId is a live tool-approval
            (clicky-ui's SessionViewer stamps this on from the matched pending
            tool) — "Reject with comment" wires the reason textarea's value
            straight into decision.message. */}
        {/* oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test control for the transcript decision callback. */}
        <button type="button" onClick={() => decide({ event: { id: 'tool-2', kind: 'tool', tool: 'Bash', approvalId: 'approval-9' }, allow: false, message: 'Looks risky' })}>Deny with reason</button>
        {/* oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test control for the transcript decision callback. */}
        <button type="button" onClick={() => decide({ event: { id: 'tool-3', kind: 'tool', tool: 'AskUserQuestion', approvalId: 'approval-9' }, allow: false, answers: { location: 'Inline' } })}>Reject answered question</button>
      </div>
    );
  },
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
    UiStop: Icon,
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
    const approveFetch = vi.fn()
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          found: true,
          state: 'approval',
          inProgress: true,
          approvals: [{ approvalId: 'approval-1', sessionId: 'session-1', tool: 'Bash', input: {}, toolUseId: 'tool-1' }],
        }),
      })
      .mockResolvedValueOnce({ ok: true, json: async () => ({ resolved: true }) });
    vi.stubGlobal('fetch', approveFetch);

    const { result } = renderHook(() => useSessionStatus('/repo', 'session-1', true), {
      wrapper: ({ children }) => <QueryClientProvider client={client}>{children}</QueryClientProvider>,
    });
    await waitFor(() => expect(result.current.approvals[0]?.approvalId).toBe('approval-1'));

    await act(() => result.current.approve('approval-1', 'approve'));

    expect(client.getQueryData<{ approvals?: unknown[] }>(statsKey)?.approvals).toEqual([]);
    const [, init] = approveFetch.mock.calls[1] as [string, RequestInit];
    expect(JSON.parse(init.body as string)).toEqual({ approvalId: 'approval-1', action: 'approve' });
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

    await expect(result.current.approve('approval-1', 'deny')).rejects.toThrow(/session approval update failed.*approval is no longer pending/i);
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

describe('TodoSession', () => {
  const todo: TodoItem = {
    ref: 'todo-1',
    title: 'Answer the agent',
    status: 'ask',
    priority: 'medium',
    sessionId: 'session-1',
    questions: [{ text: 'Proceed?' }],
  };

  function attempt(ordinal: number, overrides: Partial<TodoSessionAttempt> = {}): TodoSessionAttempt {
    return {
      promptRunId: `run-${ordinal}`,
      ordinal,
      step: 'implement',
      requested: {},
      resolved: {},
      status: 'completed',
      processActive: false,
      state: 'succeeded',
      phase: 'finished',
      queuedAt: '2026-08-02T10:00:00Z',
      admissionSessionId: `admission-${ordinal}`,
      createdAt: '2026-08-02T10:00:00Z',
      updatedAt: '2026-08-02T10:01:00Z',
      verification: null,
      ...overrides,
    };
  }

  // The attempt the todo's session id names: live, stoppable, and followed
  // through its provider session.
  const current = attempt(2, {
    status: 'running', processActive: true, state: 'running', phase: 'generate',
    executionSessionId: 'exec-2', providerSessionId: 'session-1', canStop: true,
    provider: 'anthropic', runtimeMode: 'agent', model: 'claude-opus-5', effort: 'high',
  });
  const older = attempt(1, { providerSessionId: 'provider-older' });

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

  type Route = (url: string, init?: RequestInit) => Response | undefined;

  function stubServer(route: Route, attempts: TodoSessionAttempt[] = [current, older]) {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      const routed = route(url, init);
      if (routed) return routed;
      if (url.startsWith('/api/todos/session/detail')) return Response.json({ attempts });
      throw new Error(`unexpected fetch ${url}`);
    });
    vi.stubGlobal('fetch', fetchMock);
    return fetchMock;
  }

  // A working session: nothing is asked, so the transcript is never read here.
  const statsRoute = (stats: Record<string, unknown>): Route => url => url.includes('/session/stats') ? Response.json({ ...sessionStats, state: 'working', inProgress: true, ...stats }) : undefined;

  function renderSession(props: Partial<React.ComponentProps<typeof TodoSession>> = {}, client = new QueryClient({ defaultOptions: { queries: { retry: false } } })) {
    const invalidateQueries = vi.spyOn(client, 'invalidateQueries');
    const onChanged = vi.fn();
    render(<TodoSession dir="/repo" sessionId="session-1" active todo={todo} onChanged={onChanged} {...props} />, {
      wrapper: ({ children }) => <QueryClientProvider client={client}>{children}</QueryClientProvider>,
    });
    return { client, invalidateQueries, onChanged };
  }

  const items = (inspector: HTMLElement) => JSON.parse(inspector.getAttribute('data-items') ?? '[]');

  it('hands the inspector one attempt per prompt run, each loaded from its Captain session', async () => {
    stubServer(statsRoute({}));
    renderSession();

    const inspector = await screen.findByTestId('session-inspector');
    await waitFor(() => expect(items(inspector)).toEqual([
      { id: 'run-2', label: 'Attempt #2', status: 'running', src: '/api/captain/sessions/exec-2' },
      { id: 'run-1', label: 'Attempt #1', status: 'completed', src: '/api/captain/sessions/provider-older' },
    ]));
    expect(inspector.getAttribute('data-current')).toBe('run-2');
  });

  it('reports a picked attempt so the view can route it to ?sessions=', async () => {
    stubServer(statsRoute({}));
    const onSessionIdsChange = vi.fn();
    renderSession({ sessionIds: ['run-2'], onSessionIdsChange });

    const inspector = await screen.findByTestId('session-inspector');
    await waitFor(() => expect(items(inspector)).toHaveLength(2));
    expect(inspector.getAttribute('data-selected')).toBe('run-2');
    fireEvent.click(screen.getByRole('button', { name: 'Pick Attempt #1' }));

    expect(onSessionIdsChange).toHaveBeenCalledWith(['run-1']);
  });

  it('uses the compact inspector with the attempt runtime identity and live context', async () => {
    stubServer(statsRoute({ agent: 'claude', model: 'claude-opus-5', effort: 'high', contextTokens: 14_000, contextWindow: 100_000 }));
    renderSession();

    const inspector = await screen.findByTestId('session-inspector');
    await waitFor(() => expect(inspector.getAttribute('data-context')).toBe('86'));
    expect(inspector.getAttribute('data-layout')).toBe('compact');
    expect(inspector.getAttribute('data-provider')).toBe('anthropic');
    expect(inspector.getAttribute('data-mode')).toBe('agent');
    expect(inspector.getAttribute('data-model')).toBe('claude-opus-5');
    expect(inspector.getAttribute('data-effort')).toBe('high');
    expect(screen.getByRole('button', { name: 'Copy all session details' })).toBeTruthy();
  });

  it('offers the blocking question read from the Captain transcript while the session asks', async () => {
    const question = { questions: [{ id: 'q1', question: 'Which location?', options: [{ label: 'Inline' }] }] };
    const fetchMock = stubServer(url => {
      if (url.includes('/session/stats')) return Response.json(sessionStats);
      if (url === '/api/captain/sessions/exec-2') return Response.json({
        id: 'exec-2',
        messages: [{ id: 'm1', role: 'assistant', parts: [{ type: 'tool', toolName: 'AskUserQuestion', toolCallId: 'call-7', input: question }] }],
      });
      return undefined;
    });
    renderSession();

    const inspector = await screen.findByTestId('session-inspector');
    await waitFor(() => expect(JSON.parse(inspector.getAttribute('data-pending') ?? '[]')).toEqual([
      { tool: 'AskUserQuestion', input: question, toolCallId: 'call-7', sessionId: 'session-1' },
    ]));
    expect(fetchMock.mock.calls.map(call => String(call[0]))).toContain('/api/captain/sessions/exec-2');
  });

  it('shows requested then exact resolved spec while a new session is admitted', async () => {
    stubServer(statsRoute({}));
    const { client } = renderSession();
    act(() => client.setQueryData(todoQueryKeys.launch('/repo', 'todo-1'), {
      status: 'preparing', step: 'plan', requestedSpec: { model: 'gpt-5.5' },
    }));
    await waitFor(() => expect(screen.getByText('Preparing session · plan')).toBeTruthy());
    expect(screen.getByText('Requested spec')).toBeTruthy();
    expect(screen.getByText(/"model": "gpt-5.5"/)).toBeTruthy();
    act(() => client.setQueryData(todoQueryKeys.launch('/repo', 'todo-1'), {
      status: 'resolved', step: 'plan', requestedSpec: { model: 'gpt-5.5' },
      spec: { model: 'gpt-5.5', effort: 'high' }, specYaml: 'model: gpt-5.5\neffort: high\n',
    }));
    await waitFor(() => expect(screen.getByText('Starting session · plan')).toBeTruthy());
    expect(screen.getByText('Resolved spec')).toBeTruthy();
    expect(screen.getByText(/effort: high/)).toBeTruthy();
  });

  it('shows a launch session before a never-run todo has a session ID', async () => {
    stubServer(() => undefined, []);
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    client.setQueryData(todoQueryKeys.launch('/repo', 'todo-1'), { status: 'preparing', step: 'run', requestedSpec: { model: 'gpt-5.5' } });
    renderSession({ sessionId: undefined, todo: { ...todo, sessionId: undefined } }, client);

    expect(screen.getByText('Preparing session · run')).toBeTruthy();
    const inspector = screen.getByTestId('session-inspector');
    expect(inspector.getAttribute('data-current')).toBe('pending-launch');
    expect(inspector.getAttribute('data-layout')).toBe('default');
    expect(items(inspector)).toEqual([{ id: 'pending-launch', label: 'Attempt #1', status: 'starting' }]);
  });

  it('reads a session no attempt owns straight from Captain by its provider id', async () => {
    stubServer(statsRoute({}), []);
    renderSession();

    const inspector = await screen.findByTestId('session-inspector');
    expect(inspector.getAttribute('data-src')).toBe('/api/captain/sessions/session-1');
  });

  it('answers the followed session, projects the answered todo into caches and invalidates its session reads', async () => {
    const answered = { ...todo, status: 'in_progress' as const, questions: [] };
    let answerBody: Record<string, unknown> | undefined;
    stubServer((url, init) => {
      if (url.includes('/session/stats')) return Response.json({ ...sessionStats, state: 'working', inProgress: true });
      if (url !== '/api/todos/answer') return undefined;
      answerBody = JSON.parse(init?.body as string);
      return new Response(
        `event: resolved\ndata: ${JSON.stringify({ spec: {}, specYaml: '{}\n', step: 'run' })}\n\nevent: admitted\ndata: ${JSON.stringify({ todo: answered, status: 'resumed', promptRunId: 'prompt-new', sessionId: 'session-1' })}\n\n`,
        { headers: { 'Content-Type': 'text/event-stream' } },
      );
    });
    const { invalidateQueries, onChanged } = renderSession();

    fireEvent.click(await screen.findByRole('button', { name: 'Answer question' }));

    await waitFor(() => expect(onChanged).toHaveBeenCalledWith(answered));
    expect(answerBody).toMatchObject({ dir: '/repo', ref: 'todo-1', sessionId: 'session-1' });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: todoQueryKeys.sessionStats('/repo', 'session-1') });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: todoQueryKeys.sessionDetail('/repo', 'todo-1') });
  });

  it('denies a pending approval with a reason, posting the new approve/deny/respond body', async () => {
    const fetchMock = stubServer((url, init) => {
      if (url.includes('/session/stats')) return Response.json({ ...sessionStats, state: 'approval' });
      if (!url.includes('/session/approve')) return undefined;
      expect(JSON.parse(init?.body as string)).toEqual({ approvalId: 'approval-9', action: 'deny', message: 'Looks risky' });
      return Response.json({ resolved: true });
    });
    renderSession();

    fireEvent.click(await screen.findByRole('button', { name: 'Deny with reason' }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining('/session/approve'),
      expect.objectContaining({ method: 'POST' }),
    ));
  });

  it('routes a rejected question to denial even when a stale answer map is present', async () => {
    const fetchMock = stubServer((url, init) => {
      if (url.includes('/session/stats')) return Response.json({ ...sessionStats, state: 'approval' });
      if (!url.includes('/session/approve')) return undefined;
      expect(JSON.parse(init?.body as string)).toEqual({ approvalId: 'approval-9', action: 'deny' });
      return Response.json({ resolved: true });
    });
    renderSession();

    fireEvent.click(await screen.findByRole('button', { name: 'Reject answered question' }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining('/session/approve'), expect.objectContaining({ method: 'POST' }),
    ));
  });

  it('stops the attempt from its picker row and invalidates the attempt and session caches', async () => {
    stubServer(url => {
      if (url.includes('/session/stats')) return Response.json({ ...sessionStats, state: 'working', inProgress: true });
      if (url.includes('/session/stop')) return Response.json({ status: 'stopping', promptRunId: 'run-2' });
      return undefined;
    });
    const { invalidateQueries } = renderSession();

    fireEvent.click(await screen.findByRole('button', { name: 'Stop attempt #2' }));

    await waitFor(() => expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: todoQueryKeys.sessionDetail('/repo', 'todo-1') }));
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: todoQueryKeys.sessionStats('/repo', 'session-1') });
    expect(screen.queryByRole('button', { name: 'Stop attempt #1' })).toBeNull();
  });
});
