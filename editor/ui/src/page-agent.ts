import { useCallback, useEffect, useState } from "react";
import { request } from "./api";
import type { RunRequest } from "./run-request";

/** A comment route of the shared comments API (GET/PUT /api/v1/comments/routes), as the editor's .gavel.yaml adapter projects it. */
export type CommentRoute = {
  id: string;
  scopeUri: string;
  targetPrefix: string;
  workflowId: string;
  target: { kind: "run"; prompt: string; runOptions?: Omit<RunRequest, "prompt"> };
  enabled: boolean;
  assign?: boolean;
};

/** GET /api/file/{id}/agent: the routes configured for the page and its folder, the one that applies, and the .gavel.yaml saving writes. */
export type PageAgent = {
  scope: string;
  pagePrefix: string;
  folderPrefix: string;
  configPath: string;
  page: CommentRoute | null;
  folder: CommentRoute | null;
  effective: CommentRoute | null;
};

/** Where a saved agent applies: this page only, or every page in its folder. */
export type AgentScope = "page" | "folder";

/** The route PUT /api/v1/comments/routes saves for scope: the run's prompt, every other run choice as its runOptions. */
export function routeBodyOf({
  agent,
  scope,
  enabled,
  assign,
  run,
}: {
  agent: PageAgent;
  scope: AgentScope;
  enabled: boolean;
  assign: boolean;
  run: RunRequest;
}): Omit<CommentRoute, "id"> {
  const { prompt, ...runOptions } = run;
  if (!prompt) throw new Error("a page agent needs a prompt to run");
  return {
    scopeUri: agent.scope,
    targetPrefix: scope === "page" ? agent.pagePrefix : agent.folderPrefix,
    workflowId: "human-review",
    target: { kind: "run", prompt, ...(Object.keys(runOptions).length > 0 ? { runOptions } : {}) },
    enabled,
    assign,
  };
}

/** The run a route starts, in the new-session dialog's request shape. */
export function runOf(route: CommentRoute): RunRequest {
  return { prompt: route.target.prompt, ...(route.target.runOptions ?? {}) };
}

/** Where the effective route comes from, relative to the page. */
export function originOf(agent: PageAgent): "page" | "folder" | "inherited" | "none" {
  const effective = agent.effective;
  if (!effective) return "none";
  if (effective.targetPrefix === agent.pagePrefix) return "page";
  if (effective.targetPrefix === agent.folderPrefix) return "folder";
  return "inherited";
}

/** The toolbar label: the prompt that works new threads and how it is reached, or that none does. */
export function agentLabel(agent: PageAgent): string {
  const effective = agent.effective;
  if (!effective) return "No page agent";
  if (!effective.enabled) return `Page agent off (${originOf(agent)})`;
  return `${effective.target.prompt} · auto${effective.assign ? " + assign" : ""} (${originOf(agent)})`;
}

/** A viewer's own hand-off choice for a page: a new run of the page's configured agent, or a live session. */
export type ViewerChoice = { kind: "run" } | { kind: "session"; sessionId: string };

const VIEWER_CHOICE = "gavel-edit:page-agent:";

export function loadViewerChoice(pagePrefix: string): ViewerChoice {
  try {
    const stored = window.localStorage.getItem(VIEWER_CHOICE + pagePrefix);
    const parsed: unknown = stored ? JSON.parse(stored) : null;
    if (parsed && typeof parsed === "object" && (parsed as ViewerChoice).kind === "session" && typeof (parsed as { sessionId?: unknown }).sessionId === "string")
      return parsed as ViewerChoice;
    return { kind: "run" };
  } catch {
    // A per-viewer convenience only: unreadable storage hands off to the configured run.
    return { kind: "run" };
  }
}

export function saveViewerChoice(pagePrefix: string, choice: ViewerChoice) {
  try {
    window.localStorage.setItem(VIEWER_CHOICE + pagePrefix, JSON.stringify(choice));
  } catch {
    // Storage may be blocked; the choice still applies until the page reloads.
  }
}

/** The page's agent configuration and the viewer's hand-off choice; none until a document is open. */
export function usePageAgent(fileId: string | null, onError: (message: string) => void) {
  const [agent, setAgent] = useState<PageAgent>();
  const [choice, setChoice] = useState<ViewerChoice>({ kind: "run" });
  const refresh = useCallback(async () => {
    if (!fileId) return;
    try {
      const next = await request<PageAgent>(`/api/file/${fileId}/agent`);
      setAgent(next);
      setChoice(loadViewerChoice(next.pagePrefix));
    } catch (cause) {
      onError(String(cause));
    }
  }, [fileId]);
  useEffect(() => {
    setAgent(undefined);
    void refresh();
  }, [refresh]);
  const choose = (next: ViewerChoice) => {
    if (!agent) throw new Error("the page agent is still loading");
    saveViewerChoice(agent.pagePrefix, next);
    setChoice(next);
  };
  return { agent, choice, choose, refresh };
}
