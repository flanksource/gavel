// The platform dashboard's tenant panel: every organization of the Clerk instance with its
// members, repositories, GitHub App installations and recent sandbox runs.
import { useQuery } from '@tanstack/react-query';
import { fetchJSON, queryKeys } from '../../query';
import { PanelHeading } from './PanelHeading';

export interface PlatformTenant {
  id?: string;
  clerkOrgId: string;
  slug: string;
  name: string;
  createdAt: string;
  members: number;
  repos: number;
  pushedRepos: number;
  githubInstallations: number;
  sandboxRuns: number;
  lastRunAt?: string;
  inGavel: boolean;
  clerkMembers?: number;
}

export interface PlatformTenants {
  tenants: PlatformTenant[];
  clerkError?: string;
}

// tenantWindowDays is how far back the panel counts sandbox runs.
export const tenantWindowDays = 30;

function members(tenant: PlatformTenant): string {
  return tenant.clerkMembers === undefined ? `${tenant.members}` : `${tenant.members} of ${tenant.clerkMembers}`;
}

function repos(tenant: PlatformTenant): string {
  return tenant.pushedRepos > 0 ? `${tenant.repos} (${tenant.pushedRepos} pushed)` : `${tenant.repos}`;
}

// presence marks an organization only one of gavel and Clerk knows. Without Clerk's list
// (clerkError), whether Clerk still has an organization is unknown.
function presence(tenant: PlatformTenant, clerkListed: boolean): string {
  if (!tenant.inGavel) return 'Clerk only';
  if (clerkListed && tenant.clerkMembers === undefined) return 'Not in Clerk';
  return '';
}

// useTenants reads every organization, as the tenant panel and an organization's view show them.
export function useTenants() {
  return useQuery({
    queryKey: queryKeys.platform('tenants', tenantWindowDays),
    queryFn: ({ signal }) => fetchJSON<PlatformTenants>({ url: `/api/platform/tenants?days=${tenantWindowDays}`, signal, context: 'load tenants' }),
  });
}

// TenantsPanel lists every organization; onSelect opens one gavel knows.
export function TenantsPanel({ onSelect }: { onSelect?: (org: string) => void } = {}) {
  const tenants = useTenants();
  const clerkListed = !tenants.data?.clerkError;

  return (
    <section className="flex flex-col gap-2">
      <PanelHeading>
        Tenants{tenants.data ? ` · ${tenants.data.tenants.length}` : ''}
      </PanelHeading>
      {tenants.error && <p role="alert" className="text-red-600">{tenants.error.message}</p>}
      {tenants.data?.clerkError && <p role="alert" className="text-amber-600">{tenants.data.clerkError}</p>}
      {!!tenants.data?.tenants.length && (
        <table className="w-full border-collapse text-left text-xs">
          <thead className="text-muted-foreground">
            <tr>
              {['Organization', 'Members (gavel of Clerk)', 'Repositories', 'GitHub installs', `Runs (${tenantWindowDays}d)`, 'Last run', 'Created', ''].map(heading => (
                <th key={heading} className="border-b border-border px-2 py-1.5 font-medium">{heading}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {tenants.data.tenants.map(tenant => (
              <tr key={tenant.clerkOrgId} className="border-b border-border">
                <td className="px-2 py-1.5">
                  {tenant.id && onSelect ? (
                    <button type="button" className="font-medium underline" onClick={() => onSelect(tenant.id!)}>{tenant.name || tenant.clerkOrgId}</button>
                  ) : (
                    <span className="font-medium">{tenant.name || tenant.clerkOrgId}</span>
                  )}
                  {tenant.slug && <span className="ml-1 text-muted-foreground">{tenant.slug}</span>}
                </td>
                <td className="px-2 py-1.5">{members(tenant)}</td>
                <td className="px-2 py-1.5">{repos(tenant)}</td>
                <td className="px-2 py-1.5">{tenant.githubInstallations}</td>
                <td className="px-2 py-1.5">{tenant.sandboxRuns}</td>
                <td className="px-2 py-1.5 text-muted-foreground">{tenant.lastRunAt ? new Date(tenant.lastRunAt).toLocaleString() : '—'}</td>
                <td className="px-2 py-1.5 text-muted-foreground">{new Date(tenant.createdAt).toLocaleDateString()}</td>
                <td className="px-2 py-1.5 text-muted-foreground">{presence(tenant, clerkListed)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  );
}
