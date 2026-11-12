// Tests the platform dashboard's budgets panel: each budget rule's groups with their spend
// against the rule's amount, and why a rule or a group has no total.
import { render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { queryTestWrapper } from '../todos/queryTestWrapper';
import { BudgetsPanel, type Budget } from './BudgetsPanel';

let response: { status: number; body: unknown };

beforeEach(() => {
  vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL) => {
    const url = String(input);
    if (url !== '/api/platform/budgets') throw new Error(`unexpected fetch ${url}`);
    return Promise.resolve(new Response(JSON.stringify(response.body), { status: response.status }));
  }));
});

afterEach(() => {
  vi.unstubAllGlobals();
});

const rule = (overrides: Partial<Budget>): Budget => ({
  id: overrides.name ?? 'rule', name: 'Rule', match: {}, groupBy: ['org'], amountUSD: 10, window: 'now/d',
  windowStart: '2026-11-11T00:00:00Z', groups: [], ...overrides,
});

function rows() {
  const table = screen.getByRole('table', { name: 'Budgets' });
  return within(table).getAllByRole('row').slice(1).map(row => within(row).getAllByRole('cell').map(cell => cell.textContent));
}

it('shows each group\'s spend against its rule\'s amount, and why one has no total', async () => {
  response = {
    status: 200,
    body: {
      budgets: [
        rule({ name: 'Broken', window: '2026-11-01', windowStart: undefined, windowError: 'budget window "2026-11-01" must be relative to now' }),
        rule({
          name: 'Daily per org', groups: [
            { values: { org: 'alpha' }, spendUSD: 3.5, calls: 2 },
            { values: { org: 'beta' }, spendUSD: 12, calls: 4 },
            { values: { org: 'gamma' }, spendUSD: 0, calls: 1, error: '1 completed call is priced in another currency, which can\'t be added to USD spend' },
          ],
        }),
        rule({ name: 'Monthly', window: 'now/M', windowStart: '2026-11-01T00:00:00Z', amountUSD: 500 }),
      ],
    },
  };
  render(<BudgetsPanel />, { wrapper: queryTestWrapper() });

  await screen.findByRole('table', { name: 'Budgets' });
  expect(rows()).toEqual([
    ['Broken', '2026-11-01', 'budget window "2026-11-01" must be relative to now', '—', '$10.00', '—'],
    ['Daily per org', 'now/d since 11/11/2026', 'org=alpha', '$3.50', '$10.00', '35%'],
    ['Daily per org', 'now/d since 11/11/2026', 'org=beta', '$12.00', '$10.00', '120% · over budget'],
    ['Daily per org', 'now/d since 11/11/2026', 'org=gamma', '1 completed call is priced in another currency, which can\'t be added to USD spend', '$10.00', '—'],
    ['Monthly', 'now/M since 01/11/2026', 'No spend in this window', '$0.00', '$500.00', '0%'],
  ]);
});

it('says how to add a budget when there are none', async () => {
  response = { status: 200, body: { budgets: [] } };
  render(<BudgetsPanel />, { wrapper: queryTestWrapper() });

  expect(await screen.findByText(/No budget rules\. Add one with captain budget-rule create/)).toBeTruthy();
});

it('says why the budgets didn\'t load', async () => {
  response = { status: 500, body: { error: 'read the budget rules: connection refused' } };
  render(<BudgetsPanel />, { wrapper: queryTestWrapper() });

  expect((await screen.findByRole('alert')).textContent).toContain('connection refused');
});
