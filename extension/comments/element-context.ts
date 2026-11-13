import type { ReactGrabAPI, SourceInfo } from "@flanksource/react-grab";

import {
  COMMENT_ELEMENT_HTML_LIMIT,
  COMMENT_ELEMENT_HTML_TRUNCATION,
  type CommentElementContext,
} from "../contracts/comments-model";
import { captureLocators } from "./dom-anchor";

const STACK_TIMEOUT_MS = 1500;

function sourceLocation(source: SourceInfo | null): string {
  if (!source?.filePath) return "";
  return `${source.filePath}:${source.lineNumber ?? ""}`;
}

function resolveStack(
  api: ReactGrabAPI,
  element: Element,
  fallback: string,
): Promise<string> {
  return new Promise((resolve) => {
    let settled = false;
    const finish = (value: string) => {
      if (settled) return;
      settled = true;
      window.clearTimeout(timeout);
      resolve(value || fallback);
    };
    const timeout = window.setTimeout(() => finish(fallback), STACK_TIMEOUT_MS);
    void api.getStackContext(element).then(finish, () => finish(fallback));
  });
}

export function captureElementHtml(element: Element): string {
  if (element.outerHTML.length <= COMMENT_ELEMENT_HTML_LIMIT) {
    return element.outerHTML;
  }
  return `${element.outerHTML.slice(0, COMMENT_ELEMENT_HTML_LIMIT)}${COMMENT_ELEMENT_HTML_TRUNCATION}`;
}

export async function captureElementContext(
  element: Element,
  options: {
    fallbackSource?: string;
    /** Root the locators are computed against; omit when the element has none (an overlay). */
    root?: Element;
  } = {},
): Promise<Omit<CommentElementContext, "screenshot">> {
  const { fallbackSource, root } = options;
  const html = captureElementHtml(element);
  // Synchronously, before any await, so the locators see the DOM as clicked.
  const locators = root ? captureLocators(element, root) : [];
  const api = window.__REACT_GRAB__;
  if (!api) {
    throw new Error(
      "React Grab is not ready, so the selected element cannot be attached to this comment",
    );
  }

  const directSource = await api
    .getSource(element)
    .then(sourceLocation, () => "");
  const source = await resolveStack(
    api,
    element,
    directSource || fallbackSource || "",
  );
  if (!source) {
    throw new Error(
      "React Grab could not resolve source context for the selected element",
    );
  }

  const componentName = api.getDisplayName(element);
  return {
    ...(componentName ? { componentName } : {}),
    source,
    html,
    ...(locators.length > 0 ? { locators } : {}),
  };
}
