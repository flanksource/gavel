import { useEffect, useRef, useState } from 'react';
import type { TodoSessionEvent } from '../../types';
import { GavelIcon } from '../GavelIcon';
import { Markdown } from '../Markdown';
import { todoQuery } from './format';

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
  const endRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    endRef.current?.scrollIntoView({ block: 'end' });
  }, [events.length]);

  if (!sessionId) {
    return (
      <div className="px-4 py-6 text-center text-sm text-muted-foreground">
        <GavelIcon name="codicon:comment-discussion" className="mb-2 text-3xl" />
        <p>No agent session yet. Run this todo to start one.</p>
      </div>
    );
  }

  return (
    <div className="space-y-2 px-4 py-3">
      <div className="flex items-center gap-2 text-xs text-muted-foreground">
        <GavelIcon
          name={connected ? 'svg-spinners:ring-resize' : 'codicon:debug-disconnect'}
          className="text-sm"
        />
        <span>{connected ? 'Following session' : 'Session idle'}</span>
        <span className="font-mono">{sessionId.slice(0, 8)}</span>
      </div>
      {error && <div className="text-xs text-red-600">{error}</div>}
      {events.length === 0 && !error && (
        <div className="text-sm text-muted-foreground">Waiting for session activity…</div>
      )}
      <div className="space-y-2">
        {events.map((ev, i) => (
          <SessionEventRow key={i} event={ev} />
        ))}
      </div>
      <div ref={endRef} />
    </div>
  );
}

function SessionEventRow({ event }: { event: TodoSessionEvent }) {
  switch (event.kind) {
    case 'thinking':
      return (
        <div className="flex gap-2 text-xs italic text-muted-foreground">
          <GavelIcon name="codicon:lightbulb" className="mt-0.5 shrink-0" />
          <span className="min-w-0 whitespace-pre-wrap break-words">{event.text}</span>
        </div>
      );
    case 'tool_use':
      return (
        <div className="flex items-start gap-2 text-xs">
          <GavelIcon name="codicon:tools" className="mt-0.5 shrink-0 text-blue-600" />
          <span className="min-w-0 break-words">
            <span className="font-medium text-foreground">{event.tool}</span>
            {event.action && <span className="ml-1 font-mono text-muted-foreground">{event.action}</span>}
          </span>
        </div>
      );
    case 'turn_end':
      return (
        <div className="flex items-center gap-2 border-t border-border pt-2 text-[11px] uppercase tracking-wide text-muted-foreground">
          <GavelIcon name="codicon:pass" className="shrink-0 text-green-600" />
          <span>Turn ended{event.stopReason ? ` (${event.stopReason})` : ''}</span>
        </div>
      );
    default:
      // assistant text
      return (
        <div className="flex gap-2 text-sm">
          <GavelIcon name="codicon:sparkle" className="mt-0.5 shrink-0 text-purple-600" />
          <Markdown text={event.text ?? ''} className="min-w-0 text-sm" />
        </div>
      );
  }
}
