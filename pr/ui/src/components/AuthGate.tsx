// Gates the dashboard in clerk mode: signed-out visitors sign in, and signed-in
// users pick an organization before any org-scoped request is made.
import type { ReactNode } from 'react';
import { OrganizationList, Show, SignIn, useOrganization } from '@clerk/react';

// AuthGate wraps the whole app, outside App, so pages App returns early for
// (processes, new todo) are gated too. Pending sessions count as signed out,
// so Clerk's sign-in finishes their tasks, such as choosing an organization.
export function AuthGate({ children }: { children: ReactNode }) {
  return (
    <Show when="signed-in" fallback={<Centered><SignIn routing="hash" /></Centered>}>
      <ActiveOrganization>{children}</ActiveOrganization>
    </Show>
  );
}

function ActiveOrganization({ children }: { children: ReactNode }) {
  const { organization } = useOrganization();
  if (!organization) {
    return (
      <Centered>
        <OrganizationList hidePersonal afterSelectOrganizationUrl="/" />
      </Centered>
    );
  }
  return <>{children}</>;
}

function Centered({ children }: { children: ReactNode }) {
  return <div className="flex min-h-screen items-center justify-center p-4">{children}</div>;
}
