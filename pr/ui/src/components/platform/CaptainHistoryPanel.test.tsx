// Tests the platform dashboard's Captain history panel: prompt runs per day by outcome, each
// organization's share and live sessions, and the latest failures with their errors.
import { render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { queryTestWrapper } from '../todos/queryTestWrapper';
import { CaptainHistoryPanel, runsPerDay, type CaptainHistory, type CaptainRunDay } from './CaptainHistoryPanel';

let response: { status: number; body: unknown };

beforeEach(() => {
  vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL) => {
    const url = String(input);
    if (url !== '/api/platform/captain-history?days=30') throw new Error(`unexpected fetch ${url}`);
    return Promise.resolve(new Response(JSON.stringify(response.body), { status: response.status }));
  }));
});

afterEach(() => {
  vi.unstubAllGlobals();
});

const day = (overrides: Partial<CaptainRunDay>): CaptainRunDay => ({ day: '2026-11-10', origin: 'todo', state: 'succeeded', runs: 1, runSeconds: 0, ...overrides });

function cells(name: string) {
  const table = screen.getByRole('table', { name });
  return within(table).getAllByRole('row').slice(1).map(row => within(row).getAllByRole('cell').map(cell => cell.textContent));
}

it('counts each day\'s runs by outcome, in progress for those not finished', () => {
  expect(runsPerDay([
    day({ day: '2026-11-09', state: 'failed', runs: 2 }),
    day({ day: '2026-11-10', state: 'running' }),
    day({ day: '2026-11-10', state: 'waiting' }),
    day({ day: '2026-11-10', runs: 3 }),
  ], 1, new Date('2026-11-10T12:00:00Z'))).toEqual([
    { day: '2026-11-09', succeeded: 0, failed: 2, cancelled: 0, active: 0 },
    { day: '2026-11-10', succeeded: 3, failed: 0, cancelled: 0, active: 2 },
  ]);
});

it('shows the runs, each organization\'s share and live sessions, and the latest failures', async () => {
  const history: CaptainHistory = {
    days: [
      day({ org: 'a', orgName: 'Alpha', runs: 3, runSeconds: 120 }),
      day({ org: 'a', orgName: 'Alpha', state: 'failed', runSeconds: 30 }),
      day({ origin: 'cli', state: 'failed', runSeconds: 30 }),
    ],
    failures: [
      { id: 'r2', origin: 'cli', error: 'no model key', createdAt: '2026-11-10T09:00:00Z' },
      { id: 'r1', org: 'a', orgName: 'Alpha', origin: 'todo', error: 'the model refused', createdAt: '2026-11-10T08:00:00Z' },
    ],
    live: [
      { org: 'a', orgName: 'Alpha', health: 'healthy', sessions: 2 },
      { org: 'a', orgName: 'Alpha', health: 'stalled', sessions: 1 },
    ],
  };
  response = { status: 200, body: history };
  render(<CaptainHistoryPanel />, { wrapper: queryTestWrapper() });

  expect(await screen.findByText('5 prompt runs')).toBeTruthy();
  expect(screen.getByText('2 failed (40%)')).toBeTruthy();
  expect(screen.getByText('3m run time')).toBeTruthy();
  expect(screen.getByText('3 sessions live, 1 not healthy')).toBeTruthy();
  expect(cells('Captain runs by organization')).toEqual([
    ['Alpha', '4', '1', '3m', '3 (1 stalled)'],
    ['Unattributed', '1', '1', '30s', '0'],
  ]);
  expect(cells('Latest failed prompt runs')).toEqual([
    [new Date('2026-11-10T09:00:00Z').toLocaleString(), 'Unattributed', 'cli', 'no model key'],
    [new Date('2026-11-10T08:00:00Z').toLocaleString(), 'Alpha', 'todo', 'the model refused'],
  ]);
});

it('says why Captain\'s history didn\'t load', async () => {
  response = { status: 500, body: { error: 'read Captain\'s prompt runs: connection refused' } };
  render(<CaptainHistoryPanel />, { wrapper: queryTestWrapper() });

  expect((await screen.findByRole('alert')).textContent).toContain('connection refused');
});
