import { externalUrl, type AnchorState, type HostState, type Presentation, type Selection } from "./model";

/*
 * Validation of the messages a grabber host sends its panel. Both panels (the
 * playground's and the extension's bundled gavel panel) read them through here.
 */

export function validateParentOrigin(parentOrigin: string) {
  if (parentOrigin !== "null" && (new URL(parentOrigin).origin !== parentOrigin || !/^https?:/.test(parentOrigin)))
    throw new Error("Invalid parent origin");
}

export function parseSelection(input: unknown): Selection | undefined {
  if (input === undefined) return undefined;
  if (!input || typeof input !== "object") throw new Error("Invalid element selection");
  const selection = input as Record<string, unknown>;
  const element = selection["element"];
  if (typeof selection["anchor"] !== "string" || typeof selection["label"] !== "string" || !element || typeof element !== "object")
    throw new Error("Invalid selection fields");
  const context = element as Record<string, unknown>;
  if (typeof context["source"] !== "string" || typeof context["html"] !== "string" || typeof context["pageUrl"] !== "string")
    throw new Error("Selection requires document URL, source and HTML");
  externalUrl(context["pageUrl"]);
  return input as Selection;
}

/** The document a host reports, which must belong to the page that framed the panel. */
export function parseHostState(message: unknown, parentOrigin: string): HostState {
  const input = message as Record<string, unknown> | undefined;
  if (!input || typeof input["url"] !== "string" || typeof input["title"] !== "string") throw new Error("Invalid document state");
  const documentUrl = new URL(externalUrl(input["url"]).url);
  if (parentOrigin === "null" ? documentUrl.protocol !== "file:" : documentUrl.origin !== parentOrigin)
    throw new Error("Document origin does not match the parent");
  const selection = parseSelection(input["selection"]);
  if (selection?.element.pageUrl !== undefined && selection.element.pageUrl !== input["url"]) throw new Error("Selection belongs to another URL");
  return { url: input["url"], title: input["title"], ...(selection ? { selection } : {}) };
}

export function parseAnchors(values: unknown): AnchorState[] {
  const valid =
    Array.isArray(values) &&
    values.every((value: unknown) => {
      if (!value || typeof value !== "object") return false;
      const anchor = value as Record<string, unknown>;
      return (
        typeof anchor["anchor"] === "string" &&
        typeof anchor["label"] === "string" &&
        typeof anchor["available"] === "boolean" &&
        (anchor["available"] === false || (typeof anchor["top"] === "number" && Number.isFinite(anchor["top"])))
      );
    });
  if (!valid) throw new Error("Invalid anchor state");
  return values as AnchorState[];
}

export function parsePresentation(mode: unknown): Presentation {
  if (!["compact", "draft", "rail"].includes(String(mode))) throw new Error("Invalid presentation");
  return mode as Presentation;
}

export function parseFabPosition(message: Record<string, unknown>): { x: number; y: number } {
  const { x, y } = message;
  if (typeof x !== "number" || !Number.isFinite(x) || typeof y !== "number" || !Number.isFinite(y)) throw new Error("Invalid FAB position");
  return { x, y };
}
