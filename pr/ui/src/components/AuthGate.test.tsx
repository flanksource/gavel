import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { ReactNode } from 'react';
import { AuthGate } from './AuthGate';

// Clerk's components are stand-ins: these tests cover which one AuthGate shows
// for each sign-in state, not Clerk's own behaviour.
const clerk = vi.hoisted(() => ({ signedIn: false, organization: null as null | { id: string } }));
vi.mock('@clerk/react', () => ({
  Show: ({ when, fallback, children }: { when: string; fallback?: ReactNode; children: ReactNode }) =>
    (when === 'signed-in') === clerk.signedIn ? <>{children}</> : <>{fallback}</>,
  SignIn: (props: { routing?: string }) => <div data-testid="sign-in" data-routing={props.routing} />,
  OrganizationList: (props: { hidePersonal?: boolean; afterSelectOrganizationUrl?: string }) => (
    <div data-testid="organization-list" data-hide-personal={String(props.hidePersonal)} data-after={props.afterSelectOrganizationUrl} />
  ),
  useOrganization: () => ({ organization: clerk.organization, isLoaded: true }),
}));

describe('AuthGate', () => {
  beforeEach(() => {
    clerk.signedIn = false;
    clerk.organization = null;
  });

  it('asks a signed-out visitor to sign in', () => {
    render(<AuthGate><div>dashboard</div></AuthGate>);
    expect(screen.getByTestId('sign-in').dataset.routing).toBe('hash');
    expect(screen.queryByText('dashboard')).toBeNull();
  });

  it('asks a signed-in user without an active organization to pick one', () => {
    clerk.signedIn = true;
    render(<AuthGate><div>dashboard</div></AuthGate>);
    const list = screen.getByTestId('organization-list');
    expect(list.dataset.hidePersonal).toBe('true');
    expect(list.dataset.after).toBe('/');
    expect(screen.queryByText('dashboard')).toBeNull();
  });

  it('shows the dashboard once an organization is active', () => {
    clerk.signedIn = true;
    clerk.organization = { id: 'org_a' };
    render(<AuthGate><div>dashboard</div></AuthGate>);
    expect(screen.getByText('dashboard')).toBeTruthy();
  });
});
