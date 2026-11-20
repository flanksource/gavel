// Shows the organization what it used in a window: the CPU and memory its sandbox runs held, what
// runs now, and its model calls' tokens by kind and cost. Shown as a page in Clerk's organization profile.
import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { fetchJSON, queryKeys } from '../query';

export interface Usage {
  since: string;
  compute: {
    coreSeconds: number;
    gibSeconds: number;
    runs: number;
    running: { runs: number; cpuMillicores: number; memoryBytes: number };
  };
  tokens: { input: number; output: number; reasoning: number; cacheRead: number; cacheWrite: number; counted: number };
  costUSD: number;
  calls: number;
  otherCurrencyCalls: number;
}

export interface UsageResponse {
  window: string;
  usage: Usage;
}

const windows = [
  { value: 'now/M', label: 'This month' },
  { value: 'now-30d', label: 'Last 30 days' },
  { value: 'now/w', label: 'This week' },
];

const gib = 2 ** 30;
const count = new Intl.NumberFormat('en-US');
const usd = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' });

// figure shows n with at most one decimal.
function figure(n: number): string {
  return String(Number(n.toFixed(1)));
}

function errorText(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

export function UsagePanel() {
  const [selected, setSelected] = useState(windows[0].value);
  const usage = useQuery({
    queryKey: queryKeys.usage(selected),
    queryFn: ({ signal }) => fetchJSON<UsageResponse>({ url: `/api/usage?window=${encodeURIComponent(selected)}`, signal, context: 'load the organization\'s usage' }),
  });
  const data = usage.data?.usage;

  return (
    <div className="flex flex-col gap-4 text-sm">
      <label className="flex items-center gap-2">
        <span>Window</span>
        <select className="rounded border border-border bg-background px-2 py-1 text-sm" aria-label="Window" value={selected} onChange={event => setSelected(event.target.value)}>
          {windows.map(w => <option key={w.value} value={w.value}>{w.label}</option>)}
        </select>
      </label>
      {data && <Compute usage={data} />}
      {data && <ModelCalls usage={data} />}
      {usage.error && <p role="alert" className="text-destructive">{errorText(usage.error)}</p>}
    </div>
  );
}

function Compute({ usage }: { usage: Usage }) {
  const { compute } = usage;
  return (
    <section className="flex flex-col gap-1">
      <h3 className="font-medium">Sandbox runs</h3>
      <div className="flex flex-wrap gap-x-4">
        <span>{`${figure(compute.coreSeconds / 3600)} core-hours`}</span>
        <span>{`${figure(compute.gibSeconds / 3600)} GiB-hours`}</span>
        <span>{`${count.format(compute.runs)} runs`}</span>
      </div>
      <div>
        {`${compute.running.runs} running · ${figure(compute.running.cpuMillicores / 1000)} CPU · ${figure(compute.running.memoryBytes / gib)} GiB`}
      </div>
      <p className="text-xs text-muted-foreground">
        Each run counts the CPU and memory its sandbox environment was given, for as long as it ran, whatever it used of it:
        an hour at 2 CPU is 2 core-hours.
      </p>
    </section>
  );
}

function ModelCalls({ usage }: { usage: Usage }) {
  const { tokens } = usage;
  const rows: { name: string; value: number; counted: boolean }[] = [
    { name: 'Input', value: tokens.input, counted: true },
    { name: 'Cache writes', value: tokens.cacheWrite, counted: true },
    { name: 'Output', value: tokens.output, counted: true },
    { name: 'Reasoning', value: tokens.reasoning, counted: true },
    { name: 'Cache reads', value: tokens.cacheRead, counted: false },
  ];
  return (
    <section className="flex flex-col gap-2">
      <h3 className="font-medium">Model calls</h3>
      <div className="flex flex-wrap gap-x-4">
        <span>{`${count.format(tokens.counted)} tokens counted`}</span>
        <span>{usd.format(usage.costUSD)}</span>
        <span>{`${count.format(usage.calls)} calls`}</span>
      </div>
      <table aria-label="Tokens by kind" className="w-full max-w-md border-collapse text-left text-xs">
        <tbody>
          {rows.map(row => (
            <tr key={row.name} className="border-t border-border">
              <td className="py-1">{row.name}</td>
              <td className="py-1 text-right">{count.format(row.value)}</td>
              <td className="py-1 pl-3 text-muted-foreground">{row.counted ? 'counted' : 'not counted'}</td>
            </tr>
          ))}
        </tbody>
      </table>
      <p className="text-xs text-muted-foreground">
        Counted tokens are input, cache writes, output and reasoning. Cache reads, the conversation an agent sends again on
        every turn, are left out, as model providers leave them out of their input limits.
      </p>
      <p className="text-xs text-muted-foreground">
        The cost is in USD: the cost the provider reported for each call, or else its list price, each kind of token at the
        model's price for it. The calls use this organization's own model keys, so the provider bills the organization.
        {usage.otherCurrencyCalls > 0 && ` ${count.format(usage.otherCurrencyCalls)} calls priced in another currency count their tokens but not their cost, since there is no rate to convert them.`}
      </p>
    </section>
  );
}
