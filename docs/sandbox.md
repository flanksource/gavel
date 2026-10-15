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

   The rules don't survive a reboot on their own; persist them with your distribution's iptables
   service.

`gavel serve` reads the chain with `iptables -S DOCKER-USER`, so it needs to be able to read
iptables. It finds the rules by their `gavel-sandbox` comment and checks that the ACCEPT comes
before the REJECTs.

The check was written against the format `iptables -S` prints for these rules. It hasn't yet
been run against a live firewall.

## The runner image

Sandbox jobs start from the image `--sandbox-image` names. `make runner-image` builds it from
`sandbox/runner.Dockerfile` (set `RUNNER_IMAGE` to tag it). It holds gavel, Go, Node with pnpm,
Python, git, yq and the Claude Code and Codex CLIs, all at pinned versions, and runs as uid 1000
with `GAVEL_DB=off`.

## Podman

Podman works when it runs as root. Rootless podman isn't supported: its networking bypasses the
host's iptables rules, so `gavel serve` refuses it.

## Kubernetes and agent-substrate

In a cluster, the Kubernetes executor runs each job as a Kubernetes Job and applies a
NetworkPolicy with the same egress rules. The main hosted backend will later be
[agent-substrate](https://github.com/agent-substrate/substrate) on Kubernetes.
