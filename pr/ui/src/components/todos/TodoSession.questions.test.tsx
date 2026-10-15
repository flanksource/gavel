import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, expect, it, vi } from 'vitest';
import type { TodoItem, TodoSessionAttempt } from '../../types';
import { TodoSession } from './TodoSession';

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

it('answers an envelope question in the real Transcript after failed verification', async () => {
  const question = { text: 'Which database?', context: 'Choose the staging database.', options: ['Postgres', 'SQLite'] };
  const todo: TodoItem = {
    ref: 'ask-todo', title: 'Choose the database', status: 'ask', priority: 'medium',
    sessionId: 'ask-session', questions: [question],
  };
  const attempt: TodoSessionAttempt = {
    promptRunId: 'ask-run', ordinal: 1, step: 'run', requested: {}, resolved: {},
    status: 'ask', processActive: false, state: 'waiting', phase: 'generate',
    admissionSessionId: 'ask-admission', providerSessionId: todo.sessionId,
    queuedAt: '2026-10-14T10:00:00Z', createdAt: '2026-10-14T10:00:00Z', updatedAt: '2026-10-14T10:01:00Z',
    verification: {
      kind: 'fixture', name: 'fixture', ran: true, passed: false, state: 'failed', reason: 'Fixture failed', iteration: 1,
      summary: { total: 1, passed: 0, failed: 1, warned: 0, skipped: 0, pending: 0, running: 0, timedout: 0 },
    },
  };
  const resumed: TodoItem = { ...todo, status: 'in_progress', questions: [] };
  let answerBody: unknown;
  const onChanged = vi.fn();
  vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = new URL(String(input), 'http://localhost');
    if (url.pathname === '/api/todos/session/detail') return Response.json({ attempts: [attempt] });
    if (url.pathname === '/api/todos/session/stats') return Response.json({ state: answerBody ? 'working' : 'ask', inProgress: !!answerBody });
    if (url.pathname === '/api/captain/sessions/ask-admission') return Response.json({ id: 'ask-admission', messages: [] });
    if (url.pathname === '/api/todos/answer') {
      answerBody = JSON.parse(init?.body as string);
      return new Response(
        `event: resolved\ndata: ${JSON.stringify({ spec: {}, specYaml: '{}\n', step: 'run' })}\n\nevent: admitted\ndata: ${JSON.stringify({ todo: resumed, promptRunId: 'resumed-run', sessionId: todo.sessionId })}\n\n`,
        { headers: { 'Content-Type': 'text/event-stream' } },
      );
    }
    throw new Error(`Unexpected session request: ${url.pathname}`);
  }));
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <QueryClientProvider client={client}>
      <TodoSession dir="/repo" sessionId={todo.sessionId} active todo={todo} sessionTab="transcript" onChanged={onChanged} />
    </QueryClientProvider>,
  );

  fireEvent.click(await screen.findByRole('radio', { name: 'Postgres' }));
  fireEvent.click(screen.getByRole('button', { name: 'Send answer' }));

  await waitFor(() => expect(onChanged).toHaveBeenCalledWith(resumed));
  expect(screen.queryByRole('alert')).toBeNull();
  expect(answerBody).toEqual({
    dir: '/repo', ref: todo.ref, sessionId: todo.sessionId,
    answer: 'Which database?: Postgres', answers: { '1': 'Postgres' }, rejected: false,
  });
});
