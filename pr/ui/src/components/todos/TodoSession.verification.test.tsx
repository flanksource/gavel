import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { VerifyReport } from '@flanksource/clicky-ui/data';
import type { TodoItem, TodoSessionAttempt } from '../../types';
import { TodoSession } from './TodoSession';
import { queryTestWrapper } from './queryTestWrapper';

vi.mock('@flanksource/clicky-ui/ai', async importOriginal => ({
  ...(await importOriginal<object>()),
  SessionInspector: () => <div data-testid="session-inspector" />,
}));

afterEach(() => vi.unstubAllGlobals());

describe('TODO session verification failure', () => {
  const reason = '10 definition-of-done checks: 10 failed';
  const report: VerifyReport = {
    kind: 'todo', ran: true, passed: false, state: 'failed', reason,
    summary: { total: 10, passed: 0, failed: 10, warned: 0, skipped: 0, pending: 0, running: 0, timedout: 0 },
  };
  const todo: TodoItem = { ref: 'todo-1', title: 'Failed checks', status: 'unverified', priority: 'medium', sessionId: 'provider-session', executionState: 'verification_failed' };
  const attempt: TodoSessionAttempt = {
    promptRunId: 'failed-run', admissionSessionId: 'admission', providerSessionId: todo.sessionId,
    ordinal: 1, step: 'run', requested: {}, resolved: {}, status: 'failed', processActive: false,
    state: 'failed', phase: 'verify', queuedAt: '2026-10-11T10:00:00Z',
    createdAt: '2026-10-11T10:00:00Z', updatedAt: '2026-10-11T10:01:00Z', verification: report,
  };

  function stubStats(stats: Record<string, unknown>, status = 200) {
    vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.startsWith('/api/todos/session/detail')) return Response.json({ attempts: [attempt] });
      if (url.startsWith('/api/todos/session/stats')) return Response.json({ found: true, inProgress: false, ...stats }, { status });
      throw new Error(`Unexpected request: ${url}`);
    }));
  }

  it('shows verification failure and opens its results without a session error', async () => {
    stubStats({ state: 'verification_failed', verification: report });
    const onVerification = vi.fn();
    render(<TodoSession dir="/repo" sessionId={todo.sessionId} active todo={todo} onVerification={onVerification} />, { wrapper: queryTestWrapper() });
    expect(await screen.findByText('Verification failed')).toBeTruthy();
    expect(screen.getByText(reason)).toBeTruthy();
    expect(screen.getByRole('status').className).toContain('[[data-theme=dark]_&]:text-red-400');
    expect(screen.queryByText('Session error')).toBeNull();
    expect(screen.queryByText('Session status')).toBeNull();
    expect(onVerification).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole('button', { name: 'View verification' }));
    expect(onVerification).toHaveBeenCalledOnce();
  });

  it('reads the status of an attempt without a transcript through its admission session', async () => {
    const admissionOnly: TodoSessionAttempt = { ...attempt, providerSessionId: undefined };
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.startsWith('/api/todos/session/detail')) return Response.json({ attempts: [admissionOnly] });
      if (url.startsWith('/api/todos/session/stats')) return Response.json({ found: false, inProgress: false, state: 'verification_failed', verification: report });
      throw new Error(`Unexpected request: ${url}`);
    });
    vi.stubGlobal('fetch', fetchMock);
    render(<TodoSession dir="/repo" active todo={{ ...todo, sessionId: undefined }} />, { wrapper: queryTestWrapper() });
    expect(await screen.findByText('Verification failed')).toBeTruthy();
    expect(fetchMock.mock.calls.map(([input]) => String(input))).toContainEqual(expect.stringContaining('sessionId=admission'));
  });

  it.each([
    { state: 'completed', verification: undefined, error: undefined },
    { state: 'error', verification: report, error: 'Agent connection failed' },
  ])('keeps $state independent of a failed report', async stats => {
    stubStats(stats);
    render(<TodoSession dir="/repo" sessionId={todo.sessionId} active todo={todo} />, { wrapper: queryTestWrapper() });
    await screen.findByTestId('session-inspector');
    await waitFor(() => expect(screen.queryByText('Session error') !== null).toBe(stats.state === 'error'));
    expect(screen.queryByText('Verification failed')).toBeNull();
  });

  it('keeps failed stats requests visible as session errors', async () => {
    stubStats({ error: 'Session database unavailable' }, 500);
    render(<TodoSession dir="/repo" sessionId={todo.sessionId} active todo={todo} />, { wrapper: queryTestWrapper() });
    expect(await screen.findByText('Session error')).toBeTruthy();
    expect(screen.queryByText('Verification failed')).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Show details' }));
    expect(screen.getByText(/Session database unavailable/)).toBeTruthy();
  });
});
