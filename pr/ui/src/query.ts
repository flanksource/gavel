export const queryKeys = {
  prSnapshot: () => ['prs', 'snapshot'] as const,
  prDetail: (repo: string, number: number) => ['prs', repo, number, 'detail'] as const,
  projects: () => ['projects'] as const,
  projectStatusScope: (projectName: string) => ['projects', projectName, 'status'] as const,
  projectStatus: (projectName: string, includeResults = false) => [
    ...queryKeys.projectStatusScope(projectName),
    includeResults ? 'results' : 'base',
  ] as const,
  projectDiff: (projectName: string, path: string, revision: number) => ['projects', projectName, 'diff', path, revision] as const,
  health: () => ['status', 'health'] as const,
  processStatuses: () => ['processes', 'status'] as const,
  processStatus: (projectName: string) => ['processes', projectName, 'status'] as const,
  processLogs: (projectName: string, processName: string, lines: number) => ['processes', projectName, processName, 'logs', lines] as const,
  testRuns: () => ['tests', 'runs'] as const,
  activity: () => ['activity', 'snapshot'] as const,
  activityCache: () => ['activity', 'cache'] as const,
};

interface QueryRequest {
  url: string;
  signal: AbortSignal;
  context: string;
}

interface MutationRequest {
  url: string;
  method: 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  body?: unknown;
  context: string;
}

export async function fetchJSON<T>({ url, signal, context }: QueryRequest): Promise<T> {
  const response = await request({ url, signal, context });
  let payload: unknown;
  try {
    payload = await response.json();
  } catch {
    throw new Error(`${context}: ${response.ok ? 'invalid JSON response' : `HTTP ${response.status}`}`);
  }
  if (!response.ok) throw new Error(`${context}: ${responseError(payload, response.status)}`);
  return payload as T;
}

export async function fetchText({ url, signal, context }: QueryRequest): Promise<string> {
  const response = await request({ url, signal, context });
  const payload = await response.text();
  if (!response.ok) throw new Error(`${context}: ${payload.trim() || `HTTP ${response.status}`}`);
  return payload;
}

export async function mutationJSON<T>({ url, method, body, context }: MutationRequest): Promise<T> {
  const init: RequestInit = body === undefined
    ? { method }
    : { method, headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) };
  let response: Response;
  try {
    response = await fetch(url, init);
  } catch (cause) {
    throw new Error(`${context}: ${cause instanceof Error ? cause.message : 'request failed'}`, { cause });
  }
  const text = await response.text();
  if (!response.ok) {
    let payload: unknown;
    try {
      payload = JSON.parse(text);
    } catch {
      throw new Error(`${context}: ${text.trim() || `HTTP ${response.status}`}`);
    }
    throw new Error(`${context}: ${responseError(payload, response.status)}`);
  }
  if (!text.trim()) return undefined as T;
  try {
    return JSON.parse(text) as T;
  } catch {
    throw new Error(`${context}: invalid JSON response`);
  }
}

async function request({ url, signal, context }: QueryRequest): Promise<Response> {
  try {
    return await fetch(url, { signal });
  } catch (cause) {
    if (signal.aborted) throw cause;
    throw new Error(`${context}: ${cause instanceof Error ? cause.message : 'request failed'}`, { cause });
  }
}

function responseError(payload: unknown, status: number): string {
  if (typeof payload === 'object' && payload !== null && 'error' in payload && typeof payload.error === 'string' && payload.error.trim()) {
    return payload.error;
  }
  return `HTTP ${status}`;
}
