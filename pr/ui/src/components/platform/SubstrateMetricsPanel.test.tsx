// Tests the platform dashboard's substrate metrics panel: a Mission Control View's timeseries as
// line charts with their numbers as a table, its single values, and why anything isn't shown.
import { render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { queryTestWrapper } from '../todos/queryTestWrapper';
import { SubstrateMetricsPanel, seriesLabel, timeRows, type MetricsView } from './SubstrateMetricsPanel';

let response: { status: number; body: unknown };

beforeEach(() => {
  vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL) => {
    const url = String(input);
    if (url !== '/api/platform/substrate/metrics') throw new Error(`unexpected fetch ${url}`);
    return Promise.resolve(new Response(JSON.stringify(response.body), { status: response.status }));
  }));
});

afterEach(() => {
  vi.unstubAllGlobals();
});

const at = (minute: number) => `2026-11-10T10:0${minute}:00Z`;

it('names a series by its labels, or by its panel when it has none', () => {
  expect(seriesLabel({ workerpool: 'pool-a' }, 'Ready workers')).toBe('pool-a');
  expect(seriesLabel({ workerpool: 'pool-a', state: 'active' }, 'Workers')).toBe('state=active, workerpool=pool-a');
  expect(seriesLabel({}, 'Crashes')).toBe('Crashes');
});

it('lines up the series by time, leaving a gap where one has no sample', () => {
  expect(timeRows([
    { labels: { workerpool: 'pool-a' }, points: [{ time: at(0), value: 3 }, { time: at(1), value: 4 }] },
    { labels: { workerpool: 'pool-b' }, points: [{ time: at(1), value: 2 }] },
  ])).toEqual([
    { time: at(0), s0: 3 },
    { time: at(1), s0: 4, s1: 2 },
  ]);
});

it('draws the view\'s timeseries with their numbers as a table, and its single values', async () => {
  const view: MetricsView = {
    missionControl: 'https://mc.example.com/api', title: 'Substrate', lastRefreshedAt: '2026-11-10T10:05:00Z',
    panels: [
      {
        name: 'Ready workers', description: 'Last 12h', type: 'timeseries',
        series: [
          { labels: { workerpool: 'pool-a' }, points: [{ time: at(0), value: 3 }, { time: at(1), value: 4 }] },
          { labels: { workerpool: 'pool-b' }, points: [{ time: at(1), value: 2 }] },
        ],
      },
      { name: 'Crashes today', type: 'number', unit: 'crashes', value: 7 },
      { name: 'No data yet', type: 'gauge' },
      { name: 'By pool', type: 'piechart', unsupported: 'gavel doesn\'t draw piechart panels' },
    ],
  };
  response = { status: 200, body: view };
  render(<SubstrateMetricsPanel />, { wrapper: queryTestWrapper() });

  expect(await screen.findByText('7 crashes')).toBeTruthy();
  expect(screen.getByText('no data')).toBeTruthy();
  const table = screen.getByRole('table', { name: 'Ready workers' });
  expect(within(table).getAllByRole('columnheader').map(cell => cell.textContent)).toEqual(['Time', 'pool-a', 'pool-b']);
  expect(within(table).getAllByRole('row').slice(1).map(row => within(row).getAllByRole('cell').slice(1).map(cell => cell.textContent)))
    .toEqual([['3', '—'], ['4', '2']]);
  expect(screen.getByText(/By pool: gavel doesn't draw piechart panels/)).toBeTruthy();
  expect(screen.getByRole('link', { name: 'Substrate' }).getAttribute('href')).toBe('https://mc.example.com');
});

it('says when the chart leaves series out, and the table keeps them all', async () => {
  const series = Array.from({ length: 10 }, (_, i) => ({ labels: { worker: `w${i}` }, points: [{ time: at(0), value: i }] }));
  response = { status: 200, body: { missionControl: 'https://mc.example.com/api', lastRefreshedAt: at(0), panels: [{ name: 'Actors', type: 'timeseries', series }] } };
  render(<SubstrateMetricsPanel />, { wrapper: queryTestWrapper() });

  expect(await screen.findByText('The chart shows 8 of 10 series; the table lists them all.')).toBeTruthy();
  expect(within(screen.getByRole('table', { name: 'Actors' })).getAllByRole('columnheader')).toHaveLength(11);
});

it('says when Mission Control couldn\'t refresh the view', async () => {
  response = { status: 200, body: { missionControl: 'https://mc.example.com/api', refreshError: 'prometheus: connection refused', lastRefreshedAt: at(0), panels: [] } };
  render(<SubstrateMetricsPanel />, { wrapper: queryTestWrapper() });

  expect((await screen.findByRole('status')).textContent).toContain('prometheus: connection refused');
});

it('says how to graph substrate\'s metrics when serve runs without the view', async () => {
  response = { status: 404, body: { error: 'no View graphs substrate\'s metrics: start gavel serve with --mission-control-substrate-view namespace/name' } };
  render(<SubstrateMetricsPanel />, { wrapper: queryTestWrapper() });

  expect((await screen.findByRole('alert')).textContent).toContain('--mission-control-substrate-view');
});
