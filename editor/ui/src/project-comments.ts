import { useCallback, useEffect, useRef, useState } from "react";
import { request } from "./api";
import type { CommentProject, ListedComment } from "./comments-page";

/** A generated entity list page: GET /api/v1/comments answers rows plus the paging window. */
type Paged<T> = { data: T[]; page: { limit: number; offset: number; total: number } };

/** The comments API's largest page. */
const PAGE_LIMIT = 200;
const POLL_MS = 5000;

const listUrl = (params: Record<string, string>) => `/api/v1/comments?${new URLSearchParams(params)}`;

/** Every thread under a project scope, paged through the comments API. */
export async function listScope(scope: string): Promise<ListedComment[]> {
  const rows: ListedComment[] = [];
  for (;;) {
    const page = await request<Paged<ListedComment>>(listUrl({ scope, limit: String(PAGE_LIMIT), offset: String(rows.length) }));
    rows.push(...page.data);
    if (rows.length >= page.page.total) return rows;
    if (page.data.length === 0) throw new Error(`comments for ${scope} stopped at ${rows.length} of ${page.page.total}`);
  }
}

/**
 * Polls `load` every 5s while the page is visible; `refresh` re-runs it now.
 * Stale answers are dropped; a failed poll keeps the last data beside its error.
 */
function usePolled<T>(load: (() => Promise<T>) | null) {
  const [value, setValue] = useState<{ data?: T; error?: string }>({});
  const generation = useRef(0);
  const refresh = useCallback(async () => {
    if (!load) return;
    const current = ++generation.current;
    try {
      const data = await load();
      if (current === generation.current) setValue({ data });
    } catch (cause) {
      if (current === generation.current) setValue((previous) => ({ ...previous, error: String(cause) }));
    }
  }, [load]);
  useEffect(() => {
    setValue({});
    if (!load) return;
    void refresh();
    const timer = setInterval(() => !document.hidden && void refresh(), POLL_MS);
    return () => {
      generation.current++;
      clearInterval(timer);
    };
  }, [refresh]);
  return { ...value, refresh };
}

/** GET /api/comments/projects, re-read whenever the edited files change (`filesKey`). */
export function useCommentProjects(filesKey: string): { projects?: CommentProject[]; error?: string } {
  const [state, setState] = useState<{ projects?: CommentProject[]; error?: string }>({});
  useEffect(() => {
    let current = true;
    void request<CommentProject[]>("/api/comments/projects")
      .then((projects) => current && setState({ projects }))
      .catch((cause) => current && setState({ error: `comment projects: ${String(cause)}` }));
    return () => {
      current = false;
    };
  }, [filesKey]);
  return state;
}

/** Unread threads across every project, from each scope's unread-filtered total. */
export function useProjectUnread(projects: readonly CommentProject[] | undefined, onError: (message: string) => void): number {
  const load = useCallback(async () => {
    const totals = await Promise.all(
      (projects ?? []).map((project) => request<Paged<ListedComment>>(listUrl({ scope: project.scope, unread: "true", limit: "1" }))),
    );
    return totals.reduce((sum, page) => sum + page.page.total, 0);
  }, [projects]);
  const { data, error } = usePolled(projects ? load : null);
  useEffect(() => {
    if (error) onError(`unread comments: ${error}`);
  }, [error]);
  return data ?? 0;
}

export type ProjectListing = { project: CommentProject; rows: ListedComment[] };

/** Every thread of every project, while `projects` is known; the page renders a failure itself. */
export function useProjectThreads(projects: readonly CommentProject[] | undefined) {
  const load = useCallback(
    () => Promise.all((projects ?? []).map(async (project): Promise<ProjectListing> => ({ project, rows: await listScope(project.scope) }))),
    [projects],
  );
  const { data, error, refresh } = usePolled(projects ? load : null);
  return { listings: data, error, refresh };
}
