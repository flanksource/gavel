# Clerk authentication

The `api` package implements the authentication client for the Gavel CLI. It is
not yet a hosted API server or tenant authorization layer.

## Clerk setup

1. Create a Clerk instance and enable your desired sign-in providers.
2. In Clerk's OAuth applications settings, create a **public** OAuth client.
   Never distribute a client secret in the CLI.
3. Enable **Device Authorization Grant** for the instance/application as required
   by your Clerk settings. Grant `openid`, `profile`, `email`, and `offline_access`.
   For organization identity, enable Clerk **Organizations** and also grant
   `user:org:read` to this OAuth application. Gavel requests it when the issuer
   advertises support, allowing organization selection on the consent screen.
   Existing sessions need a new `gavel login` after this change.
4. Copy the client ID and the Clerk **Frontend API URL** (not the Backend API URL
   or your application's website URL).

Clerk must advertise `device_authorization_endpoint`, `token_endpoint`, and
`jwks_uri` in `<issuer>/.well-known/oauth-authorization-server`. Instances using
older OAuth implementations may need upgrading/enabling the device flow. The
CLI fails explicitly rather than falling back to an insecure flow.

```bash
export GAVEL_CLERK_ISSUER=https://your-instance.clerk.accounts.dev
export GAVEL_CLERK_CLIENT_ID=your_public_oauth_client_id

gavel login
gavel whoami --json
gavel logout
```

You can instead pass `gavel login --issuer URL --client-id ID`. Use
`gavel login --no-browser` over SSH; open the displayed verification URL on
another device and enter the user code. Login waits up to five minutes (or the
provider's shorter device-code expiry). Ctrl-C cancels it.

`whoami` returns verified identity fields in this order: `id`, `issuer`, `name`,
`email`, `organization_id`, `organization_slug`. Missing name, email, or organization values remain
visible as empty strings; Gavel never invents an organization ID. It uses the issuer and
client ID saved at login, not changed environment variables. It refreshes expired
tokens when a refresh token is available; otherwise run login again. Identity is
verified using RS256, issuer, audience, expiration, and the issuer's JWKS.
Organization claims are informational, not proof of project authorization.
When PostgreSQL is configured, login/whoami upsert `users` and `orgs`, retaining
stored org slugs when OAuth omits them. See [identity and ownership](OWNERSHIP.md)
for schema coverage, backend slug resolution, and the remaining tenant isolation
work. Clerk's OAuth tokens are not guaranteed to include a slug.

## Credentials and boundaries

Credentials are stored in `os.UserConfigDir()/gavel/clerk-session.json` (normally
`~/.config/gavel/clerk-session.json` on Linux), using atomic replacement and
owner-only `0600` file permissions. This is an initial **file-backed store**, not
an OS keychain. Keep it outside repositories and backups shared with others.
GitHub credentials in `service/authconfig.go` are independent and unchanged.
Concurrent CLI token refreshes are not coordinated yet; avoid concurrent
`whoami` calls with rotating refresh tokens.

Logout removes local credentials only. It does not revoke Clerk sessions or
clear browser cookies. `whoami` verifies signed identity, but a still-valid JWT
is not a live revocation/membership check. Future hosted endpoints must perform
their own access-token validation and tenant authorization; an ID token is not
an API access token. Login does not upload project data, alter local test/lint
behavior, or implement tenancy.

Interactive credential commands are excluded from the automatically exposed MCP
command set.

Reference: https://clerk.com/docs/guides/configure/auth-strategies/oauth/how-clerk-implements-oauth
