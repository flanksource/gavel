import { useEffect, useState } from 'react';
import { SessionInspector, type UnifiedSessionInput } from '@flanksource/clicky-ui/ai';
import { Spinner } from '../../icons/Spinner';
import { attemptThreadSession, fetchTodoSessionDetail } from './TodoSessionDetail';

/**
 * The Session tab of one verification attempt: the agent thread that produced
 * the definition-of-done verdict. It is fetched on demand — the attempt list
 * polls attempts-only, so no transcript is loaded until a session is opened.
 */
export function VerificationAttemptSession({
  dir,
  todoRef,
  sessionId,
}: {
  dir: string;
  todoRef: string;
  sessionId: string;
}) {
  const [session, setSession] = useState<UnifiedSessionInput | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    let cancelled = false;
    setSession(null);
    setError('');
    fetchTodoSessionDetail(dir, todoRef, sessionId)
      .then((loaded) => {
        if (!cancelled) setSession(attemptThreadSession(loaded));
      })
      .catch((reason) => {
        if (!cancelled) setError(reason instanceof Error ? reason.message : String(reason));
      });
    return () => {
      cancelled = true;
    };
  }, [dir, todoRef, sessionId]);

  if (error) {
    return (
      <p role="alert" className="px-3 py-4 text-xs text-red-600">
        {error}
      </p>
    );
  }
  if (!session) {
    return (
      <p className="flex items-center gap-2 px-3 py-4 text-xs text-muted-foreground">
        <Spinner /> Loading session…
      </p>
    );
  }
  return <SessionInspector session={session} className="h-[28rem]" transcriptProps={{ showHeader: false, className: 'text-xs' }} />;
}
