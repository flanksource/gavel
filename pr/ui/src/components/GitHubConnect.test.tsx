import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { setAuthMode } from '../query';
import { ConnectGitHubButton, GitHubSetupPage } from './GitHubConnect';

// Clerk's hook is a stand-in: these tests cover what the components send with the
// token Clerk issues, not how Clerk issues it.
vi.mock('@clerk/react', () => ({
  useAuth: () => ({ getToken: async () => 'fresh-token' }),
}));

function stubFetch(status: number, payload: unknown) {
  const fetchMock = vi.fn(async (_url: string, _init?: RequestInit) => new Response(JSON.stringify(payload), { status }));
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

afterEach(() => {
  setAuthMode('none');
  vi.unstubAllGlobals();
});

describe('ConnectGitHubButton', () => {
  it("asks the server for GitHub's install page with a fresh token and goes there", async () => {
    const fetchMock = stubFetch(200, { url: 'https://github.com/apps/gavel/installations/new?state=s' });
    const navigate = vi.fn();
    render(<ConnectGitHubButton navigate={navigate} />);

    fireEvent.click(screen.getByRole('button', { name: 'Connect GitHub' }));

    await waitFor(() => expect(navigate).toHaveBeenCalledWith('https://github.com/apps/gavel/installations/new?state=s'));
    expect(fetchMock).toHaveBeenCalledWith('/api/github/install', {
      method: 'POST',
      headers: { Authorization: 'Bearer fresh-token' },
    });
  });

  it("shows why GitHub can't be connected", async () => {
    stubFetch(403, { error: 'sign in to connect GitHub' });
    const navigate = vi.fn();
    render(<ConnectGitHubButton navigate={navigate} />);

    fireEvent.click(screen.getByRole('button', { name: 'Connect GitHub' }));

    expect(await screen.findByText(/sign in to connect GitHub/)).toBeTruthy();
    expect(navigate).not.toHaveBeenCalled();
  });
});

describe('GitHubSetupPage', () => {
  it("links the installation GitHub reported and lists the organization's repositories", async () => {
    const fetchMock = stubFetch(200, { repositories: ['acme/api', 'acme/web'] });
    render(<GitHubSetupPage search="?installation_id=42&setup_action=install&state=signed" />);

    expect(await screen.findByText('acme/api')).toBeTruthy();
    expect(screen.getByText('acme/web')).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Open the dashboard' }).getAttribute('href')).toBe('/');
    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [url, init] = fetchMock.mock.calls[0];
    expect(url).toBe('/api/github/setup');
    expect(init).toMatchObject({ method: 'POST', headers: { Authorization: 'Bearer fresh-token', 'Content-Type': 'application/json' } });
    expect(JSON.parse(String(init?.body))).toEqual({ installation_id: '42', state: 'signed' });
  });

  it('says so when the installation shares no repositories', async () => {
    stubFetch(200, { repositories: [] });
    render(<GitHubSetupPage search="?installation_id=42&state=signed" />);
    expect(await screen.findByText(/shares no repositories/)).toBeTruthy();
  });

  it('shows the reason a link was refused, with a way back', async () => {
    stubFetch(409, { error: 'installation 42 is linked to another organization' });
    render(<GitHubSetupPage search="?installation_id=42&state=signed" />);
    expect(await screen.findByText(/linked to another organization/)).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Back to the dashboard' }).getAttribute('href')).toBe('/');
  });
});
