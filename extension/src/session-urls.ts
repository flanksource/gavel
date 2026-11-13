type SessionRef = { server: string; sessionId: string };

/** The session's page on the gavel dashboard. */
export function sessionDashboardUrl({ server, sessionId }: SessionRef): string {
  return new URL(`/sessions/${encodeURIComponent(sessionId)}`, server).href;
}

/** Gavel's Captain transcript of the session, which `SessionInspector` loads and follows while it is live. */
export function captainSessionUrl({ server, sessionId }: SessionRef): string {
  return new URL(`/api/captain/sessions/${encodeURIComponent(sessionId)}`, server).href;
}
