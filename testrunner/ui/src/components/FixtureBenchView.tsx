import { useMemo } from 'react';
import {
  Callout,
  DataTable,
  LabeledSparkline,
  MetricTile,
  ProcessTraceTree,
  type DataTableColumn,
  type ProcessNodeRow,
} from '@flanksource/clicky-ui/data';
import { Button } from '@flanksource/clicky-ui/components';
import { cn } from '@flanksource/clicky-ui/utils';
import { formatMiB, formatMs, formatSignedBytes, SqlStatements } from './ResourceParts';
import type {
  FixtureBenchmarkDelta,
  FixtureBenchmarkEntry,
  FixturePerformance,
  FixtureProfileSample,
  Test,
} from '../types';

interface Props {
  performance: FixturePerformance;
  tests: Test[];
  error?: string;
  sampleIndex: number | undefined;
  onSampleChange: (index: number) => void;
}

const LIFECYCLE_KINDS = ['setup', 'build', 'daemon', 'daemon-stop', 'cleanup'];

function benchmarkEntries(tests: Test[]): { phases: FixtureBenchmarkEntry[]; fixtures: FixtureBenchmarkEntry[] } {
  const phases: FixtureBenchmarkEntry[] = [];
  const fixtures: FixtureBenchmarkEntry[] = [];
  const visit = (test: Test) => {
    if (test.fixture) {
      const entry: FixtureBenchmarkEntry = {
        key: test.fixture.key, name: test.name, kind: test.fixture.kind, status: test.fixture.state,
        duration_ms: (test.duration || 0) / 1e6, command_ms: test.fixture.command_ms,
        profile: test.fixture_profile, sql_profile: test.sql_profile,
        violations: test.fixture.violations,
      };
      if (LIFECYCLE_KINDS.includes(entry.kind)) phases.push(entry);
      else fixtures.push(entry);
    }
    test.children?.forEach(visit);
  };
  tests.forEach(visit);
  return { phases, fixtures };
}

// EntryRow is a DataTable row: one phase or fixture step with its baseline delta.
type EntryRow = {
  key: string;
  name: string;
  kind: string;
  status: string;
  duration: number;
  command: number | null;
  diskRead: number | null;
  diskWrite: number | null;
  sql: number | null;
  baseline: number | null;
  change: number | null;
  entry: FixtureBenchmarkEntry;
  delta: FixtureBenchmarkDelta | null;
};

function toRows(entries: FixtureBenchmarkEntry[], deltas: FixtureBenchmarkDelta[] | undefined): EntryRow[] {
  const byKey = new Map(deltas?.map(delta => [delta.key, delta]));
  return entries.map(entry => {
    const delta = byKey.get(entry.key) ?? null;
    return {
      key: entry.key,
      name: entry.name,
      kind: entry.kind,
      status: entry.status,
      duration: entry.duration_ms,
      command: entry.command_ms ?? null,
      diskRead: entry.profile?.disk_io?.disk_read_bytes ?? null,
      diskWrite: entry.profile?.disk_io?.disk_write_bytes ?? null,
      sql: entry.sql_profile?.total_duration_ms ?? null,
      baseline: delta?.baseline_ms ?? null,
      change: delta?.delta_pct ?? null,
      entry,
      delta,
    };
  });
}

function entryColumns(compared: boolean): Array<DataTableColumn<EntryRow>> {
  const mono = 'font-mono tabular-nums';
  const columns: Array<DataTableColumn<EntryRow>> = [
    {
      key: 'name', label: 'Step', sortable: true, grow: true,
      render: (_value, row) => (
        <div>
          <span className="font-medium">{row.name}</span>
          <span className="ml-2 text-xs text-muted-foreground">{row.kind}</span>
          {row.entry.violations?.map(reason => <div key={reason} className="text-xs text-red-700 dark:text-red-400">{reason}</div>)}
        </div>
      ),
    },
    { key: 'status', label: 'Status', sortable: true, filterable: true },
    { key: 'duration', label: 'Duration', sortable: true, align: 'right', cellClassName: mono, render: (_value, row) => formatMs(row.duration) },
    {
      key: 'command', label: 'Command', sortable: true, align: 'right', cellClassName: `${mono} text-muted-foreground`,
      render: (_value, row) => row.command === null ? '—' : formatMs(row.command),
    },
    { key: 'diskRead', label: 'Disk read', sortable: true, align: 'right', cellClassName: mono, render: (_value, row) => row.diskRead === null ? '—' : formatMiB(row.diskRead) },
    { key: 'diskWrite', label: 'Disk write', sortable: true, align: 'right', cellClassName: mono, render: (_value, row) => row.diskWrite === null ? '—' : formatMiB(row.diskWrite) },
    {
      key: 'sql', label: 'SQL', sortable: true, align: 'right', cellClassName: mono,
      render: (_value, row) => {
        const sql = row.entry.sql_profile;
        if (!sql) return '—';
        return (
          <div>
            {formatMs(sql.total_duration_ms)}
            <div className="text-xs text-muted-foreground">{sql.query_count} {sql.query_count === 1 ? 'query' : 'queries'} · {sql.slow_query_count} slow</div>
            <div className="text-xs text-muted-foreground">Max {formatMs(sql.max_query_ms)}</div>
            <SqlStatements profile={sql} />
          </div>
        );
      },
    },
  ];
  if (compared) {
    columns.push(
      { key: 'baseline', label: 'Baseline', sortable: true, align: 'right', cellClassName: mono, render: (_value, row) => row.baseline === null ? '—' : formatMs(row.baseline) },
      {
        key: 'change', label: 'Change', sortable: true, align: 'right', cellClassName: mono,
        render: (_value, row) => {
          const delta = row.delta;
          if (!delta) return '—';
          return (
            <>
              <span className={cn(delta.delta_ms > 0 && 'text-red-700 dark:text-red-400', delta.delta_ms < 0 && 'text-green-700 dark:text-green-400')}>
                {delta.delta_pct === undefined ? '—' : `${delta.delta_pct >= 0 ? '+' : ''}${delta.delta_pct.toFixed(1)}%`}
              </span>
              {delta.disk_read_bytes_delta !== undefined && delta.disk_write_bytes_delta !== undefined && (
                <div className="text-xs text-muted-foreground">Disk R {formatSignedBytes(delta.disk_read_bytes_delta)} · W {formatSignedBytes(delta.disk_write_bytes_delta)}</div>
              )}
            </>
          );
        },
      },
    );
  }
  return columns;
}

function EntryTable({ title, entries, deltas }: {
  title: string;
  entries: FixtureBenchmarkEntry[];
  deltas: FixtureBenchmarkDelta[] | undefined;
}) {
  const rows = useMemo(() => toRows(entries, deltas), [entries, deltas]);
  const columns = useMemo(() => entryColumns(!!deltas), [deltas]);
  if (!entries.length) return null;
  return (
    <section className="space-y-1">
      <h2 className="text-sm font-semibold">{title}</h2>
      <DataTable<EntryRow> data={rows} columns={columns} emptyMessage={`No ${title.toLowerCase()}`} />
    </section>
  );
}

// processRows turns a sample's flat process list (pid and ppid) into the
// process-tree records the shared tree renders; a process whose parent was not
// sampled is a root. Disk IO is per sample and shown in the summary above.
function processRows(sample: FixtureProfileSample): ProcessNodeRow[] {
  const pids = new Set(sample.processes.map(process => process.pid));
  return sample.processes.map(process => ({
    key: String(process.pid),
    ...(process.ppid !== undefined && pids.has(process.ppid) ? { parentKey: String(process.ppid) } : {}),
    pid: process.pid,
    ppid: process.ppid ?? 0,
    name: process.command?.split(/\s+/, 1)[0] ?? '',
    command: process.command || `PID ${process.pid}`,
    status: process.status || 'active',
    startedAt: sample.sampled_at,
    lastSeenAt: sample.sampled_at,
    cpuPercent: process.cpuPercent ?? 0,
    rssBytes: process.rssBytes ?? 0,
    peakCpuPercent: process.cpuPercent ?? 0,
    peakRssBytes: process.rssBytes ?? 0,
  }));
}

function ProfileSection({ performance, sampleIndex, onSampleChange }: Pick<Props, 'performance' | 'sampleIndex' | 'onSampleChange'>) {
  const profile = performance.profile;
  const samples = profile?.samples ?? [];
  const selected = sampleIndex !== undefined && sampleIndex < samples.length ? sampleIndex : 0;
  const sample = samples[selected];
  const nodes = useMemo(() => (sample ? processRows(sample) : []), [sample]);

  return (
    <section className="space-y-3">
      <h2 className="text-sm font-semibold">Process profile</h2>
      {!profile ? (
        <p className="text-sm text-muted-foreground">Sampling in progress. Profile appears when the run finishes.</p>
      ) : samples.length === 0 ? (
        <p className="text-sm text-muted-foreground">No process samples were observed.</p>
      ) : (
        <>
          <div className="grid grid-cols-2 gap-density-1 sm:grid-cols-5">
            <MetricTile label="Peak CPU" value={`${profile.peak_cpu_percent.toFixed(1)}%`} />
            <MetricTile label="Peak RSS" value={formatMiB(profile.peak_rss_bytes)} />
            {profile.disk_io && <MetricTile label="Observed disk read" value={formatMiB(profile.disk_io.disk_read_bytes)} />}
            {profile.disk_io && <MetricTile label="Observed disk write" value={formatMiB(profile.disk_io.disk_write_bytes)} />}
            <MetricTile label="Samples" value={`${samples.length} at ${profile.interval_ms} ms`} />
          </div>
          <div className="grid grid-cols-2 gap-density-2">
            <LabeledSparkline label="CPU %" values={samples.map(item => item.cpu_percent)} format={value => `${value.toFixed(1)}%`} />
            <LabeledSparkline label="RSS" values={samples.map(item => item.rss_bytes)} format={formatMiB} />
          </div>
          <div className="flex flex-wrap gap-1" aria-label="Profile samples">
            {samples.map((item, index) => (
              <Button
                key={item.sampled_at}
                type="button"
                size="sm"
                variant={index === selected ? 'default' : 'outline'}
                aria-label={`Sample ${index + 1}`}
                onClick={() => onSampleChange(index)}
              >
                {new Date(item.sampled_at).toLocaleTimeString()}
              </Button>
            ))}
          </div>
          {sample && (
            <div className="space-y-1">
              <div className="text-sm text-muted-foreground">Sample {selected + 1} · CPU {sample.cpu_percent.toFixed(1)}% · RSS {formatMiB(sample.rss_bytes)}</div>
              <div className="h-64 rounded border border-border">
                <ProcessTraceTree className="h-full" nodes={nodes} selectedKey={null} expandAll={true} onSelect={() => {}} emptyLabel="No processes were sampled" />
              </div>
            </div>
          )}
        </>
      )}
    </section>
  );
}

/** A fixture benchmark or profile run: per-step timings against a baseline, and the sampled process tree. */
export function FixtureBenchView({ performance, tests, error, sampleIndex, onSampleChange }: Props) {
  const { phases, fixtures } = useMemo(() => benchmarkEntries(tests), [tests]);
  const comparison = performance.comparison;
  const profile = performance.profile;

  return (
    <div className="mx-auto max-w-7xl space-y-5 p-density-4">
      <div className="flex flex-wrap items-baseline gap-3">
        <h2 className="text-lg font-semibold">Fixture {performance.mode === 'profile' ? 'profile' : 'benchmark'}</h2>
        <span className="text-sm text-muted-foreground">{performance.status === 'running' ? 'Running' : `${performance.status} · ${formatMs(performance.duration_ms)}`}</span>
        {performance.artifact_path && <span className="break-all font-mono text-xs text-muted-foreground">{performance.artifact_path}</span>}
      </div>
      {error && <div role="alert"><Callout variant="caution">{error}</Callout></div>}

      {comparison && (
        <Callout variant="note">
          <div className="space-y-1 text-sm">
            <div>Compared with <span className="font-mono">{comparison.baseline}</span>. Changes describe these two runs only.</div>
            {!!comparison.added?.length && <div>Added: {comparison.added.join(', ')}</div>}
            {!!comparison.missing?.length && <div>Missing: {comparison.missing.join(', ')}</div>}
            {!!comparison.unmeasured?.length && <div>Unmeasured: {comparison.unmeasured.join(', ')}</div>}
            {(comparison.peak_cpu_percent_delta !== undefined || comparison.peak_rss_bytes_delta !== undefined) && (
              <div>
                Peak change: CPU {comparison.peak_cpu_percent_delta === undefined ? '—' : `${comparison.peak_cpu_percent_delta >= 0 ? '+' : ''}${comparison.peak_cpu_percent_delta.toFixed(1)} points`}; RSS {comparison.peak_rss_bytes_delta === undefined ? '—' : formatSignedBytes(comparison.peak_rss_bytes_delta)}
              </div>
            )}
          </div>
        </Callout>
      )}

      <EntryTable title="Run phases" entries={phases} deltas={comparison?.deltas} />
      <EntryTable title="Fixture steps" entries={fixtures} deltas={comparison?.deltas} />

      {performance.mode === 'profile' || profile
        ? <ProfileSection performance={performance} sampleIndex={sampleIndex} onSampleChange={onSampleChange} />
        : null}

      {!phases.length && !fixtures.length && <p className="text-sm text-muted-foreground">Waiting for setup or fixture steps…</p>}
    </div>
  );
}
