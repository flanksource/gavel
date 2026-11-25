import { useMemo, useState, type ComponentProps } from 'react';
import { SessionChatComposer, SessionInspector, type SessionInspectorLoadState } from '@flanksource/clicky-ui/ai';
import type { AgentSession } from '../../types';
import { useSessionMessageComposer } from './sessionChat';
import { sessionCollection } from './sessionCollection';

// Why the always-present composer is disabled until the inspector reports the
// session loaded: messaging a session the user cannot see would be blind.
const COMPOSER_DISABLED_REASON = {
  loading: 'Loading session…',
  error: 'This session cannot be read',
} as const;

/**
 * A session's transcript with a single-line message bar to continue it. The
 * composer goes in the inspector's `composer` slot, which pins it (padded by
 * the bottom safe area) below the scrolling transcript in every load state, so
 * the input never disappears: while loading or after a load failure it is
 * rendered disabled with the reason instead. The transcript is followed while
 * the session is live or a message just sent to it is running.
 */
export function SessionChatInspector({ session, sessions, hostTabs }: {
  session: AgentSession;
  /** Every listed session, so the inspector can offer this one's sub-sessions. */
  sessions: AgentSession[];
  hostTabs?: ComponentProps<typeof SessionInspector>['hostTabs'];
}) {
  const { follow, send, ...composerProps } = useSessionMessageComposer(session);
  const [load, setLoad] = useState<SessionInspectorLoadState>({ status: 'loading' });
  const disabledReason = load.status === 'ready' ? undefined : COMPOSER_DISABLED_REASON[load.status];
  const collection = useMemo(() => sessionCollection(session, sessions), [session, sessions]);
  return (
    <SessionInspector
      session={collection}
      follow={follow}
      layout="compact"
      className="min-h-0 flex-1 text-xs"
      onLoadStateChange={setLoad}
      {...(hostTabs ? { hostTabs } : {})}
      composer={<SessionChatComposer {...composerProps} {...(disabledReason ? { disabledReason } : {})} onSubmit={send} />}
    />
  );
}
