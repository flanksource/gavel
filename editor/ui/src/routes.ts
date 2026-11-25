/** The Comments page's filters (`?comments`), each "" (or false) when unset. */
export type CommentsFilters = {
  status: string;
  phase: string;
  author: string;
  project: string;
  unread: boolean;
  q: string;
};

export const NO_COMMENTS_FILTERS: CommentsFilters = { status: "", phase: "", author: "", project: "", unread: false, q: "" };

const COMMENTS_FILTER_KEYS = ["status", "phase", "author", "project", "q"] as const;

/** The Comments page filters a query string names, or undefined when it is not the Comments page. */
export function commentsRouteFrom(params: URLSearchParams): CommentsFilters | undefined {
  if (!params.has("comments")) return undefined;
  const filters = { ...NO_COMMENTS_FILTERS, unread: params.get("unread") === "1" };
  for (const key of COMMENTS_FILTER_KEYS) filters[key] = params.get(key) ?? "";
  return filters;
}

/** Writes the Comments page and its set filters onto `params`; unset filters leave no parameter. */
export function writeCommentsRoute(params: URLSearchParams, filters: CommentsFilters) {
  params.set("comments", "");
  for (const key of COMMENTS_FILTER_KEYS) if (filters[key]) params.set(key, filters[key]);
  if (filters.unread) params.set("unread", "1");
}

/**
 * The task run a task-manager link selects ("" for the full list). clicky's
 * TaskManagerButton links to `/tasks` and `/tasks/<id>`; the editor routes
 * through its query string instead, so each link is mapped here.
 */
export function taskRunFromHref(href: string): string {
  const match = /^\/tasks(?:\/([^/?#]+))?$/.exec(href);
  if (!match) throw new Error(`no editor page for task link ${JSON.stringify(href)}`);
  return match[1] ? decodeURIComponent(match[1]) : "";
}
