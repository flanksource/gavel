import { useCallback, useEffect, useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { SessionInspector, type SessionCollectionInput, type SessionEntry, type SessionMetadataSummary, type SessionPendingTool, type SessionToolDecision, type SessionUIMessage } from '@flanksource/clicky-ui/ai';
import type { SessionStats, TodoItem, TodoRunOptions, TodoSessionAttempt, TodoSessionDetailResponse } from '../../types';
import { todoQuery } from './format';
import { CmuxSessionButton } from './TodoSessionTimer';
import { TodoSessionStart } from './TodoSessionStart';
import { SessionErrorDetails, type SessionError } from './SessionErrorDetails';
import { CopyAllDetailsButton, SessionDiagnostics, ThreadInspector, useTodoSessionDetail } from './TodoSessionDetail';
import type { TodoRunAction } from './run';
import { invalidateTodoCaches, setTodoCaches, todoMutationJSON, useTodoSessionStop } from './todoMutations';
import { sessionStatsQueryOptions, todoQueryKeys } from './todoQueries';
import { setTodoLaunchProgress, todoMutationStream, updateTodoLaunchProgress, useTodoLaunchProgress, type TodoLaunchProgress } from './todoLaunch';

// useTodoSession follows a TODO's agent session log over SSE. The server tails
// the on-disk Claude session log and emits each conversational line as a raw
// captain SessionEntry, which we accumulate and hand to clicky-ui's
// SessionInspector to render. The stream replays existing history on connect, then
// follows new entries until unmounted.
export function useTodoSession(dir: string, sessionId: string | undefined, active: boolean) {
  const [entries, setEntries] = useState<Array<SessionEntry | SessionUIMessage>>([]);
  const [error, setError] = useState('');

  useEffect(() => {
    setEntries([]);
    setError('');
    if (!active || !sessionId) return;

    const params = new URLSearchParams(todoQuery(dir));
    params.set('sessionId', sessionId);
    const es = new EventSource(`/api/todos/session/stream?${params.toString()}`);

    es.addEventListener('entry', (e: MessageEvent) => {
      try {
        const entry = JSON.parse(e.data) as SessionEntry | SessionUIMessage;
        setEntries((prev) => mergeSessionEntry(prev, entry));
      } catch {
        // Ignore malformed frames; the next well-formed entry recovers.
      }
    });
    es.addEventListener('error', (e: MessageEvent) => {
      // A named error frame carries data; a bare connection drop does not.
      if (e.data) {
        try {
          setError(JSON.parse(e.data).error || 'Session stream error');
        } catch {
          setError(`Session stream error\n${String(e.data)}`);
        }
      } else {
        setError((previous) => previous || 'Session stream connection failed without returning error details');
      }
    });

    return () => es.close();
  }, [dir, sessionId, active]);

  return { entries, error };
}

// TodoSessionApprovalAction is the decision the server's approval endpoint
// accepts for one pending approval: approve/deny answer a tool-permission
// request (deny optionally carrying a rejection reason), respond answers a
// question-shaped approval with the user's structured input.
export type TodoSessionApprovalAction = 'approve' | 'deny' | 'respond';

// useSessionStatus polls the session stats endpoint for the high-level agent
// state and any pending tool-permission requests, and exposes a resolver that
// POSTs the user's approve/deny/respond decision. State is server-derived (the
// same source the session timer uses), so the header badge and the approval
// banner stay in sync without re-deriving anything from the event stream.
export function useSessionStatus(dir: string, sessionId: string | undefined, active: boolean) {
  const queryClient = useQueryClient();
  const enabled = active && !!sessionId;
  const options = sessionStatsQueryOptions({ dir, sessionId: sessionId ?? '' });
  const query = useQuery({ ...options, enabled });
  const stats = enabled ? query.data : undefined;
  const approveMutation = useMutation({
    mutationKey: ['todos', 'session', 'approve', { sessionId: sessionId ?? '' }],
    mutationFn: ({ approvalId, action, message, input }: {
      approvalId: string;
      action: TodoSessionApprovalAction;
      message?: string;
      input?: Record<string, unknown>;
    }) => todoMutationJSON<{ resolved: boolean }>(
      `/api/todos/session/approve?${todoQuery(dir)}`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ approvalId, action, message, input }),
      },
      'Session approval update failed',
    ),
    onSuccess: (_data, { approvalId }) => {
      queryClient.setQueryData<SessionStats>(options.queryKey, previous => previous
        ? { ...previous, approvals: (previous.approvals ?? []).filter(item => item.approvalId !== approvalId) }
        : previous);
    },
  });

  const approve = useCallback(
    async (approvalId: string, action: TodoSessionApprovalAction, message?: string, input?: Record<string, unknown>) => {
      if (!sessionId) throw new Error('Session is unavailable');
      await approveMutation.mutateAsync({ approvalId, action, message, input });
    },
    [approveMutation.mutateAsync, sessionId]
  );

  return {
    stats,
    state: stats?.state ?? '',
    error: approveMutation.error?.message || query.error?.message || stats?.error || '',
    inProgress: stats?.inProgress ?? false,
    approvals: stats?.approvals ?? [],
    approve,
    busy: approveMutation.isPending,
  };
}

export function TodoSession({
  dir,
  sessionId,
  active,
  todo,
  onChanged,
  onResume,
  resumeDisabled,
  onRun,
  onAdvanced,
  runOptions,
  planOptions,
  onRunOptionsChange,
  onPlanOptionsChange,
  runBusy,
  runDisabled,
}: {
  dir: string;
  sessionId?: string;
  active: boolean;
  todo: TodoItem;
  onChanged?: (todo: TodoItem) => void;
  // onResume/resumeDisabled back the session toolbar's "Resume in cmux" action,
  // wired from the todo detail's run flow.
  onResume?: () => void;
  resumeDisabled?: boolean;
  // onRun/onAdvanced/runBusy/runDisabled back the never-run start hero
  // (TodoSessionStart), wired from the todo detail's run flow.
  onRun?: (options?: TodoRunOptions) => void;
  onAdvanced?: (action: TodoRunAction) => void;
  runOptions?: TodoRunOptions;
  planOptions?: TodoRunOptions;
  onRunOptionsChange?: (options: TodoRunOptions) => void;
  onPlanOptionsChange?: (options: TodoRunOptions) => void;
  runBusy?: boolean;
  runDisabled?: boolean;
}) {
  const queryClient = useQueryClient();
  const launch = useTodoLaunchProgress(dir, todo.ref);
  const { detail, error: detailError } = useTodoSessionDetail(dir, todo.ref, sessionId, active);
  const launchAttempt = detail?.attempts.find(attempt => attempt.promptRunId === launch?.promptRunId);
  const followedSessionId = launch && launch.status !== 'failed'
    ? launchAttempt?.providerSessionId || (detail?.selectedPromptRunId === launch.promptRunId ? detail?.thread?.providerSessionId : undefined)
    : detail?.thread?.providerSessionId || sessionId;
  const { entries, error } = useTodoSession(dir, followedSessionId, active);
  const { stats, state, error: statusError, inProgress, approvals, approve } = useSessionStatus(dir, followedSessionId, active);
  const answerMutation = useMutation({
    mutationKey: ['todos', 'session', 'answer', { dir: dir.trim(), ref: todo.ref }],
    mutationFn: async (decision: SessionToolDecision) => {
      const data = await todoMutationStream<{ todo?: TodoItem; status?: string; promptRunId?: string; sessionId?: string }>(
        '/api/todos/answer',
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            dir,
            ref: todo.ref,
            answer: decision.message || formatDecisionAnswers(decision.answers),
            answers: decision.answers,
            rejected: !decision.allow,
          }),
        },
        'Could not resume the agent session',
        resolved => updateTodoLaunchProgress(queryClient, dir, todo.ref, previous => ({ ...previous, status: 'resolved', step: resolved.step, spec: resolved.spec, specYaml: resolved.specYaml })),
      );
      if (!data.todo?.ref) throw new Error('Could not resume the agent session: response did not include the updated todo');
      if (!data.promptRunId) throw new Error('Could not resume the agent session: response did not include the prompt run ID');
      return data;
    },
    onMutate: () => setTodoLaunchProgress(queryClient, dir, todo.ref, { status: 'preparing', step: 'resume' }),
    onSuccess: async ({ todo: updated, promptRunId, sessionId }) => {
      if (!updated) throw new Error('Could not resume the agent session: response did not include the updated todo');
      updateTodoLaunchProgress(queryClient, dir, todo.ref, previous => ({ ...previous, status: 'admitted', promptRunId, sessionId }));
      await setTodoCaches(queryClient, dir, updated);
      onChanged?.(updated);
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: todoQueryKeys.sessionStats(dir, followedSessionId ?? '') }),
        queryClient.invalidateQueries({ queryKey: todoQueryKeys.sessionDetail(dir, todo.ref, sessionId, false) }),
      ]);
    },
    onError: async error => {
      updateTodoLaunchProgress(queryClient, dir, todo.ref, previous => ({ ...previous, status: 'failed', error: error.message }));
      await invalidateTodoCaches(queryClient, dir, todo.ref);
    },
  });
  const stopMutation = useTodoSessionStop(dir, todo.ref, followedSessionId ?? sessionId);

  const pendingTools = useMemo<SessionPendingTool[]>(() => {
    if (approvals.length > 0)
      return approvals.map(item => ({
        tool: item.tool,
        input: item.input,
        toolCallId: item.toolUseId,
        approvalId: item.approvalId,
        sessionId: item.sessionId,
      }));
    if (state !== 'ask' || inProgress) return [];
    const latest = latestQuestionTool(entries);
    if (latest)
      return [
        {
          tool: 'AskUserQuestion',
          input: latest.input,
          toolCallId: latest.toolCallId,
          sessionId: followedSessionId,
        },
      ];
    if (todo.questions?.length) {
      return [
        {
          tool: 'AskUserQuestion',
          sessionId: followedSessionId,
          input: {
            questions: todo.questions.map((question, index) => ({
              id: String(index + 1),
              question: question.text,
              header: question.context,
              options: (question.options ?? []).map((option) => ({
                label: option,
                value: option,
              })),
            })),
          },
        },
      ];
    }
    return [];
  }, [approvals, entries, followedSessionId, inProgress, state, todo.questions]);

  // decide routes a clicky-ui SessionViewer decision to whichever endpoint
  // owns it. A decision on a live approval (SessionViewer stamps the matched
  // event's approvalId onto it — see clicky-ui's SessionViewer.mergePendingTools)
  // posts approve/deny/respond to /api/todos/session/approve; the reason
  // textarea on "Reject with comment" already flows through as decision.message,
  // so denying with a reason is just this wire threading it through. A decision
  // with no approvalId is the ask-status blocking-question fallback (parsed
  // from the log or from todo.questions, neither of which is a live approval)
  // and still resumes the session via /api/todos/answer.
  const decide = useCallback(
    async (decision: SessionToolDecision) => {
      if (!followedSessionId) throw new Error('Session is unavailable');
      const approvalId = decision.event.approvalId;
      if (approvalId) {
        if (decision.answers) {
          const match = approvals.find(item => item.approvalId === approvalId);
          await approve(approvalId, 'respond', undefined, { ...match?.input, answers: decision.answers });
          return;
        }
        await approve(approvalId, decision.allow ? 'approve' : 'deny', decision.message);
        return;
      }
      await answerMutation.mutateAsync(decision);
    },
    [answerMutation.mutateAsync, approvals, approve, followedSessionId]
  );

  const stopAttempt = useCallback(
    (attempt: TodoSessionAttempt) => stopMutation.mutateAsync(attempt.promptRunId).then(() => undefined),
    [stopMutation.mutateAsync]
  );

  if (!sessionId && (!launch || launch.status === 'failed')) {
    return <>{launch && <LaunchProgress launch={launch} />}<TodoSessionStart dir={dir} todo={todo} onRun={onRun} onAdvanced={onAdvanced} runOptions={runOptions} planOptions={planOptions} onRunOptionsChange={onRunOptionsChange} onPlanOptionsChange={onPlanOptionsChange} runBusy={runBusy} runDisabled={runDisabled} /></>;
  }

  const metadata = todoSessionMetadata({ detail, stats, sessionId: followedSessionId });
  const sessionErrors: SessionError[] = [
    ...(error ? [{ source: 'Session stream', message: error }] : []),
    ...(statusError ? [{ source: 'Session status', message: statusError }] : []),
    ...(detailError ? [{ source: 'Session detail', message: detailError }] : []),
    ...(answerMutation.error ? [{ source: 'Session answer', message: answerMutation.error.message }] : []),
    ...(stopMutation.error ? [{ source: 'Session stop', message: stopMutation.error.message }] : []),
  ];

  return (
    <div className="@container flex min-h-0 flex-1 flex-col overflow-hidden bg-muted/20">
      {launch && <LaunchProgress launch={launch} />}
      {detail && <SessionDiagnostics diagnostics={detail.diagnostics} />}
      <SessionErrorDetails errors={sessionErrors} />
      {detail && (detail.thread || (launch && launch.status !== 'failed')) ? (
        <ThreadInspector
          detail={detail}
          entries={entries}
          dir={dir}
          todoRef={todo.ref}
          onStop={stopAttempt}
          pendingTools={pendingTools}
          onPendingToolDecision={decide}
          layout={launch && launch.status !== 'failed' && (!detail.thread || detail.selectedPromptRunId !== launch.promptRunId) ? 'default' : 'compact'}
          metadata={metadata}
          launch={launch}
          toolbarActions={<SessionToolbarActions dir={dir} sessionId={followedSessionId} agent={stats?.agent} detail={detail} entries={entries} onResume={onResume} resumeDisabled={resumeDisabled} />}
        />
      ) : launch && launch.status !== 'failed' ? (
        <SessionInspector
          session={launchCollection(todo.ref, launch)}
          className="min-h-0 flex-1 text-xs"
        />
      ) : (
        <SessionInspector
          session={entries as SessionEntry[] | SessionUIMessage[]}
          className="min-h-0 flex-1 text-xs"
          layout="compact"
          metadata={metadata}
          toolbarActions={<SessionToolbarActions dir={dir} sessionId={followedSessionId} agent={stats?.agent} detail={detail} entries={entries} onResume={onResume} resumeDisabled={resumeDisabled} />}
          transcriptProps={{ pendingTools, onPendingToolDecision: decide, showHeader: false, className: 'text-xs' }}
        />
      )}
    </div>
  );
}

function launchCollection(ref: string, launch: TodoLaunchProgress): SessionCollectionInput {
  const id = launch.promptRunId || 'pending-launch';
  return {
    kind: 'session-collection',
    id: `todo-attempts:${ref}`,
    currentSessionId: id,
    sessions: [{
      id,
      label: 'Attempt #1',
      mode: launch.step,
      status: launch.status === 'admitted' ? 'running' : 'starting',
      session: { id, messages: [] },
    }],
  };
}

function LaunchProgress({ launch }: { launch: TodoLaunchProgress }) {
  const title = launch.status === 'preparing' ? 'Preparing session'
    : launch.status === 'resolved' ? 'Starting session'
      : launch.status === 'failed' ? 'Session could not start' : 'Session started';
  return (
    <section aria-label="Launch progress" className="shrink-0 border-b bg-background px-3 py-2 text-xs">
      <div className="font-medium">{title} · {launch.step}</div>
      {launch.error && <div role="alert" className="mt-1 text-red-600">{launch.error}</div>}
      {(launch.specYaml || launch.requestedSpec) && (
        <details className="mt-1" open={launch.status === 'preparing' || launch.status === 'resolved'}>
          <summary>{launch.specYaml ? 'Resolved spec' : 'Requested spec'}</summary>
          <pre className="mt-1 max-h-48 overflow-auto whitespace-pre-wrap break-words">{launch.specYaml || JSON.stringify(launch.requestedSpec, null, 2)}</pre>
        </details>
      )}
    </section>
  );
}

function todoSessionMetadata({ detail, stats, sessionId }: { detail: TodoSessionDetailResponse | null; stats: SessionStats | undefined; sessionId: string | undefined }): SessionMetadataSummary {
  const attempt = detail?.attempts.find(candidate => candidate.promptRunId === detail.selectedPromptRunId || candidate.executionSessionId === detail.selectedExecutionSessionId);
  const root = detail?.thread?.root;
  const provider = root?.modelProvider || attempt?.provider || root?.provider;
  const executionMode = root?.modelMode || attempt?.runtimeMode;
  const model = stats?.model || root?.model || attempt?.model;
  const reasoningEffort = stats?.effort || root?.effort || attempt?.effort;
  const contextPercent = stats?.contextWindow ? Math.min(100, Math.max(0, stats.contextTokens / stats.contextWindow * 100)) : undefined;
  return {
    ...(sessionId ? { sessionId } : {}),
    ...(provider ? { provider } : {}),
    ...(executionMode ? { executionMode } : {}),
    ...(model ? { model } : {}),
    ...(reasoningEffort ? { reasoningEffort } : {}),
    ...(contextPercent !== undefined ? { context: { usedTokens: stats?.contextTokens, windowTokens: stats?.contextWindow, freePercent: 100 - contextPercent } } : {}),
    ...(stats ? { usage: { inputTokens: stats.inputTokens, outputTokens: stats.outputTokens, cacheReadTokens: stats.cacheReadTokens, cacheWriteTokens: stats.cacheCreationTokens, totalTokens: stats.totalTokens } } : {}),
    ...(stats?.costUsd ? { cost: { inputCost: stats.costUsd } } : {}),
  };
}

function SessionToolbarActions({ dir, sessionId, agent, detail, entries, onResume, resumeDisabled }: { dir: string; sessionId: string | undefined; agent: string | undefined; detail: TodoSessionDetailResponse | null; entries: Array<SessionEntry | SessionUIMessage>; onResume: (() => void) | undefined; resumeDisabled: boolean | undefined }) {
  return (
    <>
      {sessionId ? <CmuxSessionButton dir={dir} sessionId={sessionId} {...(agent ? { agent } : {})} {...(onResume ? { onResume } : {})} {...(resumeDisabled !== undefined ? { resumeDisabled } : {})} /> : null}
      {detail ? <CopyAllDetailsButton detail={detail} entries={entries} compact /> : null}
    </>
  );
}

function mergeSessionEntry(entries: Array<SessionEntry | SessionUIMessage>, entry: SessionEntry | SessionUIMessage): Array<SessionEntry | SessionUIMessage> {
  if (!('parts' in entry) || !entry.id) return [...entries, entry];
  const index = entries.findIndex((existing) => 'parts' in existing && existing.id === entry.id);
  if (index < 0) return [...entries, entry];
  const next = [...entries];
  next[index] = entry;
  return next;
}

function latestQuestionTool(entries: Array<SessionEntry | SessionUIMessage>): { input?: Record<string, unknown>; toolCallId?: string } | undefined {
  for (let index = entries.length - 1; index >= 0; index--) {
    const entry = entries[index];
    if ('parts' in entry) {
      for (let partIndex = entry.parts.length - 1; partIndex >= 0; partIndex--) {
        const part = entry.parts[partIndex];
        if (part.toolName === 'AskUserQuestion') {
          const input = typeof part.input === 'object' && part.input !== null && !Array.isArray(part.input) ? (part.input as Record<string, unknown>) : undefined;
          return { input, toolCallId: part.toolCallId };
        }
      }
      continue;
    }
    if (entry.tool_use?.tool === 'AskUserQuestion')
      return {
        input: entry.tool_use.input,
        toolCallId: entry.tool_use.tool_use_id,
      };
    const blocks = entry.message?.content ?? [];
    for (let blockIndex = blocks.length - 1; blockIndex >= 0; blockIndex--) {
      const block = blocks[blockIndex];
      if (block.name === 'AskUserQuestion') return { input: block.input, toolCallId: block.id };
    }
  }
  return undefined;
}

function formatDecisionAnswers(answers?: Record<string, string | string[]>): string {
  if (!answers) return '';
  return Object.entries(answers)
    .map(([question, answer]) => `${question}: ${Array.isArray(answer) ? answer.join(', ') : answer}`)
    .join('\n');
}
