import type { GitWorktreeChanges } from './projectGit';

// GET /api/sessions (pr/ui/sessions.go): Captain's top-level agent sessions,
// each joined with the worktree it worked in, that worktree's git state and
// the branch's PR.
export interface AgentSessionsResponse {
  sessions: AgentSession[];
  // Projects whose git state or todo links could not be read; their sessions
  // are still listed without the parts that failed.
  errors: { project: string; error: string }[];
}

// Captain's activity vocabulary (captain SessionActivityState).
export type AgentSessionActivity = 'idle' | 'thinking' | 'working' | 'ask' | 'approval';

export interface AgentSession {
  id: string;
  providerSessionId?: string;
  title: string;
  // The gavel project the session's directory or run belongs to; absent when
  // it ran outside every configured project.
  project?: string;
  source: string;
  provider?: string;
  model?: string;
  effort?: string;
  lifecycleStatus: string;
  activityState: AgentSessionActivity | string;
  stateReason?: string;
  processActive: boolean;
  pid?: number;
  startedAt?: string;
  lastActivityAt: string;
  durationMs?: number;
  cwd?: string;
  context?: { usedTokens: number; windowTokens: number; freePercent: number };
  tokens: { input: number; output: number; cacheRead: number; total: number };
  costUsd: number;
  todo?: { id: string; title: string; status: string };
  worktree?: { path: string; branch: string; base: string; primary: boolean };
  git?: {
    changes: GitWorktreeChanges;
    ahead: number;
    behind: number;
    baseCheckedOut: boolean;
    // False once the recorded worktree was removed (only its branch remains).
    worktreeLive: boolean;
  };
  pr?: {
    number: number;
    url: string;
    state: 'open' | 'draft' | 'merged' | 'closed' | string;
    checkStatus: 'success' | 'failure' | 'pending' | 'none';
    reviewDecision?: string;
    mergeable?: string;
  };
}
