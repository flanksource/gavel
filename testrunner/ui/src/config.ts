// When the testrunner UI is served under a prefix (e.g. /results/{repo}/{id}),
// the host page sets window.__gavelBasePath so API and route URLs resolve.
// When served standalone this is empty, so all paths remain absolute as before.
export const basePath: string = ((window as any).__gavelBasePath as string) || '';

export function apiUrl(path: string): string {
  return basePath + path;
}
