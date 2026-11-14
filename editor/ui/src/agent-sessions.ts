import { useEffect, useState } from "react";
import type { DropdownMenuItem } from "@flanksource/clicky-ui/components";
import type { Comment, CommentMentionable } from "@flanksource/clicky-ui/comments";
import { UiRobotAi } from "@flanksource/clicky-ui/icons";
import { request } from "./api";

/** GET /api/v1/comments/agent-sessions: a live Claude or Codex session working in the file's project. */
export type AgentSession = {
  id: string;
  source: string;
  title: string;
  model?: string;
  cwd?: string;
  status?: string;
  inCmux: boolean;
};

/** The comments actor a session acts as once a thread is sent to it (comments/agents/session.go). */
export function agentOf(session: AgentSession): CommentMentionable {
  return { id: `${session.source}:${session.id}`, name: session.model || session.source, kind: "agent" };
}

/** The page's agent as the viewer chose to reach it: its label, and the hand-off that reaches it. */
export type PageAgentHandOff = { label: string; send: () => Promise<void> };

/**
 * The Assign menu's agent entries for an open thread: "Send to page agent"
 * when the page has one, "Send to <session>" per live session (except the one
 * already holding it, unless its last hand-off failed), then "Run new agent
 * session…", which opens the new-session dialog.
 */
export function sendToItems(
  comment: Comment,
  sessions: readonly AgentSession[],
  { send, startNew, pageAgent }: { send: (sessionId: string) => Promise<void>; startNew: () => void; pageAgent?: PageAgentHandOff },
): DropdownMenuItem[] {
  if (comment.parentId) return [];
  const run = comment.agentRun;
  const working = run !== undefined && run.state !== "error" && comment.status === "in_progress";
  const items: DropdownMenuItem[] = pageAgent
    ? [
        {
          label: `Send to page agent: ${pageAgent.label}`,
          icon: UiRobotAi,
          group: "Agent sessions",
          disabled: working,
          title: working ? "An agent session is already working this thread" : "Hand this thread to the agent configured for this page",
          onSelect: () => void pageAgent.send(),
        },
      ]
    : [];
  items.push(...sessions.map((session) => {
    const holding = run?.sessionId === session.id && run.state !== "error";
    return {
      label: `Send to ${session.title}`,
      icon: UiRobotAi,
      group: "Agent sessions",
      disabled: holding,
      title: holding ? `Already sent to ${session.title}` : `Deliver this thread into ${session.source} session ${session.id}${session.cwd ? ` in ${session.cwd}` : ""}`,
      onSelect: () => void send(session.id),
    };
  }));
  items.push({
    label: "Run new agent session…",
    icon: UiRobotAi,
    group: "Agent sessions",
    disabled: working,
    title: working ? "An agent session is already working this thread" : "Choose a runtime and start a new agent session in this file's project to work the thread",
    onSelect: startNew,
  });
  return items;
}

/** Polls the live agent sessions working in the scope's project while the page is visible; none until the scope is known. */
export function useAgentSessions(scope: string, onError: (message: string) => void): AgentSession[] {
  const [sessions, setSessions] = useState<AgentSession[]>([]);
  useEffect(() => {
    setSessions([]);
    if (!scope) return;
    let live = true;
    const refresh = async () => {
      try {
        const next = await request<AgentSession[]>(`/api/v1/comments/agent-sessions?scope=${encodeURIComponent(scope)}`);
        if (live) setSessions(next);
      } catch (cause) {
        if (live) onError(String(cause));
      }
    };
    void refresh();
    const timer = setInterval(() => !document.hidden && void refresh(), 10000);
    return () => {
      live = false;
      clearInterval(timer);
    };
  }, [scope]);
  return sessions;
}
