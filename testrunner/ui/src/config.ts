// When the testrunner UI is served under a prefix (e.g. /results/{repo}/{id}),
// the host page sets window.__gavelBasePath so API and route URLs resolve.
// When served standalone this is empty, so all paths remain absolute as before.
export const basePath: string =
  (typeof window !== 'undefined' && ((window as unknown as { __gavelBasePath?: string }).__gavelBasePath)) || '';

/** The URL prefix of one run's server: the page's base path, then the run id when a MultiServer hosts many runs. */
export function apiBase({ baseUrl = basePath, runId }: { baseUrl?: string; runId?: string } = {}): string {
  return runId ? `${baseUrl}/${encodeURIComponent(runId)}` : baseUrl;
}

export function apiUrl(path: string, scope?: { baseUrl?: string; runId?: string }): string {
  return apiBase(scope) + path;
}
