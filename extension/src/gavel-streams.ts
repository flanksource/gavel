import { createEventHub, type EventSourceFactory } from "@flanksource/clicky-ui/hooks";

/**
 * The panel's SSE transport. Gavel serves browser streams only through its
 * event hub (a direct EventSource on a stream route is answered 410 Gone), and
 * the hub subscribes server-relative paths, so each absolute gavel URL the
 * panel's components open is handed to the hub as its path and query.
 */
export function gavelStreams({ server, hub }: { server: string; hub: EventSourceFactory }): EventSourceFactory {
  const origin = new URL(server).origin;
  return url => {
    let target: URL;
    try {
      target = new URL(url);
    } catch {
      throw new Error(`Stream ${url} is not an absolute URL on the gavel server ${origin}`);
    }
    if (target.origin !== origin) throw new Error(`Stream ${url} is not on the gavel server ${origin}`);
    return hub(`${target.pathname}${target.search}`);
  };
}

/** One hub connection to the gavel server's events endpoint. Its build id is unknown here, so it never reloads the panel. */
export function gavelEventHub(server: string): EventSourceFactory {
  return gavelStreams({ server, hub: createEventHub({ url: new URL("/api/events", server).href }) });
}
