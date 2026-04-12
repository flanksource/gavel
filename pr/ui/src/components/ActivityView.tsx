import { useState, useEffect, useMemo } from 'preact/hooks';
import type { ActivitySnapshot, ActivityEntry, ActivityKindStats, CacheStatus } from '../types';
import { timeAgo } from '../utils';

const KIND_LABELS: Record<string, string> = {
  rest: 'REST',
  graphql: 'GraphQL',
  search: 'Search',
};

const KIND_COLORS: Record<string, string> = {
  rest: 'bg-blue-100 text-blue-700',
  graphql: 'bg-purple-100 text-purple-700',
  search: 'bg-amber-100 text-amber-700',
};

export function ActivityView() {
  const [snap, setSnap] = useState<ActivitySnapshot>({
    entries: [],
    stats: { total: 0, cacheHits: 0, errors: 0, totalBytes: 0, totalNs: 0, byKind: {} },
  });
  const [cache, setCache] = useState<CacheStatus | null>(null);
  const [kindFilter, setKindFilter] = useState<string>('');
  const [, tick] = useState(0);

  const refreshCache = () => {
    fetch('/api/activity/cache')
      .then(r => r.json())
      .then((c: CacheStatus) => setCache(c))
      .catch(() => {});
  };

  useEffect(() => {
    fetch('/api/activity')
      .then(r => r.json())
      .then((s: ActivitySnapshot) => setSnap(s))
      .catch(() => {});
    refreshCache();

    const es = new EventSource('/api/activity/stream');
    es.addEventListener('message', (e: MessageEvent) => {
      try {
        setSnap(JSON.parse(e.data));
      } catch { /* ignore */ }
    });
    es.onerror = () => { /* auto-reconnect */ };

    // Cache status changes rarely — refresh every 10s.
    const cacheTimer = setInterval(refreshCache, 10000);
    const timer = setInterval(() => tick(n => n + 1), 1000);
    return () => { es.close(); clearInterval(timer); clearInterval(cacheTimer); };
  }, []);

  const filtered = useMemo(
    () => kindFilter ? snap.entries.filter(e => e.kind === kindFilter) : snap.entries,
    [snap.entries, kindFilter],
  );

  const handleReset = () => {
    fetch('/api/activity/reset', { method: 'POST' })
      .then(() => setSnap({ entries: [], stats: { total: 0, cacheHits: 0, errors: 0, totalBytes: 0, totalNs: 0, byKind: {} } }))
      .catch(() => {});
  };

  const { stats } = snap;
  const hitRate = stats.total > 0 ? (stats.cacheHits / stats.total) * 100 : 0;
  const avgMs = stats.total > 0 ? stats.totalNs / stats.total / 1e6 : 0;

  return (
    <div class="bg-gray-50 h-full overflow-y-auto p-6">
      <div class="max-w-6xl mx-auto">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-lg font-semibold text-gray-900">
            <iconify-icon icon="codicon:pulse" class="mr-1.5 text-blue-600" />
            HTTP Activity
          </h2>
          <button
            onClick={handleReset}
            class="text-xs px-3 py-1.5 bg-white border border-gray-200 rounded hover:bg-gray-50 text-gray-700"
            title="Clear all recorded activity"
          >
            <iconify-icon icon="codicon:trash" class="mr-1" />
            Reset
          </button>
        </div>

        <div class="grid grid-cols-2 md:grid-cols-4 gap-3 mb-4">
          <KPI
            label="Total requests"
            value={stats.total.toLocaleString()}
            sub={stats.errors > 0 ? `${stats.errors} errors` : 'no errors'}
            subClass={stats.errors > 0 ? 'text-red-600' : 'text-green-600'}
            icon="codicon:globe"
          />
          <KPI
            label="Cache hit rate"
            value={`${hitRate.toFixed(1)}%`}
            sub={`${stats.cacheHits.toLocaleString()} / ${stats.total.toLocaleString()}`}
            subClass="text-gray-500"
            icon="codicon:database"
          />
          <KPI
            label="Bandwidth"
            value={formatBytes(stats.totalBytes)}
            sub={stats.total > 0 ? `${formatBytes(stats.totalBytes / stats.total)} avg` : '—'}
            subClass="text-gray-500"
            icon="codicon:cloud-download"
          />
          <KPI
            label="Avg latency"
            value={`${avgMs.toFixed(0)} ms`}
            sub={`${(stats.totalNs / 1e9).toFixed(1)}s total`}
            subClass="text-gray-500"
            icon="codicon:watch"
          />
        </div>

        {cache && <CachePanel cache={cache} />}

        <div class="bg-white border border-gray-200 rounded-md mb-4 p-3">
          <div class="text-xs font-semibold text-gray-500 uppercase mb-2">By kind</div>
          <div class="flex gap-2 flex-wrap">
            <KindChip kind="" label="All" active={kindFilter === ''} onClick={() => setKindFilter('')} count={stats.total} />
            {Object.entries(stats.byKind).map(([kind, ks]) => (
              <KindChip
                key={kind}
                kind={kind}
                label={KIND_LABELS[kind] || kind}
                active={kindFilter === kind}
                onClick={() => setKindFilter(kindFilter === kind ? '' : kind)}
                count={ks.total}
                stats={ks}
              />
            ))}
          </div>
        </div>

        <div class="bg-white border border-gray-200 rounded-md overflow-hidden">
          <table class="w-full text-xs">
            <thead class="bg-gray-50 text-gray-500 uppercase">
              <tr>
                <th class="px-3 py-2 text-left font-medium">Time</th>
                <th class="px-3 py-2 text-left font-medium">Kind</th>
                <th class="px-3 py-2 text-left font-medium">Method</th>
                <th class="px-3 py-2 text-left font-medium">URL</th>
                <th class="px-3 py-2 text-right font-medium">Status</th>
                <th class="px-3 py-2 text-right font-medium">Duration</th>
                <th class="px-3 py-2 text-right font-medium">Size</th>
                <th class="px-3 py-2 text-center font-medium">Cache</th>
              </tr>
            </thead>
            <tbody>
              {filtered.length === 0 && (
                <tr>
                  <td colSpan={8} class="px-3 py-6 text-center text-gray-400">
                    No requests recorded yet. Interact with the PR dashboard to generate activity.
                  </td>
                </tr>
              )}
              {filtered.map((e, i) => (
                <ActivityRow key={`${e.timestamp}-${i}`} entry={e} />
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

function KPI({ label, value, sub, subClass, icon }: { label: string; value: string; sub: string; subClass: string; icon: string }) {
  return (
    <div class="bg-white border border-gray-200 rounded-md p-3">
      <div class="flex items-center gap-1.5 text-xs text-gray-500">
        <iconify-icon icon={icon} />
        {label}
      </div>
      <div class="text-2xl font-semibold text-gray-900 mt-1">{value}</div>
      <div class={`text-xs mt-0.5 ${subClass}`}>{sub}</div>
    </div>
  );
}

function KindChip({ kind, label, active, onClick, count, stats }: {
  kind: string; label: string; active: boolean; onClick: () => void; count: number; stats?: ActivityKindStats;
}) {
  const colorClass = kind ? KIND_COLORS[kind] || 'bg-gray-100 text-gray-700' : 'bg-gray-100 text-gray-700';
  const avgMs = stats && stats.total > 0 ? stats.totalNs / stats.total / 1e6 : 0;
  const hitRate = stats && stats.total > 0 ? (stats.cacheHits / stats.total) * 100 : 0;
  return (
    <button
      onClick={onClick}
      class={`text-xs px-2.5 py-1 rounded border ${active ? 'border-blue-500 ring-1 ring-blue-200' : 'border-gray-200'} ${colorClass} hover:opacity-90`}
    >
      <span class="font-semibold">{label}</span>
      <span class="ml-1 opacity-70">{count}</span>
      {stats && (
        <span class="ml-1.5 opacity-60">
          · {avgMs.toFixed(0)}ms · {hitRate.toFixed(0)}% hit
        </span>
      )}
    </button>
  );
}

function ActivityRow({ entry }: { entry: ActivityEntry }) {
  const ms = entry.durationNs / 1e6;
  const statusClass = entry.error || entry.statusCode >= 400
    ? 'text-red-600'
    : entry.statusCode === 304
      ? 'text-blue-600'
      : 'text-gray-700';
  return (
    <tr class={`border-t border-gray-100 ${entry.error ? 'bg-red-50' : ''}`}>
      <td class="px-3 py-1.5 text-gray-500 whitespace-nowrap">{timeAgo(entry.timestamp)}</td>
      <td class="px-3 py-1.5">
        <span class={`px-1.5 py-0.5 rounded text-[10px] font-semibold ${KIND_COLORS[entry.kind] || 'bg-gray-100 text-gray-700'}`}>
          {KIND_LABELS[entry.kind] || entry.kind}
        </span>
      </td>
      <td class="px-3 py-1.5 font-mono text-gray-600">{entry.method}</td>
      <td class="px-3 py-1.5 font-mono text-gray-700 truncate max-w-md" title={entry.url}>
        {entry.url}
        {entry.error && <div class="text-red-600 text-[10px]">{entry.error}</div>}
      </td>
      <td class={`px-3 py-1.5 text-right ${statusClass}`}>{entry.statusCode || '—'}</td>
      <td class="px-3 py-1.5 text-right text-gray-600 tabular-nums">{ms.toFixed(0)} ms</td>
      <td class="px-3 py-1.5 text-right text-gray-600 tabular-nums">{formatBytes(entry.sizeBytes)}</td>
      <td class="px-3 py-1.5 text-center">
        {entry.fromCache ? (
          <span class="text-green-600" title="Served from cache (304)">
            <iconify-icon icon="codicon:check" />
          </span>
        ) : (
          <span class="text-gray-300">—</span>
        )}
      </td>
    </tr>
  );
}

function CachePanel({ cache }: { cache: CacheStatus }) {
  const totalRows = Object.values(cache.counts || {}).reduce((a, b) => a + b, 0);
  return (
    <div class={`bg-white border rounded-md mb-4 p-3 ${cache.enabled ? 'border-gray-200' : 'border-amber-300 bg-amber-50'}`}>
      <div class="flex items-center justify-between mb-2">
        <div class="flex items-center gap-2">
          <iconify-icon icon="codicon:database" class={cache.enabled ? 'text-green-600' : 'text-amber-600'} />
          <span class="text-xs font-semibold text-gray-500 uppercase">Cache</span>
          <span class={`text-xs px-2 py-0.5 rounded ${cache.enabled ? 'bg-green-100 text-green-700' : 'bg-amber-100 text-amber-700'}`}>
            {cache.enabled ? 'ENABLED' : 'DISABLED'}
          </span>
        </div>
        {cache.error && (
          <span class="text-xs text-amber-700" title={cache.error}>
            <iconify-icon icon="codicon:warning" class="mr-1" />
            {cache.error}
          </span>
        )}
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-3 text-xs">
        <div>
          <div class="text-gray-500">Driver</div>
          <div class="font-mono text-gray-800">{cache.driver}</div>
        </div>
        <div>
          <div class="text-gray-500">DSN source</div>
          <div class="font-mono text-gray-800">
            {cache.dsnSource || <span class="text-gray-400">—</span>}
          </div>
          {cache.dsnMasked && (
            <div class="font-mono text-gray-500 text-[10px] truncate" title={cache.dsnMasked}>
              {cache.dsnMasked}
            </div>
          )}
        </div>
        <div>
          <div class="text-gray-500">Retention</div>
          <div class="font-mono text-gray-800">{formatDuration(cache.retentionSec)}</div>
        </div>
      </div>

      {cache.enabled && Object.keys(cache.counts || {}).length > 0 && (
        <div class="mt-3 pt-3 border-t border-gray-100">
          <div class="flex items-center justify-between mb-1.5">
            <div class="text-xs text-gray-500">Rows</div>
            <div class="text-xs text-gray-500">{totalRows.toLocaleString()} total</div>
          </div>
          <div class="grid grid-cols-2 md:grid-cols-4 gap-2">
            {Object.entries(cache.counts).map(([table, n]) => (
              <div key={table} class="bg-gray-50 rounded px-2 py-1.5">
                <div class="text-[10px] text-gray-500 font-mono truncate" title={table}>{table}</div>
                <div class="text-sm font-semibold text-gray-800 tabular-nums">{n.toLocaleString()}</div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function formatDuration(seconds: number): string {
  if (!seconds || seconds <= 0) return '—';
  const days = Math.floor(seconds / 86400);
  if (days >= 1) return `${days}d`;
  const hours = Math.floor(seconds / 3600);
  if (hours >= 1) return `${hours}h`;
  const mins = Math.floor(seconds / 60);
  if (mins >= 1) return `${mins}m`;
  return `${seconds}s`;
}

function formatBytes(n: number): string {
  if (!n || n < 0) return '0 B';
  if (n < 1024) return `${n.toFixed(0)} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  if (n < 1024 * 1024 * 1024) return `${(n / 1024 / 1024).toFixed(1)} MB`;
  return `${(n / 1024 / 1024 / 1024).toFixed(2)} GB`;
}
