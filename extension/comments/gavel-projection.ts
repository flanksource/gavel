import type { StoredAgentRun, StoredAuthor, StoredComment, StoredProposal } from "../contracts/comments-store";

/*
 * Pure projection between gavel's native comments and the playground's comment
 * shape. Browser safe: the playground server and the Chrome extension's gavel
 * client both import it, so the mapping lives in one place.
 */

export type Actor = { kind: string; id: string; label?: string };
export type Target = {
  uri: string;
  scopeUri: string;
  kind: string;
  label?: string;
  selector?: unknown;
};
type NativeMessage = {
  id: string;
  bodyMarkdown: string;
  author: Actor;
  createdAt: string;
  /** An agent's reply may carry only `proposal`; the playground fields are the web UI's own. */
  content?: Partial<Content> & { proposal?: NativeProposal };
};
/** A patch an agent proposed in a reply, as gavel stores it under the message's `content.proposal`. */
export type NativeProposal = Omit<StoredProposal, "decidedBy"> & { decidedBy?: Actor };
/** The hand-off gavel records under `content.agentRun`, in clicky-ui's CommentAgentRun shape. */
export type NativeAgentRun = {
  sessionId?: string;
  requestedAt: string;
  state: string;
  error?: string;
  target?: { label: string; channel?: string };
};
export type Content = {
  playground: Partial<StoredComment> & { page: string };
  /** The web page a browser extension commented on. */
  page?: { url: string; title: string };
  agentRun?: NativeAgentRun;
};
export type NativeComment = {
  id: string;
  bodyMarkdown: string;
  author: Actor;
  target: Target;
  status: string;
  /** Who holds the thread; a delivery assigns the agent it hands the thread to. */
  assignee?: Actor | null;
  version: number;
  createdAt: string;
  updatedAt: string;
  content: Content;
  messages?: NativeMessage[];
  attachments?: { id: string; name: string }[];
  outcome?: { at: string; actor: Actor };
};
export type Action = {
  id: string;
  expectedVersion: number;
  action: string;
  bodyMarkdown?: string;
  messageId?: string;
  content?: Content;
  target?: Target;
  assignee?: Actor | null;
  /** `accept-proposal`: apply the proposed patch to the working tree or stage it in the index. */
  mode?: "apply" | "stage";
};

const AGENT_RUN_STATES: readonly StoredAgentRun["state"][] = ["queued", "sent", "done", "error"];

export function commentActor(author: StoredAuthor): Actor {
  return {
    kind: author.kind ?? "user",
    id: author.id ?? `${author.kind ?? "user"}:${author.name}`,
    label: author.name,
  };
}
export const localUser: StoredAuthor = { name: "You", kind: "user" };
export const bodyFor = (comment: StoredComment) =>
  comment.body.trim()
    ? comment.body
    : comment.reviewDecision === "approved"
      ? "Approved"
      : comment.rating === "positive"
        ? "Positive feedback"
        : "Negative feedback";
export function contentFor(page: string, comment: StoredComment): Content {
  const {
    id: _id,
    version: _version,
    rootId: _rootId,
    createdAt: _created,
    updatedAt: _updated,
    status: _status,
    parentId: _parent,
    ...fields
  } = comment;
  return { playground: { ...fields, page } };
}
function authorFor(actor: Actor): StoredAuthor {
  return {
    id: actor.id,
    name: actor.label ?? actor.id,
    kind: actor.kind === "agent" ? "agent" : "user",
  };
}
function proposalFor(proposal: NativeProposal): StoredProposal {
  const { decidedBy, ...rest } = proposal;
  return { ...rest, ...(decidedBy ? { decidedBy: authorFor(decidedBy) } : {}) };
}
/** The first reply an agent wrote in the thread at or after the hand-off was requested. */
function firstAgentReply(messages: NativeMessage[], requestedAt: string): string | undefined {
  const since = Date.parse(requestedAt);
  const times = messages
    .filter((m) => m.author.kind === "agent" && Date.parse(m.createdAt) >= since)
    .map((m) => Date.parse(m.createdAt));
  return times.length > 0 ? new Date(Math.min(...times)).toISOString() : undefined;
}
function agentRunFor(run: NativeAgentRun, messages: NativeMessage[]): StoredAgentRun {
  if (!AGENT_RUN_STATES.includes(run.state as StoredAgentRun["state"]))
    throw new Error(`Unknown agent run state ${JSON.stringify(run.state)}; expected ${AGENT_RUN_STATES.join(", ")}`);
  const repliedAt = firstAgentReply(messages, run.requestedAt);
  return {
    ...(repliedAt ? { repliedAt } : {}),
    ...(run.sessionId ? { sessionId: run.sessionId } : {}),
    requestedAt: run.requestedAt,
    state: run.state as StoredAgentRun["state"],
    ...(run.error ? { error: run.error } : {}),
    ...(run.target ? { target: run.target } : {}),
  };
}

/** `attachmentUrl` is where the host serves a comment's attachment: the playground proxies it, the extension reads gavel directly. */
export function project(
  c: NativeComment,
  options: { attachmentUrl: (commentId: string, attachmentId: string) => string },
): StoredComment[] {
  if (!c.content?.playground?.page)
    throw new Error(`Comment ${c.id} is missing playground metadata`);
  // Gavel owns the assignee; a copy left in the playground fields is never the current one.
  const { page: _page, assignee: _assignee, ...fields } = c.content.playground;
  const capture = c.attachments?.find((a) => a.name === "screenshot.png");
  const root: StoredComment = {
    ...fields,
    id: c.id,
    version: c.version,
    body: fields.body === "" ? "" : c.bodyMarkdown,
    author: fields.author ?? authorFor(c.author),
    ...(c.assignee ? { assignee: authorFor(c.assignee) } : {}),
    status: c.status,
    createdAt: new Date(c.createdAt).toISOString(),
    ...(c.updatedAt !== c.createdAt
      ? { updatedAt: new Date(c.updatedAt).toISOString() }
      : {}),
    ...(c.outcome
      ? {
          closedAt: new Date(c.outcome.at).toISOString(),
          closedBy: fields.closedBy ?? authorFor(c.outcome.actor),
        }
      : {}),
    ...(capture && fields.element
      ? {
          element: {
            ...fields.element,
            screenshot: {
              status: "captured",
              filename: "screenshot.png",
              url: options.attachmentUrl(c.id, capture.id),
            },
          },
        }
      : {}),
    ...(c.content.agentRun ? { agentRun: agentRunFor(c.content.agentRun, c.messages ?? []) } : {}),
  };
  return [
    root,
    ...(c.messages ?? []).map((m) => ({
      ...m.content?.playground,
      ...(m.content?.proposal ? { proposal: proposalFor(m.content.proposal) } : {}),
      id: m.id,
      rootId: c.id,
      version: c.version,
      body: m.bodyMarkdown,
      author: m.content?.playground?.author ?? authorFor(m.author),
      createdAt: new Date(m.createdAt).toISOString(),
      parentId: c.id,
      anchor: root.anchor ?? null,
    })),
  ];
}
