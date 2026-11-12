// Tests the platform dashboard's sandbox usage panel: runs and sandbox time per day by kind,
// with days without runs kept, the headline figures, and each organization's share.
import { render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { queryTestWrapper } from '../todos/queryTestWrapper';
import { SandboxUsagePanel, perDay, type SandboxDay, type SandboxUsage } from './SandboxUsagePanel';

const day = (fields: Partial<SandboxDay>): SandboxDay => ({
  day: '2026-11-08', org: 'o-alpha', orgName: 'Alpha', kind: 'test', backend: 'substrate', status: 'succeeded',
  runs: 1, runSeconds: 0, queueSeconds: 0, ...fields,
});

const usage: SandboxUsage = {
  days: [
    day({ runs: 2, runSeconds: 180, queueSeconds: 20 }),
    day({ org: 'o-beta', orgName: 'Beta', status: 'failed', runSeconds: 60 }),
    day({ day: '2026-11-10', kind: 'agent', status: 'running' }),
  ],
  active: [{ org: 'o-alpha', orgName: 'Alpha', kind: 'agent', status: 'running', runs: 1 }, { org: 'o-beta', orgName: 'Beta', kind: 'lint', status: 'queued', runs: 2 }],
};

let response: { status: number; body: unknown };

beforeEach(() => {
  vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL) => {
    const url = String(input);
    if (url !== '/api/platform/sandbox-runs?days=30') throw new Error(`unexpected fetch ${url}`);
    return Promise.resolve(new Response(JSON.stringify(response.body), { status: response.status }));
  }));
});

afterEach(() => {
  vi.unstubAllGlobals();
});

it('keeps the days without runs and stacks each kind', () => {
  const now = new Date('2026-11-10T12:00:00Z');
  expect(perDay(usage.days, 2, now, row => row.runs)).toEqual([
    { day: '2026-11-08', test: 3, lint: 0, 'push-check': 0, agent: 0 },
    { day: '2026-11-09', test: 0, lint: 0, 'push-check': 0, agent: 0 },
    { day: '2026-11-10', test: 0, lint: 0, 'push-check': 0, agent: 1 },
  ]);
  expect(perDay(usage.days, 2, now, row => row.runSeconds / 60)[0], 'sandbox minutes').toMatchObject({ test: 4 });
});

it('shows the headline figures and each organization\'s runs', async () => {
  response = { status: 200, body: usage };
  render(<SandboxUsagePanel />, { wrapper: queryTestWrapper() });

  expect(await screen.findByText('4 runs')).toBeTruthy();
  expect(screen.getByText('1 failed (25%)')).toBeTruthy();
  expect(screen.getByText('4m sandbox time')).toBeTruthy();
  expect(screen.getByText('3 active now')).toBeTruthy();
  const alpha = screen.getByRole('row', { name: /Alpha/ });
  expect(within(alpha).getAllByRole('cell').map(cell => cell.textContent)).toEqual(['Alpha', '3', '0', '3m', '1']);
  const beta = screen.getByRole('row', { name: /Beta/ });
  expect(within(beta).getAllByRole('cell').map(cell => cell.textContent)).toEqual(['Beta', '1', '1', '1m', '2']);
});

it('gives each chart a table of its numbers, for readers who can\'t tell its colors apart', async () => {
  vi.useFakeTimers({ toFake: ['Date'] });
  vi.setSystemTime(new Date('2026-11-10T12:00:00Z'));
  try {
    response = { status: 200, body: usage };
    render(<SandboxUsagePanel />, { wrapper: queryTestWrapper() });

    const runs = await screen.findByRole('table', { name: 'Runs per day' });
    expect(within(runs).getAllByRole('columnheader').map(cell => cell.textContent)).toEqual(['Day', 'Test', 'Lint', 'Push check', 'Agent run', 'Total']);
    const cells = (day: string) => within(within(runs).getByRole('row', { name: new RegExp(`^${day}`) })).getAllByRole('cell').map(cell => cell.textContent);
    expect(cells('08/11/2026'), 'days are written day first').toEqual(['08/11/2026', '3', '0', '0', '0', '3']);
    expect(cells('09/11/2026'), 'a day without runs').toEqual(['09/11/2026', '0', '0', '0', '0', '0']);
    expect(cells('10/11/2026')).toEqual(['10/11/2026', '0', '0', '0', '1', '1']);
    const minutes = screen.getByRole('table', { name: 'Sandbox minutes per day' });
    expect(within(within(minutes).getByRole('row', { name: /^08\/11\/2026/ })).getAllByRole('cell').map(cell => cell.textContent))
      .toEqual(['08/11/2026', '4', '0', '0', '0', '4']);
  } finally {
    vi.useRealTimers();
  }
});

it('says why the usage is missing', async () => {
  response = { status: 500, body: { error: 'read the platform\'s sandbox runs: connection refused' } };
  render(<SandboxUsagePanel />, { wrapper: queryTestWrapper() });

  expect((await screen.findByRole('alert')).textContent).toContain('connection refused');
});

it('shows only one organization\'s runs when given one', async () => {
  response = { status: 200, body: usage };
  render(<SandboxUsagePanel org="o-beta" />, { wrapper: queryTestWrapper() });

  expect(await screen.findByText('1 runs')).toBeTruthy();
  expect(screen.getByText('2 active now')).toBeTruthy();
  expect(screen.queryByRole('row', { name: /Alpha/ })).toBeNull();
});
