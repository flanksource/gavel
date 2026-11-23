import type { BackgroundRequest } from "../messages";
import { serverOrigin } from "../settings";
import type { TodoSync } from "./sync";

export function todoRequest({ request, sender, todos }: { request: BackgroundRequest; sender: chrome.runtime.MessageSender; todos: TodoSync }): Promise<unknown> | undefined {
  if (typeof request?.type !== "string" || !request.type.startsWith("todo-")) return undefined;
  if (!sender.url?.startsWith(chrome.runtime.getURL(""))) return Promise.reject(new Error("Todo storage is only available to extension pages"));
  return handle(request, todos);
}

async function handle(request: BackgroundRequest, todos: TodoSync): Promise<unknown> {
  if (!("server" in request)) throw new Error("Todo request requires a server");
  const server = serverOrigin(request.server);
  if (request.type === "todo-snapshot") return todos.snapshot(server, request.dir, request.sourceUrl, request.status);
  if (request.type === "todo-refresh") return todos.refresh(server, request.dir);
  if (request.type === "todo-runtime-refresh") return todos.refreshRun(server, request.dir, request.status);
  if (request.type === "todo-run-choice") return todos.rememberRun(server, request.dir, request.run);
  if (request.type === "todo-capture") {
    const saved = await todos.capture(server, request.capture);
    // The caller only waits for durable storage. Delivery outlives its dialog.
    void todos.sync().catch(cause => console.error("Todo sync failed", cause));
    return saved;
  }
  if (request.type === "todo-sync") return request.retryId ? todos.retry(server, request.retryId) : todos.sync();
  throw new Error("Unknown todo request");
}

export function warmTodoCatalog(server: string, todos: TodoSync): void {
  void todos.refresh(serverOrigin(server), "").catch(cause => console.error("Todo catalog refresh failed; saved choices retained", cause));
}
