import { MetricTile } from '@flanksource/clicky-ui/data';
import { Section } from '@flanksource/clicky-ui/components';
import { formatBytes } from '@flanksource/clicky-ui/utils';
import type { FixtureProfile, FixtureSQLProfile } from '../types';

/** Milliseconds as "125.0 ms" or, from a second, "1.25 s". */
export function formatMs(ms: number): string {
  return ms >= 1000 ? `${(ms / 1000).toFixed(2)} s` : `${ms.toFixed(1)} ms`;
}

/** Bytes with a binary unit, as the fixture benchmark tables report disk and RSS. */
export function formatMiB(value: number): string {
  return value >= 1024 * 1024 ? `${(value / (1024 * 1024)).toFixed(1)} MiB` : value >= 1024 ? `${(value / 1024).toFixed(1)} KiB` : `${value} B`;
}

export function formatSignedBytes(value: number): string {
  return `${value >= 0 ? '+' : '-'}${formatMiB(Math.abs(value))}`;
}

/** The statements a fixture's SQL profile recorded, behind a "N queries" disclosure. */
export function SqlStatements({ profile }: { profile: FixtureSQLProfile }) {
  const statements = profile.statements;
  if (!statements?.length) return null;
  return (
    <Section title={`${statements.length} ${statements.length === 1 ? 'query' : 'queries'}`} defaultOpen={false}>
      <div className="max-h-80 space-y-2 overflow-auto">
        {statements.map((statement, index) => (
          <div key={index} className="rounded border border-border bg-muted/50 p-2 text-xs">
            <div className="mb-1 text-muted-foreground">
              {statement.duration_ms.toFixed(1)} ms · {statement.rows} {statement.rows === 1 ? 'row' : 'rows'}{statement.slow ? ' · slow' : ''}{statement.error ? ' · error' : ''}
            </div>
            <pre className="whitespace-pre-wrap break-all font-mono">{statement.sql}</pre>
            {statement.params.length > 0 && (
              <div className="mt-1">
                <span className="text-muted-foreground">Params</span>
                <ol className="list-decimal pl-5 font-mono">
                  {statement.params.map((param, position) => <li key={position} className="break-all">{param}</li>)}
                </ol>
              </div>
            )}
          </div>
        ))}
      </div>
    </Section>
  );
}

/** Peak CPU/RSS/memory and observed disk of a fixture step's process tree. */
export function ProfileTiles({ profile }: { profile: FixtureProfile }) {
  return (
    <div className="space-y-2">
      {profile.sample_count > 0 ? (
        <div className="grid grid-cols-3 gap-density-1">
          <MetricTile label="Peak CPU" value={`${profile.peak_cpu_percent.toFixed(1)}%`} />
          <MetricTile label="Peak RSS" value={formatBytes(profile.peak_rss_bytes)} />
          <MetricTile label="Peak memory" value={`${profile.peak_memory_percent.toFixed(1)}%`} />
        </div>
      ) : (
        <p className="text-sm text-muted-foreground">No process samples were captured during this test.</p>
      )}
      {profile.disk_io && (
        <div className="grid grid-cols-2 gap-density-1">
          <MetricTile label="Observed disk read" value={formatBytes(profile.disk_io.disk_read_bytes)} />
          <MetricTile label="Observed disk write" value={formatBytes(profile.disk_io.disk_write_bytes)} />
        </div>
      )}
      <p className="text-xs text-muted-foreground">
        {profile.sample_count} samples · {profile.scope === 'process_tree' ? `Command process tree · PID ${profile.pid}` : 'Run process tree during this test'}
      </p>
    </div>
  );
}
