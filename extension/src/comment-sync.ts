import type { CommentCreateInput } from "@flanksource/clicky-ui/comments";
import type { CreateDeliveryChoice, DeliverTarget } from "../comments/gavel-delivery";
import { localUser } from "../comments/gavel-projection";
import { externalUrl, type Selection } from "../host/model";
import { GavelClient, GavelError, pageKey, screenshotBase64, type PageRef } from "./gavel-client";
import type { TodoStore } from "./todos/store";

export type CommentSubmission = {
  page: PageRef;
  input: CommentCreateInput;
  selection: Selection;
} & ({ delivery?: CreateDeliveryChoice; deliver?: never } | { deliver: DeliverTarget; delivery?: never });

export type CommentDestination = { server: string; project: string };
export type CommentReceipt = { id: string } & (
  | { state: "pending" | "error"; body: string; error?: string }
  | { state: "synced"; remoteId: string; adoptSession: boolean }
);
type QueueKey = CommentDestination & { id: string; pageKey: string; savedAt: number };
export type QueuedComment = QueueKey & (
  | { state: "pending" | "error"; draft: CommentSubmission; error?: string; adoptSession?: boolean }
  | { state: "synced"; remoteId: string; adoptSession: boolean }
);

const RECEIPT_LIFETIME_MS = 24 * 60 * 60 * 1000;
const REQUEST_TIMEOUT_MS = 15_000;

export class CommentSync {
  private readonly store: TodoStore;
  private readonly fetcher: typeof fetch;
  private running: Promise<void> | undefined;
  private requested = false;

  constructor({ store, fetcher }: { store: TodoStore; fetcher?: typeof fetch }) {
    this.store = store;
    this.fetcher = (input, init) => (fetcher ?? fetch)(input, { ...init, signal: AbortSignal.timeout(REQUEST_TIMEOUT_MS) });
  }

  async capture(destination: CommentDestination, draft: CommentSubmission): Promise<string> {
    const { input, page, selection, delivery, deliver } = draft;
    if (!input.body.trim()) throw new Error("A comment body is required");
    if (input.anchor !== selection.anchor || externalUrl(selection.element.pageUrl ?? "").url !== externalUrl(page.url).url)
      throw new Error("The comment selection no longer belongs to this page");
    if (delivery && deliver) throw new Error("A comment can only have one delivery choice");
    const screenshot = selection.element.screenshot;
    if (screenshot?.status === "captured") screenshotBase64(screenshot.dataUrl);
    const id = crypto.randomUUID();
    await this.store.putComment({ ...destination, id, pageKey: pageKey(page.url), savedAt: Date.now(), state: "pending", draft });
    return id;
  }

  async list(destination: CommentDestination, url: string): Promise<CommentReceipt[]> {
    const rows = await this.store.commentQueue({ ...destination, pageKey: pageKey(url) });
    return rows.map(row => row.state === "synced"
      ? { id: row.id, state: row.state, remoteId: row.remoteId, adoptSession: row.adoptSession }
      : { id: row.id, state: row.state, body: row.draft.input.body, ...(row.error ? { error: row.error } : {}) });
  }

  async retry(destination: CommentDestination, url: string, id: string): Promise<void> {
    const row = (await this.store.commentQueue({ ...destination, pageKey: pageKey(url) })).find(entry => entry.id === id);
    if (!row || row.state !== "error") throw new Error("This comment is not awaiting a retry");
    await this.store.putComment({ ...row, state: "pending" });
  }

  sync(): Promise<void> {
    this.requested = true;
    this.running ??= this.flush().finally(() => { this.running = undefined; });
    return this.running;
  }

  private async flush(): Promise<void> {
    while (this.requested) {
      this.requested = false;
      for (const row of await this.store.commentQueue()) {
        if (row.state === "pending") await this.deliver(row);
        if (row.state === "synced" && Date.now() - row.savedAt > RECEIPT_LIFETIME_MS) await this.store.removeComment(row.id);
      }
    }
  }

  private async deliver(row: Extract<QueuedComment, { draft: CommentSubmission }>): Promise<void> {
    const { id, draft: { input, ...draft }, server, project, pageKey, savedAt } = row;
    const client = new GavelClient({ server, project, fetch: this.fetcher });
    let adoptSession = row.adoptSession;
    let remoteId: string;
    try {
      if (adoptSession === undefined && !draft.deliver) {
        const preview = await client.previewDelivery({ page: draft.page, body: input.body, ...(draft.delivery ? { delivery: draft.delivery } : {}) });
        if (preview.route?.target.kind === "session" && !preview.session) throw new GavelError(409, "The delivery session is unavailable");
        adoptSession = preview.route?.target.kind === "run" && !preview.session && !preview.mention;
        await this.store.putComment({ ...row, adoptSession });
      }
      const created = await client.create({ ...draft, comment: {
        id, body: input.body, createdAt: new Date(savedAt).toISOString(), author: localUser, anchor: input.anchor ?? null,
        ...(input.rating ? { rating: input.rating } : {}), ...(input.assignee ? { assignee: input.assignee } : {}),
      } });
      remoteId = created.id;
    } catch (cause) {
      const rejected = cause instanceof GavelError && cause.status >= 400 && cause.status < 500 && ![408, 429].includes(cause.status);
      await this.store.putComment({ ...row, ...(adoptSession !== undefined ? { adoptSession } : {}), state: rejected ? "error" : "pending", error: cause instanceof Error ? cause.message : String(cause) });
      return;
    }
    await this.store.putComment({ id, server, project, pageKey, savedAt: Date.now(), state: "synced", remoteId, adoptSession: adoptSession ?? false });
  }
}
