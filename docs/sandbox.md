# Sandboxed runs on a container host

Hosted `gavel serve` runs tests, lint and push checks in sandbox jobs, and agent runs on
agent-substrate (see [agent-substrate](#agent-substrate)). On a single host, those jobs are docker
(or rootful podman) containers. Each job runs:

- read-only, as uid 1000, with every capability dropped and `no-new-privileges`;
- with a pids limit, its CPU and memory limits, and tmpfs for `/tmp`, `/workspace` and
  `/home/gavel`;
- on the `gavel-sandbox` bridge network.

It gets its run token and model keys through an env file that only the owner can read, and that
file is removed once the job starts. It never receives database or GitHub credentials.

The container isolation doesn't stop a job from reaching your private network: the database, the
cloud metadata endpoint, other services. The host firewall does that. `gavel serve` checks the
firewall before it runs any job, and refuses to start if the rules are missing. It prints the
commands to run, but never runs them itself, because they need root.

## Setup

1. Create the network:

   ```sh
   docker network create gavel-sandbox
   ```

2. Find its bridge interface. It's `br-` plus the first 12 characters of the network ID, unless
   the network names its own bridge:

   ```sh
   docker network inspect gavel-sandbox --format '{{.Id}} {{index .Options "com.docker.network.bridge.name"}}'
   ```

3. As root, let jobs reach `gavel serve`, and reject every other private destination. The ACCEPT
   has to come before the REJECT rules. Replace `BRIDGE`, `SERVE_IP` and `SERVE_PORT`:

   ```sh
   iptables -I DOCKER-USER -i BRIDGE -d SERVE_IP -p tcp --dport SERVE_PORT -j ACCEPT -m comment --comment gavel-sandbox
   iptables -A DOCKER-USER -i BRIDGE -d 10.0.0.0/8 -j REJECT -m comment --comment gavel-sandbox
   iptables -A DOCKER-USER -i BRIDGE -d 172.16.0.0/12 -j REJECT -m comment --comment gavel-sandbox
   iptables -A DOCKER-USER -i BRIDGE -d 192.168.0.0/16 -j REJECT -m comment --comment gavel-sandbox
   iptables -A DOCKER-USER -i BRIDGE -d 100.64.0.0/10 -j REJECT -m comment --comment gavel-sandbox
   iptables -A DOCKER-USER -i BRIDGE -d 169.254.0.0/16 -j REJECT -m comment --comment gavel-sandbox
   iptables -A DOCKER-USER -i BRIDGE -d 127.0.0.0/8 -j REJECT -m comment --comment gavel-sandbox
   ```

4. As root, keep jobs off this host's own services too. DOCKER-USER only sees traffic Docker
   forwards; a job connecting to the host itself through the bridge (to a database listening on
   it, say) is delivered through INPUT. Insert these at the top of INPUT, so no other rule
   accepts the jobs' traffic first:

   ```sh
   iptables -I INPUT 1 -i BRIDGE -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT -m comment --comment gavel-sandbox
   iptables -I INPUT 2 -i BRIDGE -d SERVE_IP -p tcp --dport SERVE_PORT -j ACCEPT -m comment --comment gavel-sandbox
   iptables -I INPUT 3 -i BRIDGE -j REJECT -m comment --comment gavel-sandbox
   ```

   The rules don't survive a reboot on their own; persist them with your distribution's iptables
   service.

`gavel serve` reads both chains with `iptables -S DOCKER-USER` and `iptables -S INPUT`, so it
needs to be able to read iptables. It finds the rules by their `gavel-sandbox` comment and checks
that each ACCEPT comes before its REJECTs, and that no other INPUT rule accepts traffic from the
bridge before gavel's REJECT.

The check was written against the format `iptables -S` prints for these rules. It hasn't yet
been run against a live firewall. To check a host from inside a job, run
`GAVEL_SANDBOX_FIREWALL_TEST=1 go test -run TestTheSandboxFirewallConfinesJobs ./sandbox/`.

## The runner image

Sandbox jobs start from the runner image: the one `--sandbox-image` names for docker, podman and
kubernetes, and for substrate the one `--substrate-template`'s image template runs. `make runner-image` builds it from
`sandbox/runner.Dockerfile` (set `RUNNER_IMAGE` to tag it). It holds gavel, Go, Node with pnpm,
Python, git, yq and the Claude Code and Codex CLIs, all at pinned versions, and runs as uid 1000
with `GAVEL_DB=off`. Each release publishes it as `ghcr.io/flanksource/gavel-runner:<tag>`; on substrate, register it
by digest as an image template (`ate-env manifest template --rootfs-image
ghcr.io/flanksource/gavel-runner@sha256:<digest>`) and name that template in the Helm chart's
`sandbox.substrate.template`.

## Podman

Podman works when it runs as root. Rootless podman isn't supported: its networking bypasses the
host's iptables rules, so `gavel serve` refuses it.

## Kubernetes

In a cluster, the Kubernetes executor runs each job as a Kubernetes Job and applies a
NetworkPolicy with the same egress rules.

## agent-substrate

`--sandbox substrate` runs each job as an environment of
[agent-substrate/env](https://github.com/agent-substrate/env) on an
[agent-substrate](https://github.com/agent-substrate/substrate) cluster: a gVisor sandbox
created from the runner image. The environment is deleted once the job exits or is cancelled.

- **Jobs.** A test, lint or push-check job runs through captain's substrate sandbox as a run of
  captain's `bash` model, whose prompt is the job's command: serve hands captain the job, the
  environment checks the job's commit out of serve's copy of the repository, and runs
  `gavel runner run-task`, which runs `gavel runner exec --run <id> --source .` there. The job
  clones its source from that checkout, so what it writes stays out of it, asks serve for its spec
  and posts its results over the environment's stdout and stdin, which serve answers with its run
  API as the run, and exits with the job's code, which the run reports. `gavel runner run-task`
  serves the run API on a socket, named in `GAVEL_RUN_API_SOCKET`, to what the run starts; serve
  answers a job's spec and results only. The job's output, and the environment's, are its log.
  The environment holds neither the run token nor serve's address, and can't reach serve.

- **Agent runs.** A hosted todo's steps need `--sandbox substrate`; under docker, podman or
  kubernetes they're refused. Each runs whole in its own environment, `gavel-run-<run id>`,
  through captain's substrate sandbox: serve hands captain the step, the environment checks the
  todo's commit (or its PR's head) out of serve's copy of the repository, and runs
  `gavel runner run-task`, which rebuilds the step's hooks there and runs it. Its events, tool
  calls and approvals come back to serve over the environment's stdout and stdin; its log is the
  run's log, its transcripts are ingested as it works, and its commits come back on its own
  branch, `gavel/<todo>-<run>`, which goes to GitHub for a GitHub repository. A step on a linked
  PR reads, pushes to and resolves threads on it by asking serve over that same channel, which
  serve answers with its run API as the run, for its PR's endpoints only; the environment holds
  no run token and can't reach serve, as a job's can't.
- **Cleanup.** A run works at most its timeout, an hour, whatever its step's own budget. Every
  ten minutes serve deletes `gavel-run-` environments older than a run can live (its timeout
  plus 15 minutes), left behind when serve stopped mid-run, and the organizations' templates
  no actor runs that are no longer their organization's (an old size or an old base template),
  or whose organization gavel no longer has. Substrate's access policy must let serve's identity
  list the atespace's actors and templates for this.
- **Image.** Environments are created from `--substrate-template`, which is required: an image
  template an operator registered with `ate-env manifest template --rootfs-image
  <image@sha256:…>`, which sets the image they run and their volumes.
  `--sandbox-image` is refused, since env creates no environment from an image of its own.
- **Organization templates and sizes.** Each organization's runs, jobs and agent runs alike, are
  created from its own copy of `--substrate-template`, `gavel-org-<organization>-<hash>`, which
  gavel makes on the organization's first run. The copy runs what the base template runs at the
  organization's size, the CPU and memory substrate books on a worker and holds each environment
  to: the size a platform admin set for the organization on the platform dashboard, else
  `--substrate-cpu` and `--substrate-memory` (chart `sandbox.substrate.size.cpu` and `.memory`),
  else the base template's own. A copy builds no golden snapshot, so a run starts at once from
  the image, even right after its template was made. A template can't change, so the hash names
  the base template's UID and the size: registering the runner's next image under the same name,
  or a new size, gives each organization a new template on its next run, and the sweep deletes
  the old ones once nothing runs them. Each run records its template and size
  (`sandbox_runs.template`, `cpu_millicores`, `memory_bytes`). Substrate's metrics name an
  environment's template, so an organization's actual CPU and memory show under its template's
  name. Serve's identity must be allowed to get, create, list and delete actor templates in the
  atespace, which its `editor` role allows. A run whose organization can't get its template
  fails without being recorded, saying why.
- **Usage.** Every member sees what the organization used, this month, the last 30 days or this
  week, on the Usage and budgets page of its profile (`GET /api/usage?window=now/M`). Its runs'
  CPU and memory count what each was sized at for as long as it ran, whatever it used of it:
  core-hours and GiB-hours, and the cores and GiB its runs hold now. A run serve left running when
  it stopped counts only as long as a run can live. Its model calls count their tokens by kind
  and their cost in USD: what the provider reported for each call, else its list price, each kind
  of token at the model's price for it; a call priced in another currency counts its tokens but
  not its cost. Counted tokens are input, cache writes, output and reasoning; cache reads are
  shown but not counted, as the providers leave them out of their input limits.
  Its admins also see its measured CPU and memory there, sampled from substrate's Prometheus
  through Mission Control (`--mission-control-org-view`, see
  [platform-dashboard.md](platform-dashboard.md#measured-cpu-and-memory)).
- **Budgets.** An organization admin sets its budgets on the same page (`/api/budgets`), at most
  one per resource and kind; without one, nothing is capped. A cumulative budget caps what the
  organization uses in a window, such as this month (`now/M`): core-hours, GiB-hours, counted
  tokens or USD. A concurrent one caps the cores or GiB its runs hold at once. CPU and memory
  budgets hold back every run; token and model cost budgets hold back agent runs only, as only
  they call models. Budgets are enforced:
  - A run is refused when a cumulative budget it counts towards is used up, or when it is bigger
    than a concurrent budget on its own. Its record fails saying why, and a project action
    answers 429.
  - A run that would take the organization over a concurrent budget waits its turn, first in
    first out, recorded as queued with why, which the Runs view, a push's output and an agent
    run's log show. It starts once a run ends and it fits, or an admin raises the budget, and
    fails after `sandbox.queue.timeout` (a property, 15 minutes). Its token is signed when it
    starts. A run left queued when serve stopped fails at serve's start.
  - Running runs stop once a cumulative budget that covers them is used up, ending cancelled with
    `stopped: <why>`: serve checks every organization with a running run each
    `sandbox.budget.interval` (a property, 30 seconds), and an agent run's organization after each
    of its transcripts is ingested. An agent run's spend is also capped at what the model cost
    budget leaves, and the credential provider is refused the organization's model keys once its
    token or model cost budget is used up.
  - Lowering a concurrent budget below what runs hold stops none of them; new runs wait.
  - What runs hold at once is counted in serve's memory, so serve runs as one replica.
- **Versions.** gavel speaks Substrate 1.0.0-rc.2 and env `56424de` or later.
- **Egress.** Substrate denies every destination an actor's egress policy doesn't allow. gavel
  gives each environment, a job's or an agent run's, a policy that allows proxy.golang.org,
  sum.golang.org, registry.npmjs.org, pypi.org, files.pythonhosted.org, github.com,
  codeload.github.com and objects.githubusercontent.com, plus any `--substrate-egress` hosts,
  over TLS on 443, and never gavel serve, which a run asks over its own channel.
- **Model keys.** Each organization sets its own under Model keys in its settings; they never
  enter the environment, and `--sandbox-secret-env` is refused. Substrate's egress gateway
  injects them into requests instead: each run's policy names, for every key its organization
  has, the model API host, the header to replace and the URI
  `ate-secret://gavel/<--substrate-gavel-name>/<run id>/<provider>`, and the job gets a
  placeholder in the variable the agent reads. Substrate's credential provider resolves the URI
  by asking this gavel's `/api/credentials/fetch` with `GAVEL_CREDENTIAL_PROVIDER_TOKEN`; gavel
  answers only for an active run, asked for by that run's own environment. An agent run whose
  organization has no key is refused. gavel serve logs what each run's policy injects
  (`sandbox run <id> in substrate environment …: the egress gateway injects <host> <header> from
  <URI>`, or `no credentials`) and each key it hands out (`credential fetch for run <id>`), so a
  run that reaches the model API without its key shows whether gavel asked for the key and
  whether the gateway fetched it. The job must trust the gateway's CA for those hosts: mount
  Substrate's `egress-mitm.ate.dev` trust bundle through the template's `systemInfo` volume and
  pass its path as `--substrate-trust-bundle`.
- **Control plane.** gavel serve sets egress policies through ate-api-server
  (`--substrate-control`), authenticating with `--substrate-token-file`, such as a projected
  ServiceAccount token with the `api.ate-system.svc` audience, which Substrate's access policy must
  allow to manage actors in the atespace. `--substrate-ca-file` verifies its certificate.
- **Substrate in another cluster.** When gavel serve runs outside Substrate's cluster, both APIs sit
  behind an ingress that requires a client certificate. `--substrate-client-cert` and
  `--substrate-client-key` present it, and both APIs are then reached over TLS verified against the
  system roots, or `--substrate-ca-file`. The certificate and `--substrate-token-file` are read on
  every connection and call, so renewed ones are picked up without a restart. The token is then
  minted in Substrate's cluster for a ServiceAccount there, since this cluster's tokens aren't
  trusted by it.
- **API.** gavel reaches Substrate only through captain's `pkg/substrate`: its client, and the
  CredentialProvider API `gavel credential-provider` serves through
  `substrate.RegisterCredentialProvider`. captain generates Substrate's and env's gRPC packages at
  the versions it pins, so following a Substrate release is a captain bump.

- **User.** A job runs as the user its image names, uid 1000 in the runner image; gVisor is what
  isolates it.

### Running the substrate e2e

These tests run on a live cluster, whose runner image must be this gavel's, since the environment
runs its `gavel runner run-task`. `TestPushChecksRunOnSubstrate` (`pr/ui`) runs a failing and a
passing push check through the dispatcher: each job runs in a checkout of the pushed commit, asks
a run API on the host over its run's channel and exits with the check's verdict.
`TestSubstrateExecutorAgainstACluster` (`sandbox/substrate_e2e_test.go`) runs a job directly. It
checks the job's output and exit code, that it runs at its commit and reads its spec over its
run's channel while holding no run token, that it reaches proxy.golang.org but neither serve's
address nor example.com, which it can't resolve, that it holds only a placeholder for its model
key while httpbin.org receives the real one, and that its environment is deleted afterwards. On a kind
cluster running Substrate 1.0.0-rc.2 and env `56424de`:

1. From a Substrate `v1.0.0-rc.2` checkout, create the cluster
   (`KIND_CLUSTER_NAME=gavel-substrate-rc2 hack/create-kind-cluster.sh`) and install the release's
   images, letting env's API through the router:
   `ATE_INSTALL_KIND=true go run ./cmd/ate-setup deploy ate-system --rollout-timeout 5m
   --image-repo ghcr.io/agent-substrate/substrate --image-tag v1.0.0-rc.2 --sandbox-classes=gvisor
   --credential-provider='{"name":"k8s.io"}'
   --ingress-allowed-spiffe-ids=spiffe://cluster.local/ns/ate-system/sa/atenet-client,spiffe://cluster.local/ns/ate-env/sa/default`.
2. From an env `56424de` checkout, build `ate-env-api` and `ate-env-guest` into the cluster's
   registry (`KO_DOCKER_REPO=localhost:5001/<name> ko build --bare ./cmd/<name>`), deploy env with
   the release's worker (`go run ./cmd/ate-env manifest --api-image <ate-env-api@sha256:…>
   --worker-image ghcr.io/agent-substrate/substrate/ateworker-gvisor@sha256:<v1.0.0-rc.2's> |
   kubectl apply -f -`), make `system:serviceaccount:ate-env:default` a global `owner`
   (`kubectl ate create access-policy --global`) so it can create its atespace, and restart
   `ate-env-api`.
3. Register the image templates with `go run ./cmd/ate-env manifest template --guest-image
   <ate-env-guest@sha256:…> --snapshots-bucket gs://ate-snapshots/ate-env/ | kubectl ate create
   actor-template -f -`: `default-template` as it is, and `gavel-runner` with `--template
   gavel-runner --rootfs-image localhost:5001/gavel-runner@sha256:<digest>`, plus a `systemInfo`
   volume holding the `trustBundle` named `egress-mitm.ate.dev` as `trust-bundle.pem`, mounted on
   the guest container at `/run/ate/egress-trust-bundle`. rc.2 refuses a template's own
   `snapshotConfig.storageLocation`, which `ate-env manifest template` writes: delete it.
4. Create a Secret `model` with key `key` (value `gavel-e2e-injected-credential`) in namespace
   `gavel-e2e`, and, for the hosted agent run, a Secret `claude-oauth` with key `key` holding a
   Claude Code OAuth token. Allow atespace `ate-env` that namespace in the
   `k8s-credential-provider-namespace-policy` ConfigMap (restart the provider).
5. Create ServiceAccount `gavel-e2e/gavel`, make it an `editor` of atespace `ate-env`
   (`kubectl ate create access-policy -a ate-env`), and mint its token:
   `kubectl -n gavel-e2e create token gavel --audience api.ate-system.svc > gavel.token`.
6. Point a Service at the host so jobs reach the test's stand-in for serve: Service
   `default/gavel-host` on port 18080 with an EndpointSlice holding the kind network's gateway
   (`172.18.0.1`).
7. Port-forward the APIs: `kubectl port-forward -n ate-env svc/ate-env-api 7777:7777` and
   `kubectl port-forward -n ate-system svc/api 8443:443`.

Then run:

```sh
GAVEL_SANDBOX_SUBSTRATE_TEST=1 GAVEL_SUBSTRATE_TOKEN_FILE=gavel.token go test ./sandbox ./pr/ui \
  -run 'TestSubstrateExecutorAgainstACluster|TestPushChecksRunOnSubstrate|TestHostedTodoRunsOnSubstrate|TestOrgBudgetsHoldOnSubstrate'
```

`TestOrgBudgetsHoldOnSubstrate` (`pr/ui`) holds an organization's push checks to its budgets: they
run from its own copy of the runner template at its size (1 core, 1 GiB, read from
`/proc/meminfo`), a second check waits while the first holds the organization's one concurrent
core and starts once it ends, and a used-up CPU budget stops a running check, deleting its
environment, and refuses the next. It needs `gavel-e2e/gavel` to create and delete templates in
`ate-env` (an `editor`), and deletes the template it made.

`TestHostedTodoRunsOnSubstrate` runs a hosted todo's plan step whole in an environment against the
real model, so the `gavel-runner` template's image must hold this tree's gavel; on failure it logs the environment's
output and the tool calls that failed. After the kind node restarts, recreate env's worker pods
(`kubectl delete pods -n ate-env -l pod-template-hash=<hash>`): Substrate keeps their old
addresses, and an environment then starts but never answers.

`GAVEL_SUBSTRATE_ENV_API`, `GAVEL_SUBSTRATE_CONTROL`, `GAVEL_SUBSTRATE_TEMPLATE`,
`GAVEL_SUBSTRATE_SERVE_LISTEN`, `GAVEL_SUBSTRATE_SERVE_URL`, `GAVEL_SUBSTRATE_CREDENTIAL`,
`GAVEL_SUBSTRATE_CREDENTIAL_URI` and `GAVEL_SUBSTRATE_AGENT_CREDENTIAL_URI` override the defaults
the steps above set up.
