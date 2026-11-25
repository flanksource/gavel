import type { ReactNode } from 'react';
import { SessionInspector } from '@flanksource/clicky-ui/ai';
import { Button } from '@flanksource/clicky-ui/components';
import { UiArrowLeft } from '@flanksource/clicky-ui/icons';
import { Spinner } from '../../icons/Spinner';
import { captainSessionUrl } from '../todos/TodoSessionDetail';
import { SessionActivityIcon } from './SessionList';
import type { SessionsQueryView } from './SessionsView';
import { attentionReasons } from './sessionView';
import { SessionChatInspector } from './SessionChatInspector';

function BackButton({ onBack }: { onBack: () => void }) {
  return (
    <Button
      variant="ghost"
      size="icon"
      type="button"
      onClick={onBack}
      className="h-8 w-8 shrink-0 rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
      title="Back to sessions"
      aria-label="Back to sessions"
    >
      <UiArrowLeft className="text-base" />
    </Button>
  );
}

function ScreenFrame({ header, children }: { header: ReactNode; children: ReactNode }) {
  return (
    <div className="flex h-dvh flex-col overflow-hidden bg-background text-foreground">
      <div className="flex h-11 shrink-0 items-center gap-2 border-b border-border px-2">{header}</div>
      {children}
    </div>
  );
}

/**
 * The compact layout's full-screen session: a back header (as the PR detail
 * has), then Captain's transcript — followed while the session is live or a
 * message just sent to it is running — with a composer to continue it.
 */
export function MobileSessionScreen({ query, sessionId, onBack }: {
  query: SessionsQueryView;
  sessionId: string;
  onBack: () => void;
}) {
  const session = query.data?.sessions.find(candidate => candidate.id === sessionId);
  if (!session) {
    return (
      <ScreenFrame header={<BackButton onBack={onBack} />}>
        {query.error && !query.data ? (
          <div role="alert" className="whitespace-pre-wrap p-3 text-xs text-red-600">{query.error.message}</div>
        ) : !query.data ? (
          <div role="status" className="flex items-center gap-2 p-3 text-xs text-muted-foreground"><Spinner /> Loading session…</div>
        ) : (
          <SessionInspector key={sessionId} src={captainSessionUrl(sessionId)} className="min-h-0 flex-1 text-xs" />
        )}
      </ScreenFrame>
    );
  }
  const reasons = attentionReasons(session);
  return (
    <ScreenFrame
      header={(
        <>
          <BackButton onBack={onBack} />
          <SessionActivityIcon session={session} />
          <div className="min-w-0 flex-1">
            <div className="truncate text-sm font-semibold" title={session.title}>{session.title}</div>
            <div className="truncate text-[11px] text-muted-foreground">
              {session.stateReason || reasons.join(' · ') || [session.project, session.worktree?.branch].filter(Boolean).join(' · ')}
            </div>
          </div>
        </>
      )}
    >
      <SessionChatInspector key={session.id} session={session} sessions={query.data?.sessions ?? []} />
    </ScreenFrame>
  );
}
