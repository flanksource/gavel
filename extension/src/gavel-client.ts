import type { GitDiffPayload } from "@flanksource/clicky-ui/data";
import type { CommentMentionable } from "@flanksource/clicky-ui/comments";
import type { AgentSessionGrouping, AgentSessionRuntime } from "../contracts/agent-session-store";
import type { StoredComment } from "../contracts/comments-store";
import { externalUrl, type Selection } from "../host/model";
import {
  bodyFor,
  contentFor,
  localUser,
  project,
  type Action,
  type Content,
  type NativeComment,
  type Target,
} from "../comments/gavel-projection";
import {
  DEFAULT_RUN_PROMPT,
  type CommentAgent,
  type CommentAgentPatch,
  type CommentAgentSettings,
  type CommentEdits,
  type CreateDeliveryChoice,
  type CreateDeliveryPreview,
  type DeliverBatchRequest,
  type DeliverTarget,
  type GavelAgentSession,
  type SaveCommentAgentRequest,
  type SettingsRef,
} from "../comments/gavel-delivery";

/*
 * Browser client for gavel's comments API (`/api/v1/comments`). The extension
 * page calls the dashboard directly with its host permissions, so gavel needs no
 * CORS configuration. Gavel is the source of truth for ordering, versions and
 * delivery; this file only shapes requests and projects responses.
 */

export { DEFAULT_RUN_PROMPT, type CommentAgent, type CommentAgentSettings, type CommentEdits, type DeliverBatchRequest, type DeliverTarget, type SettingsRef, type GavelAgentSession as AgentSession };
export type RunTarget = Extract<DeliverTarget, { kind: "run" }>;
const PNG_PREFIX = "data:image/png;base64,";
const PAGE_SIZE = 200;

export type PageRef = { url: string; title: string };
export type GavelEvent = { sequence: number; commentId: string; kind: string };
export type DomSelector = { kind: "dom"; anchor: string; locators?: NonNullable<Selection["element"]["locators"]>; html: string };

export class GavelError extends Error {
  constructor(
    readonly status: number,
    message: string,
  ) {
    super(message);
  }
}

/** Go's `url.PathEscape`, which gavel compares against the scope it receives: it leaves `$&+=:@` alone and escapes `!'()*`. */
function goPathEscape(name: string): string {
  return encodeURIComponent(name)
    .replace(/[!'()*]/g, char => `%${char.charCodeAt(0).toString(16).toUpperCase()}`)
    .replace(/%(24|26|2B|3D|3A|40)/g, (_match, hex: string) => String.fromCharCode(parseInt(hex, 16)));
}

export function projectScope(name: string): string {
  return `gavel://project/${goPathEscape(name)}`;
}

/** A page is identified by its URL without the query string. */
export function pageKey(url: string): string {
  return externalUrl(url).base;
}

export function pageTargetUri(scope: string, page: PageRef): string {
  return `${scope}/pages/${encodeURIComponent(pageKey(page.url))}`;
}

export function domSelector(selection: Selection): DomSelector {
  return {
    kind: "dom",
    anchor: selection.anchor,
    ...(selection.element.locators ? { locators: selection.element.locators } : {}),
    html: selection.element.html,
  };
}

export function pageTarget(options: { scope: string; page: PageRef; selector?: DomSelector }): Target {
  return {
    uri: pageTargetUri(options.scope, options.page),
    scopeUri: options.scope,
    kind: "web-page",
    label: options.page.title,
    ...(options.selector ? { selector: options.selector } : {}),
  };
}

export function screenshotBase64(dataUrl: string): string {
  if (!dataUrl.startsWith(PNG_PREFIX)) throw new Error("Screenshots must be PNG data URLs");
  const data = dataUrl.slice(PNG_PREFIX.length);
  if (!data || data.length % 4 || !/^[A-Za-z0-9+/]+={0,2}$/.test(data)) throw new Error("Screenshot contains invalid base64 data");
  return data;
}

/** Incremental server-sent events parser; `push` accepts arbitrary chunk boundaries. */
export function parseSse(onEvent: (event: { id: string; event: string; data: string }) => void): (chunk: string) => void {
  let buffer = "";
  return chunk => {
    buffer += chunk;
    for (;;) {
      const end = /\r?\n\r?\n/.exec(buffer);
      if (!end) return;
      const block = buffer.slice(0, end.index);
      buffer = buffer.slice(end.index + end[0].length);
      let id = "";
      let event = "message";
      const data: string[] = [];
      for (const line of block.split(/\r?\n/)) {
        if (line === "" || line.startsWith(":")) continue;
        const colon = line.indexOf(":");
        const field = colon === -1 ? line : line.slice(0, colon);
        const value = colon === -1 ? "" : line.slice(colon + 1).replace(/^ /, "");
        if (field === "id") id = value;
        else if (field === "event") event = value;
        else if (field === "data") data.push(value);
      }
      if (data.length > 0) onEvent({ id, event, data: data.join("\n") });
    }
  };
}

type ClientOptions = { server: string; project: string; fetch?: typeof fetch };

async function failure(response: Response, url: URL, method: string): Promise<GavelError> {
  const text = await response.text();
  let message = text;
  try {
    const body: unknown = JSON.parse(text);
    if (body && typeof body === "object" && "error" in body) message = String(body.error);
  } catch {
    // A non-JSON error body is reported verbatim.
  }
  return new GavelError(response.status, `${message || `HTTP ${response.status}`} (${method} ${url.pathname}, HTTP ${response.status})`);
}

export async function getGavelJson(options: { server: string; path: string; fetch?: typeof fetch }): Promise<{ url: URL; body: unknown }> {
  const url = new URL(options.path, options.server);
  const response = await (options.fetch ?? fetch)(url);
  if (!response.ok) throw await failure(response, url, "GET");
  return { url, body: await response.json() };
}

/** The dashboard's project names, for choosing the scope comments are filed under. */
export async function listProjects(options: { server: string; fetch?: typeof fetch }): Promise<string[]> {
  const { url, body: projects } = await getGavelJson({ ...options, path: "/api/projects" });
  if (!Array.isArray(projects) || !projects.every((entry: unknown) => !!entry && typeof entry === "object" && typeof (entry as { name?: unknown }).name === "string"))
    throw new Error(`Gavel returned an invalid project list from ${url.href}`);
  return projects.map((entry: { name: string }) => entry.name);
}

export class GavelClient {
  readonly scope: string;
  private readonly server: string;
  private readonly fetcher: typeof fetch;

  constructor(options: ClientOptions) {
    this.server = options.server;
    this.scope = projectScope(options.project);
    this.fetcher = options.fetch ?? ((input, init) => fetch(input, init));
  }

  private attachmentUrl = (commentId: string, attachmentId: string) =>
    new URL(`/api/v1/comments/${encodeURIComponent(commentId)}/attachments/${encodeURIComponent(attachmentId)}`, this.server).href;

  private url(path: string, query?: Record<string, string>): URL {
    const url = new URL(`/api/v1/comments${path}`, this.server);
    if (query) url.search = new URLSearchParams(query).toString();
    return url;
  }

  private async send(url: URL, init: RequestInit = {}): Promise<Response> {
    const method = init.method ?? "GET";
    let response: Response;
    try {
      response = await this.fetcher(url, init);
    } catch (cause) {
      throw new Error(`Cannot reach gavel at ${this.server} (${method} ${url.pathname}): ${String(cause)}`);
    }
    if (!response.ok) throw await failure(response, url, method);
    return response;
  }

  private async json<T>(url: URL, init: RequestInit = {}): Promise<T> {
    const response = await this.send(url, init.body && typeof init.body === "string" ? { ...init, headers: { "Content-Type": "application/json", ...init.headers } } : init);
    const text = await response.text();
    try {
      return JSON.parse(text) as T;
    } catch {
      throw new Error(`Gavel returned non-JSON from ${url.href} (HTTP ${response.status}): ${text}`);
    }
  }

  private post<T>(path: string, body: unknown, method = "POST"): Promise<T> {
    return this.json<T>(this.url(path), { method, body: JSON.stringify(body) });
  }

  /**
   * A new managed session started with `runtime` (gavel's default without one).
   * Threads share it through the `groupKey`: one per page, or with `shared`
   * grouping one for the whole project scope.
   */
  runTarget(options: { page: PageRef; runtime?: AgentSessionRuntime; grouping?: AgentSessionGrouping; prompt?: string }): RunTarget {
    const { runtime, grouping = "page" } = options;
    return {
      kind: "run",
      prompt: options.prompt ?? DEFAULT_RUN_PROMPT,
      groupKey: grouping === "shared" ? `extension:shared:${this.scope}` : pageTargetUri(this.scope, options.page),
      ...(runtime ? { runOptions: { spec: runtime.spec, ...(runtime.presets ? { presets: runtime.presets } : {}) } } : {}),
    };
  }

  /** Every comment on `page`, oldest first. */
  async list(page: PageRef): Promise<StoredComment[]> {
    const target = pageTargetUri(this.scope, page);
    const comments: StoredComment[] = [];
    for (let offset = 0; ; offset += PAGE_SIZE) {
      const result = await this.json<{ data: NativeComment[] }>(this.url("", { scope: this.scope, target, limit: String(PAGE_SIZE), offset: String(offset) }));
      if (!Array.isArray(result.data)) throw new Error("Gavel returned an invalid paged comments response");
      for (const row of result.data) comments.push(...project(row, { attachmentUrl: this.attachmentUrl }));
      if (result.data.length < PAGE_SIZE) return comments.sort((left, right) => left.createdAt.localeCompare(right.createdAt));
    }
  }

  /** Creates a thread root and its screenshot atomically using the draft ID as the request ID. */
  async create(options: { page: PageRef; comment: StoredComment; selection?: Selection; deliver?: DeliverTarget; delivery?: CreateDeliveryChoice }): Promise<StoredComment> {
    const { page, selection } = options;
    let comment = options.comment;
    let screenshot: string | undefined;
    if (selection) {
      const { screenshot: capture, ...element } = selection.element;
      comment = { ...comment, element: { ...element, ...(capture?.status === "unavailable" ? { screenshot: { status: "unavailable" as const, reason: capture.reason } } : {}) } };
      if (capture?.status === "captured") screenshot = screenshotBase64(capture.dataUrl);
    }
    const content: Content = { ...contentFor(pageKey(page.url), comment), page };
    const created = await this.post<NativeComment>("", {
      target: pageTarget({ scope: this.scope, page, ...(selection ? { selector: domSelector(selection) } : {}) }),
      bodyMarkdown: bodyFor(comment),
      workflowId: "human-review",
      requestId: comment.id,
      content,
      ...(options.deliver ? { deliver: options.deliver } : {}),
      ...(options.delivery ? { delivery: options.delivery } : {}),
      ...(screenshot ? { attachments: [{ filename: "screenshot.png", contentType: "image/png", data: screenshot }] } : {}),
    });
    return project(created, { attachmentUrl: this.attachmentUrl })[0]!;
  }

  /** One lifecycle action (`resolve`, `reopen`, `close`, `acknowledge`, `assign`, `delete`, ...) on a thread root. */
  async act(options: { id: string; version: number; action: string } & Partial<Omit<Action, "id" | "expectedVersion" | "action">>): Promise<StoredComment> {
    const { id, version, action, ...extra } = options;
    const updated = await this.post<NativeComment>(`/${encodeURIComponent(id)}/${action}`, { expectedVersion: version, ...extra });
    return project(updated, { attachmentUrl: this.attachmentUrl })[0]!;
  }

  /** The unified diff of the patch proposed in reply `replyId` on thread `rootId`. */
  async proposalDiff(rootId: string, replyId: string): Promise<GitDiffPayload> {
    const diff = await this.json<GitDiffPayload>(this.url(`/${encodeURIComponent(rootId)}/proposals/${encodeURIComponent(replyId)}/diff`));
    if (typeof diff?.diff !== "string") throw new Error(`Gavel returned an invalid proposal diff for reply ${replyId}`);
    return diff;
  }

  async batch(actions: Action[]): Promise<NativeComment[]> {
    return this.post<NativeComment[]>("/batch", actions);
  }

  /** Replies on thread `root`; returns the thread, root first. */
  async reply(options: { page: PageRef; root: string; version: number; body: string }): Promise<StoredComment[]> {
    const reply: StoredComment = { id: crypto.randomUUID(), body: options.body, createdAt: new Date().toISOString(), author: localUser };
    const updated = await this.post<NativeComment>(`/${encodeURIComponent(options.root)}/reply`, {
      expectedVersion: options.version,
      bodyMarkdown: options.body,
      content: contentFor(pageKey(options.page.url), reply),
    });
    return project(updated, { attachmentUrl: this.attachmentUrl });
  }

  /** Hands a thread to an agent session; gavel records the outcome under the comment's `agentRun`. */
  async deliver(options: { id: string; version: number; requestId: string; target: DeliverTarget }): Promise<StoredComment> {
    const updated = await this.post<NativeComment>(`/${encodeURIComponent(options.id)}/deliver`, { expectedVersion: options.version, requestId: options.requestId, target: options.target });
    return project(updated, { attachmentUrl: this.attachmentUrl })[0]!;
  }

  /** Hands several open thread roots to one target in a single request; a stale or closed thread fails the whole batch. */
  async deliverBatch(options: DeliverBatchRequest): Promise<StoredComment[]> {
    const updated = await this.post<NativeComment[]>("/deliver", options);
    if (!Array.isArray(updated)) throw new Error("Gavel returned an invalid batch delivery response");
    return updated.flatMap(comment => project(comment, { attachmentUrl: this.attachmentUrl }));
  }

  /** The scope's live sessions plus the managed sessions that have ended and can be resumed. */
  async agentSessions(): Promise<GavelAgentSession[]> {
    const sessions = await this.json<GavelAgentSession[]>(this.url("/agent-sessions", { scope: this.scope }));
    if (!Array.isArray(sessions)) throw new Error("Gavel returned an invalid agent session list");
    return sessions;
  }

  /**
   * Starts a new managed session without a comment. Gavel names it `<prefix>-<6 chars>`,
   * so the prefix is required, and answers once the session id is known.
   */
  async startSession(options: { page: PageRef; prefix: string; runtime?: AgentSessionRuntime; edits?: CommentEdits }): Promise<GavelAgentSession> {
    const { page, prefix, runtime, edits } = options;
    if (prefix.trim() === "") throw new Error("A session needs a name prefix");
    const session = await this.post<GavelAgentSession>("/agent-sessions", {
      scope: this.scope,
      target: pageTargetUri(this.scope, page),
      prefix: prefix.trim(),
      prompt: DEFAULT_RUN_PROMPT,
      ...(runtime ? { runOptions: { spec: runtime.spec, ...(runtime.presets ? { presets: runtime.presets } : {}) } } : {}),
      ...(edits ? { edits } : {}),
    });
    const valid = !!session && typeof session === "object" && !Array.isArray(session)
      && [session.id, session.title].every(field => typeof field === "string" && field !== "")
      // An ended managed session has no source: captain reports only its id and model when it starts.
      && typeof session.source === "string" && typeof session.inCmux === "boolean";
    if (!valid) throw new Error(`Gavel returned an invalid started session: ${JSON.stringify(session)}`);
    return session;
  }

  async mentionables(page: PageRef): Promise<CommentMentionable[]> {
    const items = await this.json<CommentMentionable[]>(this.url("/mentionables", { scope: this.scope, target: pageTargetUri(this.scope, page) }));
    if (!Array.isArray(items)) throw new Error("Gavel returned an invalid mentionables list");
    return items;
  }

  async previewDelivery(options: { page: PageRef; body?: string; deliver?: DeliverTarget; delivery?: CreateDeliveryChoice }): Promise<CreateDeliveryPreview> {
    const { page, body, deliver, delivery } = options;
    const preview = await this.post<CreateDeliveryPreview>("/delivery/preview", {
      target: pageTarget({ scope: this.scope, page }),
      ...(body ? { bodyMarkdown: body } : {}),
      ...(deliver ? { deliver } : {}),
      ...(delivery ? { delivery } : {}),
    });
    if (!preview || typeof preview !== "object" || Array.isArray(preview))
      throw new Error("Gavel returned an invalid delivery preview");
    return preview;
  }

  /** The comment agent gavel's settings layers resolve for `page`. */
  async commentAgent(page: PageRef): Promise<CommentAgentSettings> {
    const settings = await this.json<CommentAgentSettings>(this.url("/agent", { scope: this.scope, uri: pageTargetUri(this.scope, page) }));
    if (!settings?.agent?.agent || typeof settings.agent.agent !== "object") throw new Error("Gavel returned an invalid comment agent response");
    return settings;
  }

  /** Saves `agent` as `page`'s entry in the settings `layer` (`null` removes the entry) and returns the agent now resolved for the page. */
  async saveCommentAgent(options: { page: PageRef; layer: SettingsRef; agent: CommentAgentPatch | null }): Promise<CommentAgentSettings> {
    const body: SaveCommentAgentRequest = {
      target: { scopeUri: this.scope, uri: pageTargetUri(this.scope, options.page) },
      layer: options.layer,
      page: true,
      agent: options.agent,
    };
    const settings = await this.post<CommentAgentSettings>("/agent", body, "PUT");
    if (!settings?.agent?.agent || typeof settings.agent.agent !== "object") throw new Error("Gavel returned an invalid comment agent response");
    return settings;
  }

  /** Reads the project's event feed from `after` until the server closes it or `signal` aborts. */
  async streamEvents(options: { after: number; signal?: AbortSignal; onOpen?: () => void; onEvent: (event: GavelEvent) => void }): Promise<void> {
    const response = await this.send(this.url("/events", { scope: this.scope, after: String(options.after) }), options.signal ? { signal: options.signal } : {});
    if (!response.body) throw new Error("Gavel's event stream has no body");
    options.onOpen?.();
    const decoder = new TextDecoder();
    const parse = parseSse(event => {
      const value: unknown = JSON.parse(event.data);
      if (!value || typeof value !== "object" || typeof (value as GavelEvent).sequence !== "number" || typeof (value as GavelEvent).commentId !== "string")
        throw new Error(`Gavel sent an invalid event: ${event.data}`);
      options.onEvent(value as GavelEvent);
    });
    const reader = response.body.getReader();
    for (;;) {
      const { done, value } = await reader.read();
      if (done) return;
      parse(decoder.decode(value, { stream: true }));
    }
  }
}
