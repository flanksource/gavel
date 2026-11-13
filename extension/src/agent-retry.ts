import type { CommentAgentRun } from "@flanksource/clicky-ui/comments";
import type { DeliverTarget } from "../comments/gavel-delivery";

type RunTarget = Extract<DeliverTarget, { kind: "run" }>;

/**
 * Where to re-deliver a thread whose hand-off failed: the session the failed run
 * recorded when the panel sends to a session, otherwise the panel's run target
 * (gavel resolves its group's session from the same `groupKey`).
 */
export function retryTarget(options: { run: CommentAgentRun; delivery: DeliverTarget }): DeliverTarget {
  const { run, delivery } = options;
  return delivery.kind === "session" ? { kind: "session", sessionId: run.sessionId ?? delivery.sessionId } : delivery;
}

/** The run target that skips the group's recorded session and starts a new one. */
export function freshTarget(runTarget: DeliverTarget): RunTarget {
  if (runTarget.kind !== "run") throw new Error("A new session can only be started from a run target; a session target has no group to restart");
  return { ...runTarget, fresh: true };
}
