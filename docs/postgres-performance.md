# PostgreSQL performance diagnostics

Gavel's embedded PostgreSQL lifecycle preloads `pg_stat_statements`, enables
`track_io_timing`, and installs the extension in the `gavel` database. Startup
fails with an actionable error if a reused embedded instance was started
without those server settings; restart the Gavel service so its managed
PostgreSQL process is started with the current configuration.

An external DSN remains operator-managed. Set these values in the server's
`postgresql.conf` (or equivalent managed-service parameter group), restart the
server, and install the extension in the selected database:

```text
shared_preload_libraries = 'pg_stat_statements'
track_io_timing = on
```

```sql
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
```

Changing `shared_preload_libraries` requires a PostgreSQL restart. Gavel never
rewrites an external server's configuration or restarts it.

## Top statements

This query reports normalized statement text. Bind parameter values are not
stored by `pg_stat_statements`.

```sql
SELECT calls,
       round(total_exec_time::numeric, 1) AS total_exec_ms,
       round(mean_exec_time::numeric, 1) AS mean_exec_ms,
       rows,
       shared_blks_hit,
       shared_blks_read,
       temp_blks_written,
       round(blk_read_time::numeric, 1) AS block_read_ms,
       round(blk_write_time::numeric, 1) AS block_write_ms,
       query
FROM pg_stat_statements
WHERE dbid = (SELECT oid FROM pg_database WHERE datname = current_database())
ORDER BY total_exec_time DESC
LIMIT 25;
```

The view is cumulative. Record `pg_stat_statements_info.stats_reset` with any
measurement so rankings have a clear time window.

## Captain session storage

Captain exposes the same read-only measurement as
`database.SessionStorageStats`. The query combines exact live-row page
occupancy with PostgreSQL's update, dead-tuple, autovacuum, and autoanalyze
counters; it does not require `pgstattuple`.

Pages without a visible live tuple are a repeatable bloat signal, not a claim
that every byte is permanently unusable: PostgreSQL can reuse free space.
`VACUUM (ANALYZE) public.captain_sessions` safely refreshes statistics and makes
dead tuples reusable, and Captain's daily monitor already runs `VACUUM ANALYZE`.
Ordinary vacuum does not shrink the relation file.

`VACUUM FULL`, `CLUSTER`, `pg_repack`, and equivalent table rewrites are
operator-controlled maintenance. A rewrite needs additional disk space and can
hold a blocking table lock; schedule it only after measuring the remaining
bloat and correcting the write pattern that created it.
