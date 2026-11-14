import { DOCUMENT_ANCHOR, type AnnotationCapture } from "@flanksource/clicky-ui/comments";
import {
  elementAnchor,
  isElementAnchor,
  isPatchAnchor,
  parseElementAnchor,
  parsePatchAnchor,
  patchAnchor,
} from "@flanksource/clicky-ui/file-findings";

/** The views a stored selector may name; the server rejects any other. */
export type SelectorView = "preview" | "source" | "rich" | "form";

/** A stored comment selector, as `editor/commenting.go` validates it. */
export type Selector = Pick<AnnotationCapture, "kind" | "version" | "patch" | "anchor" | "locators"> & {
  view: SelectorView;
};

export function selectorView(view: string): SelectorView {
  return view === "source" || view === "rich" || view === "form" ? view : "preview";
}

/** The FileReview anchor a stored selector places at. */
export function anchorOf(selector: Selector): string {
  switch (selector.kind) {
    case "code-range":
      if (!selector.patch) throw new Error("code-range comment has no patch");
      return patchAnchor(selector.patch);
    case "dom":
      return selector.locators?.length ? elementAnchor(selector.locators) : DOCUMENT_ANCHOR;
    case "document":
      return DOCUMENT_ANCHOR;
  }
}

/** The selector to store for a comment FileReview created at `anchor` in `view`. */
export function selectorOf(anchor: string | null | undefined, view: string): Selector {
  if (anchor && isPatchAnchor(anchor)) {
    // A Preview selection is anchored to its source lines; the server accepts it from either view.
    const patchView = view === "source" ? "source" : "preview";
    return { kind: "code-range", version: 1, view: patchView, patch: parsePatchAnchor(anchor) };
  }
  if (anchor && isElementAnchor(anchor))
    return { kind: "dom", version: 1, view: "preview", anchor, locators: parseElementAnchor(anchor) };
  if (!anchor || anchor === DOCUMENT_ANCHOR)
    return { kind: "document", version: 1, view: selectorView(view) };
  throw new Error(`Unsupported comment anchor ${JSON.stringify(anchor)}`);
}
