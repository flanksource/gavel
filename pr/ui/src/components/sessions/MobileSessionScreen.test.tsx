import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import type { ReactElement, ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { SessionChatComposer, type SessionInspectorLoadState } from '@flanksource/clicky-ui/ai';
import type { AgentSession, AgentSessionsResponse } from '../../types';
import { MobileSessionScreen } from './MobileSessionScreen';

type InspectorProps = Record<string, unknown> & {
  composer?: ReactNode;
  onLoadStateChange?: (state: SessionInspectorLoadState) => void;
};

const inspector = vi.hoisted(() => ({ props: null as null | Record<string, unknown> }));

vi.mock('@flanksource/clicky-ui/ai', async importOriginal => ({
  ...await importOriginal<typeof import('@flanksource/clicky-ui/ai')>(),
  SessionInspector: (props: InspectorProps) => {
    inspector.props = props;
    return <div data-testid="inspector">{props.composer}</div>;
  },
}));

function reportLoadState(state: SessionInspectorLoadState) {
  act(() => (inspector.props as InspectorProps).onLoadStateChange!(state));
}

function session(overrides: Partial<AgentSession> = {}): AgentSession {
  return {
    id: 'sess-1',
    title: 'Refactor parser',
    source: 'claude',
    project: 'gavel',
    lifecycleStatus: 'succeeded',
    activityState: 'idle',
    processActive: false,
    lastActivityAt: '2026-10-06T12:00:00Z',
    tokens: { input: 0, output: 0, cacheRead: 0, total: 0 },
    costUsd: 0,
    ...overrides,
  };
}

function query(sessions: AgentSession[]): { data: AgentSessionsResponse; error: null } {
  return { data: { sessions, errors: [] } as AgentSessionsResponse, error: null };
}

function renderScreen(props: Partial<Parameters<typeof MobileSessionScreen>[0]> & { sessions?: AgentSession[] } = {}) {
  const onBack = vi.fn();
  const { sessions = [session()], ...rest } = props;
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  render(
    <QueryClientProvider client={queryClient}>
      <MobileSessionScreen query={query(sessions)} sessionId="sess-1" onBack={onBack} {...rest} />
    </QueryClientProvider>,
  );
  return { onBack };
}

const MESSAGE_URL = '/api/captain/sessions/sess-1/message';
const COMPOSER_PLACEHOLDER = 'Continue this session…';
const LOADING_REASON = 'Loading session…';
const UNREADABLE_REASON = 'This session cannot be read';

// The composer is the inspector's `composer` element; the reason is the prop it
// was given, which clicky-ui's SessionChatComposer turns into a disabled input.
function composerDisabledReason(): string | undefined {
  const composer = inspector.props?.composer as ReactElement<{ disabledReason?: string }>;
  return composer.props.disabledReason;
}

const SESSION_URL = '/api/captain/sessions/sess-1';

let fetchMock: ReturnType<typeof vi.fn>;

function jsonResponse(body: unknown, status: number) {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });
}

// GET /chat is a 404 (no attached chat) unless a test routes it; anything else
// unrouted throws. Each call builds a fresh Response (a body reads once).
function routeFetch(routes: Partial<Record<'/chat' | '/message' | '/permission-mode', (init?: RequestInit) => Response>> = {}) {
  fetchMock.mockImplementation(async (url: string, init?: RequestInit) => {
    const route = routes[url.replace(SESSION_URL, '') as '/chat'] ?? (url === `${SESSION_URL}/chat` ? () => jsonResponse({ error: 'no live chat' }, 404) : undefined);
    if (!route) throw new Error(`unexpected fetch ${url}`);
    return route(init);
  });
}

beforeEach(() => {
  inspector.props = null;
  fetchMock = vi.fn();
  routeFetch();
  vi.stubGlobal('fetch', fetchMock);
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('MobileSessionScreen', () => {
  it('shows the title behind a back button and inspects the Captain session compactly', () => {
    const { onBack } = renderScreen();

    expect(screen.getByText('Refactor parser')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Back to sessions' }));
    expect(onBack).toHaveBeenCalledOnce();
    expect(inspector.props).toMatchObject({ src: '/api/captain/sessions/sess-1', layout: 'compact', follow: false });
  });

  it('follows a live session', () => {
    renderScreen({ sessions: [session({ processActive: true, activityState: 'thinking' })] });

    expect(inspector.props).toMatchObject({ follow: true });
  });

  it('always hands the composer to the inspector slot, disabled with the reason until the session has loaded', () => {
    renderScreen();

    expect(composerDisabledReason()).toBe(LOADING_REASON);
    expect(screen.getByTestId('inspector').contains(screen.getByPlaceholderText(COMPOSER_PLACEHOLDER))).toBe(true);

    reportLoadState({ status: 'loading' });
    expect(composerDisabledReason()).toBe(LOADING_REASON);

    reportLoadState({ status: 'ready' });
    expect(composerDisabledReason()).toBeUndefined();
    expect(screen.getByTestId('inspector').contains(screen.getByPlaceholderText(COMPOSER_PLACEHOLDER))).toBe(true);
  });

  it('adds no wrapper chrome of its own around the composer: the inspector slot pins and pads it', () => {
    renderScreen();

    expect(inspector.props?.composer).toMatchObject({ type: SessionChatComposer });
    expect(screen.queryByRole('status')).toBeNull();
  });

  it('renders a composer that posts to the session message route and keeps following the run it starts', async () => {
    routeFetch({
      '/message': () => jsonResponse({
        runId: 'run-1',
        messageId: 'm-1',
        status: 'running',
        capabilities: { interrupt: true, steer: false, followUp: true, resume: true },
      }, 202),
    });
    renderScreen();
    reportLoadState({ status: 'ready' });

    const input = screen.getByPlaceholderText(COMPOSER_PLACEHOLDER);
    fireEvent.change(input, { target: { value: 'run the tests' } });
    fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => expect(inspector.props).toMatchObject({ follow: true }));
    const [, init] = fetchMock.mock.calls.find(([url]) => url === MESSAGE_URL) as [string, RequestInit];
    expect(JSON.parse(init.body as string)).toMatchObject({ text: 'run the tests' });
  });

  it('shows the live chat queue and a permission-mode picker that posts the chosen {mode}', async () => {
    const capabilities = { interrupt: true, steer: true, followUp: true, resume: false, setPermissionMode: true };
    routeFetch({
      '/chat': () => jsonResponse({
        runId: 'run-9',
        capabilities,
        state: { runId: 'run-9', status: 'running', turn: 1, capabilities, queued: [{ messageId: 'q-1', text: 'then run the linter' }] },
        permissionMode: 'default',
        permissionModes: ['default', 'plan'],
      }, 200),
      '/permission-mode': () => jsonResponse({ runId: 'run-9', permissionMode: 'plan' }, 202),
    });
    renderScreen({ sessions: [session({ processActive: true })] });
    reportLoadState({ status: 'ready' });

    expect(await screen.findByText('Queued: then run the linter')).toBeTruthy();
    fireEvent.change(screen.getByLabelText('Permission mode'), { target: { value: 'plan' } });

    await waitFor(() => expect(fetchMock.mock.calls.some(([url]) => url === `${SESSION_URL}/permission-mode`)).toBe(true));
    const [, init] = fetchMock.mock.calls.find(([url]) => url === `${SESSION_URL}/permission-mode`) as [string, RequestInit];
    expect(JSON.parse(init.body as string)).toEqual({ mode: 'plan' });
  });

  it('shows no permission picker when no chat is attached to the session', async () => {
    renderScreen({ sessions: [session({ processActive: true })] });
    reportLoadState({ status: 'ready' });
    await waitFor(() => expect(fetchMock.mock.calls.some(([url]) => url === `${SESSION_URL}/chat`)).toBe(true));

    expect(screen.queryByLabelText('Permission mode')).toBeNull();
    expect(screen.queryByText(/^Queued:/)).toBeNull();
  });

  it('shows the server rejection in the composer', async () => {
    routeFetch({ '/message': () => new Response('session id is ambiguous', { status: 409 }) });
    renderScreen();
    reportLoadState({ status: 'ready' });

    const input = screen.getByPlaceholderText(COMPOSER_PLACEHOLDER);
    fireEvent.change(input, { target: { value: 'hello' } });
    fireEvent.keyDown(input, { key: 'Enter' });

    expect(await screen.findByText('Send session message: session id is ambiguous')).toBeTruthy();
  });

  it('keeps the inspector and the composer mounted, disabled with the reason, when the session cannot be read', () => {
    renderScreen();

    reportLoadState({ status: 'error', error: new Error('Failed to load session (HTTP 404): session sess-1 has no transcript') });

    expect(screen.getByTestId('inspector').contains(screen.getByPlaceholderText(COMPOSER_PLACEHOLDER))).toBe(true);
    expect(composerDisabledReason()).toBe(UNREADABLE_REASON);
    expect(fetchMock.mock.calls.some(([url]) => url === MESSAGE_URL)).toBe(false);
  });

  it('disables the composer again, keeping it mounted, when the follow stream later reports an error', () => {
    renderScreen();
    reportLoadState({ status: 'ready' });
    expect(composerDisabledReason()).toBeUndefined();

    reportLoadState({ status: 'error', error: new Error('stream failed') });

    expect(screen.getByPlaceholderText(COMPOSER_PLACEHOLDER)).toBeTruthy();
    expect(composerDisabledReason()).toBe(UNREADABLE_REASON);
  });

  it('says so, with a way back, when the session is no longer in the recent list', () => {
    const { onBack } = renderScreen({ sessions: [], sessionId: 'gone' });

    expect(screen.getByText('That session is no longer in the recent list')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Back to sessions' }));
    expect(onBack).toHaveBeenCalledOnce();
    expect(inspector.props).toBeNull();
  });

  it('shows the load error instead of an empty screen', () => {
    renderScreen({ query: { data: undefined, error: new Error('Load agent sessions: HTTP 500') } });

    expect(screen.getByRole('alert').textContent).toContain('Load agent sessions: HTTP 500');
  });
});
