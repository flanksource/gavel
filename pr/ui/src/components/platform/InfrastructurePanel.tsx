// The platform dashboard's infrastructure panel: the config items, health checks and changes
// Mission Control has for gavel's infrastructure, with their health at a glance.
import { useQuery } from '@tanstack/react-query';
import { StatStrip } from '@flanksource/clicky-ui/data';
import { fetchJSON, queryKeys } from '../../query';
import { PanelHeading } from './PanelHeading';

// MissionControlResource is a resource as Mission Control's search returns it.
export interface MissionControlResource {
  id: string;
  agent: string;
  name: string;
  namespace: string;
  type: string;
  health?: string;
  status?: string;
  severity?: string;
  created_at?: string;
  updated_at?: string;
}

// Infrastructure is each list Mission Control has for gavel, or why it couldn't be read.
export interface Infrastructure {
  missionControl: string;
  configs: MissionControlResource[];
  configsError?: string;
  checks: MissionControlResource[];
  checksError?: string;
  changes: MissionControlResource[];
  changesError?: string;
}

const healthClass: Record<string, string> = {
  healthy: 'text-green-600',
  unhealthy: 'text-red-600',
  warning: 'text-amber-600',
};

// healthSummary counts resources by health: always healthy and unhealthy, then any others.
function healthSummary(resources: MissionControlResource[]): string {
  const counts = new Map<string, number>([['healthy', 0], ['unhealthy', 0]]);
  for (const resource of resources) {
    const health = resource.health || 'unknown';
    counts.set(health, (counts.get(health) ?? 0) + 1);
  }
  return [...counts].filter(([health, n]) => n > 0 || health === 'healthy' || health === 'unhealthy').map(([health, n]) => `${n} ${health}`).join(', ');
}

// unread stands in for a list's figures when Mission Control couldn't read it.
const unread = 'couldn\'t be read';

function when(at?: string): string {
  return at ? new Date(at).toLocaleString() : '—';
}

function ResourceTable({ label, error, headings, rows }: { label: string; error?: string; headings: string[]; rows: { id: string; cells: { text: string; className?: string }[] }[] }) {
  if (error) return <p role="alert" className="text-red-600">{error}</p>;
  if (rows.length === 0) return <p className="text-xs text-muted-foreground">No {label.toLowerCase()} match the search.</p>;
  return (
    <table aria-label={label} className="w-full border-collapse text-left text-xs">
      <thead className="text-muted-foreground">
        <tr>
          {headings.map(heading => (
            <th key={heading} className="border-b border-border px-2 py-1.5 font-medium">{heading}</th>
          ))}
        </tr>
      </thead>
      <tbody>
        {rows.map(row => (
          <tr key={row.id} className="border-b border-border">
            {row.cells.map((cell, i) => <td key={i} className={`px-2 py-1.5 ${cell.className ?? ''}`}>{cell.text}</td>)}
          </tr>
        ))}
      </tbody>
    </table>
  );
}

export function InfrastructurePanel() {
  const infrastructure = useQuery({
    queryKey: queryKeys.platform('infrastructure'),
    queryFn: ({ signal }) => fetchJSON<Infrastructure>({ url: '/api/platform/infrastructure', signal, context: 'load gavel\'s infrastructure' }),
  });
  const data = infrastructure.data;
  const health = (resource: MissionControlResource) => ({ text: resource.health || 'unknown', className: healthClass[resource.health ?? ''] });

  return (
    <section className="flex flex-col gap-3">
      <PanelHeading>
        Infrastructure
        {data && <> · <a className="underline" href={data.missionControl.replace(/\/api\/?$/, '')} target="_blank" rel="noreferrer">{data.missionControl}</a></>}
      </PanelHeading>
      {infrastructure.error && <p role="alert" className="text-red-600">{infrastructure.error.message}</p>}
      {data && (
        <>
          <StatStrip
            items={[
              { label: 'Config items', value: data.configsError ? unread : healthSummary(data.configs) },
              { label: 'Health checks', value: data.checksError ? unread : healthSummary(data.checks) },
              { label: 'Changes', value: data.changesError ? unread : `${data.changes.length}` },
            ]}
          />
          <ResourceTable
            label="Config items"
            error={data.configsError}
            headings={['Name', 'Type', 'Health', 'Status', 'Namespace', 'Updated']}
            rows={data.configs.map(config => ({
              id: config.id,
              cells: [{ text: config.name }, { text: config.type }, health(config), { text: config.status ?? '' }, { text: config.namespace }, { text: when(config.updated_at) }],
            }))}
          />
          <ResourceTable
            label="Health checks"
            error={data.checksError}
            headings={['Name', 'Type', 'Health', 'Status', 'Namespace']}
            rows={data.checks.map(check => ({
              id: check.id,
              cells: [{ text: check.name }, { text: check.type }, health(check), { text: check.status ?? '' }, { text: check.namespace }],
            }))}
          />
          <ResourceTable
            label="Recent changes"
            error={data.changesError}
            headings={['Config', 'Change', 'Severity', 'When']}
            rows={data.changes.map(change => ({
              id: change.id,
              cells: [{ text: change.name }, { text: change.type }, { text: change.severity ?? '' }, { text: when(change.created_at) }],
            }))}
          />
        </>
      )}
    </section>
  );
}
