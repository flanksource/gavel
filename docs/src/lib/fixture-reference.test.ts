import { expect, it } from "vitest";
import type { SchemaViewerProps } from "@flanksource/clicky-ui/data";
import { buildFixtureReferenceTree, isGlobalReferenceField } from "./fixture-reference";

const schema: SchemaViewerProps["schema"] = { type: "object", properties: {
  format: { type: "string", "x-cli-group": "Formatting" },
  json: { type: "boolean", "x-cli-group": "Formatting" },
  "no-color": { type: "boolean", "x-cli-group": "Formatting" },
  "log-level": { type: "string", "x-cli-group": "Logging" },
  "no-progress": { type: "boolean", "x-cli-group": "Tasks" },
  cwd: { type: "string", "x-cli-group": "Global flags" },
  linters: { type: "array", "x-cli-group": "Command flags", items: { type: "string" } },
} };

it("nests format shortcuts while keeping their original flat field paths", () => {
  const roots = buildFixtureReferenceTree(schema);
  expect(roots.map((node) => node.label)).toEqual(["format", "no-color", "log-level", "no-progress", "cwd", "linters"]);
  expect(roots[0]?.children).toEqual([expect.objectContaining({ label: "json", key: "root/json", path: "json" })]);
});

it("recognizes native formatting, logging, task and global flags separately from command fields", () => {
  expect(buildFixtureReferenceTree(schema).filter((node) => isGlobalReferenceField(node, schema)).map((node) => node.label)).toEqual(["format", "no-color", "log-level", "no-progress", "cwd"]);
});
