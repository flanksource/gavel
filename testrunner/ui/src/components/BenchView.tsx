import { useMemo } from 'react';
import { Badge, DataTable, type DataTableColumn } from '@flanksource/clicky-ui/data';
import { cn } from '@flanksource/clicky-ui/utils';
import type { BenchComparison, BenchDelta } from '../types';

type Category = 'regression' | 'improvement' | 'neutral' | 'only';

function formatNs(ns: number): string {
  if (ns >= 1e9) return `${(ns / 1e9).toFixed(2)}s`;
  if (ns >= 1e6) return `${(ns / 1e6).toFixed(2)}ms`;
  if (ns >= 1e3) return `${(ns / 1e3).toFixed(2)}µs`;
  return `${ns.toFixed(2)}ns`;
}

function categorize(d: BenchDelta, threshold: number): Category {
  if (d.only_in) return 'only';
  if (d.significant && d.delta_pct > threshold) return 'regression';
  if (d.significant && d.delta_pct < -threshold) return 'improvement';
  return 'neutral';
}

// BenchRow is a DataTable row: the comparison's delta with the figures a column reads.
type BenchRow = {
  name: string;
  base: number | null;
  head: number | null;
  baseSpread: string;
  headSpread: string;
  delta: number | null;
  pValue: number | null;
  samples: number | null;
  category: Category;
  onlyIn: string | null;
  threshold: number;
  significant: boolean;
};

function toRow(delta: BenchDelta, threshold: number): BenchRow {
  const only = !!delta.only_in;
  return {
    name: delta.name,
    base: only ? null : delta.base_mean,
    head: only ? null : delta.head_mean,
    baseSpread: delta.base_stddev ? `±${delta.base_stddev.toFixed(1)}%` : '',
    headSpread: delta.head_stddev ? `±${delta.head_stddev.toFixed(1)}%` : '',
    delta: only ? null : delta.delta_pct,
    pValue: delta.p_value !== undefined && delta.p_value > 0 ? delta.p_value : null,
    samples: delta.samples ?? null,
    category: categorize(delta, threshold),
    onlyIn: delta.only_in ?? null,
    threshold,
    significant: !!delta.significant,
  };
}

const CATEGORY_TEXT: Record<Category, string> = {
  regression: 'text-red-600 dark:text-red-400',
  improvement: 'text-green-600 dark:text-green-400',
  neutral: 'text-muted-foreground',
  only: 'text-muted-foreground',
};

const CATEGORY_BAR: Record<Category, string> = {
  regression: 'bg-red-500',
  improvement: 'bg-green-500',
  neutral: 'bg-muted-foreground/40',
  only: 'bg-muted-foreground/40',
};

// DeltaBar draws the signed delta: right of centre is a regression, left an
// improvement; its length is |delta| clamped at 50%.
function DeltaBar({ row }: { row: BenchRow }) {
  if (row.delta === null) return <span className="text-xs italic text-muted-foreground">only in {row.onlyIn}</span>;
  const clamped = Math.max(-50, Math.min(50, row.delta));
  const width = Math.abs(clamped);
  return (
    <div className="flex w-full min-w-48 items-center gap-2">
      <div className="relative h-4 min-w-20 flex-1 overflow-hidden rounded bg-muted">
        <div className="absolute inset-y-0 left-1/2 w-px bg-muted-foreground/60" />
        <div
          className={cn('absolute inset-y-0', CATEGORY_BAR[row.category])}
          style={row.delta >= 0 ? { left: '50%', width: `${width}%` } : { right: '50%', width: `${width}%` }}
        />
      </div>
      <span className={cn('w-16 text-right text-xs tabular-nums', CATEGORY_TEXT[row.category], row.significant && 'font-semibold')}>
        {row.delta >= 0 ? '+' : ''}{row.delta.toFixed(2)}%
      </span>
    </div>
  );
}

const COLUMNS: Array<DataTableColumn<BenchRow>> = [
  { key: 'name', label: 'Benchmark', sortable: true, grow: true, cellClassName: 'font-mono text-xs' },
  {
    key: 'base', label: 'Base', sortable: true, align: 'right',
    render: (_value, row) => row.base === null ? '' : (
      <div className="text-xs tabular-nums">{formatNs(row.base)}{row.baseSpread && <div className="text-[10px] text-muted-foreground">{row.baseSpread}</div>}</div>
    ),
  },
  {
    key: 'head', label: 'Head', sortable: true, align: 'right',
    render: (_value, row) => row.head === null ? '' : (
      <div className="text-xs tabular-nums">{formatNs(row.head)}{row.headSpread && <div className="text-[10px] text-muted-foreground">{row.headSpread}</div>}</div>
    ),
  },
  { key: 'delta', label: 'Δ', sortable: true, render: (_value, row) => <DeltaBar row={row} /> },
  {
    key: 'pValue', label: 'p-value', sortable: true, align: 'right',
    render: (_value, row) => (
      <span className="text-xs tabular-nums text-muted-foreground">
        {row.pValue !== null ? row.pValue.toPrecision(2) : '—'}
        {row.samples !== null && <span className="ml-1">n={row.samples}</span>}
      </span>
    ),
  },
];

/** A benchmark comparison (`gavel bench compare --ui`): per-benchmark base and head means, delta bar and significance. */
export function BenchView({ bench }: { bench: BenchComparison | undefined }) {
  const rows = useMemo(() => (bench?.deltas ?? []).map(delta => toRow(delta, bench!.threshold)), [bench]);
  const summary = useMemo(() => {
    const counts: Record<Category, number> = { regression: 0, improvement: 0, neutral: 0, only: 0 };
    rows.forEach(row => { counts[row.category] += 1; });
    return counts;
  }, [rows]);

  if (!bench) {
    return (
      <div className="p-8 text-center text-sm text-muted-foreground">
        No benchmark comparison loaded. Run <code className="rounded bg-muted px-1">gavel bench compare --ui</code>.
      </div>
    );
  }

  return (
    <div className="flex h-full min-h-0 flex-col gap-density-2 p-density-4">
      <div className="flex flex-wrap items-center gap-4 text-sm">
        <div>
          <span className="font-mono">{bench.base_label || 'base'}</span>
          <span className="mx-2 text-muted-foreground">→</span>
          <span className="font-mono">{bench.head_label || 'head'}</span>
        </div>
        <div className="text-xs text-muted-foreground">threshold ±{bench.threshold.toFixed(1)}%</div>
        {summary.regression > 0 && <Badge tone="danger" variant="soft" size="xs">{summary.regression} regression{summary.regression > 1 ? 's' : ''}</Badge>}
        {summary.improvement > 0 && <Badge tone="success" variant="soft" size="xs">{summary.improvement} improvement{summary.improvement > 1 ? 's' : ''}</Badge>}
        {summary.neutral > 0 && <Badge tone="neutral" variant="soft" size="xs">{summary.neutral} unchanged</Badge>}
        <div
          className={cn(
            'ml-auto font-mono text-sm',
            bench.geomean_delta > bench.threshold ? 'font-bold text-red-600 dark:text-red-400'
              : bench.geomean_delta < -bench.threshold ? 'font-bold text-green-600 dark:text-green-400'
                : 'text-muted-foreground',
          )}
        >
          geomean: {bench.geomean_delta >= 0 ? '+' : ''}{bench.geomean_delta.toFixed(2)}%
        </div>
      </div>
      <DataTable<BenchRow>
        data={rows}
        columns={COLUMNS}
        defaultSort={{ key: 'delta', dir: 'desc' }}
        emptyMessage="No benchmarks were compared"
        className="min-h-0 flex-1"
      />
    </div>
  );
}
