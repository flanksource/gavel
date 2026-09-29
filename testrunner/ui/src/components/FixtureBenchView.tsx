import type { FixtureBenchmarkDelta, FixtureBenchmarkEntry, FixturePerformance, FixtureProcess, FixtureProfileSample, Test } from '../types';
import { SQLProfileDetails } from './SQLProfileDetails';

interface Props {
  performance: FixturePerformance;
  tests: Test[];
  error?: string;
  sampleIndex: number | undefined;
  onSampleChange: (index: number) => void;
}

function duration(ms: number): string {
  return ms >= 1000 ? `${(ms / 1000).toFixed(2)} s` : `${ms.toFixed(1)} ms`;
}

function bytes(value: number): string {
  return value >= 1024 * 1024 ? `${(value / (1024 * 1024)).toFixed(1)} MiB` : value >= 1024 ? `${(value / 1024).toFixed(1)} KiB` : `${value} B`;
}

function signedBytes(value: number): string {
  return `${value >= 0 ? '+' : '-'}${bytes(Math.abs(value))}`;
}

function benchmarkEntries(tests: Test[]): { phases: FixtureBenchmarkEntry[]; fixtures: FixtureBenchmarkEntry[] } {
  const phases: FixtureBenchmarkEntry[] = [];
  const fixtures: FixtureBenchmarkEntry[] = [];
  const visit = (test: Test) => {
    if (test.fixture) {
      const entry: FixtureBenchmarkEntry = {
        key: test.fixture.key, name: test.name, kind: test.fixture.kind, status: test.fixture.state,
        duration_ms: (test.duration || 0) / 1e6, command_ms: test.fixture.command_ms,
        profile: test.fixture_profile, go_profiles: test.go_profiles, sql_profile: test.sql_profile,
        violations: test.fixture.violations,
      };
      if (['setup', 'build', 'daemon', 'daemon-stop', 'cleanup'].includes(entry.kind)) phases.push(entry);
      else fixtures.push(entry);
    }
    test.children?.forEach(visit);
  };
  tests.forEach(visit);
  return { phases, fixtures };
}

function EntryTable({ title, entries, deltas }: { title: string; entries: FixtureBenchmarkEntry[]; deltas?: FixtureBenchmarkDelta[] }) {
  if (!entries.length) return null;
  const byKey = new Map(deltas?.map(delta => [delta.key, delta]));
  return (
    <section className="bg-white border border-gray-200 rounded-lg overflow-hidden">
      <h2 className="font-semibold text-gray-800 px-4 py-3 border-b">{title}</h2>
      <div className="overflow-x-auto"><table className="w-full text-sm">
        <thead><tr className="text-left text-xs text-gray-500 bg-gray-50 border-b"><th className="px-4 py-2">Step</th><th className="px-4 py-2">Status</th><th className="px-4 py-2 text-right">Duration</th><th className="px-4 py-2 text-right">Command</th><th className="px-4 py-2 text-right">Disk read</th><th className="px-4 py-2 text-right">Disk write</th><th className="px-4 py-2 text-right">SQL</th>{deltas && <><th className="px-4 py-2 text-right">Baseline</th><th className="px-4 py-2 text-right">Change</th></>}</tr></thead>
        <tbody>{entries.map(entry => {
          const delta = byKey.get(entry.key);
          return <tr key={entry.key} className="border-b border-gray-100 last:border-0">
            <td className="px-4 py-2"><span className="font-medium">{entry.name}</span><span className="ml-2 text-xs text-gray-400">{entry.kind}</span>{entry.violations?.map(reason => <div key={reason} className="text-xs text-red-700">{reason}</div>)}{entry.go_profiles?.map(artifact => <div key={`${artifact.name}-${artifact.id || artifact.status}`} className="text-xs text-gray-500">{artifact.name}: {artifact.status === 'captured' && artifact.id ? <a className="text-blue-600 hover:underline" href={`/api/tests/fixture-profile?id=${encodeURIComponent(artifact.id)}`}>Download pprof</a> : artifact.status === 'not_emitted' ? 'Not emitted' : artifact.error || 'Invalid'}</div>)}</td>
            <td className="px-4 py-2 text-gray-600">{entry.status}</td>
            <td className="px-4 py-2 text-right font-mono tabular-nums">{duration(entry.duration_ms)}</td>
            <td className="px-4 py-2 text-right font-mono tabular-nums text-gray-500">{entry.command_ms === undefined ? '—' : duration(entry.command_ms)}</td>
            <td className="px-4 py-2 text-right font-mono tabular-nums">{entry.profile?.disk_io ? bytes(entry.profile.disk_io.disk_read_bytes) : '—'}</td>
            <td className="px-4 py-2 text-right font-mono tabular-nums">{entry.profile?.disk_io ? bytes(entry.profile.disk_io.disk_write_bytes) : '—'}</td>
            <td className="px-4 py-2 text-right font-mono tabular-nums">{entry.sql_profile ? <>{duration(entry.sql_profile.total_duration_ms)}<div className="text-xs text-gray-500">{entry.sql_profile.query_count} {entry.sql_profile.query_count === 1 ? 'query' : 'queries'} · {entry.sql_profile.slow_query_count} slow</div><div className="text-xs text-gray-500">Max {duration(entry.sql_profile.max_query_ms)}</div><SQLProfileDetails profile={entry.sql_profile} /></> : '—'}</td>
            {deltas && <><td className="px-4 py-2 text-right font-mono tabular-nums">{delta ? duration(delta.baseline_ms) : '—'}</td><td className="px-4 py-2 text-right font-mono tabular-nums">{delta ? <><span className={delta.delta_ms > 0 ? 'text-red-700' : delta.delta_ms < 0 ? 'text-green-700' : ''}>{delta.delta_pct === undefined ? '—' : `${delta.delta_pct >= 0 ? '+' : ''}${delta.delta_pct.toFixed(1)}%`}</span>{delta.disk_read_bytes_delta !== undefined && delta.disk_write_bytes_delta !== undefined && <div className="text-xs text-gray-500">Disk R {signedBytes(delta.disk_read_bytes_delta)} · W {signedBytes(delta.disk_write_bytes_delta)}</div>}</> : '—'}</td></>}
          </tr>;
        })}</tbody>
      </table></div>
    </section>
  );
}

function Series({ samples, field, color, label }: { samples: FixtureProfileSample[]; field: 'cpu_percent' | 'rss_bytes'; color: string; label: string }) {
  const max = Math.max(1, ...samples.map(sample => sample[field]));
  const points = samples.map((sample, index) => `${samples.length === 1 ? 0 : index * 100 / (samples.length - 1)},${34 - sample[field] * 30 / max}`).join(' ');
  return <div className="min-w-0"><div className="text-xs font-medium text-gray-600 mb-1">{label}</div><svg role="img" aria-label={`${label} over time`} viewBox="0 0 100 38" preserveAspectRatio="none" className="w-full h-20 bg-gray-50 rounded border border-gray-100"><polyline fill="none" stroke={color} strokeWidth="1.5" points={points} vectorEffect="non-scaling-stroke" /></svg></div>;
}

function ProcessRows({ processes, parent, depth = 0 }: { processes: FixtureProcess[]; parent: number | undefined; depth?: number }) {
  const children = processes.filter(process => parent === undefined ? !processes.some(other => other.pid === process.ppid) : process.ppid === parent);
  return <>{children.map(process => <div key={process.pid}>
    <div className="grid grid-cols-[1fr_5rem_6rem_6rem_6rem] gap-3 py-1.5 border-b border-gray-100 text-xs" style={{ paddingLeft: `${depth * 16}px` }}>
      <span className="truncate font-mono" title={process.command}>{process.command || `PID ${process.pid}`} <span className="text-gray-400">#{process.pid}</span></span>
      <span className="text-right tabular-nums">{(process.cpuPercent || 0).toFixed(1)}%</span>
      <span className="text-right tabular-nums">{bytes(process.rssBytes || 0)}</span>
      <span className="text-right tabular-nums">{process.io ? bytes(process.io.diskReadBytes) : '—'}</span>
      <span className="text-right tabular-nums">{process.io ? bytes(process.io.diskWriteBytes) : '—'}</span>
    </div>
    <ProcessRows processes={processes} parent={process.pid} depth={depth + 1} />
  </div>)}</>;
}

export function FixtureBenchView({ performance, tests, error, sampleIndex, onSampleChange }: Props) {
  const { phases, fixtures } = benchmarkEntries(tests);
  const comparison = performance.comparison;
  const profile = performance.profile;
  const selected = sampleIndex !== undefined && sampleIndex < (profile?.samples.length || 0) ? sampleIndex : 0;
  const sample = profile?.samples[selected];

  return <div className="p-5 max-w-7xl mx-auto space-y-5 text-gray-800">
    <div className="flex flex-wrap items-baseline gap-3">
      <h2 className="text-lg font-semibold">Fixture {performance.mode === 'profile' ? 'profile' : 'benchmark'}</h2>
      <span className="text-sm text-gray-500">{performance.status === 'running' ? 'Running' : `${performance.status} · ${duration(performance.duration_ms)}`}</span>
      {performance.artifact_path && <span className="text-xs text-gray-500 font-mono break-all">{performance.artifact_path}</span>}
    </div>
	{error && <div role="alert" className="bg-red-50 border border-red-200 text-red-800 rounded-lg p-3 text-sm">{error}</div>}

    {comparison && <section className="bg-blue-50 border border-blue-100 rounded-lg p-4 text-sm space-y-2">
      <div>Compared with <span className="font-mono">{comparison.baseline}</span>. Changes describe these two runs only.</div>
      {!!comparison.added?.length && <div>Added: {comparison.added.join(', ')}</div>}
      {!!comparison.missing?.length && <div>Missing: {comparison.missing.join(', ')}</div>}
      {!!comparison.unmeasured?.length && <div>Unmeasured: {comparison.unmeasured.join(', ')}</div>}
      {(comparison.peak_cpu_percent_delta !== undefined || comparison.peak_rss_bytes_delta !== undefined) && <div>Peak change: CPU {comparison.peak_cpu_percent_delta === undefined ? '—' : `${comparison.peak_cpu_percent_delta >= 0 ? '+' : ''}${comparison.peak_cpu_percent_delta.toFixed(1)} points`}; RSS {comparison.peak_rss_bytes_delta === undefined ? '—' : signedBytes(comparison.peak_rss_bytes_delta)}</div>}
    </section>}

    <EntryTable title="Run phases" entries={phases} deltas={comparison?.deltas} />
    <EntryTable title="Fixture steps" entries={fixtures} deltas={comparison?.deltas} />

    {performance.mode === 'profile' || profile ? <section className="bg-white border border-gray-200 rounded-lg p-4 space-y-4">
      <h2 className="font-semibold">Process profile</h2>
      {!profile ? <p className="text-sm text-gray-500">Sampling in progress. Profile appears when the run finishes.</p> : profile.samples.length === 0 ? <p className="text-sm text-gray-500">No process samples were observed.</p> : <>
        <div className="flex flex-wrap gap-6 text-sm"><div><span className="text-gray-500">Peak CPU</span><div className="font-mono font-semibold">{profile.peak_cpu_percent.toFixed(1)}%</div></div><div><span className="text-gray-500">Peak RSS</span><div className="font-mono font-semibold">{bytes(profile.peak_rss_bytes)}</div></div>{profile.disk_io && <><div><span className="text-gray-500">Observed disk read</span><div className="font-mono font-semibold">{bytes(profile.disk_io.disk_read_bytes)}</div></div><div><span className="text-gray-500">Observed disk write</span><div className="font-mono font-semibold">{bytes(profile.disk_io.disk_write_bytes)}</div></div></>}<div><span className="text-gray-500">Samples</span><div className="font-mono font-semibold">{profile.samples.length} at {profile.interval_ms} ms</div></div></div>
        <div className="grid grid-cols-2 gap-4"><Series samples={profile.samples} field="cpu_percent" color="#2563eb" label="CPU %" /><Series samples={profile.samples} field="rss_bytes" color="#059669" label="RSS bytes" /></div>
        <div className="flex flex-wrap gap-1" aria-label="Profile samples">{profile.samples.map((item, index) => <button key={item.sampled_at} type="button" aria-label={`Sample ${index + 1}`} onClick={() => onSampleChange(index)} className={`text-xs px-2 py-1 rounded border ${index === selected ? 'border-blue-500 bg-blue-50 text-blue-700' : 'border-gray-200 hover:bg-gray-50'}`}>{new Date(item.sampled_at).toLocaleTimeString()}</button>)}</div>
        {sample && <div><div className="text-sm text-gray-600 mb-2">Sample {selected + 1} · CPU {sample.cpu_percent.toFixed(1)}% · RSS {bytes(sample.rss_bytes)}</div><div className="grid grid-cols-[1fr_5rem_6rem_6rem_6rem] gap-3 text-xs font-semibold text-gray-500 border-b pb-1"><span>Process tree</span><span className="text-right">CPU</span><span className="text-right">RSS</span><span className="text-right">Disk read</span><span className="text-right">Disk write</span></div><ProcessRows processes={sample.processes} parent={undefined} /></div>}
      </>}
    </section> : null}

    {!phases.length && !fixtures.length && <p className="text-sm text-gray-500">Waiting for setup or fixture steps…</p>}
  </div>;
}
