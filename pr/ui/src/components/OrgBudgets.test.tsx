// Tests the organization's budgets section: each budget with how much of it is used, in its unit,
// and an admin setting and removing budgets, which members only see.
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, expect, it, vi } from 'vitest';
import { queryTestWrapper } from './todos/queryTestWrapper';
import { OrgBudgets, type BudgetStatus } from './OrgBudgets';

const monthlyCPU: BudgetStatus = { resource: 'cpu', kind: 'cumulative', amount: 40, window: 'now/M', used: 12.25, since: '2026-11-01T00:00:00Z' };
const memoryAtOnce: BudgetStatus = { resource: 'memory', kind: 'concurrent', amount: 16, used: 16 };
const monthlyUSD: BudgetStatus = { resource: 'usd', kind: 'cumulative', amount: 50, window: 'now-30d', used: 4.5, since: '2026-10-20T00:00:00Z' };

// stubServer answers /api/budgets from budgets, recording every request.
function stubServer(budgets: BudgetStatus[], canEdit: boolean, refusal?: string) {
  const requests: { method: string; url: string; body?: unknown }[] = [];
  vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    const method = init?.method ?? 'GET';
    const body = init?.body ? JSON.parse(String(init.body)) : undefined;
    requests.push({ method, url, body });
    if (method === 'PUT' && refusal) return new Response(JSON.stringify({ error: refusal }), { status: 400 });
    if (method === 'PUT') {
      const [, , , resource, kind] = url.split('/');
      budgets.push({ resource, kind, amount: body.amount, window: body.window, used: 0 } as BudgetStatus);
      return new Response(JSON.stringify(budgets[budgets.length - 1]));
    }
    if (method === 'DELETE') {
      budgets.splice(0, 1);
      return new Response(JSON.stringify({}));
    }
    return new Response(JSON.stringify({ budgets, canEdit }));
  }));
  return requests;
}

afterEach(() => {
  vi.unstubAllGlobals();
});

it('shows each budget\'s use in its unit, and the ones used up', async () => {
  stubServer([monthlyCPU, memoryAtOnce, monthlyUSD], false);
  render(<OrgBudgets />, { wrapper: queryTestWrapper() });

  expect(await screen.findByText('CPU this month: 12.3 of 40 core-hours')).toBeTruthy();
  expect(screen.getByText('Memory at once: 16 of 16 GiB · used up')).toBeTruthy();
  expect(screen.getByText('Model cost in the last 30 days: $4.50 of $50.00')).toBeTruthy();
  expect(screen.getByText(/Only an organization admin can change budgets/)).toBeTruthy();
  expect(screen.queryByRole('button', { name: 'Set budget' })).toBeNull();
  expect(screen.getByText(/A used-up budget refuses new runs and stops the ones running/)).toBeTruthy();
  expect(screen.getByText(/token and model cost budgets hold back agent runs only/)).toBeTruthy();
  expect(screen.getByText(/waits its turn for room/)).toBeTruthy();
});

it('says when there are no budgets', async () => {
  stubServer([], false);
  render(<OrgBudgets />, { wrapper: queryTestWrapper() });

  expect(await screen.findByText('No budgets: this organization\'s runs and model calls are not capped.')).toBeTruthy();
});

it('lets an admin set a cumulative budget with its window', async () => {
  const requests = stubServer([], true);
  render(<OrgBudgets />, { wrapper: queryTestWrapper() });
  await screen.findByRole('button', { name: 'Set budget' });

  fireEvent.change(screen.getByLabelText('Resource'), { target: { value: 'tokens' } });
  fireEvent.change(screen.getByLabelText('Amount (tokens)'), { target: { value: '2000000' } });
  fireEvent.change(screen.getByLabelText('Budget window'), { target: { value: 'now/w' } });
  fireEvent.click(screen.getByRole('button', { name: 'Set budget' }));

  await waitFor(() => expect(requests).toContainEqual({ method: 'PUT', url: '/api/budgets/tokens/cumulative', body: { amount: 2000000, window: 'now/w' } }));
});

it('offers concurrent budgets only for CPU and memory, which take no window', async () => {
  const requests = stubServer([], true);
  render(<OrgBudgets />, { wrapper: queryTestWrapper() });
  await screen.findByRole('button', { name: 'Set budget' });

  fireEvent.change(screen.getByLabelText('Resource'), { target: { value: 'usd' } });
  expect(screen.queryByLabelText('Kind'), 'model cost is spent, not held').toBeNull();
  fireEvent.change(screen.getByLabelText('Resource'), { target: { value: 'memory' } });
  fireEvent.change(screen.getByLabelText('Kind'), { target: { value: 'concurrent' } });
  expect(screen.queryByLabelText('Budget window')).toBeNull();
  fireEvent.change(screen.getByLabelText('Amount (GiB)'), { target: { value: '32' } });
  fireEvent.click(screen.getByRole('button', { name: 'Set budget' }));

  await waitFor(() => expect(requests).toContainEqual({ method: 'PUT', url: '/api/budgets/memory/concurrent', body: { amount: 32 } }));
});

it('lets an admin remove a budget', async () => {
  const requests = stubServer([monthlyCPU], true);
  render(<OrgBudgets />, { wrapper: queryTestWrapper() });

  fireEvent.click(await screen.findByRole('button', { name: 'Remove the CPU this month budget' }));
  await waitFor(() => expect(requests).toContainEqual({ method: 'DELETE', url: '/api/budgets/cpu/cumulative', body: undefined }));
  expect(await screen.findByText('No budgets: this organization\'s runs and model calls are not capped.')).toBeTruthy();
});

it('says why a budget wasn\'t set', async () => {
  stubServer([], true, 'invalid budget: amount must be more than 0, got 0');
  render(<OrgBudgets />, { wrapper: queryTestWrapper() });
  await screen.findByRole('button', { name: 'Set budget' });

  fireEvent.change(screen.getByLabelText('Amount (core-hours)'), { target: { value: '0' } });
  fireEvent.click(screen.getByRole('button', { name: 'Set budget' }));

  expect((await screen.findByRole('alert')).textContent).toContain('amount must be more than 0, got 0');
});
