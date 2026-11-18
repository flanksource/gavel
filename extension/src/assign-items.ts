import { isUnresolved, type Comment, type DropdownMenuItem } from "@flanksource/clicky-ui";
import { shortContextModelName } from "@flanksource/clicky-ui/chat";
import { agentGlyph } from "@flanksource/clicky-ui/comments";
import { UiSliders } from "@flanksource/clicky-ui/icons";
import type { AgentSessionRuntime } from "../contracts/agent-session-store";
import type { AgentSession } from "./gavel-client";
import { GAVEL_COMMENT_CONFIG } from "./useGavelComments";

const ENDED = "ended";

/** Live sessions first, then ended managed ones; each group keeps gavel's order. */
export function orderedSessions(sessions: AgentSession[]): AgentSession[] {
  return [...sessions.filter(session => session.status !== ENDED), ...sessions.filter(session => session.status === ENDED)];
}

/** An ended session is resumed by delivering to it, so its label says so. */
export function sessionLabel(session: AgentSession): string {
  return session.status === ENDED ? `${session.title} · resumes` : session.title;
}

/** Delivering to an ended session that gavel cannot resume would fail, so it is not offered. */
const unavailable = (session: AgentSession) => session.status === ENDED && !session.resumable;

/** The open thread roots of a page: the threads gavel will still hand off. */
export function openRoots(comments: Comment[]): Comment[] {
  return comments.filter(comment => !comment.parentId && isUnresolved(GAVEL_COMMENT_CONFIG, comment.status));
}

const newSessionLabel = (runtime: AgentSessionRuntime | undefined) => {
  const model = runtime?.spec.model;
  return model ? `New ${shortContextModelName(model)}` : "New session (Gavel's default)";
};

export type AgentAssignOptions = {
  /** Every session gavel offers for this page, live and ended. */
  sessions: AgentSession[];
  /** The session the thread is already with; it is not offered again. */
  currentSessionId: string | undefined;
  /** The runtime a new session starts with; `undefined` is gavel's default. */
  runtime: AgentSessionRuntime | undefined;
  /** The thread is no longer open, so gavel will not hand it off. */
  disabled: boolean;
  onSession: (session: AgentSession) => void;
  onNew: () => void;
  onAdvanced: () => void;
};

/** The agent entries of a thread's Assign… menu: any other session (ended ones resume), a new session with the configured runtime, or one picked for this thread. */
export function agentAssignItems({ sessions, currentSessionId, runtime, disabled, onSession, onNew, onAdvanced }: AgentAssignOptions): DropdownMenuItem[] {
  const model = runtime?.spec.model;
  return [
    ...orderedSessions(sessions)
      .filter(session => session.id !== currentSessionId)
      .map(session => ({
        label: sessionLabel(session),
        icon: agentGlyph(session.model ?? session.source).icon,
        title: session.status === ENDED ? `Resumes the ended session ${session.title} with this thread` : `Deliver this thread into the session ${session.title}`,
        disabled: disabled || unavailable(session),
        onSelect: () => onSession(session),
      })),
    {
      label: newSessionLabel(runtime),
      icon: agentGlyph(model ?? "").icon,
      title: "Start a new agent session for this thread with the configured runtime",
      disabled,
      onSelect: onNew,
    },
    {
      label: "Advanced…",
      icon: UiSliders,
      title: "Choose the model and settings of a new session for this thread",
      disabled,
      onSelect: onAdvanced,
    },
  ];
}

/** Open roots eligible for the current target. Failed hand-offs can be tried again. */
export function assignableRoots(threads: Comment[], sessionId: string | undefined): Comment[] {
  return openRoots(threads).filter(thread => {
    const run = thread.agentRun;
    if (run?.state === "queued") return false;
    return !sessionId || run?.sessionId !== sessionId || run.state === "error";
  });
}
