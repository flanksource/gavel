import { TODO_MESSAGE_SOURCE } from "../comments/gavel-todo";
import { bookmarkletParentOrigin } from "./origin";

const MODAL_ID = "gavel-grabber-todo";
const DONE_TYPES = ["todo-saved", "cancel"];
const CLOSE_EVENT = "close-todo-modal";

/** Closes the open todo modal, if any, as its own close button does. */
export function closeTodoModal() {
  document.getElementById(MODAL_ID)?.dispatchEvent(new CustomEvent(CLOSE_EVENT));
}

/**
 * Frames the extension's offline todo form. `embed-ready` hands over the
 * screenshot; `todo-saved` closes only after the extension has persisted it.
 */
export function openTodoModal(options: { relay: string; target: string; parentOrigin?: string; attachment?: { blob: Blob; name: string }; onClose?: () => void }): { close: () => void } {
  closeTodoModal();
  const relay = new URL(options.relay);
  // URL.origin is "null" for chrome-extension: pages outside Chrome itself.
  const relayOrigin = `${relay.protocol}//${relay.host}`;
  relay.search = new URLSearchParams({ target: options.target, parentOrigin: options.parentOrigin ?? bookmarkletParentOrigin(location.href) }).toString();

  const root = document.createElement("div");
  root.id = MODAL_ID;
  root.setAttribute("data-react-grab-ignore", "");
  const shadow = root.attachShadow({ mode: "open" });
  const style = document.createElement("style");
  style.textContent = `
    :host { all: initial; position: fixed; inset: 0; z-index: 2147483647; }
    .backdrop { position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; background: rgba(0, 0, 0, 0.45); }
    .dialog { display: flex; flex-direction: column; width: min(960px, 94vw); height: min(800px, 92vh); overflow: hidden; border-radius: 10px; background: #fff; color: #111; box-shadow: 0 12px 48px rgba(0, 0, 0, 0.35); }
    .bar { display: flex; align-items: center; justify-content: space-between; padding: 8px 12px; background: #111; color: #fff; font: 600 13px/1.4 system-ui, sans-serif; }
    .bar button { border: none; background: none; color: #fff; font-size: 16px; line-height: 1; cursor: pointer; padding: 2px 6px; }
    iframe { flex: 1; border: none; width: 100%; background: #fff; }
  `;
  const backdrop = document.createElement("div");
  backdrop.className = "backdrop";
  const dialog = document.createElement("div");
  dialog.className = "dialog";
  dialog.setAttribute("role", "dialog");
  dialog.setAttribute("aria-modal", "true");
  dialog.setAttribute("aria-label", "Gavel issue");
  const bar = document.createElement("div");
  bar.className = "bar";
  const label = document.createElement("span");
  label.textContent = "Gavel issue";
  const close = document.createElement("button");
  close.type = "button";
  close.setAttribute("aria-label", "Close");
  close.textContent = "✕";
  bar.append(label, close);
  const frame = document.createElement("iframe");
  frame.title = "Gavel issue";
  dialog.append(bar, frame);
  backdrop.append(dialog);
  shadow.append(style, backdrop);

  let closed = false;
  function teardown() {
    if (closed) return;
    closed = true;
    window.removeEventListener("message", onMessage);
    window.removeEventListener("keydown", onKeyDown, true);
    root.remove();
    options.onClose?.();
  }
  function onMessage(event: MessageEvent) {
    if (event.source !== frame.contentWindow || event.origin !== relayOrigin) return;
    const data = event.data as { source?: unknown; type?: unknown } | null;
    if (!data || data.source !== TODO_MESSAGE_SOURCE) return;
    if (data.type === "embed-ready") {
      if (options.attachment) frame.contentWindow?.postMessage({ source: TODO_MESSAGE_SOURCE, type: "attachment", blob: options.attachment.blob, name: options.attachment.name }, relayOrigin);
    } else if (DONE_TYPES.includes(String(data.type))) teardown();
  }
  function onKeyDown(event: KeyboardEvent) {
    if (event.key !== "Escape") return;
    event.preventDefault();
    event.stopPropagation();
    teardown();
  }
  close.addEventListener("click", teardown);
  root.addEventListener(CLOSE_EVENT, teardown);
  window.addEventListener("message", onMessage);
  window.addEventListener("keydown", onKeyDown, true);
  frame.src = relay.href;
  document.documentElement.append(root);
  frame.focus();
  return { close: teardown };
}
