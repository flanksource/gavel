type SessionRef = { server: string; sessionId: string };

const AMP_THREAD_PREFIX = "https://ampcode.com/threads/";

export function ampThreadId(value: string): string | undefined {
  const text = value.trim();
  const id = text.startsWith(AMP_THREAD_PREFIX) ? text.slice(AMP_THREAD_PREFIX.length) : text;
  return id.length <= 128 && /^T-[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*$/.test(id) ? id : undefined;
}

export function ampThreadUrl(sessionId: string | undefined): string | undefined {
  const id = sessionId && ampThreadId(sessionId);
  return id ? AMP_THREAD_PREFIX + id : undefined;
}

/** The session's page on the gavel dashboard. */
export function sessionDashboardUrl({ server, sessionId }: SessionRef): string {
  return new URL(`/sessions/${encodeURIComponent(sessionId)}`, server).href;
}

/** Gavel's Captain transcript of the session, which `SessionInspector` loads and follows while it is live. */
export function captainSessionUrl({ server, sessionId }: SessionRef): string {
  return new URL(`/api/captain/sessions/${encodeURIComponent(sessionId)}`, server).href;
}
