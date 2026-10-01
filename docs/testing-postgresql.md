# PostgreSQL integration tests

Gavel integration tests use `github.com/flanksource/commons-db/dbtest`. With `COMMONS_DB_URL` unset, dbtest starts or reuses a non-durable local PostgreSQL test server on port 7432 under `os.TempDir()/commons-db`. Its process and extracted binaries survive individual tests and subsequent test runs. The normal Go and operating-system temporary-directory configuration determines this location.

Each test leases an empty, isolated database. `Options.Name` labels the lease; omit `Options.DataDir` to share the default server. Open Gavel and Captain through their existing migration entry points using the leased `DSN()`. Migration tests still apply migrations themselves rather than cloning an already migrated schema.

Use `dbtest.ForT(t, dbtest.Options{Name: "gavel_example"})` for standard Go tests, or `dbtest.ForGinkgo(dbtest.Options{Name: "gavel_example"})` in a Ginkgo spec or setup node. Acquire the lease before opening consumer pools and register pool cleanup afterward, so pools close before the lease is released. A Ginkgo lease acquired in `BeforeAll` lasts for its ordered container. Both adapters release their test database automatically; they leave the shared server running. `dbtest.Open` returns a cleanup function for explicitly owned leases.

Run the database integration packages from the Gavel root:

```bash
env -u COMMONS_DB_URL -u COMMONS_DB_CREATE go test -count=1 ./internal/database ./internal/database/todoprojection ./internal/taskhistory ./todos/native ./todos/runtime ./todos/portable
```

Run only the shared-server regressions:

```bash
env -u COMMONS_DB_URL -u COMMONS_DB_CREATE go test ./todos/native -run '^TestNativePlanCreation$' -ginkgo.focus 'shared PostgreSQL test server' -count=1 -v
```

For an external disposable test server, set `COMMONS_DB_URL` to its PostgreSQL connection URL and leave `COMMONS_DB_CREATE` unset or set it to `true`. The account must be able to create and drop test databases. Never use `COMMONS_DB_CREATE=false` for this suite: its migration tests intentionally alter schemas, and that setting uses the supplied database directly. The server must contain only disposable test data.

Application-owned embedded PostgreSQL startup is separate from these leases. Test cleanup releases only its database; it must never stop the shared server or change its configuration.
