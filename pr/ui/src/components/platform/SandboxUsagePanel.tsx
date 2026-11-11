// The platform dashboard's sandbox usage panel: every organization's sandbox runs and sandbox
// time per day by kind, the headline figures, and each organization's share.
import { useQuery } from '@tanstack/react-query';
import { StatStrip } from '@flanksource/clicky-ui/data';
import { Bar, BarChart, CartesianGrid, Legend, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { fetchJSON, queryKeys } from '../../query';
import { kindLabels } from '../RunsView';
import { formatSeconds } from './formatSeconds';
import { windowDays } from './windowDays';

export interface SandboxDay {
  day: string;
  org: string;
  orgName: string;
  kind: string;
  backend: string;
  status: string;
  runs: number;
  runSeconds: number;
  queueSeconds: number;
}

export interface SandboxActive {
  org: string;
  orgName: string;
  kind: string;
  status: string;
  runs: number;
}

export interface SandboxUsage {
  days: SandboxDay[];
  active: SandboxActive[];
}

// usageWindowDays is how many days back the panel reads.
const usageWindowDays = 30;

// runKinds are the kinds of sandbox run, in the fixed order that assigns their colors.
const runKinds = ['test', 'lint', 'push-check', 'agent'] as const;


export type KindTotals = { day: string } & Record<(typeof runKinds)[number], number>;

// perDay totals value(row) per UTC day and kind for the days days before now and now's day,
// keeping days without runs so the axis shows them.
export function perDay(rows: SandboxDay[], days: number, now: Date, value: (row: SandboxDay) => number): KindTotals[] {
  const totals = new Map<string, KindTotals>();
  for (const day of windowDays(days, now)) totals.set(day, { day, test: 0, lint: 0, 'push-check': 0, agent: 0 });
  for (const row of rows) {
    const bucket = totals.get(row.day);
    const kind = runKinds.find(k => k === row.kind);
    if (bucket && kind) bucket[kind] += value(row);
  }
  return [...totals.values()];
}

interface OrgShare {
  org: string;
  name: string;
  runs: number;
  failed: number;
  seconds: number;
  active: number;
}

// orgShares totals each organization's runs, failures, sandbox time and active runs, busiest first.
function orgShares(usage: SandboxUsage): OrgShare[] {
  const shares = new Map<string, OrgShare>();
  const share = (org: string, name: string) => {
    let found = shares.get(org);
    if (!found) shares.set(org, (found = { org, name, runs: 0, failed: 0, seconds: 0, active: 0 }));
    return found;
  };
  for (const row of usage.days) {
    const s = share(row.org, row.orgName);
    s.runs += row.runs;
    s.seconds += row.runSeconds;
    if (row.status === 'failed') s.failed += row.runs;
  }
  for (const row of usage.active) share(row.org, row.orgName).active += row.runs;
  return [...shares.values()].sort((a, b) => b.runs - a.runs || a.name.localeCompare(b.name));
}

// KindBars stacks a day's totals by kind, each kind in its --kind-* color from index.css,
// with the same numbers as a table: two of the colors are faint on the light surface.
function KindBars({ title, data, unit }: { title: string; data: KindTotals[]; unit: string }) {
  return (
    <figure className="flex flex-col gap-1">
      <figcaption className="text-xs text-muted-foreground">{title}</figcaption>
      <div className="h-48 text-muted-foreground">
        {/* initialDimension sizes the first render, before the container is measured. */}
        <ResponsiveContainer width="100%" height="100%" initialDimension={{ width: 480, height: 192 }}>
          <BarChart data={data} margin={{ top: 4, right: 8, bottom: 0, left: 0 }}>
            <CartesianGrid vertical={false} stroke="currentColor" strokeOpacity={0.15} />
            <XAxis dataKey="day" tickFormatter={(day: string) => day.slice(5)} tick={{ fill: 'currentColor', fontSize: 11 }} tickLine={false} axisLine={false} />
            <YAxis allowDecimals={false} width={36} tick={{ fill: 'currentColor', fontSize: 11 }} tickLine={false} axisLine={false} />
            <Tooltip formatter={(value, kind) => [`${Math.round(Number(value))} ${unit}`, kindLabels[String(kind)] ?? String(kind)]} cursor={{ fillOpacity: 0.08 }} />
            <Legend formatter={(kind: string) => <span className="text-xs text-foreground">{kindLabels[kind] ?? kind}</span>} iconSize={8} />
            {runKinds.map(kind => (
              <Bar key={kind} dataKey={kind} stackId="kinds" fill={`var(--kind-${kind})`} stroke="var(--background)" strokeWidth={1} maxBarSize={24} />
            ))}
          </BarChart>
        </ResponsiveContainer>
      </div>
      <details className="text-xs">
        <summary className="cursor-pointer text-muted-foreground">Show as a table</summary>
        <table aria-label={title} className="mt-1 w-full border-collapse text-left">
          <thead className="text-muted-foreground">
            <tr>
              {['Day', ...runKinds.map(kind => kindLabels[kind] ?? kind), 'Total'].map(heading => (
                <th key={heading} className="border-b border-border px-2 py-1 font-medium">{heading}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {data.map(row => (
              <tr key={row.day} className="border-b border-border">
                <td className="px-2 py-1 font-mono">{row.day}</td>
                {runKinds.map(kind => <td key={kind} className="px-2 py-1">{Math.round(row[kind])}</td>)}
                <td className="px-2 py-1">{Math.round(runKinds.reduce((sum, kind) => sum + row[kind], 0))}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </details>
    </figure>
  );
}

export function SandboxUsagePanel() {
  const usage = useQuery({
    queryKey: queryKeys.platform('sandbox-runs', usageWindowDays),
    queryFn: ({ signal }) => fetchJSON<SandboxUsage>({ url: `/api/platform/sandbox-runs?days=${usageWindowDays}`, signal, context: 'load sandbox usage' }),
  });
  const data = usage.data;
  const runs = data?.days.reduce((sum, row) => sum + row.runs, 0) ?? 0;
  const failed = data?.days.filter(row => row.status === 'failed').reduce((sum, row) => sum + row.runs, 0) ?? 0;
  const seconds = data?.days.reduce((sum, row) => sum + row.runSeconds, 0) ?? 0;
  const active = data?.active.reduce((sum, row) => sum + row.runs, 0) ?? 0;
  const now = new Date();

  return (
    <section className="flex flex-col gap-3">
      <h2 className="text-xs font-semibold uppercase text-muted-foreground">Sandbox usage · {usageWindowDays} days</h2>
      {usage.error && <p role="alert" className="text-red-600">{usage.error.message}</p>}
      {data && (
        <>
          <StatStrip
            items={[
              { label: 'Runs', value: `${runs} runs` },
              { label: 'Failed', value: `${failed} failed (${runs ? Math.round((failed / runs) * 100) : 0}%)` },
              { label: 'Sandbox time', value: `${formatSeconds(seconds)} sandbox time` },
              { label: 'Queued or running', value: `${active} active now` },
            ]}
          />
          <div className="grid gap-4 lg:grid-cols-2">
            <KindBars title="Runs per day" data={perDay(data.days, usageWindowDays, now, row => row.runs)} unit="runs" />
            <KindBars title="Sandbox minutes per day" data={perDay(data.days, usageWindowDays, now, row => row.runSeconds / 60)} unit="min" />
          </div>
          <table className="w-full border-collapse text-left text-xs">
            <thead className="text-muted-foreground">
              <tr>
                {['Organization', 'Runs', 'Failed', 'Sandbox time', 'Active now'].map(heading => (
                  <th key={heading} className="border-b border-border px-2 py-1.5 font-medium">{heading}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {orgShares(data).map(share => (
                <tr key={share.org} className="border-b border-border">
                  <td className="px-2 py-1.5 font-medium">{share.name}</td>
                  <td className="px-2 py-1.5">{share.runs}</td>
                  <td className="px-2 py-1.5">{share.failed}</td>
                  <td className="px-2 py-1.5">{formatSeconds(share.seconds)}</td>
                  <td className="px-2 py-1.5">{share.active}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </>
      )}
    </section>
  );
}
