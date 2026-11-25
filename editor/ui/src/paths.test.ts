import { describe, expect, it } from "vitest";
import { groupFindingsByFile, type FileFinding } from "@flanksource/clicky-ui/file-findings";
import { commonRoot, editorFindingRef, nodePath, relativeToFile } from "./paths";

const finding = (file: string, workDir?: string): FileFinding => ({
  file,
  ...(workDir ? { workDir } : {}),
  line: 1,
  severity: "error",
  message: "m",
  source: "jsonlint",
  kind: "lint",
});

describe("commonRoot", () => {
  it.each([
    [[], ""],
    [["/repo/docs/a.md"], "/repo/docs/"],
    [["/repo/docs/a.md", "/repo/docs/img/b.png"], "/repo/docs/"],
    [["/repo/docs/a.md", "/repo/config/c.yaml"], "/repo/"],
    [["/repo/docs-old/a.md", "/repo/docs/b.md"], "/repo/"],
  ])("%j has root %j", (paths, root) => {
    expect(commonRoot(paths)).toBe(root);
  });
});

describe("editorFindingRef", () => {
  const root = "/repo/";
  it.each([
    ["relative to the root's work dir", finding("docs/a.md", "/repo"), "docs/a.md"],
    ["relative to a deeper work dir", finding("a.json", "/repo/config"), "config/a.json"],
    ["absolute", finding("/repo/docs/a.md"), "docs/a.md"],
    ["with dot segments", finding("../docs/a.md", "/repo/config"), "docs/a.md"],
    ["outside the root", finding("x.md", "/elsewhere"), "/elsewhere/x.md"],
  ])("rebases a finding %s", (_, input, file) => {
    expect(editorFindingRef(input, root)).toEqual({ file });
  });
});

describe("relativeToFile", () => {
  it.each([
    ["docs/guide.md", "docs/img/a.png", "img/a.png"],
    ["docs/guide.md", "assets/a.png", "../assets/a.png"],
    ["guide.md", "logo.svg", "logo.svg"],
    ["a/b/c.md", "a/x.png", "../x.png"],
  ])("from %s to %s is %s", (from, target, rel) => {
    expect(relativeToFile(from, target)).toBe(rel);
  });
});

describe("nodePath", () => {
  // products/schemas holds two files, so it is a folder row; docs/guide.md folds into one file row.
  const tree = groupFindingsByFile([], [
    { file: "products/schemas/a.json" },
    { file: "products/schemas/b.json" },
    { file: "docs/guide.md" },
  ]);
  const names = (kind: "file" | "folder", file: string) =>
    nodePath(tree, { kind, ref: { file } }).map((node) => `${node.kind}:${node.name}`);

  it.each([
    ["a file in a folder", "file", "products/schemas/a.json", ["folder:products/schemas", "file:a.json"]],
    ["a folder", "folder", "products/schemas", ["folder:products/schemas"]],
    ["a file folded into its folder's row", "file", "docs/guide.md", ["file:docs/guide.md"]],
    ["a path that is not in the tree", "file", "missing.md", []],
  ] as const)("lists the tree rows down to %s", (_, kind, file, expected) => {
    expect(names(kind, file)).toEqual(expected);
  });
});
