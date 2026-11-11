// The platform dashboard: how every organization uses hosted gavel, or one of them, for platform
// admins. Each panel loads its own source, so one source's outage leaves the others readable.
import { BudgetsPanel } from './BudgetsPanel';
import { CaptainHistoryPanel } from './CaptainHistoryPanel';
import { GitHubPanel } from './GitHubPanel';
import { InfrastructurePanel } from './InfrastructurePanel';
import { ModelUsagePanel } from './ModelUsagePanel';
import { OrganizationView } from './OrganizationView';
import { SandboxUsagePanel } from './SandboxUsagePanel';
import { SubstrateInventoryPanel } from './SubstrateInventoryPanel';
import { SubstrateMetricsPanel } from './SubstrateMetricsPanel';
import { TenantsPanel } from './TenantsPanel';

// PlatformView shows every organization, or only selectedOrg when one is selected; onSelectOrg
// selects one, or every organization again with "".
export function PlatformView({ selectedOrg, onSelectOrg }: { selectedOrg: string; onSelectOrg: (org: string) => void }) {
  if (selectedOrg) return <OrganizationView org={selectedOrg} onBack={() => onSelectOrg('')} />;
  return (
    <div className="flex h-full min-h-0 flex-col gap-6 overflow-y-auto p-4 text-sm">
      <TenantsPanel onSelect={onSelectOrg} />
      <SandboxUsagePanel />
      <SubstrateInventoryPanel />
      <SubstrateMetricsPanel />
      <ModelUsagePanel />
      <BudgetsPanel />
      <CaptainHistoryPanel />
      <GitHubPanel />
      <InfrastructurePanel />
    </div>
  );
}
