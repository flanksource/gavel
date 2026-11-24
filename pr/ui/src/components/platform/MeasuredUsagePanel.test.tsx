// Tests an organization's measured CPU and memory: platform admins see any organization's, its
// admins their own on the usage page, members nothing, and every viewer is told the figures are sampled.
import { render, screen } from '@testing-library/react';
import { afterEach, expect, it, vi } from 'vitest';
import { queryTestWrapper } from '../todos/queryTestWrapper';
import { MeasuredUsagePanel, OwnMeasuredUsage } from './MeasuredUsagePanel';
import type { MetricsView } from './SubstrateMetricsPanel';

const view: MetricsView = {
  missionControl: 'https://mc.example.com/api', title: 'Measured use', lastRefreshedAt: '2026-11-19T10:05:00Z',
  panels: [{ name: 'Running environments', type: 'number', value: 2 }],
};

// stubServer answers the measured use routes with view and /api/me with role, recording each URL.
function stubServer(role: string) {
  const urls: string[] = [];
  vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL) => {
    const url = String(input);
    urls.push(url);
    if (url === '/api/me') {
      const org = { id: 'org-1', slug: 'acme', name: 'Acme', role };
      return Promise.resolve(new Response(JSON.stringify({ user: { id: 'u', name: 'Ada', email: '', githubLogin: '' }, org, orgs: [org] })));
    }
    if (url === '/api/usage/measured' || url === '/api/platform/orgs/org-2/measured-usage') return Promise.resolve(new Response(JSON.stringify(view)));
    throw new Error(`unexpected fetch ${url}`);
  }));
  return urls;
}

afterEach(() => {
  vi.unstubAllGlobals();
});

it('shows a platform admin the organization\'s measured use, saying it is sampled', async () => {
  const urls = stubServer('org:member');
  render(<MeasuredUsagePanel org="org-2" />, { wrapper: queryTestWrapper() });

  expect(await screen.findByText('2')).toBeTruthy();
  expect(screen.getByText(/^Measured CPU and memory/)).toBeTruthy();
  expect(screen.getByText(/sampled from substrate every 15 seconds/)).toBeTruthy();
  expect(screen.getByText(/budgets count each run's size for as long as it runs/)).toBeTruthy();
  expect(urls).toEqual(['/api/platform/orgs/org-2/measured-usage']);
});

it('shows an organization admin their own organization\'s measured use', async () => {
  const urls = stubServer('org:admin');
  render(<OwnMeasuredUsage />, { wrapper: queryTestWrapper() });

  expect(await screen.findByText('2')).toBeTruthy();
  expect(urls).toContain('/api/usage/measured');
});

it('shows a member nothing, without asking for it', async () => {
  const urls = stubServer('org:member');
  const { container } = render(<OwnMeasuredUsage />, { wrapper: queryTestWrapper() });

  await vi.waitFor(() => expect(urls).toContain('/api/me'));
  expect(container.textContent).toBe('');
  expect(urls).not.toContain('/api/usage/measured');
});
