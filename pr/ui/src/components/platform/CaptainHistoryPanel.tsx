// The platform dashboard's Captain history panel: every organization's prompt runs per day by
// outcome, each organization's share and live sessions, and the latest failures with their errors.
import { useQuery } from '@tanstack/react-query';
import { StatStrip } from '@flanksource/clicky-ui/data';
import { fetchJSON, queryKeys } from '../../query';
import { formatSeconds } from './formatSeconds';
import { StackedDayBars, type DaySeries } from './StackedDayBars';
import { windowDays } from './windowDays';

// CaptainRunDay is one UTC day's prompt runs of an organization with one origin and state; org
// is absent for runs no organization of this Clerk instance owns.
export interface CaptainRunDay {
  day: string;
  org?: string;
  orgName?: string;
  origin: string;
  state: string;
  runs: number;
  runSeconds: number;
}

export interface CaptainRunFailure {
  id: string;
  org?: string;
  orgName?: string;
  origin: string;
  error: string;
  createdAt: string;
  finishedAt?: string;
}

export interface CaptainLiveSessions {
  org?: string;
  orgName?: string;
  health: string;
  sessions: number;
}

export interface CaptainHistory {
  days: CaptainRunDay[];
  failures: CaptainRunFailure[];
  live: CaptainLiveSessions[];
}

// historyWindowDays is how many days back the panel reads.
const historyWindowDays = 30;

// outcomes are how prompt runs end, in the fixed order that assigns their colors; pending,
// running and waiting runs are active.
const outcomes = ['succeeded', 'failed', 'cancelled', 'active'] as const;
type Outcome = (typeof outcomes)[number];

const outcomeSeries: DaySeries[] = [
  { key: 'succeeded', label: 'Succeeded', color: 'var(--series-1)' },
  { key: 'failed', label: 'Failed', color: 'var(--series-2)' },
  { key: 'cancelled', label: 'Cancelled', color: 'var(--series-3)' },
  { key: 'active', label: 'In progress', color: 'var(--series-4)' },
];

function outcome(state: string): Outcome {
  return state === 'succeeded' || state === 'failed' || state === 'cancelled' ? state : 'active';
}

export type OutcomeTotals = { day: string } & Record<Outcome, number>;

// runsPerDay counts each day's runs by outcome for the days days before now and now's day.
export function runsPerDay(rows: CaptainRunDay[], days: number, now: Date): OutcomeTotals[] {
  const totals = new Map<string, OutcomeTotals>();
  for (const day of windowDays(days, now)) totals.set(day, { day, succeeded: 0, failed: 0, cancelled: 0, active: 0 });
  for (const row of rows) {
    const bucket = totals.get(row.day);
    if (bucket) bucket[outcome(row.state)] += row.runs;
  }
  return [...totals.values()];
}

interface OrgShare {
  key: string;
  name: string;
  runs: number;
  failed: number;
  seconds: number;
  live: number;
  unhealthy: Record<string, number>;
}

// orgShares totals each organization's runs, failures, run time and live sessions, busiest first.
function orgShares(history: CaptainHistory): OrgShare[] {
  const shares = new Map<string, OrgShare>();
  const share = (org: string | undefined, name: string | undefined) => {
    const key = org ?? '';
    let found = shares.get(key);
    if (!found) shares.set(key, (found = { key, name: name || 'Unattributed', runs: 0, failed: 0, seconds: 0, live: 0, unhealthy: {} }));
    return found;
  };
  for (const row of history.days) {
    const s = share(row.org, row.orgName);
    s.runs += row.runs;
    s.seconds += row.runSeconds;
    if (row.state === 'failed') s.failed += row.runs;
  }
  for (const row of history.live) {
    const s = share(row.org, row.orgName);
    s.live += row.sessions;
    if (row.health !== 'healthy') s.unhealthy[row.health] = (s.unhealthy[row.health] ?? 0) + row.sessions;
  }
  return [...shares.values()].sort((a, b) => b.runs - a.runs || a.name.localeCompare(b.name));
}

// liveCell is an organization's live sessions, naming how many aren't healthy.
function liveCell(share: OrgShare): string {
  const unhealthy = Object.entries(share.unhealthy).map(([health, n]) => `${n} ${health}`);
  return unhealthy.length ? `${share.live} (${unhealthy.join(', ')})` : String(share.live);
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
            {row.cells.map((cell, i) => <td key={i} className="px-2 py-1.5">{cell}</td>)}
          </tr>
        ))}
      </tbody>
    </table>
  );
}

export function CaptainHistoryPanel() {
  const history = useQuery({
    queryKey: queryKeys.platform('captain-history', historyWindowDays),
    queryFn: ({ signal }) => fetchJSON<CaptainHistory>({ url: `/api/platform/captain-history?days=${historyWindowDays}`, signal, context: 'load Captain\'s history' }),
  });
  const data = history.data;
  const runs = data?.days.reduce((sum, row) => sum + row.runs, 0) ?? 0;
  const failed = data?.days.filter(row => row.state === 'failed').reduce((sum, row) => sum + row.runs, 0) ?? 0;
  const seconds = data?.days.reduce((sum, row) => sum + row.runSeconds, 0) ?? 0;
  const live = data?.live.reduce((sum, row) => sum + row.sessions, 0) ?? 0;
  const unhealthy = data?.live.filter(row => row.health !== 'healthy').reduce((sum, row) => sum + row.sessions, 0) ?? 0;

  return (
    <section className="flex flex-col gap-3">
      <h2 className="text-xs font-semibold uppercase text-muted-foreground">Captain history · {historyWindowDays} days</h2>
      {history.error && <p role="alert" className="text-red-600">{history.error.message}</p>}
      {data && (
        <>
          <StatStrip
            items={[
              { label: 'Prompt runs', value: `${runs} prompt runs` },
              { label: 'Failed', value: `${failed} failed (${runs ? Math.round((failed / runs) * 100) : 0}%)` },
              { label: 'Run time', value: `${formatSeconds(seconds)} run time` },
              { label: 'Live sessions', value: unhealthy ? `${live} sessions live, ${unhealthy} not healthy` : `${live} sessions live` },
            ]}
          />
          <StackedDayBars title="Prompt runs per day" data={runsPerDay(data.days, historyWindowDays, new Date())} series={outcomeSeries} unit="runs" />
          <Table
            title="Captain runs by organization"
            headings={['Organization', 'Runs', 'Failed', 'Run time', 'Live sessions']}
            rows={orgShares(data).map(share => ({
              key: share.key,
              cells: [share.name, String(share.runs), String(share.failed), formatSeconds(share.seconds), liveCell(share)],
            }))}
          />
          {data.failures.length > 0 && (
            <Table
              title="Latest failed prompt runs"
              headings={['When', 'Organization', 'Origin', 'Error']}
              rows={data.failures.map(failure => ({
                key: failure.id,
                cells: [new Date(failure.createdAt).toLocaleString(), failure.orgName || 'Unattributed', failure.origin, failure.error],
              }))}
            />
          )}
        </>
      )}
    </section>
  );
}
