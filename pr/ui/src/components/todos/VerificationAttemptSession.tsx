import { SessionInspector } from '@flanksource/clicky-ui/ai';
import { captainSessionUrl } from './TodoSessionDetail';

/**
 * The Session tab of one verification attempt: the agent thread that produced
 * the definition-of-done verdict, loaded from Captain's session handler only
 * once the tab is opened.
 */
export function VerificationAttemptSession({ sessionId }: { sessionId: string }) {
  return <SessionInspector src={captainSessionUrl(sessionId)} className="h-[28rem]" transcriptProps={{ showHeader: false, className: 'text-xs' }} />;
}
