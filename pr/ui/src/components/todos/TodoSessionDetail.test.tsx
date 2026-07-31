import type React from 'react';
import { fireEvent, render, renderHook, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { TodoSessionAttempt, TodoSessionDetailResponse, TodoSessionOverview } from '../../types';
import { attemptSessionCollection, inspectorSession, SessionDiagnostics, useTodoSessionDetail } from './TodoSessionDetail';

vi.mock('@flanksource/clicky-ui/ai', () => ({
  SessionInspector: () => <div data-testid="session-inspector" />,
}));

vi.mock('@flanksource/clicky-ui/chat', () => ({
  providerIcon: () => (props: React.SVGProps<SVGSVGElement>) => <svg data-testid="provider-icon" {...props} />,
}));

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, variant: _variant, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement> & { variant?: string }) => (
    // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky Button itself.
    <button type="button" {...props}>
      {children}
    </button>
  ),
}));

vi.mock('@flanksource/clicky-ui/icons', () => {
  const Icon = (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />;
  return {
    UiCheck: Icon,
    UiChevronDown: Icon,
    UiChevronRight: Icon,
    UiCopy: Icon,
    UiError: Icon,
    UiLoader: Icon,
    UiRobotAi: Icon,
    UiStop: Icon,
    UiBatteryChargingVertical: Icon,
    UiBatteryVerticalEmpty: Icon,
    UiBatteryVerticalFull: Icon,
    UiBatteryVerticalHigh: Icon,
    UiBatteryVerticalLow: Icon,
    UiBatteryVerticalMedium: Icon,
    UiWarningTriangle: Icon,
  };
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('TODO session details', () => {
  it('keeps a structured conflict response so attempts and candidates still render', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            attempts: [
              {
                promptRunId: 'run-4',
                ordinal: 4,
                step: 'plan',
                requested: {},
                resolved: {},
                state: 'failed',
                phase: 'finished',
                queuedAt: '2026-07-14T12:00:00Z',
                admissionSessionId: 'admission',
                createdAt: '2026-07-14T12:00:00Z',
              },
            ],
            diagnostics: [
              {
                severity: 'error',
                code: 'ambiguous_transcript_sessions',
                message: 'Two transcripts remain',
              },
            ],
          }),
          { status: 409, headers: { 'Content-Type': 'application/json' } }
        )
      )
    );

    const { result, unmount } = renderHook(() => useTodoSessionDetail('/repo', 'todo-1', 'provider-1', true));
    await waitFor(() => expect(result.current.detail?.attempts).toHaveLength(1));
    expect(result.current.error).toBe('');
    expect(result.current.detail?.diagnostics[0]?.code).toBe('ambiguous_transcript_sessions');
    unmount();
  });

  it('asks for attempts only when the caller does not render a transcript', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ attempts: [], diagnostics: [], attemptsOnly: true }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    );
    vi.stubGlobal('fetch', fetchMock);

    const { result, unmount } = renderHook(() =>
      useTodoSessionDetail('/repo', 'todo-1', undefined, true, { attemptsOnly: true, intervalMs: 15000 })
    );
    await waitFor(() => expect(result.current.detail?.attemptsOnly).toBe(true));
    expect(String(fetchMock.mock.calls[0]?.[0])).toContain('attempts=only');
    unmount();
  });

  it('keeps the loaded attempts when only the poll period changes', async () => {
    const body = {
      attempts: [
        {
          promptRunId: 'run-1',
          ordinal: 1,
          step: 'verify',
          requested: {},
          resolved: {},
          state: 'succeeded',
          phase: 'finished',
          queuedAt: '2026-07-14T12:00:00Z',
          admissionSessionId: 'admission',
          createdAt: '2026-07-14T12:00:00Z',
        },
      ],
      diagnostics: [],
    };
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } }))
    );

    // Opening the Verification tab drops the badge's slow poll to a fast one; the
    // already-listed attempts must survive that switch.
    const { result, rerender, unmount } = renderHook(
      ({ intervalMs }: { intervalMs: number }) =>
        useTodoSessionDetail('/repo', 'todo-1', undefined, true, { attemptsOnly: true, intervalMs }),
      { initialProps: { intervalMs: 15000 } }
    );
    await waitFor(() => expect(result.current.detail?.attempts).toHaveLength(1));

    rerender({ intervalMs: 1500 });
    expect(result.current.detail?.attempts).toHaveLength(1);
    unmount();
  });

  it('maps attempts to generic session collection rows and lazy loaders', async () => {
    const root: TodoSessionOverview = {
      id: 'root',
      source: 'claude',
      hostId: 'local',
      lifecycleStatus: 'succeeded',
      activityState: 'idle',
      healthState: 'healthy',
      provider: 'anthropic',
      backend: 'claude-agent',
      model: 'claude-opus-4-8',
      effort: 'high',
      processActive: false,
      messageCount: 1,
      turnCount: 0,
      inputTokens: 10,
      outputTokens: 5,
      totalTokens: 15,
      costUsd: 0.01,
    };
    const attempts: TodoSessionAttempt[] = [
      {
        promptRunId: 'run-current',
        ordinal: 2,
        step: 'run',
        mode: 'run',
        requested: {},
        resolved: {
          provider: 'anthropic',
          backend: 'claude-agent',
          model: 'claude-opus-4-8',
          effort: 'high',
        },
        provider: 'anthropic',
        backend: 'claude-agent',
        model: 'claude-opus-4-8',
        effort: 'high',
        status: 'completed',
        processActive: false,
        state: 'succeeded',
        phase: 'finished',
        queuedAt: '2026-07-14T12:00:00Z',
        admissionSessionId: 'admission-current',
        executionSessionId: 'root',
        createdAt: '2026-07-14T12:00:00Z',
        updatedAt: '2026-07-14T12:01:00Z',
      },
      {
        promptRunId: 'run-parallel',
        ordinal: 1,
        step: 'run',
        mode: 'api',
        requested: {
          provider: 'openai',
          backend: 'responses',
          model: 'gpt-5',
          effort: 'medium',
        },
        resolved: {},
        provider: 'openai',
        backend: 'responses',
        model: 'gpt-5',
        effort: 'medium',
        status: 'completed',
        pid: 4242,
        processActive: true,
        state: 'succeeded',
        phase: 'finished',
        queuedAt: '2026-07-14T11:00:00Z',
        admissionSessionId: 'admission-parallel',
        executionSessionId: 'parallel',
        createdAt: '2026-07-14T11:00:00Z',
        updatedAt: '2026-07-14T11:02:00Z',
      },
    ];
    const detail: TodoSessionDetailResponse = {
      attempts,
      selectedPromptRunId: 'run-current',
      diagnostics: [],
      thread: {
        id: 'root',
        status: 'completed',
        root,
        sessions: [root],
        turns: [],
        agents: [],
        costs: [],
        messages: [],
        durationMs: 10,
        inputTokens: 10,
        outputTokens: 5,
        totalTokens: 15,
        costUsd: 0.01,
      },
    };
    const loaded = inspectorSession(detail, []);
    if (Array.isArray(loaded) || typeof loaded === 'string') throw new Error('expected unified session');
    const loadAttempt = vi.fn().mockResolvedValue(loaded);

    const collection = attemptSessionCollection(detail, [], loadAttempt);

    expect(collection.currentSessionId).toBe('run-current');
    expect(collection.sessions).toMatchObject([
      {
        id: 'run-current',
        label: 'Attempt #2',
        mode: 'run',
        summary: {
          provider: 'anthropic',
          backend: 'claude-agent',
          model: 'claude-opus-4-8',
          effort: 'high',
          status: 'completed',
          updatedAt: '2026-07-14T12:01:00Z',
        },
      },
      {
        id: 'run-parallel',
        label: 'Attempt #1',
        mode: 'api',
        summary: {
          provider: 'openai',
          backend: 'responses',
          model: 'gpt-5',
          effort: 'medium',
          status: 'completed',
          pid: 4242,
          updatedAt: '2026-07-14T11:02:00Z',
        },
      },
    ]);
    expect(collection.sessions[0]?.session).toBeDefined();
    expect(collection.sessions[1]?.session).toBeUndefined();
    await collection.loadSession?.(collection.sessions[1]!);
    expect(loadAttempt).toHaveBeenCalledWith(attempts[1]);
  });

  it('copies the complete warning details including candidate metadata', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText },
    });
    render(
      <SessionDiagnostics
        diagnostics={[
          {
            severity: 'warning',
            code: 'legacy_session_identity_resolved',
            message: 'Selected the transcript-bearing row',
            details: [
              { id: 'captain-1', hostId: 'local' },
              { id: 'captain-2', hostId: 'MacBook-Pro.local' },
            ],
          },
        ]}
      />
    );

    fireEvent.click(screen.getByRole('button', { name: /Show details/ }));
    fireEvent.click(screen.getByRole('button', { name: 'Copy diagnostic details' }));
    await waitFor(() => expect(writeText).toHaveBeenCalledWith(expect.stringContaining('MacBook-Pro.local')));
  });

  it('nests provider sub-agents under the root for the shared inspector', () => {
    const root: TodoSessionOverview = {
      id: 'root',
      source: 'claude',
      hostId: 'local',
      lifecycleStatus: 'succeeded',
      activityState: 'idle',
      healthState: 'healthy',
      provider: 'anthropic',
      backend: 'claude-cmux',
      model: 'claude-opus-4-8',
      effort: 'high',
      pid: 8342,
      processActive: false,
      messageCount: 2,
      turnCount: 1,
      inputTokens: 10,
      outputTokens: 5,
      totalTokens: 15,
      costUsd: 1,
    };
    const detail: TodoSessionDetailResponse = {
      attempts: [],
      diagnostics: [],
      thread: {
        id: 'root',
        status: 'working',
        root,
        sessions: [root],
        costs: [],
        messages: [],
        durationMs: 10,
        turns: [
          {
            id: 'turn-db-1',
            sessionId: 'root',
            providerTurnId: 'turn-1',
            turnIndex: 1,
            status: 'running',
            model: 'claude-opus-4-8',
            backend: 'claude-cmux',
            effort: 'high',
            inputTokens: 10,
            outputTokens: 5,
            reasoningTokens: 0,
            cacheReadTokens: 0,
            cacheWriteTokens: 0,
            totalTokens: 15,
            costUsd: 1,
            messageCount: 2,
          },
        ],
        inputTokens: 10,
        outputTokens: 5,
        totalTokens: 15,
        costUsd: 1,
        agents: [
          {
            id: 'root',
            sessionId: 'root',
            isRoot: true,
            source: 'claude',
            lifecycleStatus: 'succeeded',
            activityState: 'idle',
            healthState: 'healthy',
            inputTokens: 10,
            outputTokens: 5,
            reasoningTokens: 0,
            cacheReadTokens: 0,
            cacheWriteTokens: 0,
            totalTokens: 15,
            costUsd: 1,
          },
          {
            id: 'child',
            sessionId: 'child',
            parentSessionId: 'root',
            isRoot: false,
            agentType: 'Explore',
            description: 'Inspect turn parsing',
            source: 'claude',
            lifecycleStatus: 'succeeded',
            activityState: 'idle',
            healthState: 'healthy',
            inputTokens: 3,
            outputTokens: 2,
            reasoningTokens: 0,
            cacheReadTokens: 0,
            cacheWriteTokens: 0,
            totalTokens: 5,
            costUsd: 0.2,
          },
        ],
      },
    };

    expect(inspectorSession(detail, [])).toMatchObject({
      provider: 'anthropic',
      backend: 'claude-cmux',
      model: 'claude-opus-4-8',
      reasoningEffort: 'high',
      live: { active: true, pid: 8342 },
      turns: [
        {
          id: 'turn-db-1',
          backend: 'claude-cmux',
          reasoningEffort: 'high',
          status: 'running',
        },
      ],
      root: {
        id: 'root',
        children: [
          {
            id: 'child',
            parentId: 'root',
            type: 'Explore',
            desc: 'Inspect turn parsing',
          },
        ],
      },
    });
  });
});
