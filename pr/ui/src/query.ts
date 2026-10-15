export const queryKeys = {
  me: () => ['me'] as const,
  sshKeys: () => ['ssh-keys'] as const,
  prSnapshot: () => ['prs', 'snapshot'] as const,
  prDetail: (repo: string, number: number) => ['prs', repo, number, 'detail'] as const,
  linkedPR: (url: string) => ['prs', 'linked', url] as const,
  // Links between a todo and other todos / PRs, and the reverse lookup from a PR.
  todoLinks: (dir: string, ref: string) => ['todos', 'links', { dir: dir.trim(), ref }] as const,
  todoPRLinks: (dir: string, ref: string) => ['todos', 'pr-links', { dir: dir.trim(), ref }] as const,
  prLinkedTodos: (url: string) => ['prs', 'todos', url] as const,
  prSearch: (query: string, repo: string) => ['prs', 'search', { query, repo }] as const,
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
  agentSessions: () => ['sessions', 'list'] as const,
  agentSessionChat: (sessionId: string) => ['sessions', 'chat', sessionId] as const,
  health: () => ['status', 'health'] as const,
  processStatuses: () => ['processes', 'status'] as const,
  processStatus: (projectName: string) => ['processes', projectName, 'status'] as const,
  processLogs: (projectName: string, processName: string, lines: number) => ['processes', projectName, processName, 'logs', lines] as const,
  testRuns: () => ['tests', 'runs'] as const,
  activity: () => ['activity', 'snapshot'] as const,
  activityCache: () => ['activity', 'cache'] as const,
  gitMetrics: () => ['activity', 'git'] as const,
};

/** Matches queryKeys.projectGit of any project, for invalidating them all. */
export function isProjectGitKey(key: readonly unknown[]): boolean {
  return key.length === 3 && key[0] === 'projects' && key[2] === 'git';
}

interface QueryRequest {
  url: string;
  signal: AbortSignal;
  context: string;
}

interface MutationRequest {
  url: string;
  method: 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  body?: unknown;
  // token is a Clerk session token sent as a bearer token instead of relying on the
  // session cookie, which may have lapsed.
  token?: string;
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

export type AuthMode = 'none' | 'clerk';

// A signed-in Clerk session token lives about a minute and is refreshed in the
// background; a 401 means it lapsed, and a reload lets Clerk issue a new one.
// Reloads are spaced so a server that keeps refusing the session can't loop.
const authReloadKey = 'gavel.authReloadAt';
const authReloadSpacingMs = 30_000;
let authMode: AuthMode = 'none';
let reloadPage = () => window.location.reload();

// setAuthMode records how the dashboard signs in, set once before the app renders.
export function setAuthMode(mode: AuthMode, reload: () => void = () => window.location.reload()) {
  authMode = mode;
  reloadPage = reload;
}

// currentAuthMode is how the dashboard signs in: "none" until setAuthMode says otherwise.
export function currentAuthMode(): AuthMode {
  return authMode;
}

function reloadWhenSignedOut(response: Response) {
  if (response.status !== 401 || authMode !== 'clerk') return;
  let last = 0;
  try {
    last = Number(sessionStorage.getItem(authReloadKey) ?? 0);
    if (Date.now() - last < authReloadSpacingMs) return;
    sessionStorage.setItem(authReloadKey, String(Date.now()));
  } catch {
    // Without storage there is no loop guard, so leave the error on screen.
    return;
  }
  reloadPage();
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

export async function mutationJSON<T>({ url, method, body, token, context }: MutationRequest): Promise<T> {
  const headers: Record<string, string> = {
    ...(body === undefined ? {} : { 'Content-Type': 'application/json' }),
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
  const init: RequestInit = { method };
  if (Object.keys(headers).length > 0) init.headers = headers;
  if (body !== undefined) init.body = JSON.stringify(body);
  let response: Response;
  try {
    response = await fetch(url, init);
  } catch (cause) {
    throw new Error(`${context}: ${cause instanceof Error ? cause.message : 'request failed'}`, { cause });
  }
  reloadWhenSignedOut(response);
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
    const response = await fetch(url, { signal });
    reloadWhenSignedOut(response);
    return response;
  } catch (cause) {
    if (signal.aborted) throw cause;
    throw new Error(`${context}: ${cause instanceof Error ? cause.message : 'request failed'}`, { cause });
  }
}

// responseError reads {"error": ...}, adding the server's "message" (why it failed)
// when it sends one alongside a stable error code.
function responseError(payload: unknown, status: number): string {
  if (typeof payload === 'object' && payload !== null && 'error' in payload && typeof payload.error === 'string' && payload.error.trim()) {
    const message = 'message' in payload && typeof payload.message === 'string' ? payload.message.trim() : '';
    return message ? `${payload.error}: ${message}` : payload.error;
  }
  return `HTTP ${status}`;
}
