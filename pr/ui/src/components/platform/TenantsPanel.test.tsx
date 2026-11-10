// Tests the platform dashboard's tenant panel: every organization with what it uses, those
// only one of gavel and Clerk knows marked, and the reason when a source can't answer.
import { render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { queryTestWrapper } from '../todos/queryTestWrapper';
import { TenantsPanel, type PlatformTenants } from './TenantsPanel';

let response: { status: number; body: unknown };

beforeEach(() => {
  vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL) => {
    const url = String(input);
    if (url !== '/api/platform/tenants?days=30') throw new Error(`unexpected fetch ${url}`);
    return Promise.resolve(new Response(JSON.stringify(response.body), { status: response.status }));
  }));
});

afterEach(() => {
  vi.unstubAllGlobals();
});

const usage = { members: 0, repos: 0, pushedRepos: 0, githubInstallations: 0, sandboxRuns: 0, createdAt: '2026-10-01T00:00:00Z' };

it('lists every organization with what it uses', async () => {
  const tenants: PlatformTenants = {
    tenants: [
      { ...usage, id: 'o-alpha', clerkOrgId: 'org_alpha', slug: 'alpha', name: 'Alpha', members: 2, repos: 3, pushedRepos: 1, githubInstallations: 1, sandboxRuns: 12, lastRunAt: '2026-11-09T10:00:00Z', inGavel: true, clerkMembers: 4 },
      { ...usage, clerkOrgId: 'org_beta', slug: 'beta', name: 'Beta', inGavel: false, clerkMembers: 1 },
      { ...usage, id: 'o-gamma', clerkOrgId: 'org_gamma', slug: 'gamma', name: 'Gamma', members: 1, inGavel: true },
    ],
  };
  response = { status: 200, body: tenants };
  render(<TenantsPanel />, { wrapper: queryTestWrapper() });

  const alpha = await screen.findByRole('row', { name: /Alpha/ });
  expect(within(alpha).getByText('2 of 4'), 'members who used gavel, of Clerk\'s').toBeTruthy();
  expect(within(alpha).getByText('3 (1 pushed)')).toBeTruthy();
  expect(within(alpha).getByText('12')).toBeTruthy();
  expect(within(screen.getByRole('row', { name: /Beta/ })).getByText('Clerk only')).toBeTruthy();
  expect(within(screen.getByRole('row', { name: /Gamma/ })).getByText('Not in Clerk')).toBeTruthy();
  expect(screen.getByRole('heading', { name: 'Tenants · 3' })).toBeTruthy();
});

it('says why Clerk\'s organizations are missing', async () => {
  response = {
    status: 200,
    body: { tenants: [{ ...usage, id: 'o-alpha', clerkOrgId: 'org_alpha', slug: 'alpha', name: 'Alpha', inGavel: true }], clerkError: 'list the organizations in Clerk: Clerk Backend API returned HTTP 401: check the Clerk secret key' },
  };
  render(<TenantsPanel />, { wrapper: queryTestWrapper() });

  expect((await screen.findByRole('alert')).textContent).toContain('check the Clerk secret key');
  expect(within(screen.getByRole('row', { name: /Alpha/ })).queryByText('Not in Clerk'), 'unknown while Clerk is unreachable').toBeNull();
});

it('tells a non-admin how to become a platform admin', async () => {
  response = { status: 403, body: { error: 'not_platform_admin', message: 'the platform dashboard is for platform admins: set "gavel_role": "platform_admin"' } };
  render(<TenantsPanel />, { wrapper: queryTestWrapper() });

  expect((await screen.findByRole('alert')).textContent).toContain('"gavel_role": "platform_admin"');
});
