import { useEffect, useState } from "react";
import type { DropdownMenuItem } from "@flanksource/clicky-ui/components";
import type { Comment, CommentMentionable } from "@flanksource/clicky-ui/comments";
import { UiRobotAi } from "@flanksource/clicky-ui/icons";
import { request } from "./api";

/** GET /api/agent-sessions: a live Claude or Codex session working on an edited file. */
export type AgentSession = {
  id: string;
  source: string;
  title: string;
  model?: string;
  cwd?: string;
  status?: string;
  inCmux: boolean;
};

/** The comments actor a session acts as once a thread is sent to it (editor/agents.go). */
export function agentOf(session: AgentSession): CommentMentionable {
  return { id: `${session.source}:${session.id}`, name: session.model || session.source, kind: "agent" };
}

/**
 * The Assign menu's agent entries for an open thread: "Send to <session>" per
 * live session (except the one already holding it, unless its last hand-off
 * failed), then "Run new agent session", which `send` gets without a session.
 */
export function sendToItems(
  comment: Comment,
  sessions: readonly AgentSession[],
  send: (sessionId?: string) => Promise<void>,
): DropdownMenuItem[] {
  if (comment.parentId) return [];
  const run = comment.agentRun;
  const items: DropdownMenuItem[] = sessions.map((session) => {
    const holding = run?.sessionId === session.id && run.state !== "error";
    return {
      label: `Send to ${session.title}`,
      icon: UiRobotAi,
      group: "Agent sessions",
      disabled: holding,
      title: holding ? `Already sent to ${session.title}` : `Deliver this thread into ${session.source} session ${session.id}${session.cwd ? ` in ${session.cwd}` : ""}`,
      onSelect: () => void send(session.id),
    };
  });
  const working = run !== undefined && run.state !== "error" && comment.status === "in_progress";
  items.push({
    label: "Run new agent session",
    icon: UiRobotAi,
    group: "Agent sessions",
    disabled: working,
    title: working ? "An agent session is already working this thread" : "Start a new agent session in this file's project to work the thread",
    onSelect: () => void send(),
  });
  return items;
}

/** Polls the live agent sessions while the page is visible. */
export function useAgentSessions(onError: (message: string) => void): AgentSession[] {
  const [sessions, setSessions] = useState<AgentSession[]>([]);
  useEffect(() => {
    let live = true;
    const refresh = async () => {
      try {
        const next = await request<AgentSession[]>("/api/agent-sessions");
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
  }, []);
  return sessions;
}
