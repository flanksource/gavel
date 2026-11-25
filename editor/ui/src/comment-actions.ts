import type { CommentCallbacks } from "@flanksource/clicky-ui/comments";
import type { GitDiffPayload } from "@flanksource/clicky-ui/data";
import { request } from "./api";

/** The thread fields an action needs: the root posted to, at the version it was read at, and its replies. */
type ThreadRow = { id: string; version: number; messages?: readonly { id: string }[] };

export type ThreadCallbacks = Required<
  Pick<
    CommentCallbacks,
    | "onReply"
    | "onUpdateStatus"
    | "onClose"
    | "onAcknowledge"
    | "onAssign"
    | "onDelete"
    | "onApplyProposal"
    | "onRejectProposal"
    | "onRequestRevision"
    | "loadProposalDiff"
  >
>;

/**
 * The comments API actions on existing threads: each posts to the thread root at
 * the version `rows` holds, then `refresh` re-reads them. A proposal decision
 * names the reply holding the patch; the root is the comment it is posted to.
 */
export function threadCallbacks(rows: readonly ThreadRow[], refresh: () => Promise<void>): ThreadCallbacks {
  const act = async (id: string, action: string, body: Record<string, unknown> = {}) => {
    const root = rows.find((row) => row.id === id || row.messages?.some((message) => message.id === id));
    if (!root) throw new Error("Comment no longer exists; refresh comments");
    await request(`/api/v1/comments/${root.id}/${action}`, { expectedVersion: root.version, ...body });
    await refresh();
  };
  const rootOfReply = (id: string) => {
    const root = rows.find((row) => row.messages?.some((message) => message.id === id));
    if (!root) throw new Error("The proposing reply no longer exists; refresh comments");
    return root;
  };
  return {
    onReply: (input) => act(input.parentId, "reply", { bodyMarkdown: input.body }),
    onUpdateStatus: (id, status) => act(id, status === "open" ? "reopen" : status === "resolved" ? "resolve" : status),
    onClose: (id) => act(id, "close"),
    onAcknowledge: (id) => act(id, "acknowledge"),
    onAssign: (id, assignee) =>
      act(id, "assign", { assignee: assignee && { kind: assignee.kind ?? "user", id: assignee.id, label: assignee.name } }),
    onDelete: async (id) => {
      if (window.confirm("Delete this comment and its replies?")) await act(id, "delete");
    },
    onApplyProposal: (id, mode) => act(id, "accept-proposal", { messageId: id, mode }),
    onRejectProposal: (id, reason) => act(id, "reject-proposal", { messageId: id, ...(reason ? { bodyMarkdown: reason } : {}) }),
    onRequestRevision: (id, feedback) => act(id, "request-revision", { messageId: id, bodyMarkdown: feedback }),
    loadProposalDiff: (id) => request<GitDiffPayload>(`/api/v1/comments/${rootOfReply(id).id}/proposals/${id}/diff`),
  };
}
