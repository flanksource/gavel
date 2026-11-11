// The platform dashboard's substrate inventory panel: substrate's workers with their capacity
// against what their actors use, and every atespace's actors by the organization whose run each is.
import { useQuery } from '@tanstack/react-query';
import { StatStrip } from '@flanksource/clicky-ui/data';
import { fetchJSON, queryKeys } from '../../query';

export interface SubstrateResources {
  cpuMillicores: number;
  memoryBytes: number;
  actors: number;
}

export interface SubstrateWorker {
  name: string;
  pool: string;
  node: string;
  state: string;
  capacity: SubstrateResources;
  allocated: SubstrateResources;
}

// SubstrateActor is one actor; run is set for a sandbox run's environment, and org when the run
// belongs to one of this Clerk instance's organizations.
export interface SubstrateActor {
  atespace: string;
  name: string;
  state: string;
  created?: string;
  run?: string;
  org?: { org: string; orgName: string; kind: string };
}

export interface SubstrateInventory {
  workers: SubstrateWorker[];
  workersError?: string;
  actors: SubstrateActor[];
  actorsError?: string;
}

const gib = 1 << 30;

// figure shows n with at most one decimal.
function figure(n: number): string {
  return String(Number(n.toFixed(1)));
}

const shownStates = ['running', 'suspended', 'crashed'] as const;

interface Owner {
  key: string;
  name: string;
  order: number;
  states: Record<string, number>;
  total: number;
}

// owners counts the actors of each owner by state: organizations by name, then runs gavel
// doesn't know, then each atespace's actors that aren't gavel runs.
function owners(actors: SubstrateActor[]): Owner[] {
  const byKey = new Map<string, Owner>();
  for (const actor of actors) {
    const [key, name, order] = actor.org
      ? [actor.org.org, actor.org.orgName, 0]
      : actor.run
        ? ['unknown', 'A run gavel doesn\'t know', 1]
        : [`atespace/${actor.atespace}`, `Not a gavel run, in ${actor.atespace}`, 2];
    let owner = byKey.get(key);
    if (!owner) byKey.set(key, (owner = { key, name, order, states: {}, total: 0 }));
    const state = (shownStates as readonly string[]).includes(actor.state) ? actor.state : 'other';
    owner.states[state] = (owner.states[state] ?? 0) + 1;
    owner.total++;
  }
  return [...byKey.values()].sort((a, b) => a.order - b.order || a.name.localeCompare(b.name));
}

function Table({ title, headings, rows }: { title: string; headings: string[]; rows: { key: string; cells: string[] }[] }) {
  return (
    <table aria-label={title} className="w-full border-collapse text-left text-xs">
      <thead className="text-muted-foreground">
        <tr>
          {headings.map(heading => <th key={heading} className="border-b border-border px-2 py-1.5 font-medium">{heading}</th>)}
        </tr>
      </thead>
      <tbody>
        {rows.map(row => (
          <tr key={row.key} className="border-b border-border">
            {row.cells.map((cell, i) => <td key={i} className={`px-2 py-1.5${i === 0 ? ' font-medium' : ''}`}>{cell}</td>)}
          </tr>
        ))}
      </tbody>
    </table>
  );
}

// SubstrateInventoryPanel shows substrate's workers and every actor, or only the actors of org's
// runs when given one: workers belong to no one organization.
export function SubstrateInventoryPanel({ org }: { org?: string } = {}) {
  const inventory = useQuery({
    queryKey: queryKeys.platform('substrate-inventory'),
    queryFn: ({ signal }) => fetchJSON<SubstrateInventory>({ url: '/api/platform/substrate/inventory', signal, context: 'load substrate\'s inventory' }),
  });
  const data = inventory.data && (org ? { ...inventory.data, actors: inventory.data.actors.filter(actor => actor.org?.org === org) } : inventory.data);
  const workers = data?.workers ?? [];
  const sum = (pick: (w: SubstrateWorker) => number) => workers.reduce((total, worker) => total + pick(worker), 0);

  return (
    <section className="flex flex-col gap-3">
      <h2 className="text-xs font-semibold uppercase text-muted-foreground">Substrate inventory</h2>
      {inventory.error && <p role="alert" className="text-red-600">{inventory.error.message}</p>}
      {data?.workersError && <p role="alert" className="text-red-600">{data.workersError}</p>}
      {data?.actorsError && <p role="alert" className="text-red-600">{data.actorsError}</p>}
      {data && !org && !data.workersError && (
        <>
          <StatStrip
            items={[
              { label: 'Workers', value: `${workers.filter(w => w.state === 'active').length} of ${workers.length} workers active` },
              { label: 'Actor slots', value: `${sum(w => w.allocated.actors)} of ${sum(w => w.capacity.actors)} actor slots used` },
              { label: 'CPU', value: `${figure(sum(w => w.allocated.cpuMillicores) / 1000)} of ${figure(sum(w => w.capacity.cpuMillicores) / 1000)} cores allocated` },
              { label: 'Memory', value: `${figure(sum(w => w.allocated.memoryBytes) / gib)} of ${figure(sum(w => w.capacity.memoryBytes) / gib)} GiB allocated` },
            ]}
          />
          <Table
            title="Substrate workers"
            headings={['Worker', 'Pool', 'Node', 'State', 'Cores', 'Memory (GiB)', 'Actors']}
            rows={workers.map(w => ({
              key: w.name,
              cells: [w.name, w.pool, w.node, w.state,
                `${figure(w.allocated.cpuMillicores / 1000)} / ${figure(w.capacity.cpuMillicores / 1000)}`,
                `${figure(w.allocated.memoryBytes / gib)} / ${figure(w.capacity.memoryBytes / gib)}`,
                `${w.allocated.actors} / ${w.capacity.actors}`],
            }))}
          />
        </>
      )}
      {data && !data.actorsError && (data.actors.length === 0
        ? <p className="text-xs text-muted-foreground">No actors in any atespace.</p>
        : (
          <Table
            title="Substrate actors by owner"
            headings={['Owner', 'Running', 'Suspended', 'Crashed', 'Other', 'Total']}
            rows={owners(data.actors).map(owner => ({
              key: owner.key,
              cells: [owner.name, ...[...shownStates, 'other'].map(state => String(owner.states[state] ?? 0)), String(owner.total)],
            }))}
          />
        ))}
    </section>
  );
}
