import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { ReactNode } from 'react';
import { setAuthMode } from '../query';
import { UserMenu } from './UserMenu';

// Clerk's components are stand-ins: these tests cover what UserMenu renders in
// each auth mode, not Clerk's own behaviour.
vi.mock('@clerk/react', () => {
  const UserButton = ({ children }: { children?: ReactNode }) => <div data-testid="user-button">{children}</div>;
  UserButton.UserProfilePage = ({ label, url }: { label: string; url: string }) => (
    <div data-testid="profile-page" data-label={label} data-url={url} />
  );
  const OrganizationSwitcher = (props: { hidePersonal?: boolean; afterSelectOrganizationUrl?: string; children?: ReactNode }) => (
    <div data-testid="organization-switcher" data-hide-personal={String(props.hidePersonal)} data-after={props.afterSelectOrganizationUrl}>
      {props.children}
    </div>
  );
  OrganizationSwitcher.OrganizationProfilePage = ({ label, url }: { label: string; url: string }) => (
    <div data-testid="organization-profile-page" data-label={label} data-url={url} />
  );
  return {
    OrganizationSwitcher,
    UserButton,
    useAuth: () => ({ getToken: async () => 'fresh-token' }),
  };
});

function renderMenu() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <UserMenu />
    </QueryClientProvider>,
  );
}

function stubMe(user: { name: string } | null) {
  const fetchMock = vi.fn(async () => new Response(JSON.stringify({ user, org: { slug: 'local' }, orgs: [] })));
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

afterEach(() => {
  setAuthMode('none');
  vi.unstubAllGlobals();
});

describe('UserMenu', () => {
  it('shows the local login in none mode, without an org switcher', async () => {
    const fetchMock = stubMe({ name: 'Ada Lovelace' });
    renderMenu();
    expect(await screen.findByText('Ada Lovelace')).toBeTruthy();
    expect(fetchMock).toHaveBeenCalledWith('/api/me', { signal: expect.any(AbortSignal) });
    expect(screen.queryByTestId('organization-switcher')).toBeNull();
    expect(screen.queryByRole('button', { name: 'Connect GitHub' })).toBeNull();
  });

  it('shows "local" in none mode when nobody is logged in', async () => {
    stubMe(null);
    renderMenu();
    expect(await screen.findByText('local')).toBeTruthy();
  });

  it('shows the Clerk org switcher and user button in clerk mode', () => {
    const fetchMock = stubMe(null);
    setAuthMode('clerk', vi.fn());
    renderMenu();
    const switcher = screen.getByTestId('organization-switcher');
    expect(switcher.dataset.hidePersonal).toBe('true');
    expect(switcher.dataset.after).toBe('/');
    expect(screen.getByTestId('user-button')).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Connect GitHub' })).toBeTruthy();
    const page = screen.getByTestId('profile-page');
    expect(page.dataset.label).toBe('SSH keys');
    expect(page.dataset.url).toBe('ssh-keys');
    const orgPage = screen.getByTestId('organization-profile-page');
    expect(orgPage.dataset.label).toBe('Model keys');
    expect(orgPage.dataset.url).toBe('model-keys');
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
