import { expect, it } from "vitest";
import { parseAnchors, parseFabPosition, parseHostState, parsePresentation, parseSelection, validateParentOrigin } from "./bridge-parse";

const element = { source: "https://example.com/", html: "<h1>Hi</h1>", pageUrl: "https://example.com/a" };

it("accepts a selection whose element carries source, html and a document URL", () => {
  expect(parseSelection({ anchor: "#hero", label: "Hero", element })).toEqual({ anchor: "#hero", label: "Hero", element });
  expect(parseSelection(undefined)).toBeUndefined();
});

it.each([
  { name: "a non-object", input: "x", error: /Invalid element selection/ },
  { name: "a missing label", input: { anchor: "#a", element }, error: /Invalid selection fields/ },
  { name: "an element without html", input: { anchor: "#a", label: "A", element: { source: "s", pageUrl: "https://example.com/" } }, error: /requires document URL, source and HTML/ },
  { name: "an element on a non-document URL", input: { anchor: "#a", label: "A", element: { ...element, pageUrl: "chrome://settings" } }, error: /HTTP\(S\) or file URL/ },
])("rejects $name", ({ input, error }) => {
  expect(() => parseSelection(input)).toThrow(error);
});

it("accepts document state that belongs to the parent origin and carries a matching selection", () => {
  const selection = { anchor: "#a", label: "A", element };
  expect(parseHostState({ url: "https://example.com/a", title: "Doc", selection }, "https://example.com")).toEqual({ url: "https://example.com/a", title: "Doc", selection });
  expect(parseHostState({ url: "file:///docs/a.html", title: "Doc" }, "null")).toEqual({ url: "file:///docs/a.html", title: "Doc" });
});

it.each([
  { name: "state from another origin", input: { url: "https://bank.example/", title: "x" }, error: /does not match the parent/ },
  { name: "state without a title", input: { url: "https://example.com/" }, error: /Invalid document state/ },
  { name: "a selection on another URL", input: { url: "https://example.com/b", title: "x", selection: { anchor: "#a", label: "A", element } }, error: /another URL/ },
])("rejects $name", ({ input, error }) => {
  expect(() => parseHostState(input, "https://example.com")).toThrow(error);
});

it("validates anchors, presentation, FAB position and the parent origin", () => {
  expect(parseAnchors([{ anchor: "#a", label: "A", available: true, top: 12 }, { anchor: "#b", label: "B", available: false }])).toHaveLength(2);
  expect(() => parseAnchors([{ anchor: "#a", label: "A", available: true }])).toThrow(/Invalid anchor state/);
  expect(parsePresentation("rail")).toBe("rail");
  expect(() => parsePresentation("huge")).toThrow(/Invalid presentation/);
  expect(parseFabPosition({ x: 1, y: 2 })).toEqual({ x: 1, y: 2 });
  expect(() => parseFabPosition({ x: Number.NaN, y: 2 })).toThrow(/Invalid FAB position/);
  expect(() => validateParentOrigin("https://example.com")).not.toThrow();
  expect(() => validateParentOrigin("null")).not.toThrow();
  expect(() => validateParentOrigin("https://example.com/path")).toThrow(/Invalid parent origin/);
  expect(() => validateParentOrigin("ftp://example.com")).toThrow(/Invalid parent origin/);
});
