// Tests the platform dashboard's model usage panel: model spend per day with its numbers as a
// table, each organization's and model's share, and calls whose cost can't be added up.
import { render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { queryTestWrapper } from '../todos/queryTestWrapper';
import { ModelUsagePanel, costPerDay, type ModelUsageDay } from './ModelUsagePanel';

let response: { status: number; body: unknown };

beforeEach(() => {
  vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL) => {
    const url = String(input);
    if (url !== '/api/platform/model-usage?days=30') throw new Error(`unexpected fetch ${url}`);
    return Promise.resolve(new Response(JSON.stringify(response.body), { status: response.status }));
  }));
});

afterEach(() => {
  vi.unstubAllGlobals();
});

const call = (overrides: Partial<ModelUsageDay>): ModelUsageDay => ({
  day: '2026-11-10', model: 'claude-sonnet', provider: 'anthropic', calls: 1,
  inputTokens: 0, outputTokens: 0, reasoningTokens: 0, cacheReadTokens: 0, cacheWriteTokens: 0,
  costUSD: 0, otherCurrencyCalls: 0, ...overrides,
});

it('totals the cost of each day of the window, keeping days without calls', () => {
  expect(costPerDay([
    call({ day: '2026-11-09', costUSD: 1.5 }),
    call({ day: '2026-11-09', costUSD: 0.25, org: 'b', orgName: 'Beta' }),
    call({ day: '2026-11-10', costUSD: 2 }),
  ], 2, new Date('2026-11-10T12:00:00Z'))).toEqual([
    { day: '2026-11-08', costUSD: 0 },
    { day: '2026-11-09', costUSD: 1.75 },
    { day: '2026-11-10', costUSD: 2 },
  ]);
});

it('shows the spend, each organization\'s and model\'s share, and calls left out of the cost', async () => {
  response = {
    status: 200,
    body: {
      days: [
        call({ org: 'a', orgName: 'Alpha', calls: 2, inputTokens: 1000, outputTokens: 500, costUSD: 1.25 }),
        call({ org: 'a', orgName: 'Alpha', model: 'gpt-5', provider: 'openai', inputTokens: 10, otherCurrencyCalls: 1 }),
        call({ calls: 3, inputTokens: 3, costUSD: 0.5 }),
      ],
    },
  };
  render(<ModelUsagePanel />, { wrapper: queryTestWrapper() });

  expect(await screen.findByText('$1.75 spent')).toBeTruthy();
  expect(screen.getByText('1,513 tokens')).toBeTruthy();
  expect(screen.getByRole('status').textContent).toBe('1 call was priced in another currency: its tokens count, its cost isn\'t in these totals.');

  const orgs = screen.getByRole('table', { name: 'Model spend by organization' });
  expect(within(orgs).getAllByRole('row').slice(1).map(row => within(row).getAllByRole('cell').map(cell => cell.textContent)))
    .toEqual([['Alpha', '3', '1,510', '$1.25'], ['Unattributed', '3', '3', '$0.50']]);
  const models = screen.getByRole('table', { name: 'Model spend by model' });
  expect(within(models).getAllByRole('row').slice(1).map(row => within(row).getAllByRole('cell')[0].textContent))
    .toEqual(['claude-sonnet', 'gpt-5']);
});

it('says why model usage didn\'t load', async () => {
  response = { status: 500, body: { error: 'read the platform\'s model usage: connection refused' } };
  render(<ModelUsagePanel />, { wrapper: queryTestWrapper() });

  expect((await screen.findByRole('alert')).textContent).toContain('connection refused');
});

it('shows only one organization\'s calls when given one', async () => {
  response = { status: 200, body: { days: [call({ org: 'a', orgName: 'Alpha', costUSD: 1.25 }), call({ costUSD: 0.5 })] } };
  render(<ModelUsagePanel org="a" />, { wrapper: queryTestWrapper() });

  expect(await screen.findByText('$1.25 spent')).toBeTruthy();
  expect(screen.queryByText('Unattributed')).toBeNull();
});
