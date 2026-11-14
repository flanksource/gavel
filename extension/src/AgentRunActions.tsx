import { Button } from "@flanksource/clicky-ui";
import type { CommentAgentRun } from "@flanksource/clicky-ui/comments";

/** Gavel hands off only active threads; a resolved or closed one must be reopened first. */
const FINAL_STATUSES = new Set(["resolved", "closed"]);

/** Recovery actions for a failed hand-off; successful delivery is shown in the header badge. */
export function AgentRunActions({ state, status, busy, onRetry, onRetryFresh }: {
  state: CommentAgentRun["state"];
  /** The thread's workflow status; omitted means active. */
  status?: string;
  busy: boolean;
  onRetry: () => void;
  onRetryFresh: () => void;
}) {
  if (state !== "error" || (status !== undefined && FINAL_STATUSES.has(status))) return null;
  return (
    <div className="mt-1 flex flex-wrap gap-2 px-1">
      <Button size="sm" variant="outline" disabled={busy} onClick={onRetry}>Retry</Button>
      <Button size="sm" variant="outline" disabled={busy} onClick={onRetryFresh}>Retry in a new session</Button>
    </div>
  );
}
