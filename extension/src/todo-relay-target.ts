import { validateParentOrigin } from "../host/bridge-parse";
import { bookmarkletParentOrigin, matchesBookmarkletParent } from "../host/origin";
import { serverOrigin } from "./settings";

/**
 * The todo relay page is web-accessible, so any page can frame it. It only
 * accepts capture parameters for the activated Gavel server and the page
 * actually framing it. The URL supplies initial values; no hosted form loads.
 */
export function todoRelayTarget(options: { target: string | null; parentOrigin: string | null; ancestorOrigin: string | undefined; server: string }): { url: URL; parentOrigin: string } {
  if (!options.target) throw new Error("Todo relay requires a target form URL");
  if (!options.parentOrigin) throw new Error("Todo relay requires the origin of the page that frames it");
  if (!options.ancestorOrigin) throw new Error("Todo relay must be framed by a page");
  validateParentOrigin(options.parentOrigin);
  if (!matchesBookmarkletParent({ expected: options.parentOrigin, received: options.ancestorOrigin }))
    throw new Error(`Todo relay belongs to ${options.parentOrigin}, not the framing page ${options.ancestorOrigin}`);
  const url = new URL(options.target);
  if (url.username || url.password || url.origin !== serverOrigin(options.server))
    throw new Error(`Todo relay target ${url.origin} is not the configured gavel server ${options.server}`);
  if (url.pathname !== "/todos/new" || url.searchParams.get("embed") !== "1") throw new Error(`Todo relay target is not the gavel new-todo form: ${url.href}`);
  const sourceUrl = url.searchParams.get("sourceUrl");
  if (sourceUrl && bookmarkletParentOrigin(sourceUrl) !== options.parentOrigin) throw new Error(`Todo relay source URL belongs to ${bookmarkletParentOrigin(sourceUrl)}, not ${options.parentOrigin}`);
  return { url, parentOrigin: options.parentOrigin };
}
