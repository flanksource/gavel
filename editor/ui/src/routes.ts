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
