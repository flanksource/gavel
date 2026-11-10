import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { ReactNode } from 'react';
import { AuthGate } from './AuthGate';
import { QueryClient, QueryClientProvider, useQueryClient } from '@tanstack/react-query';

// Clerk's components are stand-ins: these tests cover which one AuthGate shows
// for each sign-in state, not Clerk's own behaviour.
const clerk = vi.hoisted(() => ({ signedIn: false, organization: null as null | { id: string }, user: { id: 'user-a' } }));
vi.mock('@clerk/react', () => ({
  Show: ({ when, fallback, children }: { when: string; fallback?: ReactNode; children: ReactNode }) =>
    (when === 'signed-in') === clerk.signedIn ? <>{children}</> : <>{fallback}</>,
  SignIn: (props: { routing?: string }) => <div data-testid="sign-in" data-routing={props.routing} />,
  OrganizationList: (props: { hidePersonal?: boolean; afterSelectOrganizationUrl?: string }) => (
    <div data-testid="organization-list" data-hide-personal={String(props.hidePersonal)} data-after={props.afterSelectOrganizationUrl} />
  ),
  useOrganization: () => ({ organization: clerk.organization, isLoaded: true }),
  useUser: () => ({ user: clerk.user }),
}));

describe('AuthGate', () => {
  beforeEach(() => {
    clerk.signedIn = false;
    clerk.organization = null;
  });

  it('asks a signed-out visitor to sign in', () => {
    render(<QueryClientProvider client={new QueryClient()}><AuthGate><div>dashboard</div></AuthGate></QueryClientProvider>);
    expect(screen.getByTestId('sign-in').dataset.routing).toBe('hash');
    expect(screen.queryByText('dashboard')).toBeNull();
  });

  it('asks a signed-in user without an active organization to pick one', () => {
    clerk.signedIn = true;
    render(<QueryClientProvider client={new QueryClient()}><AuthGate><div>dashboard</div></AuthGate></QueryClientProvider>);
    const list = screen.getByTestId('organization-list');
    expect(list.dataset.hidePersonal).toBe('true');
    expect(list.dataset.after).toBe('/');
    expect(screen.queryByText('dashboard')).toBeNull();
  });

  it('shows the dashboard once an organization is active', () => {
    clerk.signedIn = true;
    clerk.organization = { id: 'org_a' };
    render(<QueryClientProvider client={new QueryClient()}><AuthGate><div>dashboard</div></AuthGate></QueryClientProvider>);
    expect(screen.getByText('dashboard')).toBeTruthy();
  });

  it('clears cached private data when the active organization changes', () => {
    clerk.signedIn = true;
    clerk.organization = { id: 'org_a' };
    const clients: QueryClient[] = [];
    function Probe() {
      const client = useQueryClient();
      clients.push(client);
      return <div>{String(client.getQueryData(['private']) ?? 'empty')}</div>;
    }
    const root = new QueryClient();
    const tree = <QueryClientProvider client={root}><AuthGate><Probe /></AuthGate></QueryClientProvider>;
    const view = render(tree);
    const first = clients.at(-1)!;
    first.setQueryData(['private'], 'confidential');
    clerk.organization = { id: 'org_b' };
    view.rerender(<QueryClientProvider client={root}><AuthGate><Probe /></AuthGate></QueryClientProvider>);
    expect(clients.at(-1)).not.toBe(first);
    expect(first.getQueryData(['private'])).toBeUndefined();
    expect(screen.getByText('empty')).toBeTruthy();
  });
});
