import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { AgentSession } from '../../types';
import { queryKeys } from '../../query';
import { postSessionMessage, RUN_START_TIMEOUT_MS, sessionMessageUrl, useSessionMessageComposer } from './sessionChat';

const SESSION_ID = 'sess-1234';
const SESSION_URL = `/api/captain/sessions/${SESSION_ID}`;
const RUN_RESPONSE = {
  runId: 'run-1',
  messageId: 'msg-1',
  status: 'running',
  capabilities: { interrupt: true, steer: false, followUp: true, resume: true, setPermissionMode: false },
};
const ATTACHED_CAPS = { interrupt: true, steer: true, followUp: true, resume: false, setPermissionMode: true };

function chatPayload(overrides: Record<string, unknown> = {}, caps: Partial<typeof ATTACHED_CAPS> = {}) {
  const capabilities = { ...ATTACHED_CAPS, ...caps };
  return {
    runId: 'run-9',
    capabilities,
    state: {
      runId: 'run-9',
      status: 'running',
      turn: 2,
      capabilities,
      queued: [{ messageId: 'q-1', text: 'then run the linter' }],
      permissionMode: 'default',
      permissionModes: ['default', 'plan'],
    },
    permissionMode: 'default',
    permissionModes: ['default', 'plan'],
    ...overrides,
  };
}

function session(overrides: Partial<AgentSession> = {}): AgentSession {
  return {
    id: SESSION_ID,
    title: 'Refactor parser',
    source: 'claude',
    lifecycleStatus: 'succeeded',
    activityState: 'idle',
    processActive: false,
    lastActivityAt: '2026-10-06T12:00:00Z',
    tokens: { input: 0, output: 0, cacheRead: 0, total: 0 },
    costUsd: 0,
    ...overrides,
  };
}

function jsonResponse(body: unknown, status = 202) {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });
}

type Route = (init?: RequestInit) => Response | Promise<Response>;

let fetchMock: ReturnType<typeof vi.fn>;
let queryClient: QueryClient;

// Routes the session's Captain URLs; GET /chat is a 404 (no attached chat)
// unless a test gives it a route, and any other unrouted call throws.
function routeFetch(routes: Partial<Record<'chat' | 'message' | 'interrupt' | 'permissionMode', Route>>) {
  const paths: Record<string, Route | undefined> = {
    '/chat': routes.chat ?? (() => jsonResponse({ error: `session ${SESSION_ID} has no live chat in this process` }, 404)),
    '/message': routes.message,
    '/interrupt': routes.interrupt,
    '/permission-mode': routes.permissionMode,
  };
  fetchMock.mockImplementation(async (url: string, init?: RequestInit) => {
    const route = paths[url.replace(SESSION_URL, '')];
    if (!url.startsWith(SESSION_URL) || !route) throw new Error(`unexpected fetch ${url}`);
    return route(init);
  });
}

const callsTo = (path: string) => fetchMock.mock.calls.filter(([url]) => url === `${SESSION_URL}${path}`) as [string, RequestInit | undefined][];

function wrapper({ children }: { children: ReactNode }) {
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}

beforeEach(() => {
  fetchMock = vi.fn();
  vi.stubGlobal('fetch', fetchMock);
  queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  routeFetch({});
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.useRealTimers();
});

describe('postSessionMessage', () => {
  it('POSTs {text, messageId} to the Captain session message route with the id URL-encoded', async () => {
    fetchMock.mockResolvedValue(jsonResponse(RUN_RESPONSE));

    const response = await postSessionMessage({ sessionId: 'a/b c', text: 'keep going', messageId: 'm-9' });

    expect(sessionMessageUrl('a/b c')).toBe('/api/captain/sessions/a%2Fb%20c/message');
    expect(fetchMock).toHaveBeenCalledWith('/api/captain/sessions/a%2Fb%20c/message', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text: 'keep going', messageId: 'm-9' }),
    });
    expect(response).toEqual(RUN_RESPONSE);
  });

  it('surfaces the server error text instead of swallowing it', async () => {
    fetchMock.mockResolvedValue(new Response('session id is ambiguous', { status: 409 }));

    await expect(postSessionMessage({ sessionId: SESSION_ID, text: 'hi', messageId: 'm' }))
      .rejects.toThrow('Send session message: session id is ambiguous');
  });

  it.each([
    ['a missing runId', { ...RUN_RESPONSE, runId: undefined }],
    ['missing capabilities', { ...RUN_RESPONSE, capabilities: undefined }],
    ['non-boolean capability flags', { ...RUN_RESPONSE, capabilities: { ...RUN_RESPONSE.capabilities, steer: 'yes' } }],
  ])('rejects a malformed response with %s', async (_label, body) => {
    fetchMock.mockResolvedValue(jsonResponse(body));

    await expect(postSessionMessage({ sessionId: SESSION_ID, text: 'hi', messageId: 'm' }))
      .rejects.toThrow('Send session message: invalid response');
  });
});

describe('useSessionMessageComposer', () => {
  it('is idle and resumable for a finished session, and running for a live one', () => {
    const finished = renderHook(() => useSessionMessageComposer(session()), { wrapper });
    expect(finished.result.current).toMatchObject({
      status: 'idle',
      follow: false,
      error: undefined,
      capabilities: { resume: true, followUp: false, steer: false },
    });

    const live = renderHook(() => useSessionMessageComposer(session({ processActive: true })), { wrapper });
    expect(live.result.current).toMatchObject({ status: 'running', follow: true, capabilities: { resume: false, followUp: true } });
  });

  it('posts the text, adopts the response capabilities, keeps following, and refreshes the sessions list', async () => {
    routeFetch({ message: () => jsonResponse(RUN_RESPONSE) });
    const invalidate = vi.spyOn(queryClient, 'invalidateQueries');
    const { result } = renderHook(() => useSessionMessageComposer(session()), { wrapper });

    act(() => result.current.send('please continue'));

    await waitFor(() => expect(result.current.follow).toBe(true));
    expect(JSON.parse(callsTo('/message')[0]![1]!.body as string)).toMatchObject({ text: 'please continue' });
    expect(result.current.capabilities).toEqual(RUN_RESPONSE.capabilities);
    expect(result.current.error).toBeUndefined();
    expect(invalidate).toHaveBeenCalledWith({ queryKey: queryKeys.agentSessions() });
  });

  it('reports the server error through `error` and does not start following', async () => {
    routeFetch({ message: () => new Response('unknown session', { status: 404 }) });
    const { result } = renderHook(() => useSessionMessageComposer(session()), { wrapper });

    act(() => result.current.send('hello'));

    await waitFor(() => expect(result.current.error).toBe('Send session message: unknown session'));
    expect(result.current.follow).toBe(false);
  });

  it('stops following once the run it started has been seen active and then ended', async () => {
    routeFetch({ message: () => jsonResponse(RUN_RESPONSE) });
    let live = false;
    const { result, rerender } = renderHook(() => useSessionMessageComposer(session({ processActive: live })), { wrapper });

    act(() => result.current.send('go'));
    await waitFor(() => expect(result.current.follow).toBe(true));

    live = true;
    rerender();
    expect(result.current).toMatchObject({ status: 'running', follow: true });

    live = false;
    rerender();
    await waitFor(() => expect(result.current.follow).toBe(false));
    expect(result.current).toMatchObject({ status: 'idle', capabilities: { resume: true, followUp: false } });
  });

  it('polls the sessions list while following a run it started', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    routeFetch({ message: () => jsonResponse(RUN_RESPONSE) });
    const { result } = renderHook(() => useSessionMessageComposer(session()), { wrapper });
    act(() => result.current.send('go'));
    await waitFor(() => expect(result.current.follow).toBe(true));
    const invalidate = vi.spyOn(queryClient, 'invalidateQueries');

    await act(async () => { await vi.advanceTimersByTimeAsync(3_100); });

    expect(invalidate).toHaveBeenCalledWith({ queryKey: queryKeys.agentSessions() });
  });

  describe('when the resumed run never appears live', () => {
    async function sendAndFollow(options?: Parameters<typeof useSessionMessageComposer>[1]) {
      vi.useFakeTimers({ shouldAdvanceTime: true });
      routeFetch({ message: () => jsonResponse(RUN_RESPONSE) });
      const hook = renderHook(() => useSessionMessageComposer(session(), options), { wrapper });
      act(() => hook.result.current.send('go'));
      await waitFor(() => expect(hook.result.current.follow).toBe(true));
      return hook;
    }

    it('stops following after the start window and says the run did not start', async () => {
      const { result } = await sendAndFollow({ startTimeoutMs: 5_000 });

      await act(async () => { await vi.advanceTimersByTimeAsync(4_900); });
      expect(result.current).toMatchObject({ follow: true, error: undefined });

      await act(async () => { await vi.advanceTimersByTimeAsync(200); });
      expect(result.current).toMatchObject({
        follow: false,
        status: 'idle',
        error: 'Resumed run did not start within 5s',
        capabilities: { resume: true, followUp: false },
      });
    });

    it('defaults the start window to 60s', async () => {
      expect(RUN_START_TIMEOUT_MS).toBe(60_000);
      const { result } = await sendAndFollow();

      await act(async () => { await vi.advanceTimersByTimeAsync(59_000); });
      expect(result.current.follow).toBe(true);

      await act(async () => { await vi.advanceTimersByTimeAsync(2_000); });
      expect(result.current).toMatchObject({ follow: false, error: 'Resumed run did not start within 60s' });
    });

    it('does not time out once the process has been seen live', async () => {
      vi.useFakeTimers({ shouldAdvanceTime: true });
      routeFetch({ message: () => jsonResponse(RUN_RESPONSE) });
      let live = false;
      const { result, rerender } = renderHook(
        () => useSessionMessageComposer(session({ processActive: live }), { startTimeoutMs: 5_000 }),
        { wrapper },
      );
      act(() => result.current.send('go'));
      await waitFor(() => expect(result.current.follow).toBe(true));

      await act(async () => { await vi.advanceTimersByTimeAsync(2_000); });
      live = true;
      rerender();
      await act(async () => { await vi.advanceTimersByTimeAsync(10_000); });

      expect(result.current).toMatchObject({ follow: true, status: 'running', error: undefined });
    });

    it('clears the timeout error when another message is sent', async () => {
      const { result } = await sendAndFollow({ startTimeoutMs: 1_000 });
      await act(async () => { await vi.advanceTimersByTimeAsync(1_100); });
      expect(result.current.error).toBe('Resumed run did not start within 1s');

      act(() => result.current.send('again'));

      await waitFor(() => expect(result.current).toMatchObject({ follow: true, error: undefined }));
    });
  });

  describe('Captain live chat', () => {
    it('keeps the derived controls when no chat is attached (404) and offers no interrupt or permission picker', async () => {
      const { result } = renderHook(() => useSessionMessageComposer(session({ processActive: true })), { wrapper });

      await waitFor(() => expect(callsTo('/chat')).toHaveLength(1));
      await act(async () => { await Promise.resolve(); });

      expect(result.current).toMatchObject({
        status: 'running',
        error: undefined,
        queued: [],
        capabilities: { followUp: true, steer: true, resume: false },
        onInterrupt: undefined,
        permissionMode: undefined,
        permissionModes: undefined,
        onPermissionModeChange: undefined,
      });
    });

    it('adopts an attached chat: its capabilities, status, queue, and the interrupt and permission controls', async () => {
      routeFetch({ chat: () => jsonResponse(chatPayload(), 200) });
      const { result } = renderHook(() => useSessionMessageComposer(session({ processActive: true })), { wrapper });

      await waitFor(() => expect(result.current.capabilities).toEqual(ATTACHED_CAPS));

      expect(result.current).toMatchObject({
        status: 'running',
        queued: [{ messageId: 'q-1', text: 'then run the linter' }],
        permissionMode: 'default',
        permissionModes: ['default', 'plan'],
        permissionFamily: 'claude',
      });
      expect(result.current.onInterrupt).toBeTypeOf('function');
      expect(result.current.onPermissionModeChange).toBeTypeOf('function');
    });

    it('withholds interrupt and the permission picker the attached chat cannot honour', async () => {
      routeFetch({ chat: () => jsonResponse(chatPayload({}, { interrupt: false, setPermissionMode: false }), 200) });
      const { result } = renderHook(() => useSessionMessageComposer(session({ processActive: true })), { wrapper });

      await waitFor(() => expect(result.current.capabilities).toMatchObject({ interrupt: false, setPermissionMode: false }));

      expect(result.current).toMatchObject({
        onInterrupt: undefined,
        permissionMode: undefined,
        permissionModes: undefined,
        onPermissionModeChange: undefined,
      });
    });

    it('surfaces a non-404 chat failure instead of treating it as no chat', async () => {
      routeFetch({ chat: () => jsonResponse({ error: 'chat broker exploded' }, 500) });
      const { result } = renderHook(() => useSessionMessageComposer(session()), { wrapper });

      await waitFor(() => expect(result.current.error).toBe('Load session chat: chat broker exploded'));
    });

    it.each([
      ['an unknown status', chatPayload({ state: { status: 'dancing' } })],
      ['a non-boolean capability flag', chatPayload({ capabilities: { ...ATTACHED_CAPS, interrupt: 'yes' } })],
      ['an unknown permission mode', chatPayload({ permissionMode: 'yolo' })],
    ])('rejects a malformed chat with %s', async (_label, body) => {
      routeFetch({ chat: () => jsonResponse(body, 200) });
      const { result } = renderHook(() => useSessionMessageComposer(session()), { wrapper });

      await waitFor(() => expect(result.current.error).toMatch(/^Load session chat: invalid response/));
    });

    it('re-reads the chat after a send and while following a run', async () => {
      vi.useFakeTimers({ shouldAdvanceTime: true });
      routeFetch({ message: () => jsonResponse(RUN_RESPONSE) });
      const { result } = renderHook(() => useSessionMessageComposer(session()), { wrapper });
      await waitFor(() => expect(callsTo('/chat')).toHaveLength(1));

      act(() => result.current.send('go'));
      await waitFor(() => expect(callsTo('/chat')).toHaveLength(2));

      await act(async () => { await vi.advanceTimersByTimeAsync(3_100); });
      expect(callsTo('/chat').length).toBeGreaterThanOrEqual(3);
    });

    it('does not poll the chat for an idle session', async () => {
      vi.useFakeTimers({ shouldAdvanceTime: true });
      renderHook(() => useSessionMessageComposer(session()), { wrapper });
      await waitFor(() => expect(callsTo('/chat')).toHaveLength(1));

      await act(async () => { await vi.advanceTimersByTimeAsync(10_000); });

      expect(callsTo('/chat')).toHaveLength(1);
    });

    it('interrupts with a body-less POST, then re-reads the chat', async () => {
      routeFetch({
        chat: () => jsonResponse(chatPayload(), 200),
        interrupt: () => jsonResponse({ status: 'interrupting', discardedMessageIds: ['q-1'] }),
      });
      const { result } = renderHook(() => useSessionMessageComposer(session({ processActive: true })), { wrapper });
      await waitFor(() => expect(result.current.onInterrupt).toBeTypeOf('function'));
      const chatReads = callsTo('/chat').length;

      act(() => result.current.onInterrupt!());

      await waitFor(() => expect(callsTo('/interrupt')).toHaveLength(1));
      expect(callsTo('/interrupt')[0]![1]).toEqual({ method: 'POST' });
      await waitFor(() => expect(callsTo('/chat').length).toBeGreaterThan(chatReads));
      expect(result.current.error).toBeUndefined();
    });

    it('shows the 409 when the run cannot be interrupted', async () => {
      routeFetch({
        chat: () => jsonResponse(chatPayload(), 200),
        interrupt: () => jsonResponse({ error: 'run cannot be interrupted' }, 409),
      });
      const { result } = renderHook(() => useSessionMessageComposer(session({ processActive: true })), { wrapper });
      await waitFor(() => expect(result.current.onInterrupt).toBeTypeOf('function'));

      act(() => result.current.onInterrupt!());

      await waitFor(() => expect(result.current.error).toBe('Interrupt session: run cannot be interrupted'));
    });

    it('switches the permission mode with {mode} and re-reads the chat', async () => {
      routeFetch({
        chat: () => jsonResponse(chatPayload(), 200),
        permissionMode: () => jsonResponse({ runId: 'run-9', permissionMode: 'plan' }),
      });
      const { result } = renderHook(() => useSessionMessageComposer(session({ processActive: true })), { wrapper });
      await waitFor(() => expect(result.current.onPermissionModeChange).toBeTypeOf('function'));
      const chatReads = callsTo('/chat').length;

      act(() => result.current.onPermissionModeChange!('plan'));

      await waitFor(() => expect(callsTo('/permission-mode')).toHaveLength(1));
      const [, init] = callsTo('/permission-mode')[0]!;
      expect(init).toMatchObject({ method: 'POST', body: JSON.stringify({ mode: 'plan' }) });
      await waitFor(() => expect(callsTo('/chat').length).toBeGreaterThan(chatReads));
      expect(result.current.error).toBeUndefined();
    });

    it('shows the 422 when the runtime does not support the mode', async () => {
      routeFetch({
        chat: () => jsonResponse(chatPayload(), 200),
        permissionMode: () => jsonResponse({ error: 'runtime does not support permission mode "plan"' }, 422),
      });
      const { result } = renderHook(() => useSessionMessageComposer(session({ processActive: true })), { wrapper });
      await waitFor(() => expect(result.current.onPermissionModeChange).toBeTypeOf('function'));

      act(() => result.current.onPermissionModeChange!('plan'));

      await waitFor(() => expect(result.current.error).toBe('Change session permission mode: runtime does not support permission mode "plan"'));
    });
  });
});
