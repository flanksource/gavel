import { SessionInspector } from '@flanksource/clicky-ui/ai';
import { AgentIdentity } from '@flanksource/clicky-ui/chat';
import { Button } from '@flanksource/clicky-ui/components';
import type { GavelAgentSession, NativeAgentRun } from '@flanksource/clicky-ui/comments/gavel';
import { UiExternalLink } from '@flanksource/clicky-ui/icons';
import { captainSessionUrl } from '../todos/TodoSessionDetail';

export function CommentSession({ run, session, local }: { run?: NativeAgentRun | undefined; session?: GavelAgentSession | undefined; local: boolean }) {
  const sessionId = run?.sessionId;
  if (!sessionId) return <p className="text-sm text-muted-foreground">{run?.state === 'queued' ? 'Waiting for the agent session to start…' : 'No agent session is linked to this thread.'}</p>;
  if (run?.target?.channel === 'amp-http') return <div className="flex items-center gap-2">
    <AgentIdentity label={session?.title ?? run.target.label} provider="amp" />
    <Button asChild size="sm" variant="outline">
      <a href={`https://ampcode.com/threads/${encodeURIComponent(sessionId)}`} target="_blank" rel="noreferrer">Open in Amp</a>
    </Button>
  </div>;

  return <div className="flex h-[min(30rem,60dvh)] min-h-0 flex-col gap-2">
    <div className="flex items-center justify-between gap-2">
      <AgentIdentity label={session?.title ?? run?.target?.label ?? 'Agent session'} identityId={sessionId}
        {...(session?.source ? { provider: session.source } : {})} {...(session?.model ? { model: session.model } : {})} />
      {local && <Button asChild size="icon" variant="ghost">
        <a href={`/sessions/${encodeURIComponent(sessionId)}`} target="_blank" rel="noreferrer" aria-label="Open session in Gavel"><UiExternalLink /></a>
      </Button>}
    </div>
    <SessionInspector key={sessionId} src={captainSessionUrl(sessionId)} layout="compact" className="min-h-0 flex-1" transcriptProps={{ scrollable: true }} />
  </div>;
}
