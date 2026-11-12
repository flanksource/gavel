// Tests the platform dashboard's two views: every organization, from which one opens, and one
// organization, from which every organization is a click back; and what everyone else sees.
import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { queryTestWrapper } from '../todos/queryTestWrapper';
import { PlatformView } from './PlatformView';

let platformAdmin: boolean;

beforeEach(() => {
  platformAdmin = true;
  vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL) => {
    if (String(input) === '/api/me') {
      return Promise.resolve(new Response(JSON.stringify({ user: { id: 'u1', name: 'Ada', email: 'ada@example.com', githubLogin: 'ada', platformAdmin },
        org: { id: 'o-alpha', slug: 'alpha', name: 'Alpha', role: 'org:admin' }, orgs: [] }), { status: 200 }));
    }
    if (String(input) === '/api/platform/tenants?days=30') {
      return Promise.resolve(new Response(JSON.stringify({ tenants: [{ id: 'o-alpha', clerkOrgId: 'org_alpha', slug: 'alpha', name: 'Alpha', createdAt: '2026-10-01T00:00:00Z',
        members: 1, repos: 0, pushedRepos: 0, githubInstallations: 0, sandboxRuns: 0, inGavel: true }] }), { status: 200 }));
    }
    return Promise.resolve(new Response(JSON.stringify({ error: 'not configured' }), { status: 404 }));
  }));
});

afterEach(() => {
  vi.unstubAllGlobals();
});

it('opens an organization from the tenant list', async () => {
  const onSelectOrg = vi.fn();
  render(<PlatformView selectedOrg="" onSelectOrg={onSelectOrg} />, { wrapper: queryTestWrapper() });

  fireEvent.click(await screen.findByRole('button', { name: 'Alpha' }));
  expect(onSelectOrg).toHaveBeenCalledWith('o-alpha');
});

it('shows the selected organization, and goes back to every organization', async () => {
  const onSelectOrg = vi.fn();
  render(<PlatformView selectedOrg="o-alpha" onSelectOrg={onSelectOrg} />, { wrapper: queryTestWrapper() });

  expect(await screen.findByRole('heading', { name: 'Alpha alpha' })).toBeTruthy();
  fireEvent.click(screen.getByRole('button', { name: 'All organizations' }));
  expect(onSelectOrg).toHaveBeenCalledWith('');
});

it('tells anyone who isn\'t a platform admin how to become one, once, and loads no panel', async () => {
  platformAdmin = false;
  render(<PlatformView selectedOrg="" onSelectOrg={() => {}} />, { wrapper: queryTestWrapper() });

  expect((await screen.findByRole('heading')).textContent).toBe('Platform admins only');
  expect(screen.getByText('{ "gavel_role": "platform_admin" }')).toBeTruthy();
  const urls = vi.mocked(fetch).mock.calls.map(([input]) => String(input));
  expect(urls.filter(url => url.startsWith('/api/platform/'))).toEqual([]);
});
