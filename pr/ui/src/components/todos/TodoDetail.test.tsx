import type React from 'react';
import { act, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { RunContext } from './providers';
import type { TodoItem } from '../../types';
import { TodoDetail } from './TodoDetail';
import { useSessionStats } from './TodoSessionTimer';
import { queryTestWrapper } from './queryTestWrapper';

// This test targets only the review/ask guard on Resume/Run/Plan — not full
// TodoDetail coverage — so every tab-gated or unconditionally-rendered sibling
// is stubbed to keep the surface narrow.
vi.mock('./TodoTimeline', () => ({ TodoTimeline: () => null }));
vi.mock('./TodoCommits', () => ({ TodoCommits: () => null }));
vi.mock('./TodoSession', () => ({
  TodoSession: ({ sessionId }: { sessionId?: string }) => <div data-testid="session-viewer">{sessionId}</div>,
}));
vi.mock('./TodoPlan', () => ({ TodoPlan: () => null }));
vi.mock('./TodoVerification', () => ({ TodoVerification: () => null }));
vi.mock('./TodoCompose', () => ({
  TodoTitleEditor: () => null,
  TodoBodyEditor: () => null,
  TodoCommentBox: () => null,
}));
vi.mock('./planActions', () => ({ TodoReviewBanner: () => null }));

vi.mock('./TodoSessionTimer', () => ({ useSessionStats: vi.fn(() => ({ stats: null })) }));

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({
    children,
    variant: _variant,
    size: _size,
    loading: _loading,
    ...props
  }: React.ButtonHTMLAttributes<HTMLButtonElement> & { variant?: string; size?: string; loading?: boolean }) => (
    // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky Button itself.
    <button type="button" {...props}>
      {children}
    </button>
  ),
  ListMenuItem: ({ children, ...props }: { children?: React.ReactNode }) => <div {...props}>{children}</div>,
  DropdownMenu: ({
    trigger,
    children,
  }: {
    trigger: React.ReactNode;
    children: (close: () => void) => React.ReactNode;
  }) => (
    <div>
      {trigger}
      {children(() => {})}
    </div>
  ),
  Combobox: () => null,
  Field: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
  Modal: ({ children, open }: { children?: React.ReactNode; open?: boolean }) => (open === false ? null : <div>{children}</div>),
  SegmentedControl: <T extends string>({
    options,
    onChange,
  }: {
    options: Array<{ id: T; label: string; disabled?: boolean }>;
    onChange: (value: T) => void;
  }) => (
    <div>
      {options.map(option => (
        // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky SegmentedControl itself.
        <button key={option.id} type="button" disabled={option.disabled} onClick={() => onChange(option.id)}>
          {option.label}
        </button>
      ))}
    </div>
  ),
}));

vi.mock('@flanksource/clicky-ui/data', () => ({
  Markdown: ({ text }: { text: string }) => <div>{text}</div>,
}));

vi.mock('@flanksource/clicky-ui/chat', () => ({
  providerIcon: () => (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />,
  ModelSelector: () => null,
  ProviderSelector: () => null,
}));

vi.mock('@flanksource/clicky-ui/ai', () => ({
  effortOptionsForModel: (_model: unknown, fallback: string[]) => fallback,
  PromptRunEditor: () => null,
  promptRuntimeValueToPayload: (value: unknown) => value,
  reconcileModelCapabilities: (value: unknown) => value,
}));

vi.mock('@flanksource/clicky-ui/icons', async (importOriginal) => ({
  ...(await importOriginal<object>()),
}));

const RUN_CONTEXT: RunContext = {
  defaultBackend: 'claude-cmux',
  efforts: ['low', 'medium', 'high', 'xhigh'],
  tools: [],
  backends: [
    {
      id: 'claude-cmux',
      label: 'Claude cmux',
      provider: 'anthropic',
      agent: 'claude',
      defaultModel: 'claude-sonnet-5',
      driver: 'claude-cmux',
      mechanisms: [{ value: 'cmux', label: 'cmux (TUI)', driver: 'claude-cmux' }],
      models: [{ id: 'claude-sonnet-5', provider: 'anthropic', label: 'Claude Sonnet 5', reasoning: true, configured: true }],
      configured: true,
    },
  ],
};

const RESOLVE_MESSAGE = 'Resolve the pending plan review or question first';

const baseTodo: TodoItem = {
  ref: 'todo-1',
  title: 'Ship the feature',
  status: 'pending',
  priority: 'medium',
  sessionId: 'session-1',
};

async function renderDetail(todo: TodoItem) {
  render(
    <TodoDetail
      todo={todo}
      loading={false}
      dir="/repo"
      onChanged={() => {}}
      onDeleted={() => {}}
    />,
    { wrapper: queryTestWrapper() },
  );
  // Flush the run-context fetch each TodoRunActionButton kicks off on mount so
  // its state settles before assertions run (and before the next test's stubs
  // replace the global fetch).
  await act(async () => {});
}

beforeEach(() => {
  vi.stubGlobal('fetch', vi.fn(async () => ({ ok: true, json: async () => RUN_CONTEXT }) as Response));
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
  vi.mocked(useSessionStats).mockReturnValue({ stats: null } as ReturnType<typeof useSessionStats>);
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('TodoDetail Resume/Run/Plan guard', () => {
  it('renders a deep-link database error with a menubar back action', () => {
    render(
      <TodoDetail
        todo={null}
        loading={false}
        loadError="todo reference abc123 is ambiguous"
        dir=""
        onChanged={() => {}}
        onDeleted={() => {}}
        onBack={() => {}}
      />,
      { wrapper: queryTestWrapper() },
    );

    expect(screen.getByRole('alert').textContent).toContain('todo reference abc123 is ambiguous');
    expect(screen.getByRole('button', { name: 'Back to todos' })).toBeTruthy();
  });

  it('opens the exact historical Session tab resolved from a session UUID', async () => {
    const historicalSession = '019f5b29-7890-7c11-8e7a-838e5d373e39';
    await renderDetail({
      ...baseTodo,
      sessionId: '019f5b2e-75b7-7de2-911b-de8b70266479',
      lookupSessionId: historicalSession,
    });

    expect(screen.getByTestId('session-viewer').textContent).toBe(historicalSession);
  });

  it('keeps Resume, Run, and Plan enabled for a pending todo', async () => {
    await renderDetail({ ...baseTodo, status: 'pending' });

    for (const label of ['Resume session', 'Run todo', 'Plan todo']) {
      const item = screen.getByText(label).closest('button');
      expect(item).not.toBeNull();
      expect((item as HTMLButtonElement).disabled).toBe(false);
    }
    expect(screen.queryByText(RESOLVE_MESSAGE)).toBeNull();
  });

  it('disables Resume, Run, and Plan while a todo awaits plan review', async () => {
    await renderDetail({ ...baseTodo, status: 'review' });

    for (const label of ['Resume session', 'Run todo', 'Plan todo']) {
      const item = screen.getByText(label).closest('button');
      expect((item as HTMLButtonElement).disabled).toBe(true);
    }
    expect(screen.getAllByText(RESOLVE_MESSAGE).length).toBeGreaterThan(0);
  });

  it('disables Resume, Run, and Plan while a todo awaits an answer', async () => {
    await renderDetail({ ...baseTodo, status: 'ask' });

    for (const label of ['Resume session', 'Run todo', 'Plan todo']) {
      const item = screen.getByText(label).closest('button');
      expect((item as HTMLButtonElement).disabled).toBe(true);
    }
    expect(screen.getAllByText(RESOLVE_MESSAGE).length).toBeGreaterThan(0);
  });

  it('disables Run via the existing sessionInProgress path, not the review/ask copy, while a session is live', async () => {
    vi.mocked(useSessionStats).mockReturnValue({ stats: { inProgress: true, found: true } } as ReturnType<typeof useSessionStats>);

    await renderDetail({ ...baseTodo, status: 'in_progress' });

    const runItem = screen.getByText('Stop unavailable').closest('button');
    expect((runItem as HTMLButtonElement).disabled).toBe(true);
    expect(screen.getByText('Session interrupt is not supported yet')).toBeTruthy();
    expect(screen.queryByText(RESOLVE_MESSAGE)).toBeNull();
  });
});

describe('TodoDetail verification badge', () => {
  function stubAttempts(attempts: Record<string, unknown>[]) {
    vi.stubGlobal(
      'fetch',
      vi.fn(async (input: RequestInfo | URL) => {
        if (String(input).startsWith('/api/todos/session/detail')) {
          // A real Response: fetchTodoSessionDetail clones it before parsing.
          return new Response(JSON.stringify({ attempts, diagnostics: [], attemptsOnly: true }), {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          });
        }
        return { ok: true, json: async () => RUN_CONTEXT } as Response;
      }),
    );
  }

  function attempt(ordinal: number, passed: boolean) {
    return {
      promptRunId: `run-${ordinal}`,
      ordinal,
      step: 'verify',
      requested: {},
      resolved: {},
      status: 'succeeded',
      processActive: false,
      state: 'succeeded',
      phase: 'finished',
      queuedAt: `2026-07-30T09:0${ordinal}:00Z`,
      startedAt: `2026-07-30T09:0${ordinal}:00Z`,
      admissionSessionId: `admission-${ordinal}`,
      createdAt: `2026-07-30T09:0${ordinal}:00Z`,
      updatedAt: `2026-07-30T09:0${ordinal}:30Z`,
      resultJson: { definitionOfDone: { ran: true, passed } },
    };
  }

  function verificationTab(): HTMLElement {
    const tab = screen.getByText('Verification').closest('button');
    if (!tab) throw new Error('Verification tab not rendered');
    return tab;
  }

  // The badge used to count acceptance criteria, so a todo whose last check
  // failed looked identical to one that had never run.
  it('counts attempts and tints red when the latest one failed', async () => {
    stubAttempts([attempt(1, true), attempt(2, false)]);

    await renderDetail({
      ...baseTodo,
      criteria: [
        { text: 'Criterion one', done: true },
        { text: 'Criterion two', done: false },
        { text: 'Criterion three', done: false },
      ],
    });

    await waitFor(() => expect(verificationTab().querySelector('span:last-of-type')?.textContent).toBe('2'));
    const tab = verificationTab();
    expect(tab.querySelector('span:last-of-type')?.className).toContain('bg-red-500/10');
    expect(tab.getAttribute('title')).toBe('Latest verification failed — 1 of 2 attempts failed');
  });

  it('keeps the badge neutral once the latest attempt passes', async () => {
    stubAttempts([attempt(1, false), attempt(2, true)]);

    await renderDetail(baseTodo);

    await waitFor(() => expect(verificationTab().querySelector('span:last-of-type')?.textContent).toBe('2'));
    const tab = verificationTab();
    expect(tab.querySelector('span:last-of-type')?.className).not.toContain('bg-red-500/10');
    expect(tab.getAttribute('title')).toBe('Latest verification passed — 1 of 2 attempts failed');
  });

  it('polls attempts only, so the badge never pays for a provider thread', async () => {
    stubAttempts([]);
    await renderDetail(baseTodo);

    const detailCalls = vi.mocked(fetch).mock.calls.map((call) => String(call[0])).filter((url) => url.startsWith('/api/todos/session/detail'));
    expect(detailCalls.length).toBeGreaterThan(0);
    expect(detailCalls.every((url) => url.includes('attempts=only'))).toBe(true);
    // A todo that has never been verified carries no badge at all.
    expect(verificationTab().querySelector('span:last-of-type')).toBeNull();
  });
});
