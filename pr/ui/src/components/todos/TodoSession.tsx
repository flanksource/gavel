import { useCallback, useMemo } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { SessionInspector, fetchRemoteSession, questionsFromToolInput, type SessionInspectorTab, type SessionMetadataSummary, type SessionPendingTool, type SessionToolDecision, type SessionUIMessage } from '@flanksource/clicky-ui/ai';
import type { SessionStats, TodoItem, TodoRunOptions, TodoSessionAttempt, TodoSessionDetailResponse } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { todoQuery } from './format';
import { CmuxSessionButton } from './TodoSessionTimer';
import { TodoSessionStart } from './TodoSessionStart';
import { SessionErrorDetails, type SessionError } from './SessionErrorDetails';
import { AttemptStopAction, CopyAllDetailsButton, PENDING_LAUNCH_ID, attemptCollection, attemptSessionId, captainSessionUrl, selectAttempt, useTodoSessionDetail } from './TodoSessionDetail';
import type { TodoRunAction } from './run';
import { invalidateTodoCaches, setTodoCaches, todoMutationJSON, useTodoSessionStop } from './todoMutations';
import { sessionStatsQueryOptions, todoQueryKeys } from './todoQueries';
import { setTodoLaunchProgress, todoMutationStream, updateTodoLaunchProgress, useTodoLaunchProgress, type TodoLaunchProgress } from './todoLaunch';

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
  sessionTab,
  onSessionTabChange,
  sessionIds,
  onSessionIdsChange,
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
  // The inspector's tab and selected attempts, owned by the todo detail view
  // (routed through the URL on the dashboard).
  sessionTab?: SessionInspectorTab;
  onSessionTabChange?: (tab: SessionInspectorTab) => void;
  sessionIds?: string[];
  onSessionIdsChange?: (ids: string[]) => void;
}) {
  const queryClient = useQueryClient();
  const inspectorTabProps = {
    ...(sessionTab ? { tab: sessionTab } : {}),
    ...(onSessionTabChange ? { onTabChange: onSessionTabChange } : {}),
  };
  const inspectorSelectionProps = {
    ...(sessionIds ? { selectedSessionIds: sessionIds } : {}),
    ...(onSessionIdsChange ? { onSelectedSessionIdsChange: onSessionIdsChange } : {}),
  };
  const launch = useTodoLaunchProgress(dir, todo.ref);
  const activeLaunch = launch && launch.status !== 'failed' ? launch : null;
  const { detail, error: detailError } = useTodoSessionDetail(dir, todo.ref, active);
  const attempts = useMemo(() => detail?.attempts ?? [], [detail]);
  const launchAttempt = activeLaunch?.promptRunId ? attempts.find(attempt => attempt.promptRunId === activeLaunch.promptRunId) : undefined;
  // A launch in flight owns the view; otherwise the attempt the todo's session
  // id names, else the newest.
  const attempt = activeLaunch ? launchAttempt : selectAttempt(attempts, sessionId);
  const currentId = activeLaunch ? activeLaunch.promptRunId || PENDING_LAUNCH_ID : attempt?.promptRunId;
  const followedSessionId = activeLaunch ? launchAttempt?.providerSessionId : attempt?.providerSessionId || sessionId;
  const { stats, state, error: statusError, inProgress, approvals, approve } = useSessionStatus(dir, followedSessionId, active);
  const collection = useMemo(
    () => currentId ? attemptCollection({ todoRef: todo.ref, attempts, currentId, launch: activeLaunch }) : undefined,
    [activeLaunch, attempts, currentId, todo.ref],
  );
  const attemptSession = attempt ? attemptSessionId(attempt) : undefined;
  // The blocking-question fallback needs the question's own input, which lives
  // in the transcript: read it from Captain only while the session asks.
  const askSrc = state === 'ask' && !inProgress && approvals.length === 0 && attemptSession ? captainSessionUrl(attemptSession) : undefined;
  const askQuery = useQuery({
    queryKey: todoQueryKeys.captainSession(askSrc ?? ''),
    queryFn: ({ signal }) => fetchRemoteSession(askSrc!, signal),
    enabled: active && !!askSrc,
  });
  const answerMutation = useMutation({
    mutationKey: ['todos', 'session', 'answer', { dir: dir.trim(), ref: todo.ref }],
    mutationFn: async (decision: SessionToolDecision) => {
      if (!followedSessionId) throw new Error('Could not resume the agent session: there is no session to answer');
      const data = await todoMutationStream<{ todo?: TodoItem; status?: string; promptRunId?: string; sessionId?: string }>(
        '/api/todos/answer',
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            dir,
            ref: todo.ref,
            sessionId: followedSessionId,
            answer: decision.message || (decision.allow ? formatDecisionAnswers(decision.answers, decision.event.toolInput) : ''),
            ...(decision.allow ? { answers: decision.answers } : {}),
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
        queryClient.invalidateQueries({ queryKey: todoQueryKeys.sessionDetail(dir, todo.ref) }),
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
    const latest = latestQuestionTool(askQuery.data?.messages ?? []);
    if (latest)
      return [{ tool: 'AskUserQuestion', input: latest.input, toolCallId: latest.toolCallId, sessionId: followedSessionId }];
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
  }, [approvals, askQuery.data, followedSessionId, inProgress, state, todo.questions]);

  // decide routes a clicky-ui SessionViewer decision to whichever endpoint
  // owns it. A decision on a live approval (SessionViewer stamps the matched
  // event's approvalId onto it — see clicky-ui's SessionViewer.mergePendingTools)
  // posts approve/deny/respond to /api/todos/session/approve; the reason
  // textarea on "Reject with comment" already flows through as decision.message,
  // so denying with a reason is just this wire threading it through. A decision
  // with no approvalId is the ask-status blocking-question fallback (parsed
  // from the transcript or from todo.questions, neither of which is a live
  // approval) and still resumes the session via /api/todos/answer.
  const decide = useCallback(
    async (decision: SessionToolDecision) => {
      if (!followedSessionId) throw new Error('Session is unavailable');
      const approvalId = decision.event.approvalId;
      if (approvalId) {
        if (!decision.allow) {
          await approve(approvalId, 'deny', decision.message);
          return;
        }
        if (decision.answers) {
          const match = approvals.find(item => item.approvalId === approvalId);
          await approve(approvalId, 'respond', undefined, { ...match?.input, answers: decision.answers });
          return;
        }
        await approve(approvalId, 'approve', decision.message);
        return;
      }
      await answerMutation.mutateAsync(decision);
    },
    [answerMutation.mutateAsync, approvals, approve, followedSessionId]
  );

  const stopAttempt = useCallback(
    (target: TodoSessionAttempt) => stopMutation.mutateAsync(target.promptRunId).then(() => undefined),
    [stopMutation.mutateAsync]
  );

  if (!sessionId && !activeLaunch) {
    return <>{launch && <LaunchProgress launch={launch} />}<TodoSessionStart dir={dir} todo={todo} onRun={onRun} onAdvanced={onAdvanced} runOptions={runOptions} planOptions={planOptions} onRunOptionsChange={onRunOptionsChange} onPlanOptionsChange={onPlanOptionsChange} runBusy={runBusy} runDisabled={runDisabled} /></>;
  }

  const metadata = todoSessionMetadata({ attempt, stats, sessionId: followedSessionId });
  const sessionErrors: SessionError[] = [
    ...(askQuery.error ? [{ source: 'Session transcript', message: askQuery.error.message }] : []),
    ...(statusError ? [{ source: 'Session status', message: statusError }] : []),
    ...(detailError ? [{ source: 'Session detail', message: detailError }] : []),
    ...(answerMutation.error ? [{ source: 'Session answer', message: answerMutation.error.message }] : []),
    ...(stopMutation.error ? [{ source: 'Session stop', message: stopMutation.error.message }] : []),
  ];
  const toolbarActions = <SessionToolbarActions dir={dir} sessionId={followedSessionId} agent={stats?.agent} detail={detail} attempt={attempt} onResume={onResume} resumeDisabled={resumeDisabled} />;
  const transcriptProps = { pendingTools, onPendingToolDecision: decide, showHeader: false, className: 'text-xs' };

  return (
    <div className="@container flex min-h-0 flex-1 flex-col overflow-hidden bg-muted/20">
      {launch && <LaunchProgress launch={launch} />}
      <SessionErrorDetails errors={sessionErrors} />
      {collection ? (
        <SessionInspector
          session={collection}
          className="h-full"
          layout={activeLaunch && !(launchAttempt && attemptSessionId(launchAttempt)) ? 'default' : 'compact'}
          metadata={metadata}
          toolbarActions={toolbarActions}
          transcriptProps={transcriptProps}
          renderSessionActions={item => {
            const target = attempts.find(candidate => candidate.promptRunId === item.id);
            return target ? <AttemptStopAction attempt={target} onStop={stopAttempt} /> : null;
          }}
          {...inspectorTabProps}
          {...inspectorSelectionProps}
        />
      ) : detail && sessionId ? (
        // A session no attempt owns (recorded before attempts were linked) is
        // still Captain's to read by its provider id.
        <SessionInspector
          src={captainSessionUrl(sessionId)}
          className="min-h-0 flex-1 text-xs"
          layout="compact"
          metadata={metadata}
          toolbarActions={toolbarActions}
          transcriptProps={transcriptProps}
          {...inspectorTabProps}
        />
      ) : !detailError ? (
        <p className="flex items-center gap-2 px-3 py-4 text-xs text-muted-foreground">
          <Spinner /> Loading attempts…
        </p>
      ) : null}
    </div>
  );
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

function todoSessionMetadata({ attempt, stats, sessionId }: { attempt: TodoSessionAttempt | undefined; stats: SessionStats | undefined; sessionId: string | undefined }): SessionMetadataSummary {
  const provider = attempt?.provider;
  const executionMode = attempt?.runtimeMode;
  const model = stats?.model || attempt?.model;
  const reasoningEffort = stats?.effort || attempt?.effort;
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

function SessionToolbarActions({ dir, sessionId, agent, detail, attempt, onResume, resumeDisabled }: { dir: string; sessionId: string | undefined; agent: string | undefined; detail: TodoSessionDetailResponse | null; attempt: TodoSessionAttempt | undefined; onResume: (() => void) | undefined; resumeDisabled: boolean | undefined }) {
  return (
    <>
      {sessionId ? <CmuxSessionButton dir={dir} sessionId={sessionId} {...(agent ? { agent } : {})} {...(onResume ? { onResume } : {})} {...(resumeDisabled !== undefined ? { resumeDisabled } : {})} /> : null}
      {detail ? <CopyAllDetailsButton detail={detail} attempt={attempt} compact /> : null}
    </>
  );
}

function latestQuestionTool(messages: SessionUIMessage[]): { input?: Record<string, unknown>; toolCallId?: string } | undefined {
  for (let index = messages.length - 1; index >= 0; index--) {
    const parts = messages[index]!.parts;
    for (let partIndex = parts.length - 1; partIndex >= 0; partIndex--) {
      const part = parts[partIndex]!;
      if (part.toolName !== 'AskUserQuestion') continue;
      const input = typeof part.input === 'object' && part.input !== null && !Array.isArray(part.input) ? (part.input as Record<string, unknown>) : undefined;
      return { input, toolCallId: part.toolCallId };
    }
  }
  return undefined;
}

function formatDecisionAnswers(answers?: Record<string, string | string[]>, input?: Record<string, unknown>): string {
  if (!answers) return '';
  const questions = new Map(questionsFromToolInput(input).map(question => [question.id, question.text]));
  return Object.entries(answers)
    .map(([id, answer]) => `${questions.get(id) ?? id}: ${Array.isArray(answer) ? answer.join(', ') : answer}`)
    .join('\n');
}
