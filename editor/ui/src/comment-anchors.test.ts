import { describe, expect, it } from "vitest";
import { DOCUMENT_ANCHOR } from "@flanksource/clicky-ui/comments";
import { capturePatch, elementAnchor, patchAnchor } from "@flanksource/clicky-ui/file-findings";
import { anchorOf, selectorOf, type Selector } from "./comment-anchors";

const source = ["# Guide", "", "intro", "body one", "body two", "", "outro"].join("\n");
const patch = capturePatch(source, 4, 5);
const locators = [{ kind: "css" as const, selector: "main > p:nth-of-type(2)" }];

describe("selectorOf", () => {
  it.each<[string, string | null, string, Selector]>([
    ["a Source selection", patchAnchor(patch), "source", { kind: "code-range", version: 1, view: "source", patch }],
    ["a Preview selection", patchAnchor(patch), "preview", { kind: "code-range", version: 1, view: "preview", patch }],
    ["a selection from a host tab", patchAnchor(patch), "rich", { kind: "code-range", version: 1, view: "preview", patch }],
    ["a picked element", elementAnchor(locators), "preview", { kind: "dom", version: 1, view: "preview", anchor: elementAnchor(locators), locators }],
    ["a file comment", DOCUMENT_ANCHOR, "form", { kind: "document", version: 1, view: "form" }],
    ["an unanchored comment", null, "compare", { kind: "document", version: 1, view: "preview" }],
  ])("stores %s", (_, anchor, view, selector) => {
    expect(selectorOf(anchor, view)).toEqual(selector);
  });

  it("rejects an anchor kind the server cannot store", () => {
    expect(() => selectorOf('q:["guide.md","","intro",""]', "preview")).toThrow(/Unsupported comment anchor/);
  });
});

describe("anchorOf", () => {
  it.each<[string, Selector, string]>([
    ["a code range", { kind: "code-range", version: 1, view: "source", patch }, patchAnchor(patch)],
    ["an element", { kind: "dom", version: 1, view: "rich", anchor: "legacy", locators }, elementAnchor(locators)],
    ["an element captured without locators", { kind: "dom", version: 1, view: "rich", anchor: "legacy" }, DOCUMENT_ANCHOR],
    ["a file", { kind: "document", version: 1, view: "source" }, DOCUMENT_ANCHOR],
  ])("places %s", (_, selector, anchor) => {
    expect(anchorOf(selector)).toBe(anchor);
  });

  it("round-trips every selector FileReview creates", () => {
    for (const anchor of [patchAnchor(patch), elementAnchor(locators), DOCUMENT_ANCHOR])
      expect(anchorOf(selectorOf(anchor, "preview"))).toBe(anchor);
  });
});
