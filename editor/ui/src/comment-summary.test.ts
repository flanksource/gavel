import { describe, expect, it } from "vitest";
import { commentCountsAt } from "./comment-summary";

const BY_PATH = new Map([
  ["products/schemas/a.json", { total: 3, unread: 2 }],
  ["products/schemas/b.json", { total: 1, unread: 0 }],
  ["products/annuity.yaml", { total: 2, unread: 1 }],
  ["products-old/c.yaml", { total: 5, unread: 5 }],
]);

describe("commentCountsAt", () => {
  it.each([
    ["a file's own counts", "file", "products/schemas/a.json", { total: 3, unread: 2 }],
    ["a file without comments", "file", "docs/readme.md", { total: 0, unread: 0 }],
    ["the sum over a folder's files", "folder", "products/schemas", { total: 4, unread: 2 }],
    ["a parent folder, but not a sibling sharing its prefix", "folder", "products", { total: 6, unread: 3 }],
    ["every file under the root", "folder", "", { total: 11, unread: 8 }],
  ] as const)("returns %s", (_, kind, path, expected) => {
    expect(commentCountsAt(BY_PATH, kind, path)).toEqual(expected);
  });
});
