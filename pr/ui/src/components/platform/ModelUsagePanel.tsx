// The platform dashboard's model usage panel: what every organization's agents spent on model
// calls per day, the headline figures, and each organization's and model's share.
import { useQuery } from '@tanstack/react-query';
import { StatStrip } from '@flanksource/clicky-ui/data';
import { Bar, BarChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { fetchJSON, queryKeys } from '../../query';
import { windowDays } from './windowDays';
import { PanelHeading } from './PanelHeading';

// ModelUsageDay is one UTC day's model calls of an organization with one model and provider;
// org is absent for calls no organization of this Clerk instance owns.
export interface ModelUsageDay {
  day: string;
  org?: string;
  orgName?: string;
  model: string;
  provider: string;
  calls: number;
  inputTokens: number;
  outputTokens: number;
  reasoningTokens: number;
  cacheReadTokens: number;
  cacheWriteTokens: number;
  costUSD: number;
  // otherCurrencyCalls are priced in another currency: their tokens count, their cost doesn't.
  otherCurrencyCalls: number;
}

export interface ModelUsage {
  days: ModelUsageDay[];
}

// usageWindowDays is how many days back the panel reads.
const usageWindowDays = 30;

const usd = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' });
const count = new Intl.NumberFormat('en-US');

function tokens(row: ModelUsageDay): number {
  return row.inputTokens + row.outputTokens + row.reasoningTokens + row.cacheReadTokens + row.cacheWriteTokens;
}

export interface DayCost {
  day: string;
  costUSD: number;
}

// costPerDay totals the cost of each UTC day for the days days before now and now's day,
// keeping days without calls so the axis shows them.
export function costPerDay(rows: ModelUsageDay[], days: number, now: Date): DayCost[] {
  const totals = new Map<string, DayCost>();
  for (const day of windowDays(days, now)) totals.set(day, { day, costUSD: 0 });
  for (const row of rows) {
    const bucket = totals.get(row.day);
    if (bucket) bucket.costUSD += row.costUSD;
  }
  return [...totals.values()];
}

interface Share {
  key: string;
  name: string;
  provider?: string;
  calls: number;
  tokens: number;
  costUSD: number;
}

// shares totals rows by key, most spent first.
function shares(rows: ModelUsageDay[], key: (row: ModelUsageDay) => string, label: (row: ModelUsageDay) => Omit<Share, 'calls' | 'tokens' | 'costUSD'>): Share[] {
  const byKey = new Map<string, Share>();
  for (const row of rows) {
    const k = key(row);
    let share = byKey.get(k);
    if (!share) byKey.set(k, (share = { ...label(row), calls: 0, tokens: 0, costUSD: 0 }));
    share.calls += row.calls;
    share.tokens += tokens(row);
    share.costUSD += row.costUSD;
  }
  return [...byKey.values()].sort((a, b) => b.costUSD - a.costUSD || a.name.localeCompare(b.name));
}

function ShareTable({ title, first, rows }: { title: string; first: string[]; rows: { cells: string[]; share: Share }[] }) {
  return (
    <table aria-label={title} className="w-full border-collapse text-left text-xs">
      <thead className="text-muted-foreground">
        <tr>
          {[...first, 'Calls', 'Tokens', 'Cost'].map(heading => (
            <th key={heading} className="border-b border-border px-2 py-1.5 font-medium">{heading}</th>
          ))}
        </tr>
      </thead>
      <tbody>
        {rows.map(({ cells, share }) => (
          <tr key={share.key} className="border-b border-border">
            {cells.map((cell, i) => <td key={i} className={`px-2 py-1.5${i === 0 ? ' font-medium' : ''}`}>{cell}</td>)}
            <td className="px-2 py-1.5">{count.format(share.calls)}</td>
            <td className="px-2 py-1.5">{count.format(share.tokens)}</td>
            <td className="px-2 py-1.5">{usd.format(share.costUSD)}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

// CostBars draws the cost of each day, with the same numbers as a table.
function CostBars({ data }: { data: DayCost[] }) {
  const title = 'Model spend per day (USD)';
  return (
    <figure className="flex flex-col gap-1">
      <figcaption className="text-xs text-muted-foreground">{title}</figcaption>
      <div className="h-48 text-muted-foreground">
        {/* initialDimension sizes the first render, before the container is measured. */}
        <ResponsiveContainer width="100%" height="100%" initialDimension={{ width: 480, height: 192 }}>
          <BarChart data={data} margin={{ top: 4, right: 8, bottom: 0, left: 0 }}>
            <CartesianGrid vertical={false} stroke="currentColor" strokeOpacity={0.15} />
            <XAxis dataKey="day" tickFormatter={(day: string) => day.slice(5)} tick={{ fill: 'currentColor', fontSize: 11 }} tickLine={false} axisLine={false} />
            <YAxis width={44} tickFormatter={(value: number) => usd.format(value)} tick={{ fill: 'currentColor', fontSize: 11 }} tickLine={false} axisLine={false} />
            <Tooltip formatter={value => [usd.format(Number(value)), 'Spent']} cursor={{ fillOpacity: 0.08 }} />
            <Bar dataKey="costUSD" fill="var(--series-1)" radius={[4, 4, 0, 0]} maxBarSize={24} isAnimationActive={false} />
          </BarChart>
        </ResponsiveContainer>
      </div>
      <details className="text-xs">
        <summary className="cursor-pointer text-muted-foreground">Show as a table</summary>
        <table aria-label={title} className="mt-1 w-full border-collapse text-left">
          <thead className="text-muted-foreground">
            <tr>
              {['Day', 'Spent'].map(heading => <th key={heading} className="border-b border-border px-2 py-1 font-medium">{heading}</th>)}
            </tr>
          </thead>
          <tbody>
            {data.map(row => (
              <tr key={row.day} className="border-b border-border">
                <td className="px-2 py-1 font-mono">{row.day}</td>
                <td className="px-2 py-1">{usd.format(row.costUSD)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </details>
    </figure>
  );
}

// ModelUsagePanel shows every organization's model calls, or only org's when given one.
export function ModelUsagePanel({ org }: { org?: string } = {}) {
  const usage = useQuery({
    queryKey: queryKeys.platform('model-usage', usageWindowDays),
    queryFn: ({ signal }) => fetchJSON<ModelUsage>({ url: `/api/platform/model-usage?days=${usageWindowDays}`, signal, context: 'load model usage' }),
  });
  const rows = (usage.data?.days ?? []).filter(row => !org || row.org === org);
  const spent = rows.reduce((sum, row) => sum + row.costUSD, 0);
  const calls = rows.reduce((sum, row) => sum + row.calls, 0);
  const allTokens = rows.reduce((sum, row) => sum + tokens(row), 0);
  const otherCurrency = rows.reduce((sum, row) => sum + row.otherCurrencyCalls, 0);
  const orgs = shares(rows, row => row.org ?? '', row => ({ key: row.org ?? '', name: row.orgName || 'Unattributed' }));
  const models = shares(rows, row => `${row.provider}/${row.model}`, row => ({ key: `${row.provider}/${row.model}`, name: row.model, provider: row.provider }));

  return (
    <section className="flex flex-col gap-3">
      <PanelHeading>Model usage · {usageWindowDays} days</PanelHeading>
      {usage.error && <p role="alert" className="text-red-600">{usage.error.message}</p>}
      {usage.data && (
        <>
          <StatStrip
            items={[
              { label: 'Spent', value: `${usd.format(spent)} spent` },
              { label: 'Calls', value: `${count.format(calls)} calls` },
              { label: 'Tokens', value: `${count.format(allTokens)} tokens` },
            ]}
          />
          {otherCurrency > 0 && (
            <p role="status" className="text-xs text-muted-foreground">
              {otherCurrency === 1
                ? '1 call was priced in another currency: its tokens count, its cost isn\'t in these totals.'
                : `${count.format(otherCurrency)} calls were priced in another currency: their tokens count, their cost isn't in these totals.`}
            </p>
          )}
          <CostBars data={costPerDay(rows, usageWindowDays, new Date())} />
          <div className="grid gap-4 lg:grid-cols-2">
            <ShareTable title="Model spend by organization" first={['Organization']} rows={orgs.map(share => ({ cells: [share.name], share }))} />
            <ShareTable title="Model spend by model" first={['Model', 'Provider']} rows={models.map(share => ({ cells: [share.name, share.provider ?? ''], share }))} />
          </div>
        </>
      )}
    </section>
  );
}
