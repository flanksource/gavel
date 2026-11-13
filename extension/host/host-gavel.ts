import type { ActionContext, ContextMenuActionContext, PluginConfig, ReactGrabAPI } from "@flanksource/react-grab";
import type { CommentScreenshotCapture } from "../contracts/comments-model";
import { writeClipboard } from "../comments/clipboard";
import {
  buildTodoUrl,
  grabHtml,
  openFileHook,
  todoBodiesForElements,
  todoBodyForContext,
  todoBodyForElement,
  todoBodyWithScreenshot,
  todoTitle,
  type UploadedAttachment,
} from "../comments/gavel-todo";
import { pngBlob } from "../comments/png-blob";
import type { GavelTodoActions } from "../comments/react-grab-actions";
import type { ScreenshotTarget } from "../comments/screenshot-viewport";
import type { openTodoModal } from "./host-todo-modal";

const SCREENSHOT_NAME = "screenshot.png";
const TODO_MODAL_ID = "gavel-grabber-todo";

export type GavelTodoOptions = {
  grab: ReactGrabAPI;
  /** The dashboard origin: where queued todos sync and clipboard screenshots upload. */
  server: string;
  /** The extension's bundled offline todo form. */
  todoRelay: string;
  /** Takes the screenshot with the grabber's own UI hidden. */
  capture: (target: Element | ScreenshotTarget) => Promise<CommentScreenshotCapture>;
  /** Uploads a PNG data URL to the dashboard from an extension context. */
  upload: (png: { dataUrl: string; name: string }) => Promise<UploadedAttachment>;
  openModal: typeof openTodoModal;
};

/** What a todo needs of the grabbed element: what a context-menu action gets, or what a Command+G selection can supply. */
type ElementTodoContext = Pick<ActionContext, "element" | "componentName" | "filePath" | "lineNumber" | "tagName" | "cleanup"> & { hideContextMenu?: () => void };

/** Removes react-grab's overlay and unfreezes the page, which toggle activation leaves stuck on "Grabbing…" until cleanup runs. */
function dismissGrab(context: Pick<ElementTodoContext, "hideContextMenu" | "cleanup">) {
  context.hideContextMenu?.();
  context.cleanup();
}

/**
 * Gavel's React Grab bookmarklet, as the grabber's gavel-backend behavior: the
 * todo form for a grabbed element (with an optional screenshot), the screenshot
 * link copies, the body react-grab's own Copy produces, the Command+G one-shot
 * and the `/@fs/` open-file hook. The shared plugin builder registers these.
 */
export function gavelTodoActions(options: GavelTodoOptions): GavelTodoActions {
  const { grab, server, todoRelay, capture, upload, openModal } = options;

  const captureAndUpload = async (element: Element): Promise<UploadedAttachment> => {
    const result = await capture(element);
    if (result.status !== "captured") throw new Error(`screenshot ${result.reason}`);
    return upload({ dataUrl: result.dataUrl, name: SCREENSHOT_NAME });
  };

  // The markup, selector and body are read before the overlay closes: cleanup() unfreezes the page, which may re-render it.
  function snapshot(context: ElementTodoContext) {
    const pageUrl = location.href;
    const title = todoTitle(context.element, context.componentName, context.tagName);
    const body = todoBodyForElement({ grab, pageUrl, element: context.element, componentName: context.componentName, filePath: context.filePath, lineNumber: context.lineNumber, html: grabHtml(context.element) });
    return async (attachment?: { blob: Blob; name: string }) => {
      openModal({ relay: todoRelay, target: buildTodoUrl({ server, title, body: await body, sourceUrl: pageUrl }), ...(attachment ? { attachment } : {}) });
    };
  }

  const reportFailure = (what: string) => (cause: unknown) => console.error(`[gavel] ${what} failed:`, cause);

  function addTodo(context: ElementTodoContext) {
    const open = snapshot(context);
    dismissGrab(context);
    open().catch(reportFailure("Add to gavel todo"));
  }

  function screenshotTodo(context: ElementTodoContext) {
    const open = snapshot(context);
    dismissGrab(context);
    capture(context.element)
      .then((result) => {
        if (result.status === "captured") return open({ blob: pngBlob(result.dataUrl), name: SCREENSHOT_NAME });
        console.warn("[gavel] screenshot unavailable, opening the todo without an image:", result.reason);
        return open();
      })
      .catch(reportFailure("Screenshot to gavel todo"));
  }

  // The clipboard write starts inside the action's gesture, which the Clipboard API requires; its content resolves after the capture and upload.
  function copyScreenshot(context: ContextMenuActionContext, copy: { body: boolean }) {
    const pageUrl = location.href;
    const body = copy.body ? todoBodyForContext({ grab, pageUrl, context }) : Promise.resolve("");
    dismissGrab(context);
    const text = Promise.all([body, captureAndUpload(context.element)]).then(([markdown, attachment]) => (copy.body ? todoBodyWithScreenshot(markdown, attachment) : attachment.url));
    writeClipboard(text).catch(reportFailure("Copy screenshot"));
  }

  // Command+G arms the next selection to open Add to gavel todo, whichever way it is made.
  const setup = (api: ReactGrabAPI): PluginConfig => {
    let armed = false;
    let selectionQueued = false;
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key.toLowerCase() !== "g" || !event.metaKey || event.ctrlKey || event.altKey || event.shiftKey || event.repeat) return;
      if (!api.isEnabled()) return;
      if (event.target instanceof Element && event.target.closest("input,textarea,select,[contenteditable]")) return;
      if (document.getElementById(TODO_MODAL_ID)) return;
      event.preventDefault();
      event.stopPropagation();
      if (armed) {
        armed = false;
        api.deactivate();
        return;
      }
      if (api.isActive()) api.deactivate();
      armed = true;
      api.activate();
    };
    window.addEventListener("keydown", onKeyDown, true);
    return {
      hooks: {
        onDeactivate: () => {
          armed = false;
          selectionQueued = false;
        },
        onElementSelect: (element) => {
          if (!armed) return;
          if (!selectionQueued) {
            selectionQueued = true;
            queueMicrotask(() => {
              if (!selectionQueued) return;
              armed = false;
              selectionQueued = false;
              const componentName = api.getDisplayName(element);
              addTodo({ element, ...(componentName ? { componentName } : {}), tagName: element.tagName.toLowerCase(), cleanup: () => api.deactivate() });
            });
          }
          return true;
        },
      },
      cleanup: () => {
        window.removeEventListener("keydown", onKeyDown, true);
        if (armed) api.deactivate();
        return undefined;
      },
    };
  };

  return {
    addTodo,
    screenshotTodo,
    copyScreenshot,
    copyContent: async (content, elements) => (await todoBodiesForElements({ grab, pageUrl: location.href, elements })) || content,
    onOpenFile: openFileHook({ fetcher: (url) => fetch(url), openWindow: (url, target, features) => window.open(url, target, features) }),
    setup,
  };
}
