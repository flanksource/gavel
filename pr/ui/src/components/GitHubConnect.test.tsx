import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
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

// route answers fetch by "METHOD url", recording each request it was sent.
function route(answers: Record<string, () => [number, unknown]>) {
  const calls: { key: string; init?: RequestInit }[] = [];
  const fetchMock = vi.fn(async (url: string, init?: RequestInit) => {
    const key = `${init?.method ?? 'GET'} ${url}`;
    calls.push({ key, init });
    const answer = answers[key];
    if (!answer) return new Response(JSON.stringify({ error: `unexpected ${key}` }), { status: 500 });
    const [status, payload] = answer();
    return new Response(JSON.stringify(payload), { status });
  });
  vi.stubGlobal('fetch', fetchMock);
  return calls;
}

function renderButton(navigate = vi.fn()) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <QueryClientProvider client={client}>
      <ConnectGitHubButton navigate={navigate} />
    </QueryClientProvider>,
  );
  fireEvent.click(screen.getByRole('button', { name: 'Connect GitHub' }));
  return navigate;
}

describe('ConnectGitHubButton', () => {
  it("lists the installations the user's GitHub account reaches and links one", async () => {
    let linked = false;
    const calls = route({
      'GET /api/github/installations': () => [200, [
        { id: 42, account: 'acme', accountType: 'Organization', status: 'linked_elsewhere' },
        { id: 44, account: 'ada', accountType: 'User', status: linked ? 'linked' : 'available' },
      ]],
      'POST /api/github/installations/44/link': () => { linked = true; return [200, { repositories: ['ada/tools'] }]; },
      'POST /api/prs/refresh': () => [202, {}],
    });
    renderButton();

    expect(await screen.findByText('Linked to another gavel organization')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Link ada' }));

    expect(await screen.findByText('Linked to this organization')).toBeTruthy();
    const link = calls.find(call => call.key === 'POST /api/github/installations/44/link');
    expect(link?.init?.headers).toEqual({ Authorization: 'Bearer fresh-token' });
    expect(calls.map(call => call.key)).toContain('POST /api/prs/refresh');
  });

  it('says what to do when the GitHub account reaches no installation', async () => {
    route({ 'GET /api/github/installations': () => [200, []] });
    renderButton();
    expect(await screen.findByText(/can't reach any installation/)).toBeTruthy();
  });

  it("shows why installations can't be listed or linked", async () => {
    route({
      'GET /api/github/installations': () => [200, [{ id: 44, account: 'ada', accountType: 'User', status: 'available' }]],
      'POST /api/github/installations/44/link': () => [409, { error: 'installation 44 is linked to another organization' }],
    });
    renderButton();
    fireEvent.click(await screen.findByRole('button', { name: 'Link ada' }));
    expect(await screen.findByText(/linked to another organization/)).toBeTruthy();
  });

  it("sends the user to GitHub's install page for another account, with a fresh token", async () => {
    const calls = route({
      'GET /api/github/installations': () => [200, []],
      'POST /api/github/install': () => [200, { url: 'https://github.com/apps/gavel/installations/new?state=s' }],
    });
    const navigate = renderButton();

    fireEvent.click(await screen.findByRole('button', { name: 'Install on another GitHub account' }));

    await waitFor(() => expect(navigate).toHaveBeenCalledWith('https://github.com/apps/gavel/installations/new?state=s'));
    expect(calls.find(call => call.key === 'POST /api/github/install')?.init?.headers).toEqual({ Authorization: 'Bearer fresh-token' });
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

  it('explains the page when it is opened without what GitHub sends', async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);
    render(<GitHubSetupPage search="" />);
    expect(await screen.findByText(/GitHub returns here after the gavel App is installed/)).toBeTruthy();
    expect(fetchMock).not.toHaveBeenCalled();
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
