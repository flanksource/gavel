import type { ContextMenuActionContext, ReactGrabAPI } from "@flanksource/react-grab";

/*
 * The grabbed-element todo formatting of gavel's React Grab bookmarklet, ported
 * to typed modules: the title and Markdown body of a todo for an element, the
 * dashboard's embed form URL, the screenshot attachment Markdown, and the
 * `/@fs/` open-file hook. The extension's local form captures the todo; nothing
 * here talks to gavel.
 */

/** Marks the messages the dashboard's embed form exchanges with the window that frames it. */
export const TODO_MESSAGE_SOURCE = "gavel-react-grab";
export const HTML_LIMIT = 2048;
/** getStackContext can be slow or hang for some elements; the body never waits longer than this for it. */
export const STACK_TIMEOUT_MS = 1500;

/** The part of react-grab the body needs: the same three calls the bookmarklet made on `window.__REACT_GRAB__`. */
export type TodoGrab = Pick<ReactGrabAPI, "getDisplayName" | "getSource" | "getStackContext">;
export type UploadedAttachment = { url: string; filename?: string };

/** Up to ~2KB of the element's raw outerHTML, clipped with a marker so the reader knows. */
export function grabHtml(element: Element | null | undefined): string {
  if (!element || typeof element.outerHTML !== "string") return "";
  const html = element.outerHTML;
  return html.length > HTML_LIMIT ? `${html.slice(0, HTML_LIMIT)}\n<!-- …truncated to 2KB -->` : html;
}

/**
 * Locates `element` in the page: nth-child steps up to the nearest unique id or
 * `<body>`. The source stack names the code; this names the instance, which
 * matters when one component renders many identical elements.
 */
export function cssPath(element: Element): string {
  const document = element.ownerDocument;
  const segments: string[] = [];
  let node: Element | null = element;
  while (node?.parentElement && node !== document.body) {
    const id = node.getAttribute("id");
    if (id && /^[A-Za-z][\w-]*$/.test(id) && document.querySelectorAll(`#${id}`).length === 1) {
      segments.unshift(`#${id}`);
      return segments.join(" > ");
    }
    const index = Array.prototype.indexOf.call(node.parentElement.children, node) + 1;
    segments.unshift(`${node.tagName.toLowerCase()}:nth-child(${index})`);
    node = node.parentElement;
  }
  segments.unshift("body");
  return segments.join(" > ");
}

/** The component and, when the element has one, its accessible name: the part that tells sibling instances apart. */
export function todoTitle(element: Element | null, componentName?: string, tagName?: string): string {
  const name = componentName || tagName || "UI element";
  const label = element?.getAttribute ? (element.getAttribute("aria-label") ?? "").trim() : "";
  return label ? `${name}: ${label}` : name;
}

/** The grabbed element as Markdown: the component, its page, a selector for the instance, the source stack and up to 2KB of HTML. */
export function buildBody(options: { componentName: string; pageUrl: string; selector: string; source: string; html: string }): string {
  const { componentName, pageUrl, selector, source, html } = options;
  const parts: string[] = [];
  if (componentName) parts.push(`**Component:** \`<${componentName}>\``);
  parts.push(`**Page:** ${pageUrl}`);
  if (selector) parts.push(`**Selector:** \`${selector}\``);
  if (source) parts.push("", "```", source, "```");
  if (html) parts.push("", "**HTML:**", "```html", html, "```");
  return parts.join("\n");
}

export function fileLine(filePath: string | null | undefined, lineNumber?: number | null): string {
  if (!filePath) return "";
  return `${filePath}:${lineNumber || ""}`;
}

/** Resolves `promise`'s value (or `fallback` when it is falsy, rejects, or is slower than `timeoutMs`); never rejects. */
export function resolveWithTimeout<T>(promise: Promise<T> | T, fallback: T, timeoutMs: number): Promise<T> {
  return new Promise((resolve) => {
    let done = false;
    const finish = (value: T) => {
      if (done) return;
      done = true;
      clearTimeout(timer);
      resolve(value);
    };
    const timer = setTimeout(() => finish(fallback), timeoutMs);
    Promise.resolve(promise).then(
      (value) => finish(value || fallback),
      (cause: unknown) => {
        console.warn("[gavel] could not read the component stack, using the source location instead:", cause);
        finish(fallback);
      },
    );
  });
}

async function fallbackSource(grab: TodoGrab, element: Element, fallback: string): Promise<string> {
  if (fallback) return fallback;
  try {
    const source = await grab.getSource(element);
    return source ? fileLine(source.filePath, source.lineNumber) : "";
  } catch (cause) {
    console.warn("[gavel] could not read the element's source location:", cause);
    return "";
  }
}

export type TodoElement = { element: Element; componentName?: string | null | undefined; filePath?: string | null | undefined; lineNumber?: number | null | undefined; html?: string };

/**
 * The single capture formatter, used by the todo form and by react-grab's own
 * Copy, so clipboard content and created or commented todos stay identical.
 */
export async function todoBodyForElement(options: { grab: TodoGrab; pageUrl: string } & TodoElement): Promise<string> {
  const { grab, pageUrl, element } = options;
  const html = options.html ?? grabHtml(element);
  const componentName = options.componentName || grab.getDisplayName(element) || "";
  // Snapshot the selector with the markup, before cleanup() may re-render the page.
  const selector = element.isConnected ? cssPath(element) : "";
  const fallback = await fallbackSource(grab, element, fileLine(options.filePath, options.lineNumber));
  const source = await resolveWithTimeout(Promise.resolve().then(() => grab.getStackContext(element)), fallback, STACK_TIMEOUT_MS);
  return buildBody({ componentName, pageUrl, selector, source, html });
}

export async function todoBodiesForElements(options: { grab: TodoGrab; pageUrl: string; elements: Element[] }): Promise<string> {
  const selected = options.elements.filter(Boolean);
  const parts = await Promise.all(selected.map((element) => todoBodyForElement({ grab: options.grab, pageUrl: options.pageUrl, element })));
  return parts.filter(Boolean).join("\n\n---\n\n");
}

export async function todoBodyForContext(options: { grab: TodoGrab; pageUrl: string; context: Pick<ContextMenuActionContext, "element" | "elements" | "componentName" | "filePath" | "lineNumber"> }): Promise<string> {
  const { grab, pageUrl, context } = options;
  const selected = context.elements.filter(Boolean);
  if (selected.length > 1) return todoBodiesForElements({ grab, pageUrl, elements: selected });
  return todoBodyForElement({ grab, pageUrl, element: context.element, componentName: context.componentName, filePath: context.filePath, lineNumber: context.lineNumber, html: grabHtml(context.element) });
}

/** The dashboard's embed form, prefilled. */
export function buildTodoUrl(options: { server: string; title: string; body: string; sourceUrl: string }): string {
  const url = new URL("/todos/new", options.server);
  url.searchParams.set("embed", "1");
  if (options.title) url.searchParams.set("title", options.title);
  if (options.body) url.searchParams.set("body", options.body);
  if (options.sourceUrl) url.searchParams.set("sourceUrl", options.sourceUrl);
  return url.toString();
}

export function screenshotMarkdown(attachment: UploadedAttachment): string {
  return `![${attachment.filename || "screenshot.png"}](${attachment.url})`;
}

export function todoBodyWithScreenshot(body: string, attachment: UploadedAttachment): string {
  const image = screenshotMarkdown(attachment);
  const trimmed = (body || "").trim();
  return trimmed ? `${trimmed}\n\n## Attachments\n\n${image}` : image;
}

/** Vite exposes absolute files to browser tooling as `/@fs/<absolute path>`; the editor wants the path. */
export function normalizeOpenFilePath(filePath: string): string {
  return filePath.startsWith("/@fs/") ? filePath.slice(4) : filePath;
}

/**
 * React Grab treats the `/@fs/` transport URL as the actual source path, so its
 * built-in Open request targets a nonexistent `/@fs` directory. This hook
 * claims only those paths and keeps React Grab's local-editor-then-web flow:
 * the page's dev server (`/__open-in-editor`) first, react-grab's open-file page
 * when that fails.
 */
export function openFileHook(options: {
  fetcher: (url: string) => Promise<{ ok: boolean }>;
  openWindow: (url: string, target: string, features: string) => unknown;
}): (filePath: string, lineNumber?: number) => boolean {
  const openFallback = (file: string, lineNumber?: number) =>
    options.openWindow(`https://react-grab.com/open-file?url=${encodeURIComponent(file)}${lineNumber ? `&line=${lineNumber}` : ""}`, "_blank", "noopener,noreferrer");
  return (filePath, lineNumber) => {
    const file = normalizeOpenFilePath(filePath);
    if (file === filePath) return false;
    const params = new URLSearchParams({ file, column: "1" });
    if (lineNumber) params.set("line", String(lineNumber));
    options.fetcher(`/__open-in-editor?${params}`).then(
      (response) => {
        if (!response.ok) openFallback(file, lineNumber);
      },
      () => openFallback(file, lineNumber),
    );
    return true;
  };
}
