import { useEffect, useMemo, useRef, useState } from 'react';
import type { TodoSessionEvent } from '../../types';
import { GavelIcon } from '../GavelIcon';
import { Markdown } from '../Markdown';
import { todoQuery } from './format';
import { SessionFilterMenu } from './SessionFilterMenu';
import {
  computeFacets,
  DEFAULT_HIDDEN_TOOLS,
  eventVisible,
  type SessionCategory,
} from './sessionFilter';

// Claude tools that pause the turn awaiting the user (the entry carries no
// terminal stop reason), so the session is "asking" rather than "working" until
// they answer.
const ASK_TOOLS = new Set(['AskUserQuestion', 'ExitPlanMode']);

interface SessionStateView {
  key: 'waiting' | 'thinking' | 'working' | 'ask' | 'completed' | 'error';
  label: string;
  icon: string;
  className: string;
}

// deriveSessionState rolls the streamed events up to a single high-level state
// for the session header — what is the agent doing right now?
//   waiting   - stream open but nothing emitted yet
//   thinking  - last block was extended thinking
//   ask       - paused on an AskUserQuestion/ExitPlanMode tool, awaiting the user
//   working   - actively running a tool or emitting a response mid-turn
//   completed - the turn ended (end_turn / stop_sequence / max_tokens)
//   error     - the stream reported an error
export function deriveSessionState(events: TodoSessionEvent[], error: string): SessionStateView {
  if (error) {
    return { key: 'error', label: 'Error', icon: 'codicon:error', className: 'text-red-300 bg-red-500/15 border-red-500/30' };
  }
  const last = events[events.length - 1];
  if (!last) {
    return { key: 'waiting', label: 'Waiting', icon: 'svg-spinners:ring-resize', className: 'text-gray-300 bg-white/5 border-zinc-700' };
  }
  switch (last.kind) {
    case 'thinking':
      return { key: 'thinking', label: 'Thinking', icon: 'codicon:lightbulb', className: 'text-amber-300 bg-amber-500/15 border-amber-500/30' };
    case 'tool_use':
      if (last.tool && ASK_TOOLS.has(last.tool)) {
        return { key: 'ask', label: 'Awaiting input', icon: 'codicon:comment-discussion', className: 'text-purple-300 bg-purple-500/15 border-purple-500/30' };
      }
      return { key: 'working', label: 'Working', icon: 'svg-spinners:ring-resize', className: 'text-cyan-300 bg-cyan-500/15 border-cyan-500/30' };
    case 'error':
      return { key: 'error', label: errorLabel(last), icon: 'codicon:error', className: 'text-red-300 bg-red-500/15 border-red-500/30' };
    case 'turn_end':
      return { key: 'completed', label: 'Completed', icon: 'codicon:pass', className: 'text-emerald-300 bg-emerald-500/15 border-emerald-500/30' };
    default:
      return { key: 'working', label: 'Working', icon: 'svg-spinners:ring-resize', className: 'text-cyan-300 bg-cyan-500/15 border-cyan-500/30' };
  }
}

// errorLabel renders a compact badge label for an API/network error event: the
// HTTP status when present (e.g. "Error 529"), else a plain "Error".
function errorLabel(ev: TodoSessionEvent): string {
  return ev.errorStatus ? `Error ${ev.errorStatus}` : 'Error';
}

// useTodoSession follows a TODO's agent session log over SSE. The server tails
// the on-disk Claude session log and re-parses it with captain, so the
// transcript is reconstructed live rather than stored on the issue. The stream
// replays existing history on connect, then follows new events until unmounted.
export function useTodoSession(dir: string, provider: string, sessionId: string | undefined, active: boolean) {
  const [events, setEvents] = useState<TodoSessionEvent[]>([]);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    setEvents([]);
    setError('');
    setConnected(false);
    if (!active || !sessionId) return;

    const params = new URLSearchParams(todoQuery(dir, provider));
    params.set('sessionId', sessionId);
    const es = new EventSource(`/api/todos/session/stream?${params.toString()}`);

    es.onopen = () => setConnected(true);
    es.addEventListener('event', (e: MessageEvent) => {
      try {
        const ev = JSON.parse(e.data) as TodoSessionEvent;
        setEvents(prev => [...prev, ev]);
      } catch {
        // Ignore malformed frames; the next well-formed event recovers.
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

  return { events, connected, error };
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
  const { events, connected, error } = useTodoSession(dir, provider, sessionId, active);
  const scrollRef = useRef<HTMLDivElement | null>(null);
  // Follow the tail like a terminal, but stop following once the user scrolls up
  // to read earlier history (re-engages when they scroll back to the bottom).
  const followRef = useRef(true);

  const [hiddenCategories, setHiddenCategories] = useState<Set<SessionCategory>>(() => new Set());
  // Explore subagent calls are noisy and excluded by default (see sessionFilter).
  const [hiddenTools, setHiddenTools] = useState<Set<string>>(() => new Set(DEFAULT_HIDDEN_TOOLS));

  const facets = useMemo(() => computeFacets(events), [events]);
  const visible = useMemo(
    () => events.filter(ev => eventVisible(ev, hiddenCategories, hiddenTools)),
    [events, hiddenCategories, hiddenTools],
  );

  function onScroll() {
    const el = scrollRef.current;
    if (!el) return;
    followRef.current = el.scrollHeight - el.scrollTop - el.clientHeight < 24;
  }

  useEffect(() => {
    const el = scrollRef.current;
    if (el && followRef.current) el.scrollTop = el.scrollHeight;
  }, [visible.length]);

  if (!sessionId) {
    return (
      <div className="flex min-h-0 flex-1 flex-col items-center justify-center px-4 py-6 text-center text-sm text-muted-foreground">
        <GavelIcon name="codicon:comment-discussion" className="mb-2 text-3xl" />
        <p>No agent session yet. Run this todo to start one.</p>
      </div>
    );
  }

  const toggleCategory = (key: SessionCategory) =>
    setHiddenCategories(prev => toggleSet(prev, key));
  const toggleTool = (key: string) => setHiddenTools(prev => toggleSet(prev, key));

  const hiddenCount = events.length - visible.length;

  return (
    <div className="m-3 flex min-h-0 flex-1 flex-col overflow-hidden rounded-md border border-zinc-800 bg-black font-mono">
      <div className="flex shrink-0 items-center gap-2 border-b border-zinc-800 bg-zinc-900 px-3 py-1 text-[10px] text-gray-400">
        <SessionStateBadge state={deriveSessionState(events, error)} />
        <span className="inline-flex items-center gap-1 text-gray-500">
          <GavelIcon name="octicon:dot-fill-16" className={`text-[7px] ${connected ? 'text-emerald-400' : 'text-gray-600'}`} />
          {connected ? 'Following session' : 'Session idle'}
        </span>
        <span className="text-gray-600">{sessionId.slice(0, 8)}</span>
        <SessionFilterMenu
          facets={facets}
          hiddenCategories={hiddenCategories}
          hiddenTools={hiddenTools}
          onToggleCategory={toggleCategory}
          onToggleTool={toggleTool}
        />
      </div>
      <div
        ref={scrollRef}
        onScroll={onScroll}
        className="min-h-0 flex-1 space-y-1.5 overflow-y-auto px-3 py-2 text-[11px] leading-snug text-gray-100"
      >
        {error && <div className="text-red-400">{error}</div>}
        {events.length === 0 && !error && (
          <div className="text-gray-500">Waiting for session activity…</div>
        )}
        {events.length > 0 && visible.length === 0 && !error && (
          <div className="text-gray-500">All {hiddenCount} events hidden by filters.</div>
        )}
        {visible.map((ev, i) => (
          <SessionEventRow key={i} event={ev} />
        ))}
      </div>
    </div>
  );
}

// toggleSet returns a new Set with key added when absent and removed when
// present — React state must not be mutated in place.
function toggleSet<T>(set: Set<T>, key: T): Set<T> {
  const next = new Set(set);
  if (next.has(key)) next.delete(key);
  else next.add(key);
  return next;
}

function SessionStateBadge({ state }: { state: SessionStateView }) {
  return (
    <span className={`inline-flex items-center gap-1 rounded border px-1.5 py-0.5 text-[9px] font-medium uppercase ${state.className}`}>
      <GavelIcon name={state.icon} className="text-[11px]" />
      {state.label}
    </span>
  );
}

function SessionEventRow({ event }: { event: TodoSessionEvent }) {
  switch (event.kind) {
    case 'thinking':
      return (
        <div className="flex gap-2 italic text-gray-500">
          <GavelIcon name="codicon:lightbulb" className="mt-0.5 shrink-0" />
          <span className="min-w-0 whitespace-pre-wrap break-words">{event.text}</span>
        </div>
      );
    case 'tool_use':
      return (
        <div className="flex items-start gap-2">
          <GavelIcon name="codicon:tools" className="mt-0.5 shrink-0 text-cyan-400" />
          <span className="min-w-0 break-words">
            <span className="font-medium text-cyan-300">{event.tool}</span>
            {event.action && <span className="ml-1 text-gray-400">{event.action}</span>}
          </span>
        </div>
      );
    case 'turn_end':
      return (
        <div className="flex items-center gap-2 border-t border-zinc-800 pt-1.5 text-[10px] uppercase tracking-wide text-gray-500">
          <GavelIcon name="codicon:pass" className="shrink-0 text-emerald-400" />
          <span>Turn ended{event.stopReason ? ` (${event.stopReason})` : ''}</span>
        </div>
      );
    case 'error':
      return (
        <div className="flex items-start gap-2 border-t border-red-500/30 pt-2 text-red-300">
          <GavelIcon name="codicon:error" className="mt-0.5 shrink-0 text-red-400" />
          <span className="min-w-0 whitespace-pre-wrap break-words">
            {event.text || `API error${event.errorStatus ? ` ${event.errorStatus}` : ''}`}
          </span>
        </div>
      );
    default:
      // assistant text — Markdown inherits the terminal's light text color
      return (
        <div className="flex gap-2">
          <GavelIcon name="codicon:sparkle" className="mt-0.5 shrink-0 text-purple-400" />
          <Markdown text={event.text ?? ''} className="min-w-0 text-gray-100" />
        </div>
      );
  }
}
