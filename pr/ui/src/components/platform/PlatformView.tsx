// The platform dashboard: how every organization uses hosted gavel, for platform admins.
// Each panel loads its own source, so one source's outage leaves the others readable.
import { GitHubPanel } from './GitHubPanel';
import { InfrastructurePanel } from './InfrastructurePanel';
import { SandboxUsagePanel } from './SandboxUsagePanel';
import { SubstrateMetricsPanel } from './SubstrateMetricsPanel';
import { TenantsPanel } from './TenantsPanel';

export function PlatformView() {
  return (
    <div className="flex h-full min-h-0 flex-col gap-6 overflow-y-auto p-4 text-sm">
      <TenantsPanel />
      <SandboxUsagePanel />
      <SubstrateMetricsPanel />
      <GitHubPanel />
      <InfrastructurePanel />
    </div>
  );
}
