import { Field } from '../components/Field';
import { ProfileTiles, SqlStatements } from '../components/ResourceParts';
import { MetricTile } from '@flanksource/clicky-ui/data';
import type { Test } from '../types';

/**
 * What a fixture step's resource profiling measured: the process tree's peak
 * CPU/RSS/memory and observed disk IO, and the SQL it ran. Go profiles are
 * `goprof` windows on the Traces tab.
 */
export function ResourcesPanel({ node }: { node: Test }) {
  const sql = node.sql_profile;
  return (
    <div className="space-y-4 p-density-4">
      {node.fixture_profile && <Field title="Resources"><ProfileTiles profile={node.fixture_profile} /></Field>}
      {sql && (
        <Field title="SQL profile">
          <div className="grid grid-cols-3 gap-density-1">
            <MetricTile label="Queries" value={String(sql.query_count)} />
            <MetricTile label="Total SQL" value={`${sql.total_duration_ms.toFixed(1)} ms`} />
            <MetricTile label="Slow queries" value={String(sql.slow_query_count)} />
          </div>
          <SqlStatements profile={sql} />
        </Field>
      )}
    </div>
  );
}
