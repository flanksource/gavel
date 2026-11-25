import { Button } from "@flanksource/clicky-ui";
import { SessionInspector } from "@flanksource/clicky-ui/ai";
import { AgentIdentity } from "@flanksource/clicky-ui/chat";
import { UiExternalLink } from "@flanksource/clicky-ui/icons";
import type { AgentSession } from "./gavel-client";
import { DeliveryTargetLine, type DeliveryPreviewState } from "./DeliveryPreview";
import { ampThreadUrl, captainSessionUrl, sessionDashboardUrl } from "./session-urls";
import styles from "./GavelPanel.module.css";

export type SessionTabProps = {
  server: string;
  /** The session threads are delivered to; `undefined` while the next send starts a new one. */
  sessionId: string | undefined;
  targetState?: DeliveryPreviewState;
};

/** The tab's own label: the session name and, once known, how full its context is. */
export function SessionTabLabel({ sessionId, session, label = "New session" }: { sessionId: string | undefined; session: AgentSession | undefined; label?: string }) {
  const context = session?.context;
  return (
    <span className="flex min-w-0 items-center gap-1.5" title={session?.title ?? sessionId ?? label}>
      {session && <AgentIdentity {...(session.source ? { provider: session.source } : {})} {...(session.model ? { model: session.model } : {})} iconOnly className="shrink-0" />}
      <span className={styles.sessionTitle}>{session?.title ?? sessionId ?? label}</span>
      {context && <span className="shrink-0 text-[10px] tabular-nums text-muted-foreground">{Math.round((context.usedTokens / context.windowTokens) * 100)}%</span>}
    </span>
  );
}

/** The delivery session's live transcript. */
export function SessionTab({ server, sessionId, targetState = {} }: SessionTabProps) {
  if (!sessionId) {
    return (
      <div className="flex flex-col items-start gap-3 p-3">
        <DeliveryTargetLine state={targetState} />
        <p className="text-sm text-muted-foreground">Choose an existing session in Settings to follow it here.</p>
      </div>
    );
  }
  const ampURL = ampThreadUrl(sessionId);
  if (ampURL) {
    return <div className="flex flex-col items-start gap-3 p-3">
      <DeliveryTargetLine state={targetState} />
      <Button asChild variant="outline" size="sm"><a href={ampURL} target="_blank" rel="noreferrer">Open in Amp</a></Button>
    </div>;
  }
  return (
    <SessionInspector
      key={sessionId}
      src={captainSessionUrl({ server, sessionId })}
      layout="compact"
      className="min-h-0 flex-1"
      transcriptProps={{ scrollable: true }}
      toolbarActions={
        <Button asChild variant="ghost" size="icon" className="size-8 shrink-0">
          <a href={sessionDashboardUrl({ server, sessionId })} target="_blank" rel="noreferrer" aria-label="Open session in gavel" title="Open session in gavel">
            <UiExternalLink className="size-4" aria-hidden />
          </a>
        </Button>
      }
    />
  );
}
