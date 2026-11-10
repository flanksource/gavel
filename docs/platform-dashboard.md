# Platform dashboard

Hosted `gavel serve` has a **Platform** tab that shows how every organization uses gavel. Only
platform admins see it. Each panel reads its own source, so when one source is down or not set up,
the panel says why and the other panels still load.

| Panel | Source |
|---|---|
| Tenants | gavel's organizations, with their members, repositories, GitHub installs and recent runs, merged with Clerk's organization list |
| Sandbox usage | gavel's sandbox runs: runs and sandbox minutes per day by kind, failures, active runs, and each organization's share. Each chart can be shown as a table |
| GitHub App | The App's installations on GitHub, the organization each is linked to, its repositories, core rate limit, and whether it is suspended or was uninstalled from GitHub |
| Infrastructure | The config items, health checks and config changes Mission Control has for gavel's infrastructure |

Windows are limited to 60 days, because sandbox runs are kept for 60 days by default.

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

If Mission Control refuses the token, the panel shows the HTTP status and says to check
`GAVEL_MISSION_CONTROL_TOKEN`.

To check a real Mission Control end to end:

```sh
GAVEL_MISSION_CONTROL_TEST=1 GAVEL_MISSION_CONTROL_URL=… GAVEL_MISSION_CONTROL_TOKEN=… \
  GAVEL_MISSION_CONTROL_INFRA=namespace=gavel go test ./missioncontrol/ -run RealMissionControl
```
