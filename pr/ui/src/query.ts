export const queryKeys = {
  prSnapshot: () => ['prs', 'snapshot'] as const,
  prDetail: (repo: string, number: number) => ['prs', repo, number, 'detail'] as const,
  projects: () => ['projects'] as const,
  projectStatusScope: (projectName: string) => ['projects', projectName, 'status'] as const,
  // worktree is the absolute path of a linked worktree; the main checkout
  // omits it, so its keys are unchanged.
  projectStatus: (projectName: string, includeResults = false, worktree = '') => [
    ...queryKeys.projectStatusScope(projectName),
    includeResults ? 'results' : 'base',
    ...(worktree ? [worktree] : []),
  ] as const,
  projectDiff: (projectName: string, path: string, revision: number, worktree = '') => [
    'projects', projectName, 'diff', path, revision, ...(worktree ? [worktree] : []),
  ] as const,
  projectGitSummary: () => ['projects', 'git-summary'] as const,
  projectGit: (projectName: string) => ['projects', projectName, 'git'] as const,
  // head is the branch tip: new commits make a new key, so a cached range is
  // never shown for a branch that has moved on.
  projectBranchFiles: (projectName: string, branch: string, head: string) => ['projects', projectName, 'branch', branch, head, 'files'] as const,
  projectBranchDiff: (projectName: string, branch: string, head: string, file: string) => ['projects', projectName, 'branch', branch, head, 'diff', file] as const,
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

// HttpError is a non-2xx response from fetchJSON; callers branch on `status`
// (e.g. 410 Gone) instead of matching the message.
export class HttpError extends Error {
  readonly status: number;
  // The parsed JSON error body, when the server sent one, so callers can read
  // fields beyond `error` (e.g. the `conflicts` list of a 409).
  readonly body: unknown;

  constructor(message: string, status: number, body?: unknown) {
    super(message);
    this.name = 'HttpError';
    this.status = status;
    this.body = body;
  }
}

export async function fetchJSON<T>({ url, signal, context }: QueryRequest): Promise<T> {
  const response = await request({ url, signal, context });
  let payload: unknown;
  try {
    payload = await response.json();
  } catch {
    if (!response.ok) throw new HttpError(`${context}: HTTP ${response.status}`, response.status);
    throw new Error(`${context}: invalid JSON response`);
  }
  if (!response.ok) throw new HttpError(`${context}: ${responseError(payload, response.status)}`, response.status, payload);
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
    throw new HttpError(`${context}: ${responseError(payload, response.status)}`, response.status, payload);
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
