import { useEffect, useState } from 'preact/hooks';
import type { RunIndexEntry } from '../types';
import { apiUrl } from '../config';

interface Props {
  onSelect: (entry: RunIndexEntry) => void;
}

function relativeTime(iso: string): string {
  const t = Date.parse(iso);
  if (Number.isNaN(t)) return '';
  const diff = Date.now() - t;
  if (diff < 60_000) return 'just now';
  if (diff < 3_600_000) return `${Math.floor(diff / 60_000)} min ago`;
  if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)} h ago`;
  if (diff < 7 * 86_400_000) return `${Math.floor(diff / 86_400_000)} d ago`;
  return new Date(t).toISOString().slice(0, 16).replace('T', ' ');
}

function formatDuration(startedISO: string | undefined, endedISO: string | undefined): string {
  if (!startedISO || !endedISO) return '';
  const start = Date.parse(startedISO);
  const end = Date.parse(endedISO);
  if (Number.isNaN(start) || Number.isNaN(end) || end < start) return '';
  const ms = end - start;
  if (ms < 1000) return `${ms} ms`;
  if (ms < 60_000) return `${(ms / 1000).toFixed(1)} s`;
  const m = Math.floor(ms / 60_000);
  const s = Math.floor((ms % 60_000) / 1000);
  return `${m}m ${s}s`;
}

function entryKey(entry: RunIndexEntry): string {
  return entry.pointer ? `pointer:${entry.pointer}` : `name:${entry.name}`;
}

export function RunIndex({ onSelect }: Props) {
  const [entries, setEntries] = useState<RunIndexEntry[] | null>(null);
  const [error, setError] = useState<string>('');

  useEffect(() => {
    let cancelled = false;
    fetch(apiUrl('/api/runs'))
      .then(async r => {
        if (!r.ok) throw new Error(`HTTP ${r.status}: ${(await r.text()).trim()}`);
        return r.json() as Promise<RunIndexEntry[]>;
      })
      .then(data => { if (!cancelled) setEntries(data); })
      .catch(e => { if (!cancelled) setError(e?.message || String(e)); });
    return () => { cancelled = true; };
  }, []);

  return (
    <div class="bg-gray-100 h-full overflow-auto">
      <div class="max-w-5xl mx-auto px-6 py-6">
        <div class="flex items-center gap-3 mb-4">
          <h1 class="text-xl font-bold text-gray-900">
            <iconify-icon icon="codicon:history" class="mr-1.5 text-blue-600" />
            Saved Runs
          </h1>
          {entries && (
            <span class="text-sm text-gray-500">
              {entries.length} {entries.length === 1 ? 'snapshot' : 'snapshots'} in <code class="bg-gray-200 px-1 rounded">.gavel/</code>
            </span>
          )}
        </div>

        {error && (
          <div class="rounded border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
            Failed to load run index: {error}
          </div>
        )}

        {!entries && !error && (
          <div class="p-8 text-center text-gray-400">
            <iconify-icon icon="svg-spinners:ring-resize" class="text-3xl text-blue-500" />
            <p class="mt-2">Loading runs...</p>
          </div>
        )}

        {entries && entries.length === 0 && (
          <div class="p-8 text-center text-gray-400 text-sm">
            No snapshots found in <code>.gavel/</code>. Run <code>gavel test</code> to populate it.
          </div>
        )}

        {entries && entries.length > 0 && (
          <div class="bg-white rounded-md shadow-sm overflow-hidden">
            <table class="w-full text-sm">
              <thead class="bg-gray-50 text-gray-500 text-xs uppercase">
                <tr>
                  <th class="text-left px-3 py-2 font-medium">Run</th>
                  <th class="text-left px-3 py-2 font-medium">Modified</th>
                  <th class="text-right px-3 py-2 font-medium">Tests</th>
                  <th class="text-right px-3 py-2 font-medium">Lint</th>
                  <th class="text-right px-3 py-2 font-medium">Duration</th>
                </tr>
              </thead>
              <tbody>
                {entries.map(entry => (
                  <RunIndexRow key={entryKey(entry)} entry={entry} onSelect={onSelect} />
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}

interface RowProps {
  entry: RunIndexEntry;
  onSelect: (entry: RunIndexEntry) => void;
}

function RunIndexRow({ entry, onSelect }: RowProps) {
  const counts = entry.counts;
  const failed = counts?.failed ?? 0;
  const passed = counts?.passed ?? 0;
  const skipped = counts?.skipped ?? 0;
  const pending = counts?.pending ?? 0;
  const total = counts?.total ?? 0;
  const lint = entry.lint ?? 0;

  return (
    <tr
      class="border-t border-gray-100 hover:bg-blue-50 cursor-pointer"
      onClick={() => onSelect(entry)}
    >
      <td class="px-3 py-2">
        <div class="flex items-center gap-2">
          {entry.pointer ? (
            <span class="inline-flex items-center gap-1 rounded-full bg-blue-100 text-blue-700 px-2 py-0.5 text-xs font-medium">
              <iconify-icon icon="codicon:bookmark" />
              {entry.pointer}
            </span>
          ) : (
            <iconify-icon icon="codicon:file" class="text-gray-400" />
          )}
          <span class="font-mono text-gray-700 text-xs truncate max-w-[28rem]">
            {entry.pointer ? entry.name : entry.name}
          </span>
          {entry.sha && (
            <span class="font-mono text-gray-400 text-xs">@{entry.sha}</span>
          )}
        </div>
        {entry.error && (
          <div class="text-xs text-red-600 mt-1">{entry.error}</div>
        )}
      </td>
      <td class="px-3 py-2 text-gray-600 whitespace-nowrap">
        {relativeTime(entry.modified)}
      </td>
      <td class="px-3 py-2 text-right whitespace-nowrap">
        {total > 0 ? (
          <span class="inline-flex items-center gap-1.5">
            {failed > 0 && <span class="text-red-600 font-semibold">{failed} failed</span>}
            {failed > 0 && passed > 0 && <span class="text-gray-300">·</span>}
            {passed > 0 && <span class="text-green-600">{passed} passed</span>}
            {skipped > 0 && <><span class="text-gray-300">·</span><span class="text-gray-500">{skipped} skipped</span></>}
            {pending > 0 && <><span class="text-gray-300">·</span><span class="text-blue-500">{pending} pending</span></>}
          </span>
        ) : (
          <span class="text-gray-300">—</span>
        )}
      </td>
      <td class="px-3 py-2 text-right whitespace-nowrap">
        {lint > 0 ? (
          <span class="text-yellow-700">{lint}</span>
        ) : (
          <span class="text-gray-300">—</span>
        )}
      </td>
      <td class="px-3 py-2 text-right text-gray-600 whitespace-nowrap font-mono text-xs">
        {formatDuration(entry.started, entry.ended) || <span class="text-gray-300">—</span>}
      </td>
    </tr>
  );
}
