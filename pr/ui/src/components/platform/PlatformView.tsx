// The platform dashboard: how every organization uses hosted gavel, or one of them, for platform
// admins. Each panel loads its own source, so one source's outage leaves the others readable.
import { useQuery } from '@tanstack/react-query';
import { meQuery } from '../oneShotQueries';
import { BudgetsPanel } from './BudgetsPanel';
import { CaptainHistoryPanel } from './CaptainHistoryPanel';
import { GitHubPanel } from './GitHubPanel';
import { InfrastructurePanel, infrastructureSubsections, useInfrastructure } from './InfrastructurePanel';
import { ModelUsagePanel } from './ModelUsagePanel';
import { OrganizationView } from './OrganizationView';
import { PlatformSections } from './PlatformSections';
import { SandboxUsagePanel } from './SandboxUsagePanel';
import { SubstrateInventoryPanel } from './SubstrateInventoryPanel';
import { SubstrateMetricsPanel } from './SubstrateMetricsPanel';
import { TenantsPanel } from './TenantsPanel';

// PlatformView shows every organization, or only selectedOrg when one is selected; onSelectOrg
// selects one, or every organization again with "". Anyone who isn't a platform admin is told how
// to become one instead, and no panel loads.
export function PlatformView({ selectedOrg, onSelectOrg }: { selectedOrg: string; onSelectOrg: (org: string) => void }) {
  const me = useQuery(meQuery());
  if (me.error) return <p role="alert" className="p-4 text-sm text-red-600">{me.error.message}</p>;
  if (!me.data) return null;
  if (!me.data.user?.platformAdmin) return <PlatformAdminsOnly />;
  if (selectedOrg) return <OrganizationView org={selectedOrg} onBack={() => onSelectOrg('')} />;
  return <EveryOrganization onSelectOrg={onSelectOrg} />;
}

// EveryOrganization is the dashboard of every organization, with the infrastructure panel's parts
// listed under it in the sidebar.
function EveryOrganization({ onSelectOrg }: { onSelectOrg: (org: string) => void }) {
  const infrastructure = useInfrastructure();
  return (
    <PlatformSections
      sections={[
        { id: 'tenants', label: 'Tenants', content: <TenantsPanel onSelect={onSelectOrg} /> },
        { id: 'sandbox-usage', label: 'Sandbox usage', content: <SandboxUsagePanel /> },
        { id: 'substrate-inventory', label: 'Substrate inventory', content: <SubstrateInventoryPanel /> },
        { id: 'substrate-metrics', label: 'Substrate metrics', content: <SubstrateMetricsPanel /> },
        { id: 'model-usage', label: 'Model usage', content: <ModelUsagePanel /> },
        { id: 'budgets', label: 'Budgets', content: <BudgetsPanel /> },
        { id: 'captain-history', label: 'Captain history', content: <CaptainHistoryPanel /> },
        { id: 'github-app', label: 'GitHub App', content: <GitHubPanel /> },
        { id: 'infrastructure', label: 'Infrastructure', content: <InfrastructurePanel />, subsections: infrastructureSubsections(infrastructure.data) },
      ]}
    />
  );
}

function PlatformAdminsOnly() {
  return (
    <section className="flex flex-col gap-3 p-4 text-sm">
      <h2 className="text-base font-semibold">Platform admins only</h2>
      <p>The platform dashboard shows every organization's use of gavel, so only platform admins can open it.</p>
      <p>To become one, add this to your user's public metadata in Clerk (Users → your user → Metadata):</p>
      <pre className="self-start rounded border border-border bg-muted px-3 py-2 font-mono text-xs">{'{ "gavel_role": "platform_admin" }'}</pre>
      <p className="text-muted-foreground">gavel reads it within 5 minutes; reload this page then.</p>
    </section>
  );
}
