import { UiGitBranch } from '@flanksource/clicky-ui/icons';
import type { CountStat, GitMetrics, LatencyStat } from '../types';

// Sections of the latency table, in the order a slow page is diagnosed: the
// reads pages wait on, then the scans that keep the rows current, the queue
// they wait in, and the git commands inside them.
const SECTIONS: { key: keyof Pick<GitMetrics, 'reads' | 'statusScans' | 'refsScans' | 'queueWait' | 'commands'>; label: string }[] = [
  { key: 'reads', label: 'Reads' },
  { key: 'statusScans', label: 'Status scans' },
  { key: 'refsScans', label: 'Ref scans' },
  { key: 'queueWait', label: 'Queue wait' },
  { key: 'commands', label: 'Git commands' },
];

export function GitTrackingPanel({ metrics, error }: { metrics: GitMetrics | null; error?: string }) {
  return (
    <div className="bg-card border border-border rounded-md mb-4 p-3">
      <div className="flex items-center justify-between mb-2 gap-2 flex-wrap">
        <div className="flex items-center gap-2">
          <UiGitBranch className="text-blue-600" />
          <span className="text-xs font-semibold text-muted-foreground uppercase">Git tracking</span>
          {metrics?.tracking && (
            <>
              <SettingBadge name="fsmonitor" on={metrics.fsmonitor} />
              <SettingBadge name="untracked cache" on={metrics.untrackedCache} />
            </>
          )}
        </div>
        {error && <span className="text-xs text-red-600">{error}</span>}
      </div>

      {metrics && !metrics.tracking && (
        <div className="text-xs text-muted-foreground">This process does not track git state (no database).</div>
      )}

      {metrics?.tracking && (
        <>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-2 mb-3">
            <Stat label="Repositories" value={metrics.trackedRepos} />
            <Stat
              label="Worktrees"
              value={metrics.hotWorktrees + metrics.idleWorktrees + metrics.backoffWorktrees}
              sub={`${metrics.hotWorktrees} hot · ${metrics.idleWorktrees} idle · ${metrics.backoffWorktrees} backed off`}
            />
            <Stat label="Scans running" value={metrics.scansInFlight} />
            <Stat label="Comparisons computed" value={metrics.rangesComputed} />
          </div>
          <ScanTriggers triggers={metrics.scanTriggers ?? []} />
          <LatencyTable metrics={metrics} />
        </>
      )}
    </div>
  );
}

function SettingBadge({ name, on }: { name: string; on: boolean }) {
  return (
    <span
      className={`text-[10px] px-1.5 py-0.5 rounded font-mono ${on ? 'bg-green-100 text-green-700' : 'bg-muted text-muted-foreground'}`}
      title={`Status scans run with ${name} ${on ? 'on' : 'off'}`}
    >
      {name} {on ? 'on' : 'off'}
    </span>
  );
}

function Stat({ label, value, sub }: { label: string; value: number; sub?: string }) {
  return (
    <div className="bg-muted rounded px-2 py-1.5">
      <div className="text-[10px] text-muted-foreground">{label}</div>
      <div className="text-sm font-semibold text-foreground tabular-nums">{value.toLocaleString()}</div>
      {sub && <div className="text-[10px] text-muted-foreground">{sub}</div>}
    </div>
  );
}

// ScanTriggers shows what keeps requesting scans — the cadence, .git metadata
// events, server-side mutations, focused views or changed refs.
function ScanTriggers({ triggers }: { triggers: CountStat[] }) {
  if (triggers.length === 0) return null;
  return (
    <div className="flex items-center gap-1.5 flex-wrap mb-3 text-xs">
      <span className="text-[10px] text-muted-foreground uppercase">Scans requested</span>
      {triggers.map(trigger => (
        <span key={`${trigger.labels.kind}-${trigger.labels.source}`} data-testid="scan-trigger" className="bg-muted rounded px-1.5 py-0.5 font-mono">
          <span className="text-muted-foreground">{trigger.labels.kind} ← {trigger.labels.source}</span>
          <span className="ml-1.5 tabular-nums text-foreground">{trigger.count.toLocaleString()}</span>
        </span>
      ))}
    </div>
  );
}

function LatencyTable({ metrics }: { metrics: GitMetrics }) {
  const sections = SECTIONS.map(section => ({ ...section, stats: metrics[section.key] ?? [] })).filter(section => section.stats.length > 0);
  if (sections.length === 0) {
    return <div className="text-xs text-muted-foreground">No scans recorded yet.</div>;
  }
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-xs">
        <thead className="text-muted-foreground uppercase">
          <tr>
            <th className="px-2 py-1 text-left font-medium">Series</th>
            <th className="px-2 py-1 text-right font-medium">Count</th>
            <th className="px-2 py-1 text-right font-medium">Avg</th>
            <th className="px-2 py-1 text-right font-medium">p50</th>
            <th className="px-2 py-1 text-right font-medium">p95</th>
            <th className="px-2 py-1 text-right font-medium">Total</th>
          </tr>
        </thead>
        <tbody>
          {sections.map(section => section.stats.map((stat, i) => (
            <tr key={`${section.key}-${i}`} className="border-t border-border">
              <td className="px-2 py-1">
                <span className="text-foreground">{section.label}</span>
                <span className="ml-1.5 font-mono text-muted-foreground">{formatLabels(stat)}</span>
              </td>
              <td className="px-2 py-1 text-right tabular-nums text-foreground">{stat.count.toLocaleString()}</td>
              <td className="px-2 py-1 text-right tabular-nums text-foreground">{formatMs(stat.avgMs)}</td>
              <td className="px-2 py-1 text-right tabular-nums text-muted-foreground">{formatMs(stat.p50Ms)}</td>
              <td className="px-2 py-1 text-right tabular-nums text-muted-foreground">{formatMs(stat.p95Ms)}</td>
              <td className="px-2 py-1 text-right tabular-nums text-muted-foreground">{formatMs(stat.totalMs)}</td>
            </tr>
          )))}
        </tbody>
      </table>
    </div>
  );
}

function formatLabels(stat: LatencyStat): string {
  return Object.entries(stat.labels)
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([name, value]) => `${name}=${value}`)
    .join(' ');
}

export function formatMs(ms: number): string {
  if (ms >= 1000) return `${(ms / 1000).toFixed(1)} s`;
  if (ms >= 10) return `${ms.toFixed(0)} ms`;
  return `${ms.toFixed(1)} ms`;
}
