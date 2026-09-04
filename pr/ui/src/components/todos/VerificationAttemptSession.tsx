import { useQuery } from '@tanstack/react-query';
import { SessionInspector } from '@flanksource/clicky-ui/ai';
import { Spinner } from '../../icons/Spinner';
import { attemptThreadSession } from './TodoSessionDetail';
import { sessionDetailQueryOptions } from './todoQueries';

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
  const query = useQuery({
    ...sessionDetailQueryOptions(dir, todoRef, sessionId, false, 1_500),
    select: attemptThreadSession,
  });

  if (query.error) {
    return (
      <p role="alert" className="px-3 py-4 text-xs text-red-600">
        {query.error.message}
      </p>
    );
  }
  if (!query.data) {
    return (
      <p className="flex items-center gap-2 px-3 py-4 text-xs text-muted-foreground">
        <Spinner /> Loading session…
      </p>
    );
  }
  return <SessionInspector session={query.data} className="h-[28rem]" transcriptProps={{ showHeader: false, className: 'text-xs' }} />;
}
