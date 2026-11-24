// Tests the platform dashboard's GitHub panel: the App's installations with their
// organization, repositories, rate limit and state, and the reason when a source fails.
import { render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { queryTestWrapper } from '../todos/queryTestWrapper';
import { GitHubPanel, type PlatformGitHub } from './GitHubPanel';

let response: { status: number; body: unknown };

beforeEach(() => {
  vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL) => {
    const url = String(input);
    if (url !== '/api/platform/github') throw new Error(`unexpected fetch ${url}`);
    return Promise.resolve(new Response(JSON.stringify(response.body), { status: response.status }));
  }));
});

afterEach(() => {
  vi.unstubAllGlobals();
});

const cells = (name: RegExp) => within(screen.getByRole('row', { name })).getAllByRole('cell').map(cell => cell.textContent);

it('lists the App\'s installations with their organization, repositories and rate limit', async () => {
  const github: PlatformGitHub = {
    app: { id: 123, slug: 'gavel-dev' },
    installations: [
      { id: 7, account: 'acme', accountType: 'Organization', createdAt: '2026-10-01T09:00:00Z', org: { id: 'o-alpha', slug: 'alpha', name: 'Alpha' }, repos: 3, rateLimit: { limit: 5000, remaining: 4200, used: 800, reset: 1762780000, resource: 'core' } },
      { id: 8, account: 'solo', accountType: 'User', createdAt: '2026-10-02T09:00:00Z', suspendedAt: '2026-11-01T10:00:00Z', repos: 0 },
      { id: 9, account: 'flaky', accountType: 'Organization', createdAt: '2026-10-03T09:00:00Z', repos: 0, rateLimitError: 'read the rate limit of GitHub App installation 9: Server Error' },
      { id: 11, account: 'gone', org: { id: 'o-alpha', slug: 'alpha', name: 'Alpha' }, repos: 1, uninstalled: true },
    ],
  };
  response = { status: 200, body: github };
  render(<GitHubPanel />, { wrapper: queryTestWrapper() });

  expect(await screen.findByRole('heading', { name: 'GitHub App gavel-dev · 4 installations, 2 linked, 1 suspended' })).toBeTruthy();
  expect(cells(/acme/).slice(0, 4)).toEqual(['acme', 'Alpha', '3', '4200 of 5000 left']);
  expect(cells(/solo/)[1], 'linked to no organization').toBe('Not linked');
  expect(cells(/solo/)[4]).toBe('Suspended 01/11/2026');
  expect(cells(/acme/), 'installed on, day first').toContain('01/10/2026');
  expect(cells(/flaky/)[3]).toContain('Server Error');
  expect(cells(/gone/)[4], 'linked in gavel, gone from GitHub').toBe('Uninstalled from GitHub');
});

it('says why GitHub\'s installations are missing', async () => {
  response = { status: 200, body: { app: { id: 123, slug: 'gavel-dev' }, installations: [], githubError: 'list the GitHub App\'s installations: cannot reach GitHub' } };
  render(<GitHubPanel />, { wrapper: queryTestWrapper() });

  expect((await screen.findByRole('alert')).textContent).toContain('cannot reach GitHub');
});

it('says how to configure the App when serve runs without one', async () => {
  response = { status: 404, body: { error: 'the GitHub App isn\'t configured: start gavel serve with --github-app-id' } };
  render(<GitHubPanel />, { wrapper: queryTestWrapper() });

  expect((await screen.findByRole('alert')).textContent).toContain('--github-app-id');
});

it('shows only one organization\'s installations when given one', async () => {
  response = {
    status: 200,
    body: {
      app: { id: 123, slug: 'gavel-dev' },
      installations: [
        { id: 7, account: 'acme', org: { id: 'o-alpha', slug: 'alpha', name: 'Alpha' }, repos: 3 },
        { id: 8, account: 'solo', repos: 0 },
      ],
    },
  };
  render(<GitHubPanel org="o-alpha" />, { wrapper: queryTestWrapper() });

  expect(await screen.findByRole('heading', { name: 'GitHub App gavel-dev · 1 installations, 1 linked, 0 suspended' })).toBeTruthy();
  expect(screen.queryByRole('row', { name: /solo/ })).toBeNull();
});
