import { useCallback, useEffect, useRef, useState, type ComponentType } from 'react';
import { SessionViewer, type SessionEntry } from '@flanksource/clicky-ui/ai';
import { Button } from '@flanksource/clicky-ui/components';
import { UiCircleFilled, UiComment, UiError, UiLightbulb, UiPass, UiShield, type IconProps } from '@flanksource/clicky-ui/icons';
import type { SessionStats, TodoQuestion, TodoSessionApproval } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { todoQuery } from './format';
import { TodoSessionTimer } from './TodoSessionTimer';
import { AnswerBox, QuestionsPanel } from './planActions';

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
  const [status, setStatus] = useState<{ state: string; error: string; approval: TodoSessionApproval | null }>(
    { state: '', error: '', approval: null },
  );
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    setStatus({ state: '', error: '', approval: null });
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
        if (!cancelled) setStatus({ state: stats.state ?? '', error: stats.error ?? '', approval: stats.approval ?? null });
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
    async (allow: boolean, message?: string) => {
      if (!sessionId) return;
      setBusy(true);
      try {
        await fetch('/api/todos/session/approve', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ sessionId, allow, message }),
        });
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
  onResume,
  resumeDisabled,
}: {
  dir: string;
  provider: string;
  sessionId?: string;
  active: boolean;
  // onResume/resumeDisabled back the session timer's "Resume in cmux" action,
  // wired from the todo detail's run flow.
  onResume?: () => void;
  resumeDisabled?: boolean;
}) {
  const { entries, connected, error } = useTodoSession(dir, provider, sessionId, active);
  const { state, error: statusError, approval, approve, busy: approveBusy } = useSessionStatus(dir, provider, sessionId, active);
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
      {approval && <ApprovalBanner approval={approval} busy={approveBusy} onDecide={approve} />}
      <div ref={scrollRef} onScroll={onScroll} className="min-h-0 flex-1 overflow-y-auto px-3 py-2">
        {error && <div className="text-xs text-red-500">{error}</div>}
        {entries.length === 0 && !error && (
          <div className="text-xs text-muted-foreground">Waiting for session activity…</div>
        )}
        {entries.length > 0 && (
          <SessionViewer session={entries} showHeader={false} menuContainer={menuHost} className="text-xs" />
        )}
      </div>
    </div>
  );
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
  onDecide: (allow: boolean, message?: string) => void;
}) {
  if (approval.tool === 'AskUserQuestion') {
    return <QuestionApprovalBanner approval={approval} busy={busy} onDecide={onDecide} />;
  }

  const summary = approvalInputSummary(approval.input);
  return (
    <div className="flex shrink-0 items-center gap-2 border-b border-amber-500/30 bg-amber-500/10 px-3 py-1.5 text-[11px] text-amber-700">
      <UiShield className="shrink-0 text-amber-600" />
      <span className="min-w-0 flex-1 break-words">
        Needs approval: <span className="font-medium">{approval.tool}</span>
        {summary && <span className="ml-1 opacity-80">{summary}</span>}
      </span>
      <Button
        variant="ghost"
        type="button"
        disabled={busy}
        onClick={() => onDecide(true)}
        className="h-auto rounded border border-emerald-500/40 bg-emerald-500/15 px-2 py-0.5 text-emerald-700 hover:bg-emerald-500/25 disabled:opacity-50"
      >
        Allow
      </Button>
      <Button
        variant="ghost"
        type="button"
        disabled={busy}
        onClick={() => onDecide(false)}
        className="h-auto rounded border border-red-500/40 bg-red-500/15 px-2 py-0.5 text-red-700 hover:bg-red-500/25 disabled:opacity-50"
      >
        Deny
      </Button>
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
  onDecide: (allow: boolean, message?: string) => void;
}) {
  const [answer, setAnswer] = useState('');
  const questions = approvalQuestions(approval.input);

  useEffect(() => {
    setAnswer('');
  }, [approval.sessionId, approval.toolUseId, approval.tool]);

  const submit = () => {
    const trimmed = answer.trim();
    if (!trimmed) return;
    onDecide(true, trimmed);
  };

  return (
    <div className="flex shrink-0 flex-col gap-2 border-b border-purple-500/30 bg-purple-500/5 px-3 py-2 text-[11px] text-purple-700">
      <div className="flex items-center gap-2">
        <UiComment className="shrink-0 text-purple-600" />
        <span className="font-medium">The agent needs your input to continue</span>
      </div>
      {questions.length > 0 ? (
        <QuestionsPanel questions={questions} />
      ) : (
        <div className="rounded border border-purple-500/30 bg-background/60 p-2 text-sm text-foreground">
          The agent is waiting for an answer.
        </div>
      )}
      <AnswerBox
        value={answer}
        onChange={setAnswer}
        onSubmit={submit}
        busy={busy}
        label="Send answer"
        placeholder="Answer the agent's question... (Cmd/Ctrl+Enter to send)"
      />
      <div className="flex items-center justify-end">
        <Button
          variant="ghost"
          type="button"
          disabled={busy}
          onClick={() => onDecide(false, answer.trim() || 'No answer provided')}
          className="h-auto rounded border border-red-500/40 bg-red-500/10 px-2 py-0.5 text-red-700 hover:bg-red-500/20 disabled:opacity-50"
        >
          Deny
        </Button>
      </div>
    </div>
  );
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

export function approvalQuestions(input?: Record<string, unknown>): TodoQuestion[] {
  if (!input) return [];
  const rawQuestions = input.questions;
  if (Array.isArray(rawQuestions)) {
    return rawQuestions.map(questionFromValue).filter((q): q is TodoQuestion => q !== null);
  }

  const text = stringField(input, 'question') || stringField(input, 'prompt') || stringField(input, 'text');
  if (!text) return [];
  const context = stringField(input, 'header') || stringField(input, 'context') || stringField(input, 'description');
  const options = optionLabels(input.options);
  return [{
    text,
    ...(context ? { context } : {}),
    ...(options.length ? { options } : {}),
  }];
}

function questionFromValue(value: unknown): TodoQuestion | null {
  if (typeof value === 'string') {
    const text = value.trim();
    return text ? { text } : null;
  }
  if (!isRecord(value)) return null;
  const text = stringField(value, 'question') || stringField(value, 'text') || stringField(value, 'prompt') || stringField(value, 'label');
  if (!text) return null;
  const context = stringField(value, 'header') || stringField(value, 'context') || stringField(value, 'description');
  const options = optionLabels(value.options);
  return {
    text,
    ...(context ? { context } : {}),
    ...(options.length ? { options } : {}),
  };
}

function optionLabels(value: unknown): string[] {
  if (!Array.isArray(value)) return [];
  return value.flatMap(option => {
    if (typeof option === 'string') {
      const label = option.trim();
      return label ? [label] : [];
    }
    if (!isRecord(option)) return [];
    const label = stringField(option, 'label') || stringField(option, 'value') || stringField(option, 'id');
    const description = stringField(option, 'description');
    if (label && description) return [`${label} - ${description}`];
    if (label) return [label];
    if (description) return [description];
    return [];
  });
}

function stringField(record: Record<string, unknown>, key: string): string {
  const value = record[key];
  return typeof value === 'string' ? value.trim() : '';
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return !!value && typeof value === 'object' && !Array.isArray(value);
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
