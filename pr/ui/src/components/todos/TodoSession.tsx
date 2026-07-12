import { useCallback, useEffect, useMemo, useRef, useState, type ComponentType } from 'react';
import { SessionViewer, questionsFromToolInput, type SessionEntry, type SessionPendingTool, type SessionQuestion, type SessionToolDecision } from '@flanksource/clicky-ui/ai';
import { Button } from '@flanksource/clicky-ui/components';
import { UiCancel, UiCheck, UiCircleFilled, UiComment, UiError, UiLightbulb, UiPass, UiShield, type IconProps } from '@flanksource/clicky-ui/icons';
import type { SessionStats, TodoItem, TodoSessionApproval } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { inputClass, todoQuery } from './format';
import { TodoSessionTimer } from './TodoSessionTimer';

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
  '': { label: 'Waiting', icon: Spinner, className: 'text-muted-foreground bg-muted/50 border-border' },
  thinking: { label: 'Thinking', icon: UiLightbulb, className: 'text-amber-600 bg-amber-500/15 border-amber-500/30' },
  working: { label: 'Working', icon: Spinner, className: 'text-cyan-600 bg-cyan-500/15 border-cyan-500/30' },
  ask: { label: 'Awaiting input', icon: UiComment, className: 'text-purple-600 bg-purple-500/15 border-purple-500/30' },
  approval: { label: 'Needs approval', icon: UiShield, className: 'text-amber-600 bg-amber-500/15 border-amber-500/30' },
  completed: { label: 'Completed', icon: UiPass, className: 'text-emerald-600 bg-emerald-500/15 border-emerald-500/30' },
  error: { label: 'Error', icon: UiError, className: 'text-red-600 bg-red-500/15 border-red-500/30' },
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
export function useTodoSession(dir: string, provider: string, sessionId: string | undefined, active: boolean) {
  const [entries, setEntries] = useState<SessionEntry[]>([]);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    setEntries([]);
    setError('');
    setConnected(false);
    if (!active || !sessionId) return;

    const params = new URLSearchParams(todoQuery(dir, provider));
    params.set('sessionId', sessionId);
    const es = new EventSource(`/api/todos/session/stream?${params.toString()}`);

    es.onopen = () => setConnected(true);
    es.addEventListener('entry', (e: MessageEvent) => {
      try {
        const entry = JSON.parse(e.data) as SessionEntry;
        setEntries(prev => [...prev, entry]);
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
          setError('Session stream error');
        }
      }
      setConnected(false);
    });

    return () => es.close();
  }, [dir, provider, sessionId, active]);

  return { entries, connected, error };
}

// useSessionStatus polls the session stats endpoint for the high-level agent
// state and any pending tool-permission request, and exposes a resolver that
// POSTs the user's Allow/Deny. State is server-derived (the same source the
// session timer uses), so the header badge and the approval banner stay in sync
// without re-deriving anything from the event stream.
export function useSessionStatus(dir: string, provider: string, sessionId: string | undefined, active: boolean) {
  const [status, setStatus] = useState<{ state: string; error: string; inProgress: boolean; approval: TodoSessionApproval | null }>(
    { state: '', error: '', inProgress: false, approval: null },
  );
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    setStatus({ state: '', error: '', inProgress: false, approval: null });
    if (!active || !sessionId) return;
    let cancelled = false;
    const params = new URLSearchParams(todoQuery(dir, provider));
    params.set('sessionId', sessionId);
    const url = `/api/todos/session/stats?${params.toString()}`;
    const poll = async () => {
      try {
        const res = await fetch(url);
        if (!res.ok) return;
        const stats = (await res.json()) as SessionStats;
        if (!cancelled) setStatus({ state: stats.state ?? '', error: stats.error ?? '', inProgress: stats.inProgress ?? false, approval: stats.approval ?? null });
      } catch {
        // Ignore transient fetch errors; the next tick recovers.
      }
    };
    poll();
    const id = setInterval(poll, 1500);
    return () => {
      cancelled = true;
      clearInterval(id);
    };
  }, [dir, provider, sessionId, active]);

  const approve = useCallback(
    async (allow: boolean, message?: string, updatedInput?: Record<string, unknown>) => {
      if (!sessionId) return;
      setBusy(true);
      try {
        const res = await fetch('/api/todos/session/approve', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ sessionId, allow, message, updatedInput }),
        });
        if (!res.ok) {
          let detail = 'Approval update failed';
          try {
            const data = await res.json();
            detail = data.error || detail;
          } catch {
            detail = await res.text() || detail;
          }
          throw new Error(detail);
        }
        setStatus(prev => ({ ...prev, approval: null }));
      } finally {
        setBusy(false);
      }
    },
    [sessionId],
  );

  return { ...status, approve, busy };
}

export function TodoSession({
  dir,
  provider,
  sessionId,
  active,
  todo,
  onChanged,
  onResume,
  resumeDisabled,
}: {
  dir: string;
  provider: string;
  sessionId?: string;
  active: boolean;
  todo: TodoItem;
  onChanged?: (todo: TodoItem) => void;
  // onResume/resumeDisabled back the session timer's "Resume in cmux" action,
  // wired from the todo detail's run flow.
  onResume?: () => void;
  resumeDisabled?: boolean;
}) {
  const { entries, connected, error } = useTodoSession(dir, provider, sessionId, active);
  const { state, error: statusError, inProgress, approval, approve } = useSessionStatus(dir, provider, sessionId, active);
  const scrollRef = useRef<HTMLDivElement | null>(null);
  // Host element for the SessionViewer's 3-dot menu: the viewer portals its menu
  // into this header slot so the filter/density controls sit in the same fixed
  // header as the status badge and session timer instead of scrolling with the log.
  const [menuHost, setMenuHost] = useState<HTMLElement | null>(null);
  // Follow the tail like a terminal, but stop following once the user scrolls up
  // to read earlier history (re-engages when they scroll back to the bottom).
  const followRef = useRef(true);

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
    if (approval) return [{ tool: approval.tool, input: approval.input, toolCallId: approval.toolUseId, sessionId: approval.sessionId }];
    if (state !== 'ask' || inProgress) return [];
    const latest = latestQuestionTool(entries);
    if (latest) return [{ tool: 'AskUserQuestion', input: latest.input, toolCallId: latest.toolCallId, sessionId }];
    if (todo.questions?.length) {
      return [{
        tool: 'AskUserQuestion',
        sessionId,
        input: { questions: todo.questions.map((question, index) => ({
          id: String(index + 1), question: question.text, header: question.context,
          options: (question.options ?? []).map(option => ({ label: option, value: option })),
        })) },
      }];
    }
    return [];
  }, [approval, entries, inProgress, sessionId, state, todo.questions]);

  const decide = useCallback(async (decision: SessionToolDecision) => {
    if (!sessionId) throw new Error('Session is unavailable');
    if (approval) {
      const updatedInput = decision.answers
        ? { ...(approval.input ?? {}), answers: decision.answers }
        : undefined;
      await approve(decision.allow, decision.message, updatedInput);
      return;
    }
    const res = await fetch('/api/todos/answer', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        provider, dir, ref: todo.ref,
        answer: decision.message || formatDecisionAnswers(decision.answers),
        answers: decision.answers,
        rejected: !decision.allow,
      }),
    });
    const body = await res.json().catch(() => ({}));
    if (!res.ok) {
      throw new Error(body.error || 'Could not resume the agent session');
    }
    onChanged?.(body.todo ?? todo);
  }, [approval, approve, dir, onChanged, provider, sessionId, todo.ref]);

  if (!sessionId) {
    return (
      <div className="flex min-h-0 flex-1 flex-col items-center justify-center px-4 py-6 text-center text-sm text-muted-foreground">
        <UiComment className="mb-2 text-3xl" />
        <p>No agent session yet. Run this todo to start one.</p>
      </div>
    );
  }

  const view = stateView(state, error || statusError);

  return (
    <div className="m-3 flex min-h-0 flex-1 flex-col overflow-hidden rounded-md border border-border bg-card">
      <div className="flex shrink-0 flex-wrap items-center gap-x-3 gap-y-1 border-b border-border bg-muted/40 px-3 py-1.5 text-[11px] text-muted-foreground">
        <SessionStateBadge view={view} />
        <span className="inline-flex items-center gap-1">
          <UiCircleFilled className={`text-[7px] ${connected ? 'text-emerald-500' : 'text-muted-foreground'}`} />
          {connected ? 'Following session' : 'Session idle'}
        </span>
        <span className="font-mono">{sessionId.slice(0, 8)}</span>
        <TodoSessionTimer dir={dir} provider={provider} sessionId={sessionId} active={active} onResume={onResume} resumeDisabled={resumeDisabled} />
        <span ref={setMenuHost} className="ml-auto inline-flex items-center" />
      </div>
      <div ref={scrollRef} onScroll={onScroll} className="min-h-0 flex-1 overflow-y-auto px-3 py-2">
        {error && <div className="text-xs text-red-500">{error}</div>}
        {entries.length === 0 && !error && (
          <div className="text-xs text-muted-foreground">Waiting for session activity…</div>
        )}
        {entries.length > 0 && (
          <SessionViewer session={entries} pendingTools={pendingTools} onPendingToolDecision={decide} showHeader={false} menuContainer={menuHost} className="text-xs" />
        )}
        {entries.length === 0 && pendingTools.length > 0 && (
          <SessionViewer session={[]} pendingTools={pendingTools} onPendingToolDecision={decide} showHeader={false} menuContainer={menuHost} className="text-xs" />
        )}
      </div>
    </div>
  );
}

function latestQuestionTool(entries: SessionEntry[]): { input?: Record<string, unknown>; toolCallId?: string } | undefined {
  for (let index = entries.length - 1; index >= 0; index--) {
    const entry = entries[index];
    if (entry.tool_use?.tool === 'AskUserQuestion') return { input: entry.tool_use.input, toolCallId: entry.tool_use.tool_use_id };
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
  return Object.entries(answers).map(([question, answer]) => `${question}: ${Array.isArray(answer) ? answer.join(', ') : answer}`).join('\n');
}

// ApprovalBanner shows a pending tool-permission request with Allow/Deny. The
// driver is blocked awaiting the decision, so the buttons drive the run forward.
export function ApprovalBanner({
  approval,
  busy,
  onDecide,
}: {
  approval: TodoSessionApproval;
  busy: boolean;
  onDecide: (allow: boolean, message?: string) => Promise<void> | void;
}) {
  if (approval.tool === 'AskUserQuestion') {
    return <QuestionApprovalBanner approval={approval} busy={busy} onDecide={onDecide} />;
  }

  return <PermissionApprovalBanner approval={approval} busy={busy} onDecide={onDecide} />;
}

function PermissionApprovalBanner({
  approval,
  busy,
  onDecide,
}: {
  approval: TodoSessionApproval;
  busy: boolean;
  onDecide: (allow: boolean, message?: string) => Promise<void> | void;
}) {
  const [commentOpen, setCommentOpen] = useState(false);
  const [comment, setComment] = useState('');
  const [error, setError] = useState('');

  const summary = approvalInputSummary(approval.input);
  const command = approvalCommand(approval.input);
  const details = approvalDetailEntries(approval.input);

  useEffect(() => {
    setCommentOpen(false);
    setComment('');
    setError('');
  }, [approval.sessionId, approval.toolUseId, approval.tool]);

  const decide = async (allow: boolean, message?: string) => {
    setError('');
    try {
      if (message === undefined) await onDecide(allow);
      else await onDecide(allow, message);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Approval update failed');
    }
  };

  return (
    <div className="flex shrink-0 flex-col gap-2 border-b border-amber-500/30 bg-amber-500/10 px-3 py-2 text-[11px] text-amber-800">
      <div className="flex flex-wrap items-center gap-2">
        <UiShield className="shrink-0 text-amber-600" />
        <span className="font-medium">
          Needs approval: <span className="text-foreground">{approval.tool}</span>
        </span>
        {summary && <span className="min-w-0 flex-1 truncate opacity-80">{summary}</span>}
      </div>
      <div className="space-y-2 rounded border border-amber-500/20 bg-background/70 p-2 text-foreground">
        {command && (
          <pre className="max-h-40 overflow-auto rounded bg-muted px-2 py-1.5 font-mono text-xs leading-relaxed">
            {command}
          </pre>
        )}
        {details.length > 0 && (
          <dl className="grid gap-1 text-xs sm:grid-cols-[max-content_1fr]">
            {details.map(detail => (
              <div key={detail.name} className="contents">
                <dt className="text-muted-foreground">{detail.name}</dt>
                <dd className="min-w-0 break-words font-mono">{detail.value}</dd>
              </div>
            ))}
          </dl>
        )}
        {approval.input && (
          <details className="text-xs">
            <summary className="cursor-pointer text-muted-foreground hover:text-foreground">Raw request</summary>
            <pre className="mt-1 max-h-52 overflow-auto rounded bg-muted px-2 py-1.5 font-mono leading-relaxed">
              {JSON.stringify(approval.input, null, 2)}
            </pre>
          </details>
        )}
      </div>
      {commentOpen && (
        <div className="space-y-1.5">
          <textarea
            className={`${inputClass} min-h-[4rem] resize-y bg-background text-sm text-foreground`}
            value={comment}
            onChange={event => setComment(event.currentTarget.value)}
            placeholder="Tell the agent why this request is rejected..."
          />
          <div className="flex justify-end gap-1.5">
            <Button
              type="button"
              variant="ghost"
              disabled={busy}
              onClick={() => {
                setCommentOpen(false);
                setComment('');
              }}
              className="h-auto rounded px-2 py-0.5 text-muted-foreground hover:bg-muted disabled:opacity-50"
            >
              Cancel
            </Button>
            <Button
              variant="ghost"
              type="button"
              disabled={busy || !comment.trim()}
              onClick={() => void decide(false, comment.trim())}
              className="h-auto rounded border border-red-500/40 bg-red-500/15 px-2 py-0.5 text-red-700 hover:bg-red-500/25 disabled:opacity-50"
            >
              <UiCancel className="text-xs" />
              Reject with comment
            </Button>
          </div>
        </div>
      )}
      {error && <div className="text-xs text-red-600">{error}</div>}
      <div className="flex flex-wrap items-center justify-end gap-1.5">
        <Button
          variant="ghost"
          type="button"
          disabled={busy}
          onClick={() => void decide(true)}
          className="h-auto rounded border border-emerald-500/40 bg-emerald-500/15 px-2 py-0.5 text-emerald-700 hover:bg-emerald-500/25 disabled:opacity-50"
        >
          <UiCheck className="text-xs" />
          Allow
        </Button>
        <Button
          variant="ghost"
          type="button"
          disabled={busy}
          onClick={() => void decide(false)}
          className="h-auto rounded border border-red-500/40 bg-red-500/15 px-2 py-0.5 text-red-700 hover:bg-red-500/25 disabled:opacity-50"
        >
          <UiCancel className="text-xs" />
          Reject
        </Button>
        {!commentOpen && (
          <Button
            variant="ghost"
            type="button"
            disabled={busy}
            onClick={() => setCommentOpen(true)}
            className="h-auto rounded border border-red-500/30 bg-background px-2 py-0.5 text-red-700 hover:bg-red-500/10 disabled:opacity-50"
          >
            <UiComment className="text-xs" />
            Reject with comment
          </Button>
        )}
      </div>
    </div>
  );
}

function QuestionApprovalBanner({
  approval,
  busy,
  onDecide,
}: {
  approval: TodoSessionApproval;
  busy: boolean;
  onDecide: (allow: boolean, message?: string) => Promise<void> | void;
}) {
  const questions = questionsFromToolInput(approval.input);
  const [answers, setAnswers] = useState<Record<string, string | string[]>>({});
  const [details, setDetails] = useState('');
  const [commentOpen, setCommentOpen] = useState(false);
  const [comment, setComment] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    setAnswers({});
    setDetails('');
    setCommentOpen(false);
    setComment('');
    setError('');
  }, [approval.sessionId, approval.toolUseId, approval.tool]);

  const message = formatQuestionApprovalMessage(questions, answers, details);

  const decide = async (allow: boolean, nextMessage?: string) => {
    setError('');
    try {
      if (nextMessage === undefined) await onDecide(allow);
      else await onDecide(allow, nextMessage);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Approval update failed');
    }
  };

  const submit = () => {
    if (!message) return;
    void decide(true, message);
  };

  return (
    <div className="flex shrink-0 flex-col gap-2 border-b border-purple-500/30 bg-purple-500/5 px-3 py-2 text-[11px] text-purple-700">
      <div className="flex items-center gap-2">
        <UiComment className="shrink-0 text-purple-600" />
        <span className="font-medium">The agent needs your input to continue</span>
      </div>
      {questions.length > 0 ? questions.map((question, index) => (
        <QuestionPrompt
          key={`${question.id}-${index}`}
          approval={approval}
          question={question}
          index={index}
          value={answers[questionKey(question, index)]}
          disabled={busy}
          onChange={value => setAnswers(prev => ({ ...prev, [questionKey(question, index)]: value }))}
        />
      )) : (
        <div className="rounded border border-purple-500/30 bg-background/60 p-2 text-sm text-foreground">
          The agent is waiting for an answer.
        </div>
      )}
      {questions.length === 0 && (
        <textarea
          className={`${inputClass} min-h-[4rem] resize-y bg-background text-sm text-foreground`}
          value={details}
          disabled={busy}
          onChange={event => setDetails(event.currentTarget.value)}
          placeholder="Answer the agent's question..."
        />
      )}
      {questions.some(question => question.options.length > 0) && (
        <textarea
          className={`${inputClass} min-h-[3rem] resize-y bg-background text-sm text-foreground`}
          value={details}
          disabled={busy}
          onChange={event => setDetails(event.currentTarget.value)}
          placeholder="Additional details for the agent..."
        />
      )}
      {commentOpen && (
        <div className="space-y-1.5">
          <textarea
            className={`${inputClass} min-h-[4rem] resize-y bg-background text-sm text-foreground`}
            value={comment}
            disabled={busy}
            onChange={event => setComment(event.currentTarget.value)}
            placeholder="Tell the agent why this question is rejected..."
          />
          <div className="flex justify-end gap-1.5">
            <Button
              type="button"
              variant="ghost"
              disabled={busy}
              onClick={() => {
                setCommentOpen(false);
                setComment('');
              }}
              className="h-auto rounded px-2 py-0.5 text-muted-foreground hover:bg-muted disabled:opacity-50"
            >
              Cancel
            </Button>
            <Button
              variant="ghost"
              type="button"
              disabled={busy || !comment.trim()}
              onClick={() => void decide(false, comment.trim())}
              className="h-auto rounded border border-red-500/40 bg-red-500/15 px-2 py-0.5 text-red-700 hover:bg-red-500/25 disabled:opacity-50"
            >
              <UiCancel className="text-xs" />
              Reject with comment
            </Button>
          </div>
        </div>
      )}
      {error && <div className="text-xs text-red-600">{error}</div>}
      <div className="flex flex-wrap items-center justify-end gap-1.5">
        <Button
          variant="ghost"
          type="button"
          disabled={busy || !message}
          onClick={submit}
          className="h-auto rounded border border-emerald-500/40 bg-emerald-500/15 px-2 py-0.5 text-emerald-700 hover:bg-emerald-500/25 disabled:opacity-50"
        >
          <UiCheck className="text-xs" />
          Send answer
        </Button>
        <Button
          variant="ghost"
          type="button"
          disabled={busy}
          onClick={() => void decide(false)}
          className="h-auto rounded border border-red-500/40 bg-red-500/10 px-2 py-0.5 text-red-700 hover:bg-red-500/20 disabled:opacity-50"
        >
          <UiCancel className="text-xs" />
          Reject
        </Button>
        {!commentOpen && (
          <Button
            variant="ghost"
            type="button"
            disabled={busy}
            onClick={() => setCommentOpen(true)}
            className="h-auto rounded border border-red-500/30 bg-background px-2 py-0.5 text-red-700 hover:bg-red-500/10 disabled:opacity-50"
          >
            <UiComment className="text-xs" />
            Reject with comment
          </Button>
        )}
      </div>
    </div>
  );
}

function QuestionPrompt({
  approval,
  question,
  index,
  value,
  disabled,
  onChange,
}: {
  approval: TodoSessionApproval;
  question: SessionQuestion;
  index: number;
  value: string | string[] | undefined;
  disabled?: boolean;
  onChange: (value: string | string[]) => void;
}) {
  const key = questionKey(question, index);
  const label = question.context || (question.id ? `Question ${question.id}` : `Question ${index + 1}`);
  const fieldName = `approval-${approval.toolUseId || approval.sessionId}-${key}`;

  return (
    <fieldset className="rounded-md border border-purple-500/30 bg-background/70 p-2 text-sm text-foreground">
      <legend className="px-1 text-[11px] font-medium uppercase text-muted-foreground">{label}</legend>
      <div className="whitespace-pre-wrap break-words font-medium">{question.text}</div>
      {question.options.length > 0 ? (
        <div className="mt-2 grid gap-1.5">
          {question.options.map(option => {
            const checked = Array.isArray(value)
              ? value.includes(option.value)
              : value === option.value;
            return (
              <label key={option.value} className="flex cursor-pointer items-start gap-2 rounded border border-border bg-muted/30 px-2 py-1.5 hover:bg-muted">
                <input
                  className="mt-0.5"
                  type={question.multiSelect ? 'checkbox' : 'radio'}
                  name={fieldName}
                  value={option.value}
                  checked={checked}
                  disabled={disabled}
                  onChange={event => {
                    if (question.multiSelect) {
                      const current = Array.isArray(value) ? value : [];
                      onChange(event.currentTarget.checked
                        ? [...current, option.value]
                        : current.filter(item => item !== option.value));
                    } else {
                      onChange(option.value);
                    }
                  }}
                />
                <span className="min-w-0">
                  <span className="block font-medium">{option.label}</span>
                  {option.description && <span className="block text-xs text-muted-foreground">{option.description}</span>}
                </span>
              </label>
            );
          })}
        </div>
      ) : (
        <textarea
          className={`${inputClass} mt-2 min-h-[4rem] resize-y bg-background text-sm`}
          value={typeof value === 'string' ? value : ''}
          disabled={disabled}
          onChange={event => onChange(event.currentTarget.value)}
          placeholder="Answer this question..."
        />
      )}
    </fieldset>
  );
}

export function formatQuestionApprovalMessage(
  questions: SessionQuestion[],
  answers: Record<string, string | string[]>,
  details = '',
): string {
  if (questions.length === 0) return details.trim();
  const lines: string[] = [];
  questions.forEach((question, index) => {
    const value = answers[questionKey(question, index)];
    const values = Array.isArray(value) ? value : value ? [value] : [];
    if (values.length === 0) return;
    const label = question.context || question.id || `Question ${index + 1}`;
    const rendered = question.options.length > 0
      ? values.map(optionValue => question.options.find(option => option.value === optionValue)?.label || optionValue).join(', ')
      : values.join(', ');
    if (rendered.trim()) lines.push(`${label}: ${rendered.trim()}`);
  });
  const trimmedDetails = details.trim();
  if (trimmedDetails) lines.push(`Additional details: ${trimmedDetails}`);
  return lines.join('\n');
}

function questionKey(question: SessionQuestion, index: number): string {
  return question.id || String(index + 1);
}

// approvalInputSummary picks the most descriptive field of a tool's input for a
// one-line preview (the command, the file, etc.), truncated.
function approvalInputSummary(input?: Record<string, unknown>): string {
  if (!input) return '';
  for (const key of ['command', 'file_path', 'path', 'pattern', 'query', 'url']) {
    const v = input[key];
    if (typeof v === 'string' && v) {
      return v.length > 120 ? `${v.slice(0, 120)}…` : v;
    }
  }
  return '';
}

function approvalCommand(input?: Record<string, unknown>): string {
  const command = input?.command;
  return typeof command === 'string' ? command : '';
}

function approvalDetailEntries(input?: Record<string, unknown>): Array<{ name: string; value: string }> {
  if (!input) return [];
  const consumed = new Set(['command']);
  return Object.entries(input)
    .filter(([name, value]) => !consumed.has(name) && value !== undefined && value !== null && value !== '')
    .map(([name, value]) => ({ name, value: detailValue(value) }));
}

function detailValue(value: unknown): string {
  if (typeof value === 'string') return value;
  try {
    return JSON.stringify(value);
  } catch {
    return String(value);
  }
}

function SessionStateBadge({ view }: { view: SessionStateView }) {
  const Icon = view.icon;
  return (
    <span className={`inline-flex items-center gap-1 rounded border px-1.5 py-0.5 text-[9px] font-medium uppercase ${view.className}`}>
      <Icon className="text-[11px]" />
      {view.label}
    </span>
  );
}
