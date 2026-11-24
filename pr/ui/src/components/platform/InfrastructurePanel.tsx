// The platform dashboard's infrastructure panel: gavel's catalog by type, its health checks when it
// has any, and the catalog's changes, from Mission Control, each a subsection the sidebar lists.
import { useQuery } from '@tanstack/react-query';
import { Badge, StatStrip, type BadgeTone } from '@flanksource/clicky-ui/data';
import { fetchJSON, queryKeys } from '../../query';
import { PanelHeading, PanelSubheading } from './PanelHeading';
import { sectionAnchor } from './PlatformSections';
import { formatDateTime } from './dates';

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

// MissionControlChange is a change to a config item as Mission Control's changes search returns
// it: name and type are the config item's.
export interface MissionControlChange {
  id: string;
  config_id: string;
  name: string;
  type?: string;
  change_type: string;
  summary?: string;
  severity?: string;
  created_at?: string;
}

// Infrastructure is each list Mission Control has for gavel, or why it couldn't be read; changes
// reach changesDays days back.
export interface Infrastructure {
  missionControl: string;
  configs: MissionControlResource[];
  configsError?: string;
  checks: MissionControlResource[];
  checksError?: string;
  changes: MissionControlChange[];
  changesDays: number;
  changesError?: string;
}

// healthTone is the badge color of a resource's health.
export function healthTone(health: string): BadgeTone {
  return ({ healthy: 'success', unhealthy: 'danger', warning: 'warning' } as Record<string, BadgeTone>)[health] ?? 'neutral';
}

// severityTone is the badge color of a change's severity.
export function severityTone(severity: string): BadgeTone {
  return ({ critical: 'danger', high: 'danger', medium: 'warning', low: 'info' } as Record<string, BadgeTone>)[severity] ?? 'neutral';
}

// Cell is a table cell: plain text, a badge of tone, or a link to href.
interface Cell {
  text: string;
  tone?: BadgeTone;
  href?: string;
}

// badgeText keeps a badge's text the page's text color, black in light mode, over its tone's background.
const badgeText = 'text-foreground dark:text-foreground';

function CellContent({ cell }: { cell: Cell }) {
  if (cell.tone) return cell.text ? <Badge tone={cell.tone} variant="soft" size="sm" className={badgeText}>{cell.text}</Badge> : null;
  if (cell.href) return <a className="underline" href={cell.href} target="_blank" rel="noreferrer">{cell.text}</a>;
  return <>{cell.text}</>;
}

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
  return at ? formatDateTime(at) : '—';
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
  rows: { id: string; cells: Cell[] }[];
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
            {row.cells.map((cell, i) => <td key={i} className="px-2 py-1.5"><CellContent cell={cell} /></td>)}
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
  const health = (resource: MissionControlResource): Cell => ({ text: resource.health || 'unknown', tone: healthTone(resource.health ?? '') });
  // ui is Mission Control's web UI, which serves its backend under /api.
  const ui = data?.missionControl.replace(/\/api\/?$/, '') ?? '';
  const catalogLink = (id: string) => `${ui}/catalog/${encodeURIComponent(id)}`;

  return (
    <section className="flex flex-col gap-3">
      <PanelHeading>
        Infrastructure
        {data && <> · <a className="underline" href={ui} target="_blank" rel="noreferrer">{data.missionControl}</a></>}
      </PanelHeading>
      {infrastructure.error && <p role="alert" className="text-red-600">{infrastructure.error.message}</p>}
      {data && (
        <>
          <StatStrip
            items={[
              { label: 'Catalog', value: data.configsError ? unread : healthSummary(data.configs) },
              ...(parts.has('infrastructure-checks') ? [{ label: 'Health checks', value: data.checksError ? unread : healthSummary(data.checks) }] : []),
              { label: 'Changes', value: data.changesError ? unread : `${data.changes.length} in ${data.changesDays} days` },
            ]}
          />
          <div id={sectionAnchor('infrastructure-catalog')} className="flex scroll-mt-4 flex-col gap-2">
            <PanelSubheading>Catalog</PanelSubheading>
            <ResourceTable
              label="Catalog"
              error={data.configsError}
              empty="No config items match the search."
              headings={['Name', 'Type', 'Health', 'Status', 'Namespace', 'Updated']}
              rows={[...data.configs].sort(byType).map(config => ({
                id: config.id,
                cells: [{ text: config.name, href: catalogLink(config.id) }, { text: config.type, tone: 'neutral' }, health(config), { text: config.status ?? '' }, { text: config.namespace }, { text: when(config.updated_at) }],
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
              empty={`No changes to these config items in the last ${data.changesDays} days.`}
              headings={['Config', 'Type', 'Change', 'Summary', 'Severity', 'When']}
              rows={data.changes.map(change => ({
                id: change.id,
                cells: [
                  { text: change.name, href: catalogLink(change.config_id) },
                  { text: change.type ?? '', tone: 'neutral' },
                  { text: change.change_type, tone: 'neutral' },
                  { text: change.summary ?? '' },
                  { text: change.severity ?? '', tone: severityTone(change.severity ?? '') },
                  { text: when(change.created_at) },
                ],
              }))}
            />
          </div>
        </>
      )}
    </section>
  );
}
