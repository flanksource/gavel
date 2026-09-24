import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { fetchRemoteSession, type SessionCollectionInput, type SessionCollectionItem } from '@flanksource/clicky-ui/ai';
import { Button } from '@flanksource/clicky-ui/components';
import { UiCheck, UiCopy, UiStop } from '@flanksource/clicky-ui/icons';
import type { TodoSessionAttempt, TodoSessionDetailResponse } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { copyText } from '../../clipboard';
import { sessionDetailQueryOptions } from './todoQueries';
import type { TodoLaunchProgress } from './todoLaunch';

export interface TodoSessionDetailOptions {
  /** Poll period while an attempt is live; settled attempts are re-read slowly. */
  intervalMs?: number;
}

/** The todo's attempts, newest first — one shared poll per todo. */
export function useTodoSessionDetail(dir: string, ref: string, active: boolean, { intervalMs = 1500 }: TodoSessionDetailOptions = {}) {
  const enabled = active && !!ref;
  const query = useQuery({ ...sessionDetailQueryOptions(dir, ref, intervalMs), enabled });
  return {
    detail: enabled ? query.data ?? null : null,
    error: enabled && query.error
      ? `Session detail request failed\n${query.error instanceof Error ? query.error.stack || query.error.message : String(query.error)}`
      : '',
  };
}

/** Captain's session handler, mounted on the dashboard (pr/ui/handler.go). */
export function captainSessionUrl(sessionId: string) {
  return `/api/captain/sessions/${encodeURIComponent(sessionId)}`;
}

/** The Captain session holding an attempt's transcript, once the run has one. */
export function attemptSessionId(attempt: TodoSessionAttempt) {
  return attempt.executionSessionId || attempt.providerSessionId || undefined;
}

/**
 * The attempt a session id names — its prompt run, admission, execution or
 * provider session (the same four ids the server's RunMatchesSession accepts)
 * — else the newest attempt.
 */
export function selectAttempt(attempts: TodoSessionAttempt[], sessionId: string | undefined) {
  const named = sessionId
    ? attempts.find(attempt => [attempt.promptRunId, attempt.admissionSessionId, attempt.executionSessionId, attempt.providerSessionId].includes(sessionId))
    : undefined;
  return named ?? attempts[0];
}

export const PENDING_LAUNCH_ID = 'pending-launch';

/**
 * One collection item per attempt. An attempt with a Captain session loads it
 * from `src` (SessionInspector fetches, and follows while live); one that has
 * none yet — or a launch still being admitted — shows as an empty session.
 */
export function attemptCollection({ todoRef, attempts, currentId, launch }: {
  todoRef: string;
  attempts: TodoSessionAttempt[];
  currentId: string;
  launch: TodoLaunchProgress | null;
}): SessionCollectionInput {
  const launchItem: SessionCollectionItem[] = launch && !attempts.some(attempt => attempt.promptRunId === currentId)
    ? [{
      id: currentId,
      label: `Attempt #${attempts.length + 1}`,
      mode: launch.step,
      status: launch.status === 'admitted' ? 'running' : 'starting',
      session: { id: currentId, messages: [] },
    }]
    : [];
  return {
    kind: 'session-collection',
    id: `todo-attempts:${todoRef}`,
    currentSessionId: currentId,
    sessions: [...launchItem, ...attempts.map(attemptItem)],
  };
}

function attemptItem(attempt: TodoSessionAttempt): SessionCollectionItem {
  const status = attempt.stopping ? 'stopping' : attempt.status;
  const mode = attempt.mode || attempt.step;
  const sessionId = attemptSessionId(attempt);
  return {
    id: attempt.promptRunId,
    label: `Attempt #${attempt.ordinal}`,
    mode,
    status,
    summary: {
      provider: attempt.provider,
      modelMode: attempt.runtimeMode,
      model: attempt.model,
      effort: attempt.effort,
      mode,
      status,
      pid: attempt.pid,
      durationMs: attempt.durationMs,
      updatedAt: attempt.updatedAt,
    },
    ...(sessionId ? { src: captainSessionUrl(sessionId) } : { session: { id: attempt.promptRunId, messages: [] } }),
  };
}

export function AttemptStopAction({ attempt, onStop }: { attempt: TodoSessionAttempt; onStop: (attempt: TodoSessionAttempt) => Promise<void> }) {
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

/**
 * Copies the attempt list plus the selected attempt's Captain session, read
 * fresh from the session URL at click time.
 */
export function CopyAllDetailsButton({ detail, attempt, compact = false }: { detail: TodoSessionDetailResponse; attempt: TodoSessionAttempt | undefined; compact?: boolean }) {
  const [state, setState] = useState<'idle' | 'copied' | 'error'>('idle');
  const [error, setError] = useState('');
  const sessionId = attempt ? attemptSessionId(attempt) : undefined;
  const copy = async () => {
    setError('');
    try {
      const session = sessionId ? await fetchRemoteSession(captainSessionUrl(sessionId)) : null;
      await copyText(JSON.stringify({ ...detail, session }, null, 2));
      setState('copied');
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : String(reason));
      setState('error');
    }
    window.setTimeout(() => setState('idle'), 1800);
  };
  const label = state === 'error' ? `Copy failed: ${error}` : 'Copy all session details';
  return (
    <Button
      variant="ghost"
      {...(compact ? { size: 'icon' as const } : {})}
      type="button"
      aria-label="Copy all session details"
      title={label}
      className={compact ? 'hidden size-7 shrink-0 text-muted-foreground hover:bg-muted hover:text-foreground @min-[48rem]:inline-flex' : 'h-8 gap-1 px-2 text-[11px]'}
      onClick={() => void copy()}
    >
      {state === 'copied' ? <UiCheck className="text-emerald-600" /> : <UiCopy className={state === 'error' ? 'text-red-600' : undefined} />}
      {!compact ? (state === 'copied' ? 'Copied' : state === 'error' ? 'Copy failed' : 'Copy all') : null}
    </Button>
  );
}
