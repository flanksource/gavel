# Progress from database routines

These native `kind: query` fixtures report progress from inside a PostgreSQL function and a SQL Server stored procedure. Each routine processes five batches with a one-second delay, updates progress from 0/5 to 5/5, sets descriptions and status messages, and returns a count checked by CEL.

Run from the repository root with a Gavel binary built from this checkout and the accompanying commons-db notice support:

```bash
./.bin/gavel fixtures outline examples/postgres-progress.fixture.yaml
./.bin/gavel fixtures examples/postgres-progress.fixture.yaml --ui
./.bin/gavel fixtures outline examples/sqlserver-progress.fixture.yaml
./.bin/gavel fixtures examples/sqlserver-progress.fixture.yaml --ui
```

Keep each fixture next to its matching `.sql` file when copying it. Omit `--ui` for terminal output. While the query runs, its progress count, description, status, and batch-completion output should update before the final response contains `completed: 5`. Workflow command lines are consumed as updates and are absent from the retained stdout; ordinary completion messages remain.

Each fixture opens the existing connection picker, filtered to PostgreSQL or SQL Server respectively. Select a saved connection from the catalog. No databases, containers or credentials are provisioned, and no `psql` or `sqlcmd` client is required. The selected account needs permission to create temporary routines. The routines only sleep and emit messages; they do not access application tables.

For noninteractive execution, supply a saved connection name:

```bash
./.bin/gavel fixtures examples/postgres-progress.fixture.yaml --param connection=example-postgres
./.bin/gavel fixtures examples/sqlserver-progress.fixture.yaml --param connection=example-sqlserver
```

Replace these names with entries in your catalog. PostgreSQL creates its function in `pg_temp`; SQL Server uses a local temporary procedure (`#report_progress`). Both disappear when their query session closes, including after a failure. Cancelling the connection form skips the query.

PostgreSQL emits raw `RAISE NOTICE` messages with `client_min_messages = NOTICE`. SQL Server uses severity-zero `RAISERROR ... WITH NOWAIT` to flush each message immediately; plain `PRINT` may buffer output. SQL Server functions cannot use `RAISERROR`, so that example uses a procedure.
