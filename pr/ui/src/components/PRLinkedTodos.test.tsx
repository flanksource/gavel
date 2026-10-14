import { act, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { PRLinkedTodo } from '../types';
import { PRLinkedTodos } from './PRLinkedTodos';
import { queryTestWrapper } from './todos/queryTestWrapper';

const PR_URL = 'https://github.com/acme/ledger/pull/42';

const linked: PRLinkedTodo[] = [
  { id: 'full-1', shortId: 'ab12', title: 'Fix ledger rounding', status: 'in_progress', priority: 'high', relation: 'implements', source: 'manual' },
  { id: 'full-2', shortId: 'cd34', title: 'Audit precision', status: 'completed', priority: 'low', relation: 'related', source: 'triage' },
];

let response: () => Promise<Response>;
let fetchMock: ReturnType<typeof vi.fn>;

function json(body: unknown, status = 200) {
  return Promise.resolve({ ok: status < 400, status, json: async () => body, text: async () => JSON.stringify(body), clone() { return this; } } as unknown as Response);
}

async function renderLinked() {
  const view = render(<PRLinkedTodos prUrl={PR_URL} />, { wrapper: queryTestWrapper() });
  await waitFor(() => expect(fetchMock).toHaveBeenCalled());
  await act(async () => { await new Promise(resolve => setTimeout(resolve, 0)); });
  return view;
}

beforeEach(() => {
  response = () => json(linked);
  fetchMock = vi.fn(() => response());
  vi.stubGlobal('fetch', fetchMock);
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('PRLinkedTodos', () => {
  it('asks /api/prs/todos for the PR by URL', async () => {
    await renderLinked();

    const url = new URL(String(fetchMock.mock.calls[0]![0]), 'http://x');
    expect(url.pathname).toBe('/api/prs/todos');
    expect(Object.fromEntries(url.searchParams)).toEqual({ url: PR_URL });
  });

  it('lists each linked todo with its status, relation and a link to the todo route', async () => {
    await renderLinked();

    const section = screen.getByRole('heading', { name: 'Linked TODOs' }).closest('section')!;
    const first = screen.getByRole('link', { name: 'Fix ledger rounding' });
    expect(first.getAttribute('href')).toBe('/todos/ab12');
    expect(screen.getByRole('link', { name: 'Audit precision' }).getAttribute('href')).toBe('/todos/cd34');
    const rows = within(section).getAllByRole('listitem');
    expect(rows).toHaveLength(2);
    expect(within(rows[0]!).getByText('implements')).toBeTruthy();
    expect(within(rows[0]!).getByRole('img', { name: 'in progress' })).toBeTruthy();
    expect(within(rows[1]!).getByText('related')).toBeTruthy();
    expect(within(rows[1]!).getByText('triage')).toBeTruthy();
    expect(within(rows[1]!).getByRole('img', { name: 'completed' })).toBeTruthy();
  });

  it('renders nothing for a PR with no linked todos', async () => {
    response = () => json([]);
    const { container } = await renderLinked();

    expect(container.textContent).toBe('');
  });

  it('reports a failed lookup instead of pretending there are no links', async () => {
    response = () => json({ error: 'database disabled' }, 500);
    await renderLinked();

    expect(screen.getByText(/Failed to load todos linked to .*database disabled/)).toBeTruthy();
  });
});
