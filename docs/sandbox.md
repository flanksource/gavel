# Sandboxed runs on a container host

Hosted `gavel serve` runs tests, lint and push checks in sandbox jobs. On a single host, those jobs
are docker (or rootful podman) containers. Each job runs:

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

Sandbox jobs start from the image `--sandbox-image` names. `make runner-image` builds it from
`sandbox/runner.Dockerfile` (set `RUNNER_IMAGE` to tag it). It holds gavel, Go, Node with pnpm,
Python, git, yq and the Claude Code and Codex CLIs, all at pinned versions, and runs as uid 1000
with `GAVEL_DB=off`.

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
created from the runner image, whose one process is the job. The environment is deleted once the
job exits or is cancelled.

- **Image.** Substrate runs images pinned by digest, so `--sandbox-image` must be
  `repo@sha256:<digest>`. Environments derive from `--substrate-template` (env's
  `default-template` when empty), which sets their CPU, memory and volumes.
- **Egress.** Substrate denies every destination an actor's egress policy doesn't allow. gavel
  gives each environment a policy that allows gavel serve (`--public-url`, which must be a DNS
  name), proxy.golang.org, sum.golang.org, registry.npmjs.org, pypi.org, files.pythonhosted.org,
  github.com, codeload.github.com and objects.githubusercontent.com, plus any
  `--substrate-egress` hosts.
- **Model keys.** They never enter the environment, and `--sandbox-secret-env` is refused.
  Substrate's egress gateway injects them into requests instead: each `--substrate-credential`
  names the host, the header to replace (and an optional `prefix`, such as `Bearer `), the
  variable the agent reads (set to a placeholder in the job), and the credential's `ate-secret://`
  URI, for example
  `host=api.anthropic.com,header=x-api-key,env=ANTHROPIC_API_KEY,uri=ate-secret://k8s.io/default/gavel/anthropic/key`.
  The job must trust the gateway's CA for those hosts: mount Substrate's `egress-mitm.ate.dev`
  trust bundle through the template's `systemInfo` volume and pass its path as
  `--substrate-trust-bundle`. Grant the atespace the Secret's namespace in Substrate's
  credential provider.
- **Control plane.** gavel serve sets egress policies through ate-api-server
  (`--substrate-control`), authenticating with `--substrate-token-file`, such as a projected
  ServiceAccount token with the `api.ate-system.svc` audience, which Substrate's access policy must
  allow to manage actors in the atespace. `--substrate-ca-file` verifies its certificate.
- **API.** gavel carries copies of Substrate's and env's generated gRPC packages in
  `sandbox/internal/substrate`, refreshed by its `sync.sh`, rather than depending on their Go
  modules.

- **User.** env's guest starts the job as root, whatever user the image names; gVisor is what
  isolates it.

### Running the substrate e2e

`TestSubstrateExecutorAgainstACluster` (`sandbox/substrate_e2e_test.go`) runs a job on a live
cluster. It checks the job's output and exit code, that it reaches serve and proxy.golang.org but
can't resolve example.com, that it holds only a placeholder for its model key while
httpbin.org receives the real one, and that its environment is deleted afterwards. On a kind
cluster with Substrate installed (`ate-setup deploy ate-system`):

1. Deploy env (`ate-env manifest --worker-image <ateworker-gvisor@sha256:…> --api-image
   <ate-env-api@sha256:…> | kubectl apply -f -`), and make its ServiceAccount
   `system:serviceaccount:ate-env:default` a global `owner` in Substrate's access policy so it
   can create its atespace.
2. Register `default-template` (`ate-env manifest template`), then a `gavel-runner` template:
   the same, plus a `systemInfo` volume holding the `egress-mitm.ate.dev` trust bundle as
   `trust-bundle.pem`, mounted on the guest container at `/run/ate/egress-trust-bundle`. A
   Substrate that no longer has `ActorTemplate.sandboxConfig` (removed in agent-substrate#2780)
   rejects the field `ate-env manifest template` writes; delete it from the manifest.
3. Create a Secret `model` with key `key` (value `gavel-e2e-injected-credential`) in namespace
   `gavel-e2e`, and allow atespace `ate-env` that namespace in the
   `k8s-credential-provider-namespace-policy` ConfigMap (restart the provider).
4. Create ServiceAccount `gavel-e2e/gavel`, make it an `editor` of atespace `ate-env`, and mint its
   token: `kubectl -n gavel-e2e create token gavel --audience api.ate-system.svc > gavel.token`.
5. Point a Service at the host so jobs reach the test's stand-in for serve: Service
   `default/gavel-host` on port 18080 with an EndpointSlice holding the kind network's gateway
   (`172.18.0.1`).
6. Push a runner image to the cluster's registry and port-forward the APIs:
   `kubectl port-forward -n ate-env svc/ate-env-api 7777:7777` and
   `kubectl port-forward -n ate-system svc/api 8443:443`.

Then run:

```sh
GAVEL_SANDBOX_SUBSTRATE_TEST=1 GAVEL_SUBSTRATE_IMAGE=localhost:5001/gavel-runner@sha256:<digest> \
  GAVEL_SUBSTRATE_TOKEN_FILE=gavel.token go test ./sandbox -run TestSubstrateExecutorAgainstACluster
```

`GAVEL_SUBSTRATE_ENV_API`, `GAVEL_SUBSTRATE_CONTROL`, `GAVEL_SUBSTRATE_TEMPLATE`,
`GAVEL_SUBSTRATE_SERVE_LISTEN`, `GAVEL_SUBSTRATE_SERVE_URL`, `GAVEL_SUBSTRATE_CREDENTIAL` and
`GAVEL_SUBSTRATE_CREDENTIAL_URI` override the defaults the steps above set up.
