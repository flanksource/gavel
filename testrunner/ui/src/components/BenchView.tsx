import { useMemo, useState } from 'preact/hooks';
import type { BenchComparison, BenchDelta } from '../types';
import { BenchDeltaBar } from './BenchDeltaBar';

interface Props {
  bench: BenchComparison | undefined;
}

type SortKey = 'name' | 'delta' | 'base' | 'head' | 'p';

function formatNs(ns: number): string {
  if (ns >= 1e9) return `${(ns / 1e9).toFixed(2)}s`;
  if (ns >= 1e6) return `${(ns / 1e6).toFixed(2)}ms`;
  if (ns >= 1e3) return `${(ns / 1e3).toFixed(2)}µs`;
  return `${ns.toFixed(2)}ns`;
}

function categorize(d: BenchDelta, threshold: number): 'regression' | 'improvement' | 'neutral' | 'only' {
  if (d.only_in) return 'only';
  if (d.significant && d.delta_pct > threshold) return 'regression';
  if (d.significant && d.delta_pct < -threshold) return 'improvement';
  return 'neutral';
}

export function BenchView({ bench }: Props) {
  const [sortKey, setSortKey] = useState<SortKey>('delta');
  const [sortDesc, setSortDesc] = useState(true);

  if (!bench) {
    return (
      <div class="p-8 text-center text-gray-400 text-sm">
        No benchmark comparison loaded. Run <code class="px-1 bg-gray-200 rounded">gavel bench compare --ui</code>.
      </div>
    );
  }

  const sorted = useMemo(() => {
    const deltas = [...(bench.deltas || [])];
    const cmp = (a: BenchDelta, b: BenchDelta) => {
      switch (sortKey) {
        case 'name': return a.name.localeCompare(b.name);
        case 'delta': return (a.delta_pct || 0) - (b.delta_pct || 0);
        case 'base': return (a.base_mean || 0) - (b.base_mean || 0);
        case 'head': return (a.head_mean || 0) - (b.head_mean || 0);
        case 'p': return (a.p_value || 1) - (b.p_value || 1);
      }
    };
    deltas.sort((a, b) => {
      const v = cmp(a, b);
      return sortDesc ? -v : v;
    });
    return deltas;
  }, [bench, sortKey, sortDesc]);

  const summary = useMemo(() => {
    let regressions = 0, improvements = 0, neutral = 0, only = 0;
    for (const d of bench.deltas || []) {
      const c = categorize(d, bench.threshold);
      if (c === 'regression') regressions++;
      else if (c === 'improvement') improvements++;
      else if (c === 'only') only++;
      else neutral++;
    }
    return { regressions, improvements, neutral, only };
  }, [bench]);

  const onSort = (key: SortKey) => {
    if (sortKey === key) setSortDesc(!sortDesc);
    else { setSortKey(key); setSortDesc(true); }
  };

  const sortIcon = (key: SortKey) => {
    if (sortKey !== key) return <iconify-icon icon="codicon:chevron-down" class="opacity-20" />;
    return <iconify-icon icon={sortDesc ? 'codicon:chevron-down' : 'codicon:chevron-up'} />;
  };

  return (
    <div class="p-4">
      <div class="mb-3 flex items-center gap-4 text-sm">
        <div class="text-gray-600">
          <span class="font-mono">{bench.base_label || 'base'}</span>
          <span class="mx-2 text-gray-400">→</span>
          <span class="font-mono">{bench.head_label || 'head'}</span>
        </div>
        <div class="text-gray-500 text-xs">threshold ±{bench.threshold.toFixed(1)}%</div>
        {summary.regressions > 0 && (
          <span class="px-2 py-0.5 bg-red-100 text-red-700 text-xs rounded font-semibold">
            {summary.regressions} regression{summary.regressions > 1 ? 's' : ''}
          </span>
        )}
        {summary.improvements > 0 && (
          <span class="px-2 py-0.5 bg-green-100 text-green-700 text-xs rounded font-semibold">
            {summary.improvements} improvement{summary.improvements > 1 ? 's' : ''}
          </span>
        )}
        {summary.neutral > 0 && (
          <span class="px-2 py-0.5 bg-gray-100 text-gray-600 text-xs rounded">
            {summary.neutral} unchanged
          </span>
        )}
        <div class={`ml-auto text-sm font-mono ${
          bench.geomean_delta > bench.threshold ? 'text-red-600 font-bold' :
          bench.geomean_delta < -bench.threshold ? 'text-green-600 font-bold' :
          'text-gray-500'
        }`}>
          geomean: {bench.geomean_delta >= 0 ? '+' : ''}{bench.geomean_delta.toFixed(2)}%
        </div>
      </div>

      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-gray-200 text-left text-xs text-gray-500 uppercase">
            <SortHeader onClick={() => onSort('name')} icon={sortIcon('name')}>Benchmark</SortHeader>
            <SortHeader onClick={() => onSort('base')} icon={sortIcon('base')} align="right">Base</SortHeader>
            <SortHeader onClick={() => onSort('head')} icon={sortIcon('head')} align="right">Head</SortHeader>
            <SortHeader onClick={() => onSort('delta')} icon={sortIcon('delta')}>Δ</SortHeader>
            <SortHeader onClick={() => onSort('p')} icon={sortIcon('p')} align="right">p-value</SortHeader>
          </tr>
        </thead>
        <tbody>
          {sorted.map(d => <BenchRow key={d.name} delta={d} threshold={bench.threshold} />)}
        </tbody>
      </table>
    </div>
  );
}

function SortHeader({ children, onClick, icon, align }: {
  children: any;
  onClick: () => void;
  icon: any;
  align?: 'right';
}) {
  return (
    <th
      class={`py-2 px-2 font-semibold cursor-pointer hover:bg-gray-50 select-none ${align === 'right' ? 'text-right' : ''}`}
      onClick={onClick}
    >
      <span class="inline-flex items-center gap-1">{children}{icon}</span>
    </th>
  );
}

function BenchRow({ delta, threshold }: { delta: BenchDelta; threshold: number }) {
  const cat = categorize(delta, threshold);
  const rowBg = cat === 'regression' ? 'bg-red-50' : cat === 'improvement' ? 'bg-green-50' : '';

  if (delta.only_in) {
    return (
      <tr class="border-b border-gray-100 text-gray-500">
        <td class="py-1.5 px-2 font-mono text-xs">{delta.name}</td>
        <td class="py-1.5 px-2 text-right text-xs italic" colSpan={4}>
          only in {delta.only_in}
        </td>
      </tr>
    );
  }

  const stddevBase = delta.base_stddev ? `±${delta.base_stddev.toFixed(1)}%` : '';
  const stddevHead = delta.head_stddev ? `±${delta.head_stddev.toFixed(1)}%` : '';

  return (
    <tr class={`border-b border-gray-100 hover:bg-gray-50 ${rowBg}`}>
      <td class="py-1.5 px-2 font-mono text-xs truncate max-w-[24rem]" title={delta.name}>
        {delta.name}
      </td>
      <td class="py-1.5 px-2 text-right tabular-nums text-xs">
        <div>{formatNs(delta.base_mean)}</div>
        {stddevBase && <div class="text-gray-400 text-[10px]">{stddevBase}</div>}
      </td>
      <td class="py-1.5 px-2 text-right tabular-nums text-xs">
        <div>{formatNs(delta.head_mean)}</div>
        {stddevHead && <div class="text-gray-400 text-[10px]">{stddevHead}</div>}
      </td>
      <td class="py-1.5 px-2 w-64">
        <BenchDeltaBar delta={delta} threshold={threshold} />
      </td>
      <td class="py-1.5 px-2 text-right tabular-nums text-xs text-gray-500">
        {delta.p_value !== undefined && delta.p_value > 0
          ? delta.p_value.toPrecision(2)
          : <span class="text-gray-300">—</span>}
        {delta.samples !== undefined && (
          <span class="ml-1 text-gray-400">n={delta.samples}</span>
        )}
      </td>
    </tr>
  );
}
