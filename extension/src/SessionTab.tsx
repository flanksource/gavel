import { Badge, Button, DropdownMenu } from "@flanksource/clicky-ui";
import { SessionInspector } from "@flanksource/clicky-ui/ai";
import { UiChevronDown, UiExternalLink, UiRobotAi } from "@flanksource/clicky-ui/icons";
import { AgentPicker, type AgentPickerProps } from "../agents/AgentPicker";
import type { AgentSession } from "./gavel-client";
import { captainSessionUrl, sessionDashboardUrl } from "./session-urls";

export type SessionTabProps = {
  server: string;
  /** The session threads are delivered to; `undefined` while the next send starts a new one. */
  sessionId: string | undefined;
  /** How many live sessions the Agent menu offers; unknown while gavel has not listed them. */
  liveCount: number | undefined;
  picker: AgentPickerProps;
};

/** The tab's own label: the session name and, once known, how full its context is. */
export function SessionTabLabel({ sessionId, session }: { sessionId: string | undefined; session: AgentSession | undefined }) {
  const context = session?.context;
  return (
    <span className="flex min-w-0 max-w-40 items-center gap-1.5">
      <span className="truncate">{sessionId ? (session?.title ?? sessionId) : "New session"}</span>
      {context && <span className="shrink-0 text-[10px] tabular-nums text-muted-foreground">{Math.round((context.usedTokens / context.windowTokens) * 100)}%</span>}
    </span>
  );
}

/** Where threads go: switch to a live session, start a new one, or change the model and auto-run. */
function AgentMenu({ liveCount, picker }: Pick<SessionTabProps, "liveCount" | "picker">) {
  return (
    <DropdownMenu
      role="dialog"
      align="right"
      menuLabel="Agent"
      menuClassName="w-96 max-w-[calc(100vw-16px)] max-h-[80dvh] overflow-auto"
      trigger={
        <Button variant="outline" size="sm" aria-label="Agent" title="Agent settings and active sessions">
          <UiRobotAi className="size-4" />
          Agent
          {liveCount !== undefined && <Badge variant="soft" clickToCopy={false}>{liveCount}</Badge>}
          <UiChevronDown className="size-3" />
        </Button>
      }
    >
      {close => (
        <AgentPicker
          {...picker}
          {...(picker.sessions
            ? {
                sessions: {
                  ...picker.sessions,
                  onSelect: session => { close(); picker.sessions!.onSelect(session); },
                  ...(picker.sessions.onNew ? { onNew: () => { close(); picker.sessions!.onNew!(); } } : {}),
                },
              }
            : {})}
        />
      )}
    </DropdownMenu>
  );
}

/** The delivery session's live transcript, with the Agent menu that switches it. */
export function SessionTab({ server, sessionId, liveCount, picker }: SessionTabProps) {
  const menu = <AgentMenu liveCount={liveCount} picker={picker} />;
  if (!sessionId) {
    return (
      <div className="flex flex-col items-start gap-3 p-3">
        {menu}
        <p className="text-sm text-muted-foreground">Send to agent starts a new session until you pick a live session from Agent to follow here. Auto-run always starts its own run and never uses the picked session.</p>
      </div>
    );
  }
  return (
    <SessionInspector
      key={sessionId}
      src={captainSessionUrl({ server, sessionId })}
      layout="compact"
      className="min-h-0 flex-1"
      transcriptProps={{ scrollable: true }}
      toolbarActions={
        <>
          {menu}
          <Button asChild variant="ghost" size="icon" className="size-8 shrink-0">
            <a href={sessionDashboardUrl({ server, sessionId })} target="_blank" rel="noreferrer" aria-label="Open session in gavel" title="Open session in gavel">
              <UiExternalLink className="size-4" aria-hidden />
            </a>
          </Button>
        </>
      }
    />
  );
}
