import { describe, expect, it } from "vitest";
import type { MdxEntry } from "./mdx";
import { buildDocTree } from "./docs";

const doc = (slug: string, parent?: string): MdxEntry => ({
  slug, frontmatter: { title: slug, parent }, Component: () => null,
});

describe("documentation hierarchy", () => {
  it("nests descendants even when they precede their parents", () => {
    const leaf = doc("exec", "steps");
    const steps = doc("steps", "fixtures");
    const fixtures = doc("fixtures");
    expect(buildDocTree([leaf, fixtures, steps])).toEqual([
      { doc: fixtures, children: [{ doc: steps, children: [{ doc: leaf, children: [] }] }] },
    ]);
  });

  it.each([
    { docs: [doc("exec", "missing")], error: /missing parent/ },
    { docs: [doc("exec"), doc("exec")], error: /duplicate/ },
    { docs: [doc("exec", "steps"), doc("steps", "exec")], error: /cycle/ },
  ])("rejects invalid hierarchy: $error", ({ docs, error }) => {
    expect(() => buildDocTree(docs)).toThrow(error);
  });
});
