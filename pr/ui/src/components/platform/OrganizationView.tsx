// One organization's view of the platform dashboard: its summary from the tenant list, and the
// panels that can show one organization, each with only that organization's rows.
import { StatStrip } from '@flanksource/clicky-ui/data';
import { CaptainHistoryPanel } from './CaptainHistoryPanel';
import { GitHubPanel } from './GitHubPanel';
import { ModelUsagePanel } from './ModelUsagePanel';
import { SandboxUsagePanel } from './SandboxUsagePanel';
import { SubstrateInventoryPanel } from './SubstrateInventoryPanel';
import { tenantWindowDays, useTenants } from './TenantsPanel';

export function OrganizationView({ org, onBack }: { org: string; onBack: () => void }) {
  const tenants = useTenants();
  const tenant = tenants.data?.tenants.find(t => t.id === org);

  return (
    <div className="flex h-full min-h-0 flex-col gap-6 overflow-y-auto p-4 text-sm">
      <button type="button" className="self-start text-xs underline text-muted-foreground" onClick={onBack}>All organizations</button>
      {tenants.error && <p role="alert" className="text-red-600">{tenants.error.message}</p>}
      {tenants.data && !tenant && <p role="alert" className="text-red-600">gavel has no organization {org} in this Clerk instance.</p>}
      {tenant && (
        <>
          <section className="flex flex-col gap-2">
            <h2 className="text-base font-semibold">
              {tenant.name || tenant.clerkOrgId} <span className="text-xs font-normal text-muted-foreground">{tenant.slug}</span>
            </h2>
            <StatStrip
              items={[
                { label: 'Members', value: tenant.clerkMembers === undefined ? `${tenant.members} members` : `${tenant.members} of ${tenant.clerkMembers} members` },
                { label: 'Repositories', value: tenant.pushedRepos > 0 ? `${tenant.repos} repositories (${tenant.pushedRepos} pushed)` : `${tenant.repos} repositories` },
                { label: 'GitHub', value: `${tenant.githubInstallations} GitHub installations` },
                { label: 'Last run', value: tenant.lastRunAt ? `Last run ${new Date(tenant.lastRunAt).toLocaleString()}` : `No runs in ${tenantWindowDays} days` },
              ]}
            />
          </section>
          <SandboxUsagePanel org={org} />
          <ModelUsagePanel org={org} />
          <CaptainHistoryPanel org={org} />
          <SubstrateInventoryPanel org={org} />
          <GitHubPanel org={org} />
        </>
      )}
    </div>
  );
}
