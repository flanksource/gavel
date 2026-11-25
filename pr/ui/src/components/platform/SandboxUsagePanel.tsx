// The platform dashboard's sandbox usage panel: every organization's sandbox runs and sandbox
// time per day by kind, the headline figures, and each organization's share.
import { useQuery } from '@tanstack/react-query';
import { StatStrip } from '@flanksource/clicky-ui/data';
import { fetchJSON, queryKeys } from '../../query';
import { kindLabels } from '../RunsView';
import { formatSeconds } from './formatSeconds';
import { StackedDayBars, type DaySeries } from './StackedDayBars';
import { windowDays } from './windowDays';
import { PanelHeading } from './PanelHeading';

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
type RunKind = (typeof runKinds)[number];
const knownKinds: ReadonlySet<string> = new Set(runKinds);

function isRunKind(kind: string): kind is RunKind {
  return knownKinds.has(kind);
}

export type KindTotals = { day: string } & Record<RunKind, number>;

// perDay totals value(row) per UTC day and kind for the days days before now and now's day,
// keeping days without runs so the axis shows them.
export function perDay(rows: SandboxDay[], days: number, now: Date, value: (row: SandboxDay) => number): KindTotals[] {
  const totals = new Map<string, KindTotals>();
  for (const day of windowDays(days, now)) totals.set(day, { day, test: 0, lint: 0, 'push-check': 0, agent: 0 });
  for (const row of rows) {
    const bucket = totals.get(row.day);
    if (bucket && isRunKind(row.kind)) bucket[row.kind] += value(row);
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

// kindSeries are the run kinds as stacked segments, each in its --kind-* color from index.css.
const kindSeries: DaySeries[] = runKinds.map(kind => ({ key: kind, label: kindLabels[kind] ?? kind, color: `var(--kind-${kind})` }));

// SandboxUsagePanel shows every organization's runs, or only org's when given one.
export function SandboxUsagePanel({ org }: { org?: string } = {}) {
  const usage = useQuery({
    queryKey: queryKeys.platform('sandbox-runs', usageWindowDays),
    queryFn: ({ signal }) => fetchJSON<SandboxUsage>({ url: `/api/platform/sandbox-runs?days=${usageWindowDays}`, signal, context: 'load sandbox usage' }),
  });
  const data = usage.data && (org ? { days: usage.data.days.filter(row => row.org === org), active: usage.data.active.filter(row => row.org === org) } : usage.data);
  const runs = data?.days.reduce((sum, row) => sum + row.runs, 0) ?? 0;
  const failed = data?.days.filter(row => row.status === 'failed').reduce((sum, row) => sum + row.runs, 0) ?? 0;
  const seconds = data?.days.reduce((sum, row) => sum + row.runSeconds, 0) ?? 0;
  const active = data?.active.reduce((sum, row) => sum + row.runs, 0) ?? 0;
  const now = new Date();

  return (
    <section className="flex flex-col gap-3">
      <PanelHeading>Sandbox usage · {usageWindowDays} days</PanelHeading>
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
            <StackedDayBars title="Runs per day" data={perDay(data.days, usageWindowDays, now, row => row.runs)} series={kindSeries} unit="runs" />
            <StackedDayBars title="Sandbox minutes per day" data={perDay(data.days, usageWindowDays, now, row => row.runSeconds / 60)} series={kindSeries} unit="min" />
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
