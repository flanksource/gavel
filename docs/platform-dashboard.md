# Platform dashboard

Hosted `gavel serve` has a **Platform** tab that shows how every organization uses gavel. Only
platform admins see it. Each panel reads its own source, so when one source is down or not set up,
the panel says why and the other panels still load.

| Panel | Source |
|---|---|
| Tenants | gavel's organizations, with their members, repositories, GitHub installs and recent runs, merged with Clerk's organization list |
| Sandbox usage | gavel's sandbox runs: runs and sandbox minutes per day by kind, failures, active runs, and each organization's share. Each chart can be shown as a table |
| GitHub App | The App's installations on GitHub, the organization each is linked to, its repositories, core rate limit, and whether it is suspended or was uninstalled from GitHub |
| Substrate inventory | Substrate's workers with their capacity against what their actors are allocated, and every atespace's actors by the organization whose run each is |
| Substrate metrics | The panels of a Mission Control View graphing substrate's Prometheus metrics: timeseries as line charts, each with its numbers as a table, and number and gauge panels as figures |
| Model usage | Captain's model-call ledger: spend per day, calls and tokens, and each organization's and model's share |
| Budgets | Captain's budget rules, each group's completed spend in the rule's current window against its amount |
| Captain history | Captain's prompt runs per day by outcome, each organization's runs, failures, run time and live sessions (with how many are stalled or zombie), and the latest failures with their errors |
| Infrastructure | The config items, health checks and config changes Mission Control has for gavel's infrastructure |

Windows are limited to 60 days, because sandbox runs are kept for 60 days by default. Each window
starts at UTC midnight, so every day on a chart is a whole day.

Click an organization on the Tenants panel to open its own view at `/platform/<organization ID>`:
its members, repositories, GitHub installations and last run, its sandbox size (the CPU and memory
each of its substrate environments gets, which a platform admin sets there or returns to serve's
default, `--substrate-cpu` and `--substrate-memory`), its measured CPU and memory (below), then
the sandbox usage, model usage,
Captain history, substrate actors and GitHub installations panels with only its rows. Budgets,
substrate's workers, its metrics and gavel's infrastructure belong to no one organization, so they
show only on the dashboard of every organization.

## Platform admins

A platform admin is a Clerk user with this in their **public metadata** (Clerk dashboard → Users →
the user → Metadata):

```json
{ "gavel_role": "platform_admin" }
```

gavel reads it with the rest of the user's Clerk profile. It applies within 5 minutes and works for
both browser sessions and CLI tokens. Other users don't see the tab. If they open `/platform`
directly, the page says how to become a platform admin. The API (`/api/platform/…`) answers them
with `403 not_platform_admin`.

## Model usage and budgets

Both panels read Captain's tables in gavel's database, so they need no configuration.

- **Owner:** a model call belongs to the organization that owns its session, or its root session
  when the session has no owner. Calls of sessions without an owner, or owned by the local
  organization or another Clerk instance's, are counted as **Unattributed**.
- **Cost:** a call costs what its provider reported, else its list price (the same rule Captain's
  own totals use). A call priced in another currency adds its tokens but not its cost, since
  there's no rate to convert it, and the panel says how many calls that left out.
- **Budget spend:** only completed calls count, from where the rule's window starts until now.
  The window is the rule's datemath (`now/M`, `now-30d`), evaluated as Captain evaluates it. A
  group holding a call in another currency shows why it has no total, as Captain refuses to add
  it up too. Rules are created with `captain budget-rule create` or on `captain serve`'s Budgets
  page.

## Substrate inventory

With `--sandbox substrate`, the panel lists substrate's workers and the actors of every atespace
through ate-api-server, with serve's `--substrate-*` credentials. Listing either needs
substrate's global `viewer` role, so grant it to gavel's identity in substrate's global access
policy:

```yaml
bindings:
- role: viewer
  members:
  - user:system:serviceaccount:gavel:gavel-beta
```

Without it, the panel shows substrate's `PermissionDenied` for the list it couldn't read. An actor
named `gavel-run-<run ID>` is that sandbox run's environment, listed under the run's organization;
other actors are listed by atespace.

## Mission Control

The Infrastructure panel reads a Mission Control instance through its client SDK. Without one, the
panel says which settings are missing. To set one up:

1. In Mission Control, create an access token with the viewer role while signed in. One way is
   `POST /auth/create_token` with a body like this:

   ```json
   { "name": "gavel-platform", "expiry": "8760h", "auto_renew": true,
     "roles": [{ "namespace": "mission-control-builtin", "name": "viewer" }] }
   ```

2. Pick the Mission Control resource search that selects gavel's infrastructure, e.g.
   `namespace=gavel`. It selects config items and health checks; the changes shown are the newest
   200 of the config items it selects in the last 7 days, not of their children, since Mission
   Control can't search changes by fields like `namespace`. Each list is read on its own, so one
   that fails shows why and the others still load.

3. Start `gavel serve` with these settings:

   | Setting | Flag / env | Chart value |
   |---|---|---|
   | Backend URL, e.g. `https://mc.example.com/api` | `--mission-control-url` or `GAVEL_MISSION_CONTROL_URL` | `missionControl.url` |
   | The search | `--mission-control-infra` | `missionControl.infra` |
   | The token | `GAVEL_MISSION_CONTROL_TOKEN` (environment only) | `missionControl.token.secretName` / `.key` |
| Optional: substrate's metrics View, as `namespace/name` | `--mission-control-substrate-view` | `missionControl.substrateView` |
| Optional: the organizations' measured use View, as `namespace/name` | `--mission-control-org-view` | `missionControl.orgView` |

If Mission Control refuses the token, the panel shows the HTTP status and says to check
`GAVEL_MISSION_CONTROL_TOKEN`.

To check a real Mission Control end to end:

```sh
GAVEL_MISSION_CONTROL_TEST=1 GAVEL_MISSION_CONTROL_URL=… GAVEL_MISSION_CONTROL_TOKEN=… \
  GAVEL_MISSION_CONTROL_INFRA=namespace=gavel GAVEL_MISSION_CONTROL_SUBSTRATE_VIEW=substrate/capacity \
  go test ./missioncontrol/ -run RealMissionControl
```

## Substrate metrics

Substrate runs in its own cluster, so gavel reads its metrics through Mission Control, from a View
whose queries reach substrate's Prometheus. Create a Prometheus connection to it in Mission Control
(below, `connection://substrate/prometheus`), apply this View, and pass it to
`--mission-control-substrate-view substrate/capacity`.

The Substrate metrics panel draws `timeseries` panels as line charts (one line per label set),
and shows `number` and `gauge` panels as figures. It lists other panel types by name without
drawing them. A panel's query names the `timestamp` and `value` columns, or the panel's
`timeseries.timeKey`/`valueKey` do; every other column is a label.

The View answers four questions: is there room (workers by state, unschedulable workers), what is
running (actors, their memory and CPU, by template), how fast do sandboxes start and stop
(scheduling, actor operations, snapshot restore and checkpoint), and is anything churning (actor
state changes, image cache hits, cache sizes).

- **Replicas:** every `ate-api-server` replica exports the same worker counts, so worker queries
  take the `max` per pool and state before adding pools up.
- **Labels:** queries rename substrate's labels to short ones (`template`, `state`) and keep
  only those with `selectLabels`, so each query's rows have exactly its columns.
- **No data:** a percentile or ratio over a window with no events is `NaN`, which Mission Control
  can't send as JSON; each such query ends in `>= 0`, which drops those samples, so the panel
  is empty rather than broken.

```yaml
apiVersion: mission-control.flanksource.com/v1
kind: View
metadata:
  name: capacity
  namespace: substrate
spec:
  display:
    title: Substrate
  panels:
    # Room
    - name: Workers
      type: number
      query: SELECT value FROM workers_now
    - name: Unschedulable workers
      type: number
      query: SELECT value FROM unschedulable_workers
    - name: Workers by state
      description: Idle, partly used, full or unschedulable, last 12h
      type: timeseries
      query: SELECT timestamp, state, value FROM workers_by_state
    # Running
    - name: Running actors
      type: number
      query: SELECT value FROM actors_now
    - name: Running actors by template
      description: Last 12h
      type: timeseries
      query: SELECT timestamp, template, value FROM running_actors
    - name: Actor memory working set (GiB)
      description: Per template, last 12h
      type: timeseries
      query: SELECT timestamp, template, value FROM working_set
    - name: Actor CPU (cores)
      description: Per template, last 12h
      type: timeseries
      query: SELECT timestamp, template, value FROM cpu
    # Speed
    - name: Scheduling p95 (s)
      description: From asking for an actor to a worker taking it, last 12h
      type: timeseries
      query: SELECT timestamp, value FROM scheduling_p95
    - name: Actor operations p95 (s)
      description: Create, resume, suspend and delete, last 12h
      type: timeseries
      query: SELECT timestamp, operation, value FROM operation_p95
    - name: Snapshot restore p95 (s)
      description: Per template, last 12h
      type: timeseries
      query: SELECT timestamp, template, value FROM restore_p95
    - name: Snapshot checkpoint p95 (s)
      description: Per template, last 12h
      type: timeseries
      query: SELECT timestamp, template, value FROM checkpoint_p95
    - name: Failed scheduling, last 24h
      type: number
      query: SELECT value FROM unassigned
    # Churn
    - name: Actor state changes per hour
      description: By the state entered, last 12h
      type: timeseries
      query: SELECT timestamp, state, value FROM state_changes
    - name: Image cache hit rate, last 24h (%)
      type: number
      query: SELECT value FROM image_cache_hits
    - name: Cache size (GiB)
      description: Snapshot and sandbox asset caches, last 12h
      type: timeseries
      query: SELECT timestamp, cache, value FROM file_cache
  queries:
    workers_now:
      columns: { value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: sum(max by (ate_workerpool_namespace, ate_workerpool_name, ate_sandbox_class, ate_worker_state) (ate_workerpool_workers))
    unschedulable_workers:
      columns: { value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: sum(max by (ate_workerpool_namespace, ate_workerpool_name, ate_sandbox_class) (ate_workerpool_workers{ate_worker_state="unschedulable"})) or vector(0)
    workers_by_state:
      columns: { timestamp: datetime, state: string, value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: label_replace(sum by (ate_worker_state) (max by (ate_workerpool_namespace, ate_workerpool_name, ate_sandbox_class, ate_worker_state) (ate_workerpool_workers)), "state", "$1", "ate_worker_state", "(.*)")
        selectLabels: [state]
        range: { start: now-12h, end: now, step: 5m }
    actors_now:
      columns: { value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: sum(ate_actor_stats_sampled_actors) or vector(0)
    running_actors:
      columns: { timestamp: datetime, template: string, value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: label_replace(sum by (ate_template_name) (ate_actor_stats_sampled_actors), "template", "$1", "ate_template_name", "(.*)")
        selectLabels: [template]
        range: { start: now-12h, end: now, step: 5m }
    working_set:
      columns: { timestamp: datetime, template: string, value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: label_replace(sum by (ate_template_name) (ate_actor_stats_memory_working_set_bytes) / 2^30, "template", "$1", "ate_template_name", "(.*)")
        selectLabels: [template]
        range: { start: now-12h, end: now, step: 5m }
    cpu:
      columns: { timestamp: datetime, template: string, value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: label_replace(sum by (ate_template_name) (rate(ate_actor_stats_cpu_time_seconds_total[5m])), "template", "$1", "ate_template_name", "(.*)")
        selectLabels: [template]
        range: { start: now-12h, end: now, step: 5m }
    scheduling_p95:
      columns: { timestamp: datetime, value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: histogram_quantile(0.95, sum by (le) (rate(ate_scheduler_assignment_duration_seconds_bucket[30m]))) >= 0
        range: { start: now-12h, end: now, step: 5m }
    operation_p95:
      columns: { timestamp: datetime, operation: string, value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: label_replace(histogram_quantile(0.95, sum by (le, ate_actor_operation_name) (rate(ate_actor_lifecycle_operation_duration_seconds_bucket[30m]))) >= 0, "operation", "$1", "ate_actor_operation_name", "(.*)")
        selectLabels: [operation]
        range: { start: now-12h, end: now, step: 5m }
    restore_p95:
      columns: { timestamp: datetime, template: string, value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: label_replace(histogram_quantile(0.95, sum by (le, ate_template_name) (rate(ate_actor_restore_duration_seconds_bucket{ate_snapshot_phase="total"}[30m]))) >= 0, "template", "$1", "ate_template_name", "(.*)")
        selectLabels: [template]
        range: { start: now-12h, end: now, step: 5m }
    checkpoint_p95:
      columns: { timestamp: datetime, template: string, value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: label_replace(histogram_quantile(0.95, sum by (le, ate_template_name) (rate(ate_actor_checkpoint_duration_seconds_bucket{ate_snapshot_phase="total"}[30m]))) >= 0, "template", "$1", "ate_template_name", "(.*)")
        selectLabels: [template]
        range: { start: now-12h, end: now, step: 5m }
    unassigned:
      columns: { value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: sum(increase(ate_scheduler_assignment_duration_seconds_count{ate_scheduler_outcome!="assigned"}[24h])) or vector(0)
    state_changes:
      columns: { timestamp: datetime, state: string, value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: label_replace(sum by (ate_actor_state) (increase(substrate_actor_state_changes_total[1h])), "state", "$1", "ate_actor_state", "(.*)")
        selectLabels: [state]
        range: { start: now-12h, end: now, step: 5m }
    image_cache_hits:
      columns: { value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: 100 * sum(increase(ate_imagecache_requests_total{ate_imagecache_outcome="hit"}[24h])) / sum(increase(ate_imagecache_requests_total[24h])) >= 0
    file_cache:
      columns: { timestamp: datetime, cache: string, value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: label_replace(sum by (ate_filecache_name) (ate_filecache_size_bytes) / 2^30, "cache", "$1", "ate_filecache_name", "(.*)")
        selectLabels: [cache]
        range: { start: now-12h, end: now, step: 5m }
```

The metric names below are the ones substrate's Prometheus exports, checked against a running
substrate:

| Metric | What it measures | Labels |
|---|---|---|
| `ate_workerpool_workers` | Workers by pool, state (`idle`, `partial`, `at_capacity`, `unschedulable`) and sandbox class, from each `ate-api-server` replica | `ate_workerpool_namespace`, `ate_workerpool_name`, `ate_worker_state`, `ate_sandbox_class` |
| `ate_actor_stats_sampled_actors` | Running actors with a resource sample | `ate_template_name`, `ate_template_atespace`, `ate_sandbox_class`, `ate_workerpool_name` |
| `ate_actor_stats_memory_working_set_bytes`, `ate_actor_stats_memory_usage_bytes` | Running actors' memory | as above |
| `ate_actor_stats_cpu_time_seconds_total` | Running actors' CPU time | as above |
| `ate_scheduler_assignment_duration_seconds` | Time to place an actor on a worker (histogram) | `ate_scheduler_outcome`, `ate_workerpool_name`, `ate_sandbox_class` |
| `ate_actor_lifecycle_operation_duration_seconds` | Create, resume, suspend and delete times (histogram) | `ate_actor_operation_name`, `ate_template_name`, `ate_snapshot_kind` |
| `ate_actor_restore_duration_seconds`, `ate_actor_checkpoint_duration_seconds` | Snapshot restore and checkpoint times by phase; `total` is the whole (histogram) | `ate_snapshot_phase`, `ate_template_name`, `ate_snapshot_kind` |
| `substrate_actor_state_changes_total` | Actors entering a state | `ate_actor_state` |
| `ate_imagecache_requests_total` | Image cache lookups | `ate_imagecache_outcome` (`hit`, `miss`) |
| `ate_filecache_size_bytes` | Snapshot and sandbox asset cache sizes | `ate_filecache_name` |

Substrate exports no crash metric; a crashed actor shows only in the Substrate inventory panel's
actor states.

If Mission Control can't refresh the View, for example because Prometheus is unreachable, the
panel says why, and shows whatever data Mission Control still has cached for it.

## Measured CPU and memory

An organization's environments run from its own templates, `gavel-org-<organization ID>-<hash>`,
so substrate's metrics labelled with those templates are what the organization used. gavel reads
them from a second View, `--mission-control-org-view substrate/org-usage`, passing the
organization's ID as its variable `org`, which its queries read as `$(.var.org)`. Mission Control
caches the View once per organization. Platform admins see any organization's on its platform view
(`/api/platform/orgs/{id}/measured-usage`); an organization's admins see their own on its Usage and
budgets page (`/api/usage/measured`); its other members don't see it.

The figures are sampled every scrape (15 seconds on substrate's Prometheus), so a short burst of
CPU may not show, and they reset when that Prometheus restarts. Budgets and usage count each run's
size for as long as it runs instead.

- **CPU:** substrate keeps one CPU counter per template and worker node, which stays a while after
  its environments end, and whose first sample already holds a short run's CPU. `increase()`
  misses that first sample, so the window's CPU is the counter's latest value less its value when
  the window began, or less nothing when the counter began within it.
- **Peak memory** samples at the scrape interval (`[24h:15s]`), so an environment that lives a
  minute still counts.

```yaml
apiVersion: mission-control.flanksource.com/v1
kind: View
metadata:
  name: org-usage
  namespace: substrate
spec:
  display:
    title: Measured use
  # The organization's ID, which gavel passes.
  templating:
    - key: org
      label: Organization ID
      # The View CRD requires values or valueFrom. Mission Control uses a passed value only when
      # the variable has no options, or the value is one of them, else the first option, so an
      # empty list keeps any organization's ID.
      values: []
  panels:
    - name: Running environments
      type: number
      query: SELECT value FROM environments_now
    - name: CPU, last 24h (core-hours)
      type: number
      query: SELECT value FROM cpu_core_hours
    - name: Peak memory, last 24h (GiB)
      type: number
      query: SELECT value FROM peak_memory
    - name: CPU (cores)
      description: Last 12h
      type: timeseries
      query: SELECT timestamp, value FROM cpu
    - name: Memory working set (GiB)
      description: Last 12h, per minute
      type: timeseries
      query: SELECT timestamp, value FROM working_set
  queries:
    environments_now:
      columns: { value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: sum(ate_actor_stats_sampled_actors{ate_template_name=~"gavel-org-$(.var.org)-.*"}) or vector(0)
    cpu_core_hours:
      columns: { value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        # One counter per template and worker node, kept a while after its environments end: its
        # first sample already holds a short run's CPU, which increase() misses. CPU in the window
        # is its latest value less its value when the window began, or less 0 when it began since.
        query: (sum(max_over_time(ate_actor_stats_cpu_time_seconds_total{ate_template_name=~"gavel-org-$(.var.org)-.*"}[24h]) - (last_over_time(ate_actor_stats_cpu_time_seconds_total{ate_template_name=~"gavel-org-$(.var.org)-.*"}[5m] offset 24h) or max_over_time(ate_actor_stats_cpu_time_seconds_total{ate_template_name=~"gavel-org-$(.var.org)-.*"}[24h]) * 0)) / 3600) or vector(0)
    # Sampled at the scrape interval, so an environment of a minute still counts.
    peak_memory:
      columns: { value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: (max_over_time(sum(ate_actor_stats_memory_working_set_bytes{ate_template_name=~"gavel-org-$(.var.org)-.*"})[24h:15s]) / 2^30) or vector(0)
    cpu:
      columns: { timestamp: datetime, value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: sum(rate(ate_actor_stats_cpu_time_seconds_total{ate_template_name=~"gavel-org-$(.var.org)-.*"}[5m])) or vector(0)
        range: { start: now-12h, end: now, step: 5m }
    working_set:
      columns: { timestamp: datetime, value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: (sum(ate_actor_stats_memory_working_set_bytes{ate_template_name=~"gavel-org-$(.var.org)-.*"}) / 2^30) or vector(0)
        range: { start: now-12h, end: now, step: 1m }
```
