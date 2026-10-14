import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { Button } from '@flanksource/clicky-ui/components';
import { TodoPRStatus } from './TodoPRStatus';
import type { PRDetail } from '../../types';

vi.mock('../PRActions', () => ({
  PRActions: ({ disabled, onChanged }: { disabled?: boolean; onChanged: () => void }) => (
    <Button disabled={disabled} onClick={onChanged}>PR actions</Button>
  ),
}));

const pr = { number: 27, url: 'https://github.com/acme/service/pull/27' };
const status: PRDetail & { repo: string } = {
  repo: 'acme/service',
  pr: {
    number: 27, title: 'Fix validation', author: { login: 'dev' },
    headRefName: 'fix/validation', baseRefName: 'main', headRefOid: 'abc123',
    state: 'OPEN', isDraft: false, reviewDecision: 'APPROVED', mergeable: 'MERGEABLE', url: pr.url,
    statusCheckRollup: [{ name: 'Tests', status: 'COMPLETED', conclusion: 'SUCCESS', detailsUrl: pr.url }],
    reviewThreads: [{ id: 1, threadId: 'thread-1', body: 'Validate input', author: 'reviewer', url: pr.url, createdAt: '', isResolved: false }],
  },
};

function renderStatus(disabled = false) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(<QueryClientProvider client={client}><TodoPRStatus pr={pr} disabled={disabled} /></QueryClientProvider>);
}

afterEach(() => vi.unstubAllGlobals());

describe('linked TODO PR status', () => {
  it('loads the linked URL only on demand, then refreshes after an action', async () => {
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => new Response(JSON.stringify(status)));
    vi.stubGlobal('fetch', fetchMock);
    renderStatus();
    expect(screen.getByRole('link', { name: 'PR #27' }).getAttribute('href')).toBe(pr.url);
    expect(fetchMock).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole('button', { name: 'PR status and actions' }));
    expect(await screen.findByText('Fix validation')).toBeTruthy();
    expect(screen.getByText('fix/validation → main')).toBeTruthy();
    expect(screen.getByText('1 unresolved review thread')).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Tests: SUCCESS' })).toBeTruthy();
    expect(fetchMock.mock.calls[0][0]).toBe(`/api/prs/linked?url=${encodeURIComponent(pr.url)}`);
    fireEvent.click(screen.getByRole('button', { name: 'PR actions' }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
  });

  it('shows provider errors and offers refresh without enabling actions', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ error: 'Token expired' }), { status: 502 })));
    renderStatus();
    fireEvent.click(screen.getByRole('button', { name: 'PR status and actions' }));
    expect((await screen.findByRole('alert')).textContent).toContain('Token expired');
    expect(screen.queryByRole('button', { name: 'PR actions' })).toBeNull();
    expect(screen.getByRole('button', { name: 'Refresh PR status' })).toBeTruthy();
  });

  it('keeps status readable but prevents actions while a TODO turn is running', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify(status))));
    renderStatus(true);
    fireEvent.click(screen.getByRole('button', { name: 'PR status and actions' }));
    expect(await screen.findByText('Fix validation')).toBeTruthy();
    expect((screen.getByRole('button', { name: 'PR actions' }) as HTMLButtonElement).disabled).toBe(true);
  });
});
