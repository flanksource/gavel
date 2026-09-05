import type { SessionUIMessage } from '@flanksource/clicky-ui/ai';
import type { VerifyReport } from '@flanksource/clicky-ui/data';

// Rolled-up stats for a TODO's agent session (see /api/todos/session/stats):
// identity (agent/model/effort), elapsed time, token usage and derived cost.
// Mirrors cmux.SessionStats. found=false means the session produced no log yet.
export interface SessionStats {
  sessionId?: string;
  agent?: string;
  model?: string;
  effort?: string;
  startedAt?: string;
  updatedAt?: string;
  durationMs: number;
  inputTokens: number;
  outputTokens: number;
  cacheReadTokens: number;
  cacheCreationTokens: number;
  totalTokens: number;
  // Live context-window occupancy (latest turn's input + cache), reset by each
  // compaction — surfaced as the token figure instead of the ever-growing total.
  contextTokens: number;
  // Total context-window size (tokens) for the model, from captain's pricing
  // registry; 0 when the model is unknown. Denominator for the context bar.
  contextWindow: number;
  turns: number;
  // Number of context compactions seen so far; each shrinks contextTokens.
  compactions: number;
  costUsd: number;
  inProgress: boolean;
  found: boolean;
  // High-level agent state from the latest session-log event: thinking | working
  // | ask | completed | error. 'approval' is set when a tool-permission request
  // is pending (see approval). Empty before the first event.
  state?: 'thinking' | 'working' | 'ask' | 'approval' | 'completed' | 'error';
  // API/network failure reason when state === 'error' (the "API Error: …" message).
  error?: string;
  // Pending tool-permission requests awaiting the user's approve/deny/respond.
  // Non-empty (and state === 'approval') only while a driver is blocked on at
  // least one. Also served standalone by GET /api/todos/session/approvals.
  approvals?: TodoSessionApproval[];
}

export interface SessionRuntimeSelection {
  provider?: string;
  // mode is the runtime mechanism: api | agent | cli | cmux.
  mode?: string;
  model?: string;
  effort?: string;
}

export interface TodoSessionAttempt {
  promptRunId: string;
  ordinal: number;
  step: string;
  mode?: string;
  driver?: string;
  requested: SessionRuntimeSelection;
  resolved: SessionRuntimeSelection;
  provider?: string;
  // runtimeMode is the mechanism the model ran on; `mode` above is the run mode
  // (run/plan) — two different axes that share a word.
  runtimeMode?: string;
  model?: string;
  effort?: string;
  status: string;
  pid?: number;
  processActive: boolean;
  state: string;
  phase: string;
  queuedAt: string;
  startedAt?: string;
  finishedAt?: string;
  durationMs?: number;
  error?: string;
  resultText?: string;
  resultJson?: Record<string, unknown>;
  // The step's captain VerifyReport — a running snapshot while the attempt is
  // live, the terminal result once it finishes. Always sent now (null for an
  // attempt that was never verified) — see todo_session_detail.go's
  // todoAttemptDetail.Verification.
  verification: VerifyReport | null;
  admissionSessionId: string;
  executionSessionId?: string;
  providerSessionId?: string;
  canStop?: boolean;
  stopping?: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface TodoSessionDiagnostic {
  severity: 'warning' | 'error';
  code: string;
  message: string;
  details?: unknown;
}

export interface TodoSessionOverview {
  id: string;
  providerSessionId?: string;
  source: string;
  provider?: string;
  hostId: string;
  parentSessionId?: string;
  rootSessionId?: string;
  path?: string;
  historyFile?: string;
  project?: string;
  cwd?: string;
  title?: string;
  agentType?: string;
  description?: string;
  lifecycleStatus: string;
  activityState: string;
  healthState: string;
  startedAt?: string;
  endedAt?: string;
  lastActivityAt?: string;
  processActive: boolean;
  processStatus?: string;
  pid?: number;
  model?: string;
  modelProvider?: string;
  modelMode?: string;
  effort?: string;
  messageCount: number;
  turnCount: number;
  inputTokens: number;
  outputTokens: number;
  totalTokens: number;
  costUsd: number;
}

export interface TodoSessionTurn {
  id: string;
  sessionId: string;
  providerTurnId?: string;
  turnIndex: number;
  status: string;
  stopReason?: string;
  error?: string;
  startedAt?: string;
  endedAt?: string;
  model?: string;
  modelProvider?: string;
  modelMode?: string;
  effort?: string;
  inputTokens: number;
  outputTokens: number;
  reasoningTokens: number;
  cacheReadTokens: number;
  cacheWriteTokens: number;
  totalTokens: number;
  costUsd: number;
  messageCount: number;
}

export interface TodoSessionAgent {
  id: string;
  sessionId: string;
  parentSessionId?: string;
  rootSessionId?: string;
  isRoot: boolean;
  agentType?: string;
  description?: string;
  historyFile?: string;
  source: string;
  provider?: string;
  lifecycleStatus: string;
  activityState: string;
  healthState: string;
  startedAt?: string;
  endedAt?: string;
  inputTokens: number;
  outputTokens: number;
  reasoningTokens: number;
  cacheReadTokens: number;
  cacheWriteTokens: number;
  totalTokens: number;
  costUsd: number;
}

export interface TodoSessionCost {
  id: string;
  sessionId: string;
  model: string;
  provider?: string;
  modelMode?: string;
  effort?: string;
  currency: string;
  modelCallCount: number;
  inputTokens: number;
  outputTokens: number;
  reasoningTokens: number;
  cacheReadTokens: number;
  cacheWriteTokens: number;
  totalTokens: number;
  totalCost: number;
  firstCallAt?: string;
  lastCallAt?: string;
}

export interface TodoProviderThread {
  id: string;
  providerSessionId?: string;
  status: string;
  root: TodoSessionOverview;
  sessions: TodoSessionOverview[];
  turns: TodoSessionTurn[];
  agents: TodoSessionAgent[];
  costs: TodoSessionCost[];
  messages: SessionUIMessage[];
  startedAt?: string;
  lastActivityAt?: string;
  durationMs: number;
  inputTokens: number;
  outputTokens: number;
  totalTokens: number;
  costUsd: number;
}

export interface TodoSessionDetailResponse {
  attempts: TodoSessionAttempt[];
  selectedPromptRunId?: string;
  selectedExecutionSessionId?: string;
  thread?: TodoProviderThread;
  diagnostics: TodoSessionDiagnostic[];
  // Set when the request asked for attempts only: the provider thread was
  // deliberately not resolved, so a missing `thread` is not a missing session.
  attemptsOnly?: boolean;
}

// TodoSessionApproval is a tool-permission request a driver surfaced for human
// review; the dashboard answers it via POST /api/todos/session/approve with
// `{approvalId, action: "approve" | "deny" | "respond", message?, input?}`.
export interface TodoSessionApproval {
  approvalId: string;
  sessionId: string;
  toolUseId?: string;
  tool: string;
  input?: Record<string, unknown>;
}
