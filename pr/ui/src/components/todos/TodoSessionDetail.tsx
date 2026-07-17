import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  SessionInspector,
  type SessionAgent,
  type SessionCollectionInput,
  type SessionCollectionItem,
  type SessionEntry,
  type SessionInput,
  type SessionPendingTool,
  type SessionToolDecision,
  type SessionUIMessage,
  type UnifiedSessionInput,
} from '@flanksource/clicky-ui/ai';
import { Button } from '@flanksource/clicky-ui/components';
import { UiCheck, UiChevronDown, UiChevronRight, UiCopy, UiError, UiStop, UiWarningTriangle } from '@flanksource/clicky-ui/icons';
import type { TodoSessionAttempt, TodoSessionDetailResponse, TodoSessionDiagnostic } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { copyText } from '../../clipboard';
import { sessionResponseError } from './SessionErrorDetails';

export function useTodoSessionDetail(dir: string, ref: string, sessionId: string | undefined, active: boolean) {
  const [detail, setDetail] = useState<TodoSessionDetailResponse | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    setDetail(null);
    setError('');
    if (!active || !ref) return;
    let cancelled = false;
    const poll = async () => {
      try {
        const body = await fetchTodoSessionDetail(dir, ref, sessionId);
        if (!cancelled) {
          setDetail(body);
          setError('');
        }
      } catch (reason) {
        if (!cancelled) setError(`Session detail request failed\n${reason instanceof Error ? reason.stack || reason.message : String(reason)}`);
      }
    };
    void poll();
    const timer = window.setInterval(poll, 1500);
    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, [active, dir, ref, sessionId]);

  return { detail, error };
}

export async function fetchTodoSessionDetail(dir: string, ref: string, sessionId?: string) {
  const params = new URLSearchParams();
  if (dir.trim()) params.set('dir', dir.trim());
  params.set('ref', ref);
  if (sessionId) params.set('sessionId', sessionId);
  const response = await fetch(`/api/todos/session/detail?${params.toString()}`);
  const body = (await response
    .clone()
    .json()
    .catch(() => null)) as TodoSessionDetailResponse | null;
  if ((response.ok || response.status === 409) && body?.attempts && body.diagnostics) return body;
  throw new Error(await sessionResponseError(response, 'Session detail request failed'));
}

export function SessionDiagnostics({ diagnostics }: { diagnostics: TodoSessionDiagnostic[] }) {
  const hasError = diagnostics.some((diagnostic) => diagnostic.severity === 'error');
  const [expanded, setExpanded] = useState(hasError);
  const [copied, setCopied] = useState(false);
  if (diagnostics.length === 0) return null;
  const Chevron = expanded ? UiChevronDown : UiChevronRight;
  const Icon = hasError ? UiError : UiWarningTriangle;
  const text = JSON.stringify(diagnostics, null, 2);
  const tone = hasError ? 'border-red-500/25 bg-red-500/10 text-red-700 dark:text-red-300' : 'border-amber-500/25 bg-amber-500/10 text-amber-800 dark:text-amber-300';

  return (
    <section className={`shrink-0 border-b px-3 py-2 text-xs ${tone}`}>
      <div className="flex items-center gap-2">
        <Icon className="shrink-0" />
        <span className="min-w-0 flex-1 font-medium">{diagnostics.map((item) => item.message).join(' ')}</span>
        <Button variant="ghost" type="button" onClick={() => setExpanded((value) => !value)} aria-expanded={expanded} className="h-8 gap-1 px-2 text-[11px]">
          <Chevron /> {expanded ? 'Hide details' : 'Show details'}
        </Button>
      </div>
      {expanded && (
        <div className="mt-2 rounded border border-current/15 bg-background/80 p-2 text-foreground">
          <div className="mb-1 flex justify-end">
            <Button
              variant="ghost"
              type="button"
              aria-label="Copy diagnostic details"
              className="h-8 gap-1 px-2 text-[11px]"
              onClick={async () => {
                await copyText(text);
                setCopied(true);
                window.setTimeout(() => setCopied(false), 1800);
              }}
            >
              {copied ? <UiCheck className="text-emerald-600" /> : <UiCopy />} {copied ? 'Copied' : 'Copy details'}
            </Button>
          </div>
          <pre className="max-h-48 overflow-auto whitespace-pre-wrap break-words font-mono text-[10px]">{text}</pre>
        </div>
      )}
    </section>
  );
}

export function ThreadInspector({
  detail,
  entries,
  dir,
  todoRef,
  onStop,
  pendingTools,
  onPendingToolDecision,
}: {
  detail: TodoSessionDetailResponse;
  entries: Array<SessionEntry | SessionUIMessage>;
  dir: string;
  todoRef: string;
  onStop: (attempt: TodoSessionAttempt) => Promise<void>;
  pendingTools: SessionPendingTool[];
  onPendingToolDecision: (decision: SessionToolDecision) => Promise<void> | void;
}) {
  const loadAttempt = useCallback(
    async (attempt: TodoSessionAttempt) => {
      const sessionId = attempt.executionSessionId || attempt.providerSessionId;
      if (!sessionId) throw new Error(`Attempt #${attempt.ordinal} has no execution session`);
      const loaded = await fetchTodoSessionDetail(dir, todoRef, sessionId);
      if (!loaded.thread) throw new Error(`Attempt #${attempt.ordinal} has no provider thread`);
      return requireUnifiedSession(inspectorSession(loaded, loaded.thread.messages));
    },
    [dir, todoRef]
  );
  const collection = useMemo(() => attemptSessionCollection(detail, entries, loadAttempt), [detail, entries, loadAttempt]);
  return (
    <SessionInspector
      session={collection}
      className="h-full"
      transcriptProps={{
        pendingTools,
        onPendingToolDecision,
        showHeader: false,
        className: 'text-xs',
      }}
      renderSessionActions={(item) => {
        const attempt = detail.attempts.find((candidate) => candidate.promptRunId === item.id);
        return attempt ? <AttemptStopAction attempt={attempt} onStop={onStop} /> : null;
      }}
    />
  );
}

export function attemptSessionCollection(detail: TodoSessionDetailResponse, entries: Array<SessionEntry | SessionUIMessage>, loadAttempt: (attempt: TodoSessionAttempt) => Promise<UnifiedSessionInput>): SessionCollectionInput {
  if (!detail.thread || !detail.selectedPromptRunId) {
    throw new Error('A selected provider thread is required for session hierarchy');
  }
  const currentMessages = mergeTranscriptMessages(detail.thread.messages, entries);
  const current = requireUnifiedSession(inspectorSession(detail, currentMessages));
  return {
    kind: 'session-collection',
    id: `todo-attempts:${detail.thread.id}`,
    currentSessionId: detail.selectedPromptRunId,
    sessions: detail.attempts.map((attempt) => {
      const status = attempt.stopping ? 'stopping' : attempt.status;
      const item: SessionCollectionItem = {
        id: attempt.promptRunId,
        label: `Attempt #${attempt.ordinal}`,
        mode: attempt.mode || attempt.step,
        status,
        summary: {
          provider: attempt.provider,
          backend: attempt.backend,
          model: attempt.model,
          effort: attempt.effort,
          mode: attempt.mode || attempt.step,
          status,
          pid: attempt.pid,
          durationMs: attempt.durationMs,
          updatedAt: attempt.updatedAt,
        },
      };
      if (attempt.promptRunId === detail.selectedPromptRunId) item.session = current;
      return item;
    }),
    loadSession: (item) => {
      const attempt = detail.attempts.find((candidate) => candidate.promptRunId === item.id);
      if (!attempt) throw new Error(`Unknown attempt ${item.id}`);
      return loadAttempt(attempt);
    },
  };
}

function AttemptStopAction({ attempt, onStop }: { attempt: TodoSessionAttempt; onStop: (attempt: TodoSessionAttempt) => Promise<void> }) {
  const [stopping, setStopping] = useState(attempt.stopping);
  const [error, setError] = useState('');
  if (!attempt.canStop) return null;
  const stop = async () => {
    setStopping(true);
    setError('');
    try {
      await onStop(attempt);
    } catch (reason) {
      setStopping(false);
      setError(reason instanceof Error ? reason.message : String(reason));
    }
  };
  return (
    <span className="inline-flex items-center gap-1">
      {error ? (
        <span role="alert" className="max-w-32 truncate text-[10px] text-red-600" title={error}>
          {error}
        </span>
      ) : null}
      <Button
        variant="ghost"
        size="icon"
        type="button"
        aria-label={`Stop attempt #${attempt.ordinal}`}
        title={`Stop attempt #${attempt.ordinal}`}
        disabled={stopping}
        onClick={() => void stop()}
        className="size-7 text-red-600 hover:bg-red-500/10 hover:text-red-700"
      >
        {stopping ? <Spinner className="size-3.5" /> : <UiStop className="size-3.5" />}
      </Button>
    </span>
  );
}

export function CopyAllDetailsButton({ detail, entries }: { detail: TodoSessionDetailResponse; entries: Array<SessionEntry | SessionUIMessage> }) {
  const [copied, setCopied] = useState(false);
  return (
    <Button
      variant="ghost"
      type="button"
      aria-label="Copy all session details"
      className="h-8 gap-1 px-2 text-[11px]"
      onClick={async () => {
        await copyText(JSON.stringify({ ...detail, transcript: entries }, null, 2));
        setCopied(true);
        window.setTimeout(() => setCopied(false), 1800);
      }}
    >
      {copied ? <UiCheck className="text-emerald-600" /> : <UiCopy />} {copied ? 'Copied' : 'Copy all'}
    </Button>
  );
}

export function inspectorSession(detail: TodoSessionDetailResponse, entries: Array<SessionEntry | SessionUIMessage>): SessionInput {
  const thread = detail.thread;
  if (!thread) {
    const unified = entries.filter((entry): entry is SessionUIMessage => 'parts' in entry);
    if (unified.length > 0) return unified;
    return entries.filter((entry): entry is SessionEntry => !('parts' in entry));
  }
  const messages = entries.filter((entry): entry is SessionUIMessage => 'parts' in entry);
  const agents: SessionAgent[] = thread.agents.map((agent) => ({
    id: agent.id,
    parentId: agent.parentSessionId,
    type: agent.agentType,
    desc: agent.description,
    isRoot: agent.isRoot,
    historyFile: agent.historyFile,
    usage: {
      inputTokens: agent.inputTokens,
      outputTokens: agent.outputTokens,
      reasoningTokens: agent.reasoningTokens,
      cacheReadTokens: agent.cacheReadTokens,
      cacheWriteTokens: agent.cacheWriteTokens,
      totalTokens: agent.totalTokens,
    },
    cost: { inputCost: agent.costUsd },
  }));
  const agentsById = new Map(agents.flatMap((agent) => (agent.id ? [[agent.id, agent] as const] : [])));
  for (const agent of agents) {
    if (!agent.parentId) continue;
    const parent = agentsById.get(agent.parentId);
    if (parent) (parent.children ??= []).push(agent);
  }
  const root = agents.find((agent) => agent.isRoot) ?? agents.find((agent) => !agent.parentId);
  return {
    id: thread.id,
    source: thread.root.source,
    provider: thread.root.provider,
    backend: thread.root.backend,
    model: thread.root.model,
    reasoningEffort: thread.root.effort,
    project: thread.root.project,
    cwd: thread.root.cwd,
    historyFile: thread.root.historyFile || thread.root.path,
    startedAt: thread.startedAt,
    endedAt: thread.lastActivityAt,
    messages,
    turns: thread.turns.map((turn) => ({
      id: turn.id,
      index: turn.turnIndex,
      startedAt: turn.startedAt,
      endedAt: turn.endedAt,
      stopReason: turn.stopReason,
      model: turn.model,
      backend: turn.backend,
      reasoningEffort: turn.effort,
      status: turn.status,
      error: turn.error,
      usage: {
        inputTokens: turn.inputTokens,
        outputTokens: turn.outputTokens,
        reasoningTokens: turn.reasoningTokens,
        cacheReadTokens: turn.cacheReadTokens,
        cacheWriteTokens: turn.cacheWriteTokens,
        totalTokens: turn.totalTokens,
      },
      cost: { inputCost: turn.costUsd },
      messageIds: Array.from({ length: turn.messageCount }, (_, index) => `${turn.id}:${index}`),
    })),
    root,
    agents,
    usage: {
      inputTokens: thread.inputTokens,
      outputTokens: thread.outputTokens,
      totalTokens: thread.totalTokens,
    },
    cost: { inputCost: thread.costUsd },
    toolCosts: thread.costs.map((cost) => ({
      model: cost.model,
      inputTokens: cost.inputTokens,
      outputTokens: cost.outputTokens,
      reasoningTokens: cost.reasoningTokens,
      cacheReadTokens: cost.cacheReadTokens,
      cacheWriteTokens: cost.cacheWriteTokens,
      totalTokens: cost.totalTokens,
      inputCost: cost.totalCost,
    })),
    health: detail.diagnostics.map((diagnostic) => ({
      kind: diagnostic.code,
      severity: diagnostic.severity,
      message: diagnostic.message,
    })),
    live: {
      active: thread.status === 'working',
      pid: thread.root.pid,
      status: thread.root.processStatus,
    },
    prompt: {
      attempts: detail.attempts,
      diagnostics: detail.diagnostics,
      thread,
    },
  };
}

function mergeTranscriptMessages(snapshot: SessionUIMessage[], entries: Array<SessionEntry | SessionUIMessage>) {
  const messages = new Map(snapshot.map((message) => [message.id, message]));
  for (const entry of entries) {
    if ('parts' in entry) messages.set(entry.id, entry);
  }
  return [...messages.values()];
}

function requireUnifiedSession(session: SessionInput): UnifiedSessionInput {
  if (typeof session === 'string' || Array.isArray(session)) {
    throw new Error('Provider thread did not produce a unified session');
  }
  return session;
}
