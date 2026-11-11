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
its members, repositories, GitHub installations and last run, then the sandbox usage, model usage,
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
   `namespace=gavel`. It selects config items, health checks and config changes.

3. Start `gavel serve` with these settings:

   | Setting | Flag / env | Chart value |
   |---|---|---|
   | Backend URL, e.g. `https://mc.example.com/api` | `--mission-control-url` or `GAVEL_MISSION_CONTROL_URL` | `missionControl.url` |
   | The search | `--mission-control-infra` | `missionControl.infra` |
   | The token | `GAVEL_MISSION_CONTROL_TOKEN` (environment only) | `missionControl.token.secretName` / `.key` |
| Optional: substrate's metrics View, as `namespace/name` | `--mission-control-substrate-view` | `missionControl.substrateView` |

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
(below, `connection://substrate/prometheus`), apply a View like this one, and pass it to
`--mission-control-substrate-view substrate/capacity`.

The Substrate metrics panel draws `timeseries` panels as line charts (one line per label set),
and shows `number` and `gauge` panels as figures. It lists other panel types by name without
drawing them. A panel's query names the `timestamp` and `value` columns, or the panel's
`timeseries.timeKey`/`valueKey` do; every other column is a label.

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
    - name: Ready workers
      description: Per worker pool, last 12h
      type: timeseries
      query: SELECT timestamp, pool, value FROM ready_workers
    - name: Running actors
      description: Per template, last 12h
      type: timeseries
      query: SELECT timestamp, template, value FROM running_actors
    - name: Actor memory working set (GiB)
      description: Per template, last 12h
      type: timeseries
      query: SELECT timestamp, template, value FROM working_set
    - name: Actor crashes, last 24h
      type: number
      query: SELECT value FROM crashes
  queries:
    ready_workers:
      columns: { timestamp: datetime, pool: string, value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: sum by (pool) (label_replace(ate_workerpool_ready_workers, "pool", "$1", "ate_workerpool_name", "(.*)"))
        range: { start: now-12h, end: now, step: 5m }
    running_actors:
      columns: { timestamp: datetime, template: string, value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: sum by (template) (label_replace(ate_actor_stats_sampled_actors, "template", "$1", "ate_template_name", "(.*)"))
        range: { start: now-12h, end: now, step: 5m }
    working_set:
      columns: { timestamp: datetime, template: string, value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: sum by (template) (label_replace(ate_actor_stats_memory_working_set_bytes, "template", "$1", "ate_template_name", "(.*)")) / 2^30
        range: { start: now-12h, end: now, step: 5m }
    crashes:
      columns: { value: decimal }
      prometheus:
        connection: connection://substrate/prometheus
        query: sum(increase(ate_actor_crashes_total[24h]))
```

The metric names are substrate's OpenTelemetry instruments as its Prometheus exporter renders
them, from substrate's source; check them against your Prometheus before relying on a panel:

| Metric | What it measures | Labels |
|---|---|---|
| `ate_workerpool_desired_workers`, `ate_workerpool_ready_workers` | Worker pods requested and ready, per WorkerPool | `ate_workerpool_namespace`, `ate_workerpool_name` |
| `ate_workerpool_workers` | Workers by pool, state and sandbox class | `ate_workerpool_name`, `ate_worker_state`, `ate_sandbox_class` |
| `ate_actor_crashes_total` | Actors that crashed | the failure reason |
| `ate_actor_stats_sampled_actors` | Running actors with a resource sample | `ate_template_name`, `ate_sandbox_class`, `ate_workerpool_name` |
| `ate_actor_stats_memory_working_set_bytes`, `ate_actor_stats_memory_usage_bytes` | Running actors' memory | as above |
| `ate_actor_stats_cpu_time_seconds_total` | Running actors' CPU time | as above |

If Mission Control can't refresh the View, for example because Prometheus is unreachable, the
panel says why, and shows whatever data Mission Control still has cached for it.
