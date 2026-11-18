import { describe, expect, it } from "vitest";
import { previewDataOf } from "./data-preview";

describe("previewDataOf", () => {
  it.each([
    ["YAML under a gavel header", "data.yaml", "# gavel: preview=./Preview.tsx\nname: example\ntags: [a, b]\n", { name: "example", tags: ["a", "b"] }],
    ["JSON without its $preview key", "data.json", '{"$schema":"./schema.json","$preview":"./Preview.tsx","name":"example"}', { $schema: "./schema.json", name: "example" }],
    ["YAML without its $preview key", "data.yml", "$preview: ./Preview.tsx\ncount: 3\n", { count: 3 }],
    ["a top-level list as it is", "data.yaml", "- one\n- two\n", ["one", "two"]],
  ])("is the parsed document of %s", (_case, path, source, data) => {
    expect(previewDataOf(source, path)).toEqual(data);
  });

  it("throws on a syntax error rather than rendering stale data as current", () => {
    expect(() => previewDataOf('{"name": ', "data.json")).toThrow();
  });
});
