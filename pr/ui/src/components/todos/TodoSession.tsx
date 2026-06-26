import { useCallback, useEffect, useRef, useState } from 'react';
import { SessionViewer, type SessionEntry } from '@flanksource/clicky-ui/ai';
import { Button } from '@flanksource/clicky-ui/components';
import type { SessionStats, TodoSessionApproval } from '../../types';
import { GavelIcon } from '../GavelIcon';
import { todoQuery } from './format';

interface SessionStateView {
  label: string;
  icon: string;
  className: string;
}

// STATE_VIEWS maps the server-derived session state (cmux.SessionStats.State) to
// the header badge's label, icon and colour. Tints are semi-transparent so they
// read on both the light and dark dashboard themes. The empty key is the initial
// "no event yet" state.
const STATE_VIEWS: Record<string, SessionStateView> = {
  '': { label: 'Waiting', icon: 'svg-spinners:ring-resize', className: 'text-muted-foreground bg-muted/50 border-border' },
  thinking: { label: 'Thinking', icon: 'codicon:lightbulb', className: 'text-amber-600 bg-amber-500/15 border-amber-500/30' },
  working: { label: 'Working', icon: 'svg-spinners:ring-resize', className: 'text-cyan-600 bg-cyan-500/15 border-cyan-500/30' },
  ask: { label: 'Awaiting input', icon: 'codicon:comment-discussion', className: 'text-purple-600 bg-purple-500/15 border-purple-500/30' },
  approval: { label: 'Needs approval', icon: 'codicon:shield', className: 'text-amber-600 bg-amber-500/15 border-amber-500/30' },
  completed: { label: 'Completed', icon: 'codicon:pass', className: 'text-emerald-600 bg-emerald-500/15 border-emerald-500/30' },
  error: { label: 'Error', icon: 'codicon:error', className: 'text-red-600 bg-red-500/15 border-red-500/30' },
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
}: {
  dir: string;
  provider: string;
  sessionId?: string;
  active: boolean;
}) {
  const { entries, connected, error } = useTodoSession(dir, provider, sessionId, active);
  const { state, error: statusError, approval, approve, busy: approveBusy } = useSessionStatus(dir, provider, sessionId, active);
  const scrollRef = useRef<HTMLDivElement | null>(null);
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
        <GavelIcon name="codicon:comment-discussion" className="mb-2 text-3xl" />
        <p>No agent session yet. Run this todo to start one.</p>
      </div>
    );
  }

  const view = stateView(state, error || statusError);

  return (
    <div className="m-3 flex min-h-0 flex-1 flex-col overflow-hidden rounded-md border border-border bg-card">
      <div className="flex shrink-0 items-center gap-2 border-b border-border bg-muted/40 px-3 py-1 text-[10px] text-muted-foreground">
        <SessionStateBadge view={view} />
        <span className="inline-flex items-center gap-1">
          <GavelIcon name="octicon:dot-fill-16" className={`text-[7px] ${connected ? 'text-emerald-500' : 'text-muted-foreground'}`} />
          {connected ? 'Following session' : 'Session idle'}
        </span>
        <span className="font-mono">{sessionId.slice(0, 8)}</span>
      </div>
      {approval && <ApprovalBanner approval={approval} busy={approveBusy} onDecide={approve} />}
      <div ref={scrollRef} onScroll={onScroll} className="min-h-0 flex-1 overflow-y-auto px-3 py-2">
        {error && <div className="text-xs text-red-500">{error}</div>}
        {entries.length === 0 && !error && (
          <div className="text-xs text-muted-foreground">Waiting for session activity…</div>
        )}
        {entries.length > 0 && <SessionViewer session={entries} showHeader={false} className="text-xs" />}
      </div>
    </div>
  );
}

// ApprovalBanner shows a pending tool-permission request with Allow/Deny. The
// driver is blocked awaiting the decision, so the buttons drive the run forward.
function ApprovalBanner({
  approval,
  busy,
  onDecide,
}: {
  approval: TodoSessionApproval;
  busy: boolean;
  onDecide: (allow: boolean) => void;
}) {
  const summary = approvalInputSummary(approval.input);
  return (
    <div className="flex shrink-0 items-center gap-2 border-b border-amber-500/30 bg-amber-500/10 px-3 py-1.5 text-[11px] text-amber-700">
      <GavelIcon name="codicon:shield" className="shrink-0 text-amber-600" />
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

function SessionStateBadge({ view }: { view: SessionStateView }) {
  return (
    <span className={`inline-flex items-center gap-1 rounded border px-1.5 py-0.5 text-[9px] font-medium uppercase ${view.className}`}>
      <GavelIcon name={view.icon} className="text-[11px]" />
      {view.label}
    </span>
  );
}
