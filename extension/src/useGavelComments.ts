import { useCallback, useEffect, useRef, useState } from "react";
import {
  DEFAULT_COMMENT_STATUSES,
  type CommentAuthor,
  type CommentConfig,
  type CommentCreateInput,
  type CommentReplyInput,
} from "@flanksource/clicky-ui/comments";
import type { StoredComment } from "../contracts/comments-store";
import { commentActor, localUser } from "../comments/gavel-projection";
import type { Selection } from "../host/model";
import type { DeliverTarget, GavelClient, PageRef } from "./gavel-client";

export const GAVEL_COMMENT_CONFIG: CommentConfig = {
  statuses: DEFAULT_COMMENT_STATUSES,
  // Agents are not listed: assigning a thread to one delivers it, which the panel's own Assign… entries do.
  assignees: [{ id: "you", name: localUser.name, kind: "user" }],
};

const REFRESH_DEBOUNCE_MS = 200;
const RECONNECT_MS = 3000;

function describe(cause: unknown): string {
  return cause instanceof Error ? cause.message : String(cause);
}

// The event cursor only spares replaying the project's whole history on every
// open; storage that is unavailable just means starting from the beginning.
function readCursor(key: string): number {
  try {
    const stored = window.localStorage.getItem(key);
    return stored === null ? 0 : Number(stored) || 0;
  } catch {
    return 0;
  }
}

function writeCursor(key: string, sequence: number) {
  try {
    window.localStorage.setItem(key, String(sequence));
  } catch {
    // See readCursor.
  }
}

export type GavelComments = {
  comments: StoredComment[];
  loading: boolean;
  /** A broken backend or event stream must never look like "no comments". */
  error: string | null;
  create: (input: CommentCreateInput, selection: Selection) => Promise<void>;
  reply: (input: CommentReplyInput) => Promise<void>;
  updateStatus: (id: string, status: string) => Promise<void>;
  acknowledge: (id: string) => Promise<void>;
  assign: (id: string, assignee: CommentAuthor | null) => Promise<void>;
  close: (id: string) => Promise<void>;
  remove: (id: string) => Promise<void>;
  /** Hands a thread to an agent and returns gavel's recorded hand-off. */
  deliver: (id: string, target: DeliverTarget) => Promise<StoredComment>;
  /** Hands several open thread roots to one target in a single batch; gavel refuses the whole batch if one is stale or closed. */
  deliverAll: (ids: string[], target: DeliverTarget) => Promise<StoredComment[]>;
};

/**
 * The comments of one page held in gavel: loaded through the client, kept
 * current by the project's event stream, mutated with the version the thread
 * was last seen at.
 */
export function useGavelComments(options: { client: GavelClient; page: PageRef | undefined }): GavelComments {
  const { client, page } = options;
  const pageKey = page?.url;
  const [stored, setStored] = useState<StoredComment[]>([]);
  const [loading, setLoading] = useState(page !== undefined);
  const [error, setError] = useState<string | null>(null);
  const [streamError, setStreamError] = useState<string | null>(null);
  const latest = useRef(stored);
  latest.current = stored;
  const pageRef = useRef(page);
  pageRef.current = page;

  const refresh = useCallback(async () => {
    const current = pageRef.current;
    if (!current) return;
    try {
      const comments = await client.list(current);
      if (pageRef.current?.url !== current.url) return;
      setStored(comments);
      setError(null);
    } catch (cause) {
      setError(describe(cause));
    } finally {
      setLoading(false);
    }
  }, [client]);

  useEffect(() => {
    setStored([]);
    setLoading(pageKey !== undefined);
    void refresh();
  }, [refresh, pageKey]);

  useEffect(() => {
    const controller = new AbortController();
    const cursorKey = `gavel.events.${client.scope}`;
    let after = readCursor(cursorKey);
    let timer: number | undefined;
    const scheduled = () => {
      window.clearTimeout(timer);
      timer = window.setTimeout(() => void refresh(), REFRESH_DEBOUNCE_MS);
    };
    void (async () => {
      while (!controller.signal.aborted) {
        try {
          await client.streamEvents({
            after,
            signal: controller.signal,
            onOpen: () => setStreamError(null),
            onEvent: event => {
              after = event.sequence;
              writeCursor(cursorKey, after);
              scheduled();
            },
          });
        } catch (cause) {
          if (controller.signal.aborted) return;
          setStreamError(`Live updates stopped: ${describe(cause)}`);
        }
        await new Promise(resolve => window.setTimeout(resolve, RECONNECT_MS));
      }
    })();
    return () => {
      controller.abort();
      window.clearTimeout(timer);
    };
  }, [client, refresh]);

  const root = (id: string): StoredComment => {
    const comment = latest.current.find(entry => entry.id === id);
    if (!comment) throw new Error(`Comment ${id} is not loaded; refresh and try again`);
    if (comment.parentId) throw new Error("Only thread roots can change lifecycle stage");
    if (comment.version === undefined) throw new Error(`Comment ${id} has no gavel version`);
    return comment;
  };
  const mutate = async (work: () => Promise<unknown>) => {
    await work();
    await refresh();
  };
  const requirePage = (): PageRef => {
    if (!pageRef.current) throw new Error("The grabber has not reported the page yet");
    return pageRef.current;
  };
  const act = (id: string, action: string, extra: { assignee?: ReturnType<typeof commentActor> | null } = {}) => {
    const comment = root(id);
    return mutate(() => client.act({ id, version: comment.version!, action, ...extra }));
  };

  return {
    comments: stored,
    loading,
    error: error ?? streamError,
    create: async (input, selection) => {
      const comment: StoredComment = {
        id: crypto.randomUUID(),
        body: input.body,
        createdAt: new Date().toISOString(),
        author: localUser,
        anchor: input.anchor ?? null,
        ...(input.rating ? { rating: input.rating } : {}),
        ...(input.assignee ? { assignee: input.assignee } : {}),
      };
      await mutate(() => client.create({ page: requirePage(), comment, selection }));
    },
    reply: async input => {
      const parent = latest.current.find(entry => entry.id === input.parentId);
      if (!parent) throw new Error(`Comment ${input.parentId} is not loaded; refresh and try again`);
      const rootId = parent.rootId ?? parent.id;
      const thread = root(rootId);
      await mutate(() => client.reply({ page: requirePage(), root: rootId, version: thread.version!, body: input.body }));
    },
    updateStatus: (id, status) => {
      if (status === "resolved") return act(id, "resolve");
      if (status === "open") return act(id, "reopen");
      throw new Error(`Status ${JSON.stringify(status)} cannot be set directly; use acknowledge, assign, resolve or reopen`);
    },
    acknowledge: id => act(id, "acknowledge"),
    assign: (id, assignee) => act(id, "assign", { assignee: assignee ? commentActor(assignee) : null }),
    close: id => act(id, "close"),
    remove: async id => {
      const comment = latest.current.find(entry => entry.id === id);
      if (!comment) throw new Error(`Comment ${id} is not loaded; refresh and try again`);
      if (!comment.parentId) return act(id, "delete");
      const thread = root(comment.rootId ?? comment.parentId);
      await mutate(() => client.batch([{ id: thread.id, expectedVersion: thread.version!, action: "delete-reply", messageId: id }]));
    },
    deliver: async (id, target) => {
      const comment = root(id);
      try {
        return await client.deliver({ id, version: comment.version!, requestId: crypto.randomUUID(), target });
      } finally {
        // A failed delivery is recorded on the thread's agentRun, so it is reloaded either way.
        await refresh();
      }
    },
    deliverAll: async (ids, target) => {
      const threads = ids.map(id => ({ id, expectedVersion: root(id).version! }));
      try {
        return await client.deliverBatch({ requestId: crypto.randomUUID(), target, threads });
      } finally {
        await refresh();
      }
    },
  };
}
