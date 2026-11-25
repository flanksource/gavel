import { isUnresolved, type Comment, type DropdownMenuItem } from "@flanksource/clicky-ui";
import { commentAgentAssignmentItems, type CommentAgentAssignmentOptions } from "@flanksource/clicky-ui/comments/gavel";
import type { AgentSessionRuntime } from "../contracts/agent-session-store";
import { GAVEL_COMMENT_CONFIG } from "./useGavelComments";

export { orderedCommentAgentSessions as orderedSessions, commentAgentSessionLabel as sessionLabel } from "@flanksource/clicky-ui/comments/gavel";

/** The open thread roots of a page: the threads gavel will still hand off. */
export function openRoots(comments: Comment[]): Comment[] {
  return comments.filter(comment => !comment.parentId && isUnresolved(GAVEL_COMMENT_CONFIG, comment.status));
}

type AgentAssignOptions = Omit<CommentAgentAssignmentOptions, "model"> & {
  runtime: AgentSessionRuntime | undefined;
  onAmp?: () => void;
};

export function agentAssignItems({ runtime, onAmp, ...options }: AgentAssignOptions): DropdownMenuItem[] {
  const items = commentAgentAssignmentItems({ ...options, model: runtime?.spec.model });
  if (onAmp) items.push({ label: "Send to Amp thread…", disabled: Boolean(options.disabled), onSelect: onAmp });
  return items;
}

/** Open roots eligible for the current target. Failed hand-offs can be tried again. */
export function assignableRoots(threads: Comment[], sessionId: string | undefined): Comment[] {
  return openRoots(threads).filter(thread => {
    const run = thread.agentRun;
    if (run?.state === "queued") return false;
    return !sessionId || run?.sessionId !== sessionId || run.state === "error";
  });
}
