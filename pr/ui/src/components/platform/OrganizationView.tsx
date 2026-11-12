// One organization's view of the platform dashboard: its summary from the tenant list, and the
// panels that can show one organization, each with only that organization's rows.
import { StatStrip } from '@flanksource/clicky-ui/data';
import { CaptainHistoryPanel } from './CaptainHistoryPanel';
import { GitHubPanel } from './GitHubPanel';
import { ModelUsagePanel } from './ModelUsagePanel';
import { PlatformSections } from './PlatformSections';
import { SandboxUsagePanel } from './SandboxUsagePanel';
import { SubstrateInventoryPanel } from './SubstrateInventoryPanel';
import { tenantWindowDays, useTenants } from './TenantsPanel';
import { formatDateTime } from './dates';

export function OrganizationView({ org, onBack }: { org: string; onBack: () => void }) {
  const tenants = useTenants();
  const tenant = tenants.data?.tenants.find(t => t.id === org);

  const summary = tenant && (
    <section className="flex flex-col gap-2">
      <h2 className="text-xl font-semibold">
        {tenant.name || tenant.clerkOrgId} <span className="text-xs font-normal text-muted-foreground">{tenant.slug}</span>
      </h2>
      <StatStrip
        items={[
          { label: 'Members', value: tenant.clerkMembers === undefined ? `${tenant.members} members` : `${tenant.members} of ${tenant.clerkMembers} members` },
          { label: 'Repositories', value: tenant.pushedRepos > 0 ? `${tenant.repos} repositories (${tenant.pushedRepos} pushed)` : `${tenant.repos} repositories` },
          { label: 'GitHub', value: `${tenant.githubInstallations} GitHub installations` },
          { label: 'Last run', value: tenant.lastRunAt ? `Last run ${formatDateTime(tenant.lastRunAt)}` : `No runs in ${tenantWindowDays} days` },
        ]}
      />
    </section>
  );

  return (
    <PlatformSections
      header={(
        <>
          <button type="button" className="self-start text-xs underline text-muted-foreground" onClick={onBack}>All organizations</button>
          {tenants.error && <p role="alert" className="text-red-600">{tenants.error.message}</p>}
          {tenants.data && !tenant && <p role="alert" className="text-red-600">gavel has no organization {org} in this Clerk instance.</p>}
        </>
      )}
      sections={tenant
        ? [
            { id: 'summary', label: tenant.name || tenant.clerkOrgId, content: summary },
            { id: 'sandbox-usage', label: 'Sandbox usage', content: <SandboxUsagePanel org={org} /> },
            { id: 'model-usage', label: 'Model usage', content: <ModelUsagePanel org={org} /> },
            { id: 'captain-history', label: 'Captain history', content: <CaptainHistoryPanel org={org} /> },
            { id: 'substrate-inventory', label: 'Substrate inventory', content: <SubstrateInventoryPanel org={org} /> },
            { id: 'github-app', label: 'GitHub App', content: <GitHubPanel org={org} /> },
          ]
        : []}
    />
  );
}
