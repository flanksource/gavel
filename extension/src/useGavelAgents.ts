import { useCallback, useEffect, useRef, useState } from "react";
import type { Comment } from "@flanksource/clicky-ui";
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
    const adopted = watched.map(handedOff).find(run => run?.sessionId);
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
  const cancel = useCallback(() => setWatched([]), []);
  return { watch, cancel };
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

const PROJECT_LAYER = { kind: "project" } as const;

/**
 * Whether gavel starts an agent run for every new human comment on the current
 * page: the page's comment agent in gavel's settings, saved at the project
 * layer. It describes a run, never a live session. Turning it off removes the
 * page entry, or, when an inherited default would still enable the page, saves
 * an explicit disabled agent. While it is on, a changed run target is re-saved.
 */
export function useAutoRun(options: { client: GavelClient; page: PageRef | undefined; run: RunTarget | undefined }) {
  const { client, page, run } = options;
  const prefix = page ? pageTargetUri(client.scope, page) : undefined;
  const runKey = JSON.stringify(run ?? null);
  const [agent, setAgent] = useState<CommentAgentSettings["agent"]>();
  const enabled = agent?.agent.enabled === true;
  const [ready, setReady] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const synced = useRef(runKey);

  useEffect(() => {
    if (!page) return;
    let cancelled = false;
    setReady(false);
    setAgent(undefined);
    client.commentAgent(page).then(
      settings => {
        if (cancelled) return;
        setAgent(settings.agent);
        setError(null);
        setReady(true);
      },
      (cause: unknown) => { if (!cancelled) setError(describe(cause)); },
    );
    return () => { cancelled = true; };
  }, [client, prefix]);

  const edits = agent?.agent.edits ?? "direct";
  const set = useCallback(async (on: boolean, chosenEdits: CommentEdits | undefined = agent?.agent.edits) => {
    if (!page || !run) throw new Error("The grabber has not reported the page yet");
    try {
      let saved = await client.saveCommentAgent({ page, layer: PROJECT_LAYER, agent: on ? { ...commentAgentFor(run), ...(chosenEdits ? { edits: chosenEdits } : {}) } : null });
      if (!on && saved.agent.agent.enabled === true) saved = await client.saveCommentAgent({ page, layer: PROJECT_LAYER, agent: { enabled: false } });
      setAgent(saved.agent);
      setError(null);
    } catch (cause) {
      setError(describe(cause));
    }
  }, [client, page, run, agent]);

  /** How an enabled page's agent changes the project: edits it directly or proposes patches. Saving while off would turn auto-run on. */
  const setEdits = useCallback(async (next: CommentEdits) => {
    if (!enabled) throw new Error("Cannot choose how the agent edits while auto-run is off; turn auto-run on first");
    await set(true, next);
  }, [enabled, set]);

  useEffect(() => {
    if (synced.current === runKey) return;
    synced.current = runKey;
    if (enabled && ready) void set(true);
  }, [runKey]);

  return { agent, enabled, edits, ready, error, set, setEdits };
}
