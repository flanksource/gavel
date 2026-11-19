// Tests the organization's usage page: its sandbox runs' CPU and memory, what runs now, its model
// calls' tokens by bucket and cost, how the cost is computed, and the window it covers.
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, expect, it, vi } from 'vitest';
import { queryTestWrapper } from './todos/queryTestWrapper';
import { UsagePanel, type UsageResponse } from './UsagePanel';

const gib = 2 ** 30;

function usage(window: string, otherCurrencyCalls = 0): UsageResponse {
  return {
    window,
    usage: {
      since: '2026-11-01T00:00:00Z',
      compute: { coreSeconds: 9 * 3600, gibSeconds: 18 * 3600, runs: 12, running: { runs: 2, cpuMillicores: 3000, memoryBytes: 6 * gib } },
      tokens: { input: 1200, output: 300, reasoning: 0, cacheRead: 90000, cacheWrite: 500, counted: 2000 },
      costUSD: 4.5,
      calls: 40,
      otherCurrencyCalls,
    },
  };
}

// stubServer answers /api/usage for each window it is asked for, recording the URLs.
function stubServer(otherCurrencyCalls = 0) {
  const urls: string[] = [];
  vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL) => {
    const url = String(input);
    urls.push(url);
    const window = new URL(url, 'http://gavel.test').searchParams.get('window') ?? 'now/M';
    return new Response(JSON.stringify(usage(window, otherCurrencyCalls)));
  }));
  return urls;
}

afterEach(() => {
  vi.unstubAllGlobals();
});

it('shows the CPU and memory the organization\'s runs held, and what runs now', async () => {
  stubServer();
  render(<UsagePanel />, { wrapper: queryTestWrapper() });

  expect(await screen.findByText('9 core-hours')).toBeTruthy();
  expect(screen.getByText('18 GiB-hours')).toBeTruthy();
  expect(screen.getByText('12 runs')).toBeTruthy();
  expect(screen.getByText('2 running · 3 CPU · 6 GiB')).toBeTruthy();
  expect(screen.getByText(/for as long as it ran, whatever it used of it/)).toBeTruthy();
});

it('shows every token bucket and which ones count', async () => {
  stubServer();
  render(<UsagePanel />, { wrapper: queryTestWrapper() });

  const table = await screen.findByRole('table', { name: 'Tokens by kind' });
  const row = (name: string) => within(table).getByRole('row', { name: new RegExp(`^${name}`) }).textContent;
  expect(row('Input')).toContain('1,200');
  expect(row('Cache writes')).toContain('500');
  expect(row('Output')).toContain('300');
  expect(row('Cache reads')).toContain('90,000');
  expect(row('Cache reads')).toContain('not counted');
  expect(screen.getByText('2,000 tokens counted')).toBeTruthy();
});

it('says how the cost is computed and what it leaves out', async () => {
  stubServer(3);
  render(<UsagePanel />, { wrapper: queryTestWrapper() });

  expect(await screen.findByText('$4.50')).toBeTruthy();
  expect(screen.getByText(/the cost the provider reported for each call, or else its list price/)).toBeTruthy();
  expect(screen.getByText(/3 calls priced in another currency/)).toBeTruthy();
});

it('changes the window it covers', async () => {
  const urls = stubServer();
  render(<UsagePanel />, { wrapper: queryTestWrapper() });
  await screen.findByText('9 core-hours');

  fireEvent.change(screen.getByLabelText('Window'), { target: { value: 'now-30d' } });
  await waitFor(() => expect(urls).toContain('/api/usage?window=now-30d'));
  expect(urls[0]).toBe('/api/usage?window=now%2FM');
});

it('says why there is no usage', async () => {
  vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ error: 'sandbox runs exist only on a hosted dashboard with a sandbox' }), { status: 404 })));
  render(<UsagePanel />, { wrapper: queryTestWrapper() });

  expect((await screen.findByRole('alert')).textContent).toContain('sandbox runs exist only on a hosted dashboard with a sandbox');
});
