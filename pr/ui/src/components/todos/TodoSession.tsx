import { useCallback, useEffect, useMemo, useRef, useState, type ComponentType } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { SessionViewer, type SessionEntry, type SessionPendingTool, type SessionToolDecision, type SessionUIMessage } from '@flanksource/clicky-ui/ai';
import { UiCancel, UiCircleFilled, UiComment, UiError, UiLightbulb, UiPass, UiShield, type IconProps } from '@flanksource/clicky-ui/icons';
import type { SessionStats, TodoItem, TodoRunOptions, TodoSessionAttempt } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { todoQuery } from './format';
import { TodoSessionTimer } from './TodoSessionTimer';
import { TodoSessionStart } from './TodoSessionStart';
import { SessionErrorDetails, type SessionError } from './SessionErrorDetails';
import { CopyAllDetailsButton, SessionDiagnostics, ThreadInspector, useTodoSessionDetail } from './TodoSessionDetail';
import type { TodoRunAction } from './run';
import { invalidateTodoCollections, setTodoCaches, todoMutationJSON } from './todoMutations';
import { sessionStatsQueryOptions, todoQueryKeys } from './todoQueries';

interface SessionStateView {
  label: string;
  icon: ComponentType<IconProps>;
  className: string;
}

// STATE_VIEWS maps the server-derived session state (cmux.SessionStats.State) to
// the header badge's label, icon and colour. Tints are semi-transparent so they
// read on both the light and dark dashboard themes. The empty key is the initial
// "no event yet" state.
const STATE_VIEWS: Record<string, SessionStateView> = {
  '': {
    label: 'Waiting',
    icon: Spinner,
    className: 'text-muted-foreground bg-muted/50 border-border',
  },
  thinking: {
    label: 'Thinking',
    icon: UiLightbulb,
    className: 'text-amber-600 bg-amber-500/15 border-amber-500/30',
  },
  working: {
    label: 'Working',
    icon: Spinner,
    className: 'text-cyan-600 bg-cyan-500/15 border-cyan-500/30',
  },
  ask: {
    label: 'Awaiting input',
    icon: UiComment,
    className: 'text-purple-600 bg-purple-500/15 border-purple-500/30',
  },
  approval: {
    label: 'Needs approval',
    icon: UiShield,
    className: 'text-amber-600 bg-amber-500/15 border-amber-500/30',
  },
  completed: {
    label: 'Completed',
    icon: UiPass,
    className: 'text-emerald-600 bg-emerald-500/15 border-emerald-500/30',
  },
  idle: {
    label: 'Idle',
    icon: UiCircleFilled,
    className: 'text-muted-foreground bg-muted/50 border-border',
  },
  failed: {
    label: 'Failed',
    icon: UiError,
    className: 'text-red-600 bg-red-500/15 border-red-500/30',
  },
  cancelled: {
    label: 'Cancelled',
    icon: UiCancel,
    className: 'text-muted-foreground bg-muted/50 border-border',
  },
  interrupted: {
    label: 'Interrupted',
    icon: UiError,
    className: 'text-amber-600 bg-amber-500/15 border-amber-500/30',
  },
  error: {
    label: 'Error',
    icon: UiError,
    className: 'text-red-600 bg-red-500/15 border-red-500/30',
  },
};

function stateView(state: string, error: string): SessionStateView {
  if (error) return STATE_VIEWS.error;
  return STATE_VIEWS[state] ?? STATE_VIEWS[''];
}

// useTodoSession follows a TODO's agent session log over SSE. The server tails
// the on-disk Claude session log and emits each conversational line as a raw
// captain SessionEntry, which we accumulate and hand to clicky-ui's
// SessionViewer to render. The stream replays existing history on connect, then
// follows new entries until unmounted.
export function useTodoSession(dir: string, sessionId: string | undefined, active: boolean) {
  const [entries, setEntries] = useState<Array<SessionEntry | SessionUIMessage>>([]);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    setEntries([]);
    setError('');
    setConnected(false);
    if (!active || !sessionId) return;

    const params = new URLSearchParams(todoQuery(dir));
    params.set('sessionId', sessionId);
    const es = new EventSource(`/api/todos/session/stream?${params.toString()}`);

    es.onopen = () => setConnected(true);
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
      setConnected(false);
    });

    return () => es.close();
  }, [dir, sessionId, active]);

  return { entries, connected, error };
}

// useSessionStatus polls the session stats endpoint for the high-level agent
// state and any pending tool-permission request, and exposes a resolver that
// POSTs the user's Allow/Deny. State is server-derived (the same source the
// session timer uses), so the header badge and the approval banner stay in sync
// without re-deriving anything from the event stream.
export function useSessionStatus(dir: string, sessionId: string | undefined, active: boolean) {
  const queryClient = useQueryClient();
  const enabled = active && !!sessionId;
  const options = sessionStatsQueryOptions({ dir, sessionId: sessionId ?? '' });
  const query = useQuery({ ...options, enabled });
  const stats = enabled ? query.data : undefined;
  const approveMutation = useMutation({
    mutationKey: ['todos', 'session', 'approve', { sessionId: sessionId ?? '' }],
    mutationFn: ({ allow, message, updatedInput }: {
      allow: boolean;
      message?: string;
      updatedInput?: Record<string, unknown>;
    }) => todoMutationJSON<{ resolved: boolean; allow: boolean }>(
      '/api/todos/session/approve',
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ sessionId, allow, message, updatedInput }),
      },
      'Session approval update failed',
    ),
    onSuccess: () => {
      queryClient.setQueryData<SessionStats>(options.queryKey, previous => previous
        ? { ...previous, approval: undefined }
        : previous);
    },
  });

  const approve = useCallback(
    async (allow: boolean, message?: string, updatedInput?: Record<string, unknown>) => {
      if (!sessionId) throw new Error('Session is unavailable');
      await approveMutation.mutateAsync({ allow, message, updatedInput });
    },
    [approveMutation.mutateAsync, sessionId]
  );

  return {
    state: stats?.state ?? '',
    error: approveMutation.error?.message || query.error?.message || stats?.error || '',
    inProgress: stats?.inProgress ?? false,
    approval: stats?.approval ?? null,
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
  // onResume/resumeDisabled back the session timer's "Resume in cmux" action,
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
  const { detail, error: detailError } = useTodoSessionDetail(dir, todo.ref, sessionId, active);
  const followedSessionId = detail?.thread?.providerSessionId || sessionId;
  const { entries, connected, error } = useTodoSession(dir, followedSessionId, active);
  const { state, error: statusError, inProgress, approval, approve } = useSessionStatus(dir, followedSessionId, active);
  const scrollRef = useRef<HTMLDivElement | null>(null);
  // Host element for the SessionViewer's 3-dot menu: the viewer portals its menu
  // into this header slot so the filter/density controls sit in the same fixed
  // header as the status badge and session timer instead of scrolling with the log.
  const [menuHost, setMenuHost] = useState<HTMLElement | null>(null);
  // Follow the tail like a terminal, but stop following once the user scrolls up
  // to read earlier history (re-engages when they scroll back to the bottom).
  const followRef = useRef(true);
  const answerMutation = useMutation({
    mutationKey: ['todos', 'session', 'answer', { dir: dir.trim(), ref: todo.ref }],
    mutationFn: async (decision: SessionToolDecision) => {
      const data = await todoMutationJSON<{ todo?: TodoItem; status?: string }>(
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
      );
      if (!data.todo?.ref) throw new Error('Could not resume the agent session: response did not include the updated todo');
      return data.todo;
    },
    onSuccess: async (updated) => {
      await setTodoCaches(queryClient, dir, updated);
      onChanged?.(updated);
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: todoQueryKeys.sessionStats(dir, followedSessionId ?? '') }),
        queryClient.invalidateQueries({ queryKey: todoQueryKeys.sessionDetail(dir, todo.ref, sessionId, false) }),
      ]);
    },
  });
  const stopMutation = useMutation({
    mutationKey: ['todos', 'session', 'stop', { dir: dir.trim(), ref: todo.ref }],
    mutationFn: (attempt: TodoSessionAttempt) => todoMutationJSON<{ status: string; promptRunId: string }>(
      `/api/todos/session/stop?${todoQuery(dir)}`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ref: todo.ref, promptRunId: attempt.promptRunId }),
      },
      'Could not stop the attempt',
    ),
    onSuccess: async () => {
      await Promise.all([
        invalidateTodoCollections(queryClient, dir),
        queryClient.invalidateQueries({ queryKey: todoQueryKeys.sessionStats(dir, followedSessionId ?? '') }),
        queryClient.invalidateQueries({ queryKey: todoQueryKeys.sessionDetail(dir, todo.ref, sessionId, false) }),
        queryClient.invalidateQueries({ queryKey: todoQueryKeys.sessionDetail(dir, todo.ref, undefined, true) }),
      ]);
    },
  });

  function onScroll() {
    const el = scrollRef.current;
    if (!el) return;
    followRef.current = el.scrollHeight - el.scrollTop - el.clientHeight < 24;
  }

  useEffect(() => {
    const el = scrollRef.current;
    if (el && followRef.current) el.scrollTop = el.scrollHeight;
  }, [entries.length]);

  const pendingTools = useMemo<SessionPendingTool[]>(() => {
    if (approval)
      return [
        {
          tool: approval.tool,
          input: approval.input,
          toolCallId: approval.toolUseId,
          sessionId: approval.sessionId,
        },
      ];
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
  }, [approval, entries, followedSessionId, inProgress, state, todo.questions]);

  const decide = useCallback(
    async (decision: SessionToolDecision) => {
      if (!followedSessionId) throw new Error('Session is unavailable');
      if (approval) {
        const updatedInput = decision.answers ? { ...approval.input, answers: decision.answers } : undefined;
        await approve(decision.allow, decision.message, updatedInput);
        return;
      }
      await answerMutation.mutateAsync(decision);
    },
    [answerMutation.mutateAsync, approval, approve, followedSessionId]
  );

  const stopAttempt = useCallback(
    (attempt: TodoSessionAttempt) => stopMutation.mutateAsync(attempt).then(() => undefined),
    [stopMutation.mutateAsync]
  );

  if (!sessionId) {
    return <TodoSessionStart dir={dir} todo={todo} onRun={onRun} onAdvanced={onAdvanced} runOptions={runOptions} planOptions={planOptions} onRunOptionsChange={onRunOptionsChange} onPlanOptionsChange={onPlanOptionsChange} runBusy={runBusy} runDisabled={runDisabled} />;
  }

  const threadState = detail?.thread?.status || state;
  const view = stateView(threadState, error || statusError || detailError);
  const sessionErrors: SessionError[] = [
    ...(error ? [{ source: 'Session stream', message: error }] : []),
    ...(statusError ? [{ source: 'Session status', message: statusError }] : []),
    ...(detailError ? [{ source: 'Session detail', message: detailError }] : []),
    ...(answerMutation.error ? [{ source: 'Session answer', message: answerMutation.error.message }] : []),
    ...(stopMutation.error ? [{ source: 'Session stop', message: stopMutation.error.message }] : []),
  ];

  return (
    <div className="m-3 flex min-h-0 flex-1 flex-col overflow-hidden rounded-md border border-border bg-card">
      <div className="flex shrink-0 flex-wrap items-center gap-x-3 gap-y-1 border-b border-border bg-muted/40 px-3 py-1.5 text-[11px] text-muted-foreground">
        <SessionStateBadge view={view} />
        <span className="inline-flex items-center gap-1">
          <UiCircleFilled className={`text-[7px] ${connected ? 'text-emerald-500' : 'text-muted-foreground'}`} />
          {connected ? 'Following session' : 'Session idle'}
        </span>
        <span className="font-mono">{(detail?.thread?.providerSessionId || sessionId).slice(0, 8)}</span>
        <TodoSessionTimer dir={dir} sessionId={followedSessionId} active={active} onResume={onResume} resumeDisabled={resumeDisabled} />
        {detail && <CopyAllDetailsButton detail={detail} entries={entries} />}
        <span ref={setMenuHost} className="ml-auto inline-flex items-center" />
      </div>
      {detail && <SessionDiagnostics diagnostics={detail.diagnostics} />}
      <SessionErrorDetails errors={sessionErrors} />
      <div ref={scrollRef} onScroll={onScroll} className={`min-h-0 flex-1 ${detail?.thread ? 'overflow-hidden' : 'overflow-y-auto px-3 py-2'}`}>
        {entries.length === 0 && sessionErrors.length === 0 && <div className="text-xs text-muted-foreground">Waiting for session activity…</div>}
        {detail?.thread ? (
          <ThreadInspector detail={detail} entries={entries} dir={dir} todoRef={todo.ref} onStop={stopAttempt} pendingTools={pendingTools} onPendingToolDecision={decide} />
        ) : (
          entries.length > 0 && <SessionViewer session={entries as SessionEntry[] | SessionUIMessage[]} pendingTools={pendingTools} onPendingToolDecision={decide} showHeader={false} menuContainer={menuHost} className="text-xs" />
        )}
        {!detail?.thread && entries.length === 0 && pendingTools.length > 0 && <SessionViewer session={[]} pendingTools={pendingTools} onPendingToolDecision={decide} showHeader={false} menuContainer={menuHost} className="text-xs" />}
      </div>
    </div>
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

function SessionStateBadge({ view }: { view: SessionStateView }) {
  const Icon = view.icon;
  return (
    <span className={`inline-flex items-center gap-1 rounded border px-1.5 py-0.5 text-[10px] font-medium uppercase ${view.className}`}>
      <Icon className="text-[11px]" />
      {view.label}
    </span>
  );
}
