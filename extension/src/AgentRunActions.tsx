import { Button } from "@flanksource/clicky-ui";
import type { CommentAgentRun } from "@flanksource/clicky-ui/comments";

/** Gavel hands off only active threads; a resolved or closed one must be reopened first. */
const FINAL_STATUSES = new Set(["resolved", "closed"]);

/** What a thread's recorded hand-off lets the user do next: retry a failed one, or send to a new session once it went through. */
export function AgentRunActions({ state, status, busy, onRetry, onRetryFresh, onSendFresh }: {
  state: CommentAgentRun["state"];
  /** The thread's workflow status; omitted means active. */
  status?: string;
  busy: boolean;
  onRetry: () => void;
  onRetryFresh: () => void;
  onSendFresh: () => void;
}) {
  if (state === "queued" || (status !== undefined && FINAL_STATUSES.has(status))) return null;
  return (
    <div className="mt-1 flex flex-wrap gap-2 px-1">
      {state === "error" ? (
        <>
          <Button size="sm" variant="outline" disabled={busy} onClick={onRetry}>Retry</Button>
          <Button size="sm" variant="outline" disabled={busy} onClick={onRetryFresh}>Retry in a new session</Button>
        </>
      ) : (
        <Button size="sm" variant="outline" disabled={busy} onClick={onSendFresh}>Send to a new session</Button>
      )}
    </div>
  );
}
