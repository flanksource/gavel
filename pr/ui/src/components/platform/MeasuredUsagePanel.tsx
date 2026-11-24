// An organization's measured CPU and memory, sampled from substrate's Prometheus through Mission
// Control: platform admins see any organization's, an organization's admins their own.
import { useQuery } from '@tanstack/react-query';
import { meQuery } from '../oneShotQueries';
import { MetricsViewPanel } from './SubstrateMetricsPanel';

const sampled = 'What the organization\'s environments used, sampled from substrate every 15 seconds, so a short burst may not show. Its budgets count each run\'s size for as long as it runs instead.';

// MeasuredUsagePanel shows org's measured use to a platform admin.
export function MeasuredUsagePanel({ org }: { org: string }) {
  return (
    <MetricsViewPanel
      heading="Measured CPU and memory"
      source={`measured-usage/${org}`}
      url={`/api/platform/orgs/${encodeURIComponent(org)}/measured-usage`}
      context="load the organization's measured CPU and memory"
      note={sampled}
    />
  );
}

// OwnMeasuredUsage shows an organization admin their organization's measured use, and members nothing.
export function OwnMeasuredUsage() {
  const me = useQuery(meQuery());
  if (me.data?.org.role !== 'org:admin') return null;
  return (
    <MetricsViewPanel
      heading="Measured CPU and memory"
      source="measured-usage"
      url="/api/usage/measured"
      context="load the organization's measured CPU and memory"
      note={sampled}
    />
  );
}
