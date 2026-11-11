// Tests one organization's view of the platform dashboard: its summary, the panels showing only
// its rows, the way back to every organization, and an organization gavel doesn't know.
import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { queryTestWrapper } from '../todos/queryTestWrapper';
import { OrganizationView } from './OrganizationView';

const bodies: Record<string, unknown> = {
  '/api/platform/tenants?days=30': {
    tenants: [
      { id: 'o-alpha', clerkOrgId: 'org_alpha', slug: 'alpha', name: 'Alpha', createdAt: '2026-10-01T00:00:00Z', members: 2, repos: 3, pushedRepos: 1,
        githubInstallations: 1, sandboxRuns: 12, lastRunAt: '2026-11-09T10:00:00Z', inGavel: true, clerkMembers: 4 },
    ],
  },
  '/api/platform/sandbox-runs?days=30': {
    days: [
      { day: '2026-11-10', org: 'o-alpha', orgName: 'Alpha', kind: 'test', backend: 'substrate', status: 'succeeded', runs: 2, runSeconds: 0, queueSeconds: 0 },
      { day: '2026-11-10', org: 'o-beta', orgName: 'Beta', kind: 'test', backend: 'substrate', status: 'succeeded', runs: 9, runSeconds: 0, queueSeconds: 0 },
    ],
    active: [],
  },
  '/api/platform/model-usage?days=30': { days: [] },
  '/api/platform/captain-history?days=30': { days: [], failures: [], live: [] },
  '/api/platform/substrate/inventory': { workers: [], actors: [] },
  '/api/platform/github': { app: { id: 1, slug: 'gavel-dev' }, installations: [] },
};

beforeEach(() => {
  vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL) => {
    const url = String(input);
    if (!(url in bodies)) throw new Error(`unexpected fetch ${url}`);
    return Promise.resolve(new Response(JSON.stringify(bodies[url]), { status: 200 }));
  }));
});

afterEach(() => {
  vi.unstubAllGlobals();
});

it('shows the organization\'s summary and only its rows in each panel', async () => {
  render(<OrganizationView org="o-alpha" onBack={() => {}} />, { wrapper: queryTestWrapper() });

  expect(await screen.findByRole('heading', { name: 'Alpha alpha' })).toBeTruthy();
  expect(screen.getByText('2 of 4 members')).toBeTruthy();
  expect(screen.getByText('3 repositories (1 pushed)')).toBeTruthy();
  expect(await screen.findByText('2 runs'), 'Beta\'s 9 runs are left out').toBeTruthy();
});

it('goes back to every organization', async () => {
  const onBack = vi.fn();
  render(<OrganizationView org="o-alpha" onBack={onBack} />, { wrapper: queryTestWrapper() });

  fireEvent.click(await screen.findByRole('button', { name: 'All organizations' }));
  expect(onBack).toHaveBeenCalled();
});

it('says when gavel has no such organization', async () => {
  render(<OrganizationView org="o-missing" onBack={() => {}} />, { wrapper: queryTestWrapper() });

  expect((await screen.findByRole('alert')).textContent).toBe('gavel has no organization o-missing in this Clerk instance.');
});
