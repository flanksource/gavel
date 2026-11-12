// The platform dashboard's infrastructure panel: gavel's catalog by type, its health checks when it
// has any, and the catalog's changes, from Mission Control, each a subsection the sidebar lists.
import { useQuery } from '@tanstack/react-query';
import { StatStrip } from '@flanksource/clicky-ui/data';
import { fetchJSON, queryKeys } from '../../query';
import { PanelHeading, PanelSubheading } from './PanelHeading';
import { sectionAnchor } from './PlatformSections';

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

// infrastructureSubsections are the panel's parts, as the sidebar lists them: none until Mission
// Control answers, and health checks only when gavel has some or they couldn't be read.
export function infrastructureSubsections(data: Infrastructure | undefined): { id: string; label: string }[] {
  if (!data) return [];
  return [
    { id: 'infrastructure-catalog', label: 'Catalog' },
    ...(data.checks.length > 0 || data.checksError ? [{ id: 'infrastructure-checks', label: 'Health checks' }] : []),
    { id: 'infrastructure-changes', label: 'Changes' },
  ];
}

// byType orders config items by type, then name.
function byType(a: MissionControlResource, b: MissionControlResource): number {
  return a.type.localeCompare(b.type) || a.name.localeCompare(b.name);
}

function ResourceTable({ label, error, empty, headings, rows }: {
  label: string;
  error?: string;
  empty: string;
  headings: string[];
  rows: { id: string; cells: { text: string; className?: string }[] }[];
}) {
  if (error) return <p role="alert" className="text-red-600">{error}</p>;
  if (rows.length === 0) return <p className="text-xs text-muted-foreground">{empty}</p>;
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

// useInfrastructure reads what Mission Control has for gavel's infrastructure.
export function useInfrastructure() {
  return useQuery({
    queryKey: queryKeys.platform('infrastructure'),
    queryFn: ({ signal }) => fetchJSON<Infrastructure>({ url: '/api/platform/infrastructure', signal, context: 'load gavel\'s infrastructure' }),
  });
}

export function InfrastructurePanel() {
  const infrastructure = useInfrastructure();
  const data = infrastructure.data;
  const parts = new Set(infrastructureSubsections(data).map(part => part.id));
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
              { label: 'Catalog', value: data.configsError ? unread : healthSummary(data.configs) },
              ...(parts.has('infrastructure-checks') ? [{ label: 'Health checks', value: data.checksError ? unread : healthSummary(data.checks) }] : []),
              { label: 'Changes', value: data.changesError ? unread : `${data.changes.length}` },
            ]}
          />
          <div id={sectionAnchor('infrastructure-catalog')} className="flex scroll-mt-4 flex-col gap-2">
            <PanelSubheading>Catalog</PanelSubheading>
            <ResourceTable
              label="Catalog"
              error={data.configsError}
              empty="No config items match the search."
              headings={['Type', 'Name', 'Health', 'Status', 'Namespace', 'Updated']}
              rows={[...data.configs].sort(byType).map(config => ({
                id: config.id,
                cells: [{ text: config.type }, { text: config.name }, health(config), { text: config.status ?? '' }, { text: config.namespace }, { text: when(config.updated_at) }],
              }))}
            />
          </div>
          {parts.has('infrastructure-checks') && (
            <div id={sectionAnchor('infrastructure-checks')} className="flex scroll-mt-4 flex-col gap-2">
              <PanelSubheading>Health checks</PanelSubheading>
              <ResourceTable
                label="Health checks"
                error={data.checksError}
                empty="No health checks match the search."
                headings={['Name', 'Type', 'Health', 'Status', 'Namespace']}
                rows={data.checks.map(check => ({
                  id: check.id,
                  cells: [{ text: check.name }, { text: check.type }, health(check), { text: check.status ?? '' }, { text: check.namespace }],
                }))}
              />
            </div>
          )}
          <div id={sectionAnchor('infrastructure-changes')} className="flex scroll-mt-4 flex-col gap-2">
            <PanelSubheading>Changes</PanelSubheading>
            <ResourceTable
              label="Changes"
              error={data.changesError}
              empty="No changes to these config items."
              headings={['Config', 'Change', 'Severity', 'When']}
              rows={data.changes.map(change => ({
                id: change.id,
                cells: [{ text: change.name }, { text: change.type }, { text: change.severity ?? '' }, { text: when(change.created_at) }],
              }))}
            />
          </div>
        </>
      )}
    </section>
  );
}
