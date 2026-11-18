import { useCallback, useEffect, useRef, useState } from "react";
import type { Comment } from "@flanksource/clicky-ui";
import { ownLayerOf } from "@flanksource/clicky-ui/comments/gavel";
import type { CommentAgentPatch } from "../comments/gavel-delivery";
import type { RuntimeCatalog } from "../agents/runtime-families";
import { commentAgentFor, fetchRuntimeCatalog, pageTargetUri, type AgentSession, type CommentAgentSettings, type CommentEdits, type GavelClient, type PageRef, type RunTarget } from "./gavel-client";

const SESSION_POLL_MS = 5000;

function describe(cause: unknown): string {
  return cause instanceof Error ? cause.message : String(cause);
}

/**
 * The agent sessions threads can be delivered to (live, plus ended managed ones
 * that a delivery resumes), refreshed while the panel is visible and on demand
 * through `refresh`, which settles once the reload has been applied.
 */
export function useAgentSessions(client: GavelClient): { sessions: AgentSession[] | undefined; error: string | null; refresh: () => Promise<void> } {
  const [sessions, setSessions] = useState<AgentSession[]>();
  const [error, setError] = useState<string | null>(null);
  const generation = useRef(0);
  const refresh = useCallback(async () => {
    const mine = ++generation.current;
    try {
      const next = await client.agentSessions();
      if (mine === generation.current) { setSessions(next); setError(null); }
    } catch (cause) {
      if (mine === generation.current) setError(describe(cause));
    }
  }, [client]);
  useEffect(() => {
    const update = () => { if (document.visibilityState === "visible") void refresh(); };
    void refresh();
    const timer = window.setInterval(update, SESSION_POLL_MS);
    window.addEventListener("focus", update);
    return () => {
      generation.current++;
      window.clearInterval(timer);
      window.removeEventListener("focus", update);
    };
  }, [refresh]);
  return { sessions, error, refresh };
}

type WatchedThread = { id: string; requestedAt: string | undefined };

/**
 * Gavel records a run's session on the delivered thread after the deliver
 * request has answered, so it arrives with the thread's comment stream, not in
 * the response. `watch` the delivered thread ids and `adopt` is called once with
 * the first session recorded by that hand-off (a hand-off is recognised by its
 * `requestedAt` differing from the one the thread had when watched, so a session
 * an earlier run left is never adopted). A thread whose hand-off failed stops
 * being watched; `cancel` stops all.
 */
export function useAdoptRunSession(options: { comments: Pick<Comment, "id" | "agentRun">[]; adopt: (sessionId: string) => void }) {
  const { comments } = options;
  const [watched, setWatched] = useState<WatchedThread[]>([]);
  const latest = useRef(comments);
  latest.current = comments;
  const adopt = useRef(options.adopt);
  adopt.current = options.adopt;

  useEffect(() => {
    if (watched.length === 0) return;
    const recorded = new Map(comments.map(comment => [comment.id, comment.agentRun]));
    const handedOff = (thread: WatchedThread) => {
      const run = recorded.get(thread.id);
      return run && run.requestedAt !== thread.requestedAt ? run : undefined;
    };
    const adopted = watched.map(handedOff).find(run => run?.sessionId && run.state !== "error");
    if (adopted?.sessionId) {
      setWatched([]);
      adopt.current(adopted.sessionId);
      return;
    }
    const pending = watched.filter(thread => handedOff(thread)?.state !== "error");
    if (pending.length !== watched.length) setWatched(pending);
  }, [comments, watched]);

  const watch = useCallback((ids: string[]) => {
    setWatched(ids.map(id => ({ id, requestedAt: latest.current.find(comment => comment.id === id)?.agentRun?.requestedAt })));
  }, []);
  const watchNew = useCallback((id: string) => setWatched([{ id, requestedAt: undefined }]), []);
  const cancel = useCallback(() => setWatched([]), []);
  return { watch, watchNew, cancel };
}

/** The models and agent runtimes a new session can start with, loaded once from the dashboard. */
export function useRuntimeCatalog(server: string): { catalog: RuntimeCatalog | undefined; error: string | null } {
  const [catalog, setCatalog] = useState<RuntimeCatalog>();
  const [error, setError] = useState<string | null>(null);
  useEffect(() => {
    let cancelled = false;
    fetchRuntimeCatalog({ server }).then(
      next => { if (!cancelled) setCatalog(next); },
      (cause: unknown) => { if (!cancelled) setError(describe(cause)); },
    );
    return () => { cancelled = true; };
  }, [server]);
  return { catalog, error };
}

/** Reads Gavel's effective rule and patches only the settings the user changes. */
export function useAutoRun(options: { client: GavelClient; page: PageRef | undefined; run: RunTarget | undefined }) {
  const { client, page, run } = options;
  const prefix = page ? pageTargetUri(client.scope, page) : undefined;
  const [loaded, setLoaded] = useState<{ settings: CommentAgentSettings; prefix: string }>();
  const [error, setError] = useState<string | null>(null);
  const generation = useRef(0);
  const revision = useRef(0);
  const writes = useRef(Promise.resolve());
  const current = loaded?.prefix === prefix ? loaded?.settings : undefined;
  const agent = current?.agent;

  useEffect(() => {
    const mine = ++generation.current;
    setLoaded(undefined);
    setError(null);
    if (!page || !prefix) return;
    void writes.current.then(async () => {
      if (generation.current !== mine) return;
      try {
        const settings = await client.commentAgent(page);
        if (generation.current === mine) setLoaded({ settings, prefix });
      } catch (cause) {
        if (generation.current === mine) setError(describe(cause));
      }
    });
    return () => { generation.current++; };
  }, [client, prefix]);

  const save = (patch: CommentAgentPatch) => {
    if (!page || !prefix || !current) throw new Error("The page's comment agent settings have not loaded");
    const mine = generation.current;
    const change = ++revision.current;
    writes.current = writes.current.then(async () => {
      if (generation.current !== mine) return;
      try {
        const settings = await client.saveCommentAgent({ page, layer: ownLayerOf(current.settings), agent: patch });
        if (generation.current !== mine || revision.current !== change) return;
        setLoaded({ settings, prefix });
        setError(null);
      } catch (cause) {
        if (generation.current === mine && revision.current === change) setError(describe(cause));
      }
    });
    return writes.current;
  };

  const set = (on: boolean) => {
    if (!on || !run) return save({ enabled: on });
    const existing = agent?.agent ?? {};
    const fields = Object.entries(commentAgentFor(run)).filter(([field]) => field === "enabled" || existing[field as keyof typeof existing] === undefined);
    return save(Object.fromEntries(fields));
  };

  return {
    agent,
    enabled: agent?.agent.enabled === true,
    edits: agent?.agent.edits ?? "direct",
    ready: current !== undefined,
    error,
    set,
    setEdits: (edits: CommentEdits) => save({ edits }),
  };
}
