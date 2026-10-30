import { useMemo } from 'react';
import { Badge, DataTable, type DataTableColumn } from '@flanksource/clicky-ui/data';
import { formatDuration } from '@flanksource/clicky-ui/utils';
import type { RunHistory as RunHistoryState } from '../hooks/use-run-history';
import { runDurationMs, runViewHref, type RunInfo } from '../runHistory';

type Row = RunInfo & Record<string, unknown>;

function ResultBadges({ run }: { run: RunInfo }) {
  return (
    <span className="inline-flex flex-wrap items-center gap-1">
      {run.total > 0 && <Badge tone={run.failed > 0 ? 'danger' : 'success'} variant="outline" size="xs">{run.failed > 0 ? `${run.failed} failed` : `${run.passed} passed`}</Badge>}
      {run.total > 0 && <span className="text-xs text-muted-foreground">of {run.total}</span>}
      {run.kind !== 'test' && <Badge tone={run.lintViolations > 0 ? 'warning' : 'neutral'} variant="outline" size="xs">{run.lintViolations} lint</Badge>}
    </span>
  );
}

/** Every recorded run, newest first; picking one opens it. */
export function RunHistory({ history, navigate }: { history: RunHistoryState; navigate: (href: string) => void }) {
  const columns = useMemo<DataTableColumn<Row>[]>(() => [
    {
      key: 'started',
      label: 'Started',
      kind: 'timestamp',
      sortable: true,
      render: (_value, run) => (
        <span className="inline-flex items-center gap-1.5">
          <a href={runViewHref(run)} className="text-primary no-underline hover:underline" onClick={event => event.stopPropagation()}>
            {run.started ? new Date(run.started).toLocaleString() : run.runId}
          </a>
          {history.current === run.runId && <Badge tone="info" variant="soft" size="xs">viewing</Badge>}
          {!history.current && history.latest.includes(run.runId) && <Badge tone="info" variant="soft" size="xs">latest</Badge>}
        </span>
      ),
    },
    { key: 'kind', label: 'Kind', filterable: true, shrink: true },
    { key: 'results', label: 'Results', accessor: run => run.failed + run.lintViolations, sortable: true, render: (_value, run) => <ResultBadges run={run} /> },
    { key: 'duration', label: 'Duration', accessor: run => runDurationMs(run) ?? null, sortable: true, align: 'right', shrink: true, render: value => (typeof value === 'number' ? formatDuration(value) : '') },
    { key: 'frameworks', label: 'Frameworks', kind: 'tags', accessor: run => run.frameworks ?? [] },
    { key: 'sha', label: 'Commit', shrink: true, render: value => <code className="text-xs">{String(value ?? '').slice(0, 8)}</code> },
  ], [history.current, history.latest]);

  return (
    <DataTable<Row>
      data={history.runs as Row[]}
      columns={columns}
      getRowId={run => run.runId}
      onRowClick={run => navigate(runViewHref(run))}
      getRowClassName={run => (history.current === run.runId ? 'bg-primary/5' : undefined)}
      autoFilter
      infinite={{ hasMore: history.hasMore, loading: history.loading, onLoadMore: history.loadMore }}
      loading={history.loading && history.runs.length === 0}
      loadingMessage="Loading run history"
      error={history.error}
      emptyMessage="No runs recorded in .gavel yet"
    />
  );
}
