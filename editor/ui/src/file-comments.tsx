import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { SessionInspector, SessionTail } from "@flanksource/clicky-ui/ai";
import type { Comment, CommentAgentRun, CommentAuthor, CommentCreateInput } from "@flanksource/clicky-ui/comments";
import { Button, Modal } from "@flanksource/clicky-ui/components";
import { Badge } from "@flanksource/clicky-ui/data";
import type { FileReviewComments } from "@flanksource/clicky-ui/file-findings";
import type { Autosave } from "./autosave";
import { ApiError, request } from "./api";
import { sendToItems, useAgentSessions, type AgentSession, type PageAgentHandOff } from "./agent-sessions";
import { anchorOf, selectorOf, type Selector } from "./comment-anchors";
import { runOf, usePageAgent, type PageAgent, type ViewerChoice } from "./page-agent";
import { runTargetOf, type RunPreview, type RunRequest, type RunTarget } from "./run-request";

type Author = { label?: string; id: string; kind: "user" | "agent" };
export type NativeComment = {
  id: string;
  bodyMarkdown: string;
  version: number;
  status: string;
  createdAt: string;
  author: Author;
  assignee?: Author;
  /** Host-owned JSON; the editor keeps a captured-screenshot note and the agent hand-off here. */
  content?: { agentRun?: CommentAgentRun } | null;
  target: { selector: Selector; revision: string };
  messages?: { id: string; bodyMarkdown: string; createdAt: string; author: Author }[];
  attachments?: { id: string; name: string }[];
};
/** GET /api/file/{id}/comments: the file's threads under its project scope, and the target new ones use. */
type Listing = { comments: NativeComment[]; revision: string; scope: string; target: string };
const EMPTY_LISTING: Listing = { comments: [], revision: "", scope: "", target: "" };

export type ProjectInput = { name: string; dir: string };
/** A 409 no_project answer: the file is outside every gavel project, and the project to create for it. */
export type MissingProject = { error: string; suggestion: ProjectInput };

/** The comments FileReview renders, or the project the file needs before it can have any. */
export type FileCommentsState = {
  comments?: FileReviewComments;
  missingProject?: MissingProject;
  createProject: (project: ProjectInput) => Promise<void>;
  /** The thread "Run new agent session…" was chosen for, while its dialog is open. */
  newSession?: {
    thread: string;
    fileId: string;
    preview: (run: RunRequest) => Promise<RunPreview>;
    start: (run: RunRequest) => Promise<void>;
    cancel: () => void;
  };
  /** The page's configured agent, the viewer's hand-off choice and the live sessions it can pick from. */
  pageAgent?: {
    agent: PageAgent;
    choice: ViewerChoice;
    sessions: readonly AgentSession[];
    choose: (choice: ViewerChoice) => void;
    refresh: () => Promise<void>;
  };
};

/** A thread's session as one live line, expanding into the full clicky session inspector. */
function SessionPreview({ sessionId }: { sessionId: string }) {
  const [expanded, setExpanded] = useState(false);
  const src = `/api/captain/sessions/${encodeURIComponent(sessionId)}`;
  return (
    <>
      <div className="flex min-w-0 items-center gap-1">
        <SessionTail src={src} emptyLabel="Waiting for the session…" className="min-w-0 flex-1" />
        <Button size="sm" variant="ghost" onClick={() => setExpanded(true)} title="Show the full session">
          Expand
        </Button>
      </div>
      {expanded && (
        <Modal open onClose={() => setExpanded(false)} title="Agent session" size="full">
          <SessionInspector src={src} className="h-[75vh] min-h-0" layout="compact" />
        </Modal>
      )}
    </>
  );
}

export function missingProjectOf(cause: unknown): MissingProject | undefined {
  if (!(cause instanceof ApiError) || cause.status !== 409) return undefined;
  const body = cause.json<MissingProject & { code?: string }>();
  return body?.code === "no_project" ? { error: body.error, suggestion: body.suggestion } : undefined;
}

const authorOf = (author: Author): CommentAuthor => ({ id: author.id, name: author.label ?? author.id, kind: author.kind });

/** A native comment and its replies as FileReview items, the root carrying its assignee and agent hand-off. */
export function threadItems(row: NativeComment): Comment[] {
  const anchor = anchorOf(row.target.selector);
  const agentRun = row.content?.agentRun;
  return [
    {
      id: row.id,
      body: row.bodyMarkdown,
      status: row.status,
      createdAt: row.createdAt,
      anchor,
      author: authorOf(row.author),
      ...(row.assignee ? { assignee: authorOf(row.assignee) } : {}),
      ...(agentRun ? { agentRun } : {}),
    },
    ...(row.messages ?? []).map((message) => ({
      id: message.id,
      parentId: row.id,
      body: message.bodyMarkdown,
      createdAt: message.createdAt,
      anchor,
      author: authorOf(message.author),
    })),
  ];
}

const RUN_TONE = { queued: "info", sent: "info", done: "success", error: "danger" } as const;

/** "Sent to <session> via <channel> · sent", or the failure, for a thread handed to an agent. */
function AgentRunChip({ run }: { run: CommentAgentRun }) {
  const where = `Sent to ${run.target?.label ?? run.sessionId ?? "agent session"}${run.target?.channel ? ` via ${run.target.channel}` : ""}`;
  return (
    <Badge tone={RUN_TONE[run.state]} variant="soft" clickToCopy={false} className="max-w-full overflow-hidden [&>span]:truncate">
      <span title={run.state === "error" ? run.error : where}>{run.state === "error" ? `Send failed: ${run.error ?? "no error recorded"}` : `${where} · ${run.state}`}</span>
    </Badge>
  );
}

/**
 * The open document's native comments as FileReview threads: selections are
 * stored as git-style code ranges, picked elements as dom selectors, and file
 * comments as document selectors, with an optional screenshot attachment.
 */
export function useFileComments({
  fileId,
  queue,
  view,
  onError,
}: {
  fileId: string | null;
  queue: Autosave | undefined;
  view: string;
  onError: (message: string) => void;
}): FileCommentsState {
  const [listing, setListing] = useState<Listing>(EMPTY_LISTING);
  const [missingProject, setMissingProject] = useState<MissingProject | undefined>(undefined);
  const [newSessionFor, setNewSessionFor] = useState<Comment | undefined>(undefined);
  const generation = useRef(0);
  const pending = useRef<{ key: string; requestId: string; created?: NativeComment } | undefined>(undefined);

  const refresh = useCallback(async () => {
    if (!fileId) return;
    const current = ++generation.current;
    try {
      const result = await request<Listing>(`/api/file/${fileId}/comments`);
      if (current !== generation.current) return;
      setMissingProject(undefined);
      setListing(result);
    } catch (cause) {
      if (current !== generation.current) return;
      const missing = missingProjectOf(cause);
      if (missing) setMissingProject(missing);
      else onError(String(cause));
    }
  }, [fileId]);

  useEffect(() => {
    setListing(EMPTY_LISTING);
    setMissingProject(undefined);
    void refresh();
    const timer = setInterval(() => !document.hidden && void refresh(), 5000);
    return () => {
      generation.current++;
      clearInterval(timer);
    };
  }, [refresh]);
  const revision = queue?.state().revision;
  useEffect(() => void refresh(), [revision, refresh]);

  const sessions = useAgentSessions(listing.scope, onError);
  const page = usePageAgent(listing.scope ? fileId : null, onError);
  const items = useMemo<Comment[]>(() => listing.comments.flatMap(threadItems), [listing]);

  const createProject = async (project: ProjectInput) => {
    await request("/api/projects", project);
    await refresh();
  };
  if (!fileId || !queue || missingProject) return { missingProject, createProject };

  const act = async (id: string, action: string, body: Record<string, unknown> = {}) => {
    const root = listing.comments.find((row) => row.id === id || row.messages?.some((message) => message.id === id));
    if (!root) throw new Error("Comment no longer exists; refresh comments");
    await request(`/api/v1/comments/${root.id}/${action}`, { expectedVersion: root.version, ...body });
    await refresh();
  };

  // The shared comments delivery: the thread at its current version, to a live session or a new run.
  const deliver = async (commentId: string, target: { kind: "session"; sessionId: string } | RunTarget) => {
    const root = listing.comments.find((row) => row.id === commentId);
    if (!root) throw new Error("Comment no longer exists; refresh comments");
    await request(`/api/v1/comments/${commentId}/deliver`, { expectedVersion: root.version, requestId: crypto.randomUUID(), target });
  };
  // A failed hand-off is recorded on the thread (agentRun.state "error") as well as reported.
  const send = async (commentId: string, sessionId: string) => {
    try {
      await deliver(commentId, { kind: "session", sessionId });
    } catch (cause) {
      onError(String(cause));
    } finally {
      await refresh();
    }
  };
  // "Send to page agent": the viewer's live session for the page, else a new run of its configured agent.
  const pageHandOff = (commentId: string): PageAgentHandOff | undefined => {
    const effective = page.agent?.effective;
    if (page.choice.kind === "session") {
      const sessionId = page.choice.sessionId;
      const session = sessions.find((candidate) => candidate.id === sessionId);
      return { label: session?.title ?? sessionId, send: () => send(commentId, sessionId) };
    }
    if (!effective) return undefined;
    return {
      label: effective.target.prompt,
      send: async () => {
        try {
          await deliver(commentId, runTargetOf(runOf(effective)));
        } catch (cause) {
          onError(String(cause));
        } finally {
          await refresh();
        }
      },
    };
  };
  // A new managed session in the file's project; the dialog reports a failed launch itself.
  const startRun = async (commentId: string, run: RunRequest) => {
    try {
      await deliver(commentId, runTargetOf(run));
      setNewSessionFor(undefined);
    } finally {
      await refresh();
    }
  };
  const newSession = newSessionFor && {
    thread: newSessionFor.body,
    fileId,
    preview: (run: RunRequest) => request<RunPreview>(`/api/v1/comments/${newSessionFor.id}/deliver/preview`, { target: runTargetOf(run) }),
    start: (run: RunRequest) => startRun(newSessionFor.id, run),
    cancel: () => setNewSessionFor(undefined),
  };

  const uploadScreenshot = async (created: NativeComment, dataUrl: string) => {
    const latest = await request<NativeComment>(`/api/v1/comments/${created.id}`);
    if (latest.attachments?.some((attachment) => attachment.name === "screenshot.png")) return;
    const response = await fetch(
      `/api/v1/comments/${created.id}/attachments?name=screenshot.png&expectedVersion=${latest.version}`,
      { method: "POST", headers: { "Content-Type": "image/png" }, body: await (await fetch(dataUrl)).blob() },
    );
    if (!response.ok) throw new Error(`Comment saved but screenshot upload failed: ${await response.text()}`);
  };

  const create = async (input: CommentCreateInput) => {
    await queue.flush();
    if (queue.state().status !== "saved") throw new Error("Save the file before commenting");
    if (!listing.scope) throw new Error("Comments are still loading; try again");
    const selector = selectorOf(input.anchor, view);
    const key = JSON.stringify({ selector, body: input.body });
    if (pending.current?.key !== key) pending.current = { key, requestId: crypto.randomUUID() };
    const attempt = pending.current;
    attempt.created ??= await request<NativeComment>("/api/v1/comments", {
      target: { uri: listing.target, scopeUri: listing.scope, kind: selector.kind, revision: queue.state().revision, selector },
      bodyMarkdown: input.body,
      workflowId: "human-review",
      requestId: attempt.requestId,
      ...(input.screenshot?.status === "unavailable" ? { content: { screenshot: input.screenshot } } : {}),
    });
    try {
      if (input.screenshot?.status === "captured") await uploadScreenshot(attempt.created, input.screenshot.dataUrl);
    } finally {
      await refresh();
    }
    pending.current = undefined;
  };

  const pageAgent = page.agent && { agent: page.agent, choice: page.choice, sessions, choose: page.choose, refresh: page.refresh };
  return { createProject, ...(newSession ? { newSession } : {}), ...(pageAgent ? { pageAgent } : {}), comments: {
    items,
    anchoring: "patch",
    targets: { element: true, document: true, screenshot: true },
    callbacks: {
      onCreate: create,
      onReply: (input) => act(input.parentId, "reply", { bodyMarkdown: input.body }),
      onUpdateStatus: (id, status) => act(id, status === "open" ? "reopen" : status === "resolved" ? "resolve" : status),
      onClose: (id) => act(id, "close"),
      onAcknowledge: (id) => act(id, "acknowledge"),
      onAssign: (id, assignee) =>
        act(id, "assign", { assignee: assignee && { kind: assignee.kind ?? "user", id: assignee.id, label: assignee.name } }),
      onDelete: async (id) => {
        if (window.confirm("Delete this comment and its replies?")) await act(id, "delete");
      },
    },
    extraAssignItems: (comment) => {
      const handOff = pageHandOff(comment.id);
      return sendToItems(comment, sessions, {
        send: (sessionId) => send(comment.id, sessionId),
        startNew: () => setNewSessionFor(comment),
        ...(handOff ? { pageAgent: handOff } : {}),
      });
    },
    renderRootDetails: (comment) => {
      const row = listing.comments.find((candidate) => candidate.id === comment.id);
      if (!row) return null;
      const shot = row.attachments?.find((attachment) => attachment.name === "screenshot.png");
      const url = shot && `/api/v1/comments/${row.id}/attachments/${shot.id}`;
      const run = comment.agentRun;
      if (!url && !run) return null;
      return (
        <div className="mt-2 flex min-w-0 flex-col gap-1.5">
          {url && (
            <a href={url} target="_blank" rel="noreferrer" className="block">
              <img src={url} alt="Screenshot" className="max-h-40 rounded border border-border" />
            </a>
          )}
          {run && <AgentRunChip run={run} />}
          {run?.sessionId && run.state !== "error" && <SessionPreview sessionId={run.sessionId} />}
        </div>
      );
    },
  } };
}
