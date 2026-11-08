import { buildSchemaTree, type SchemaViewerNode, type SchemaViewerProps } from "@flanksource/clicky-ui/data";
import bundle from "@/data/fixtures.schema.json";

type Schema = SchemaViewerProps["schema"];

const branches = bundle.document.properties.steps.items.oneOf;
const envelopeFields = new Set(Object.keys(branches[0].properties).filter((name) =>
  name !== "kind" && branches.every((branch) => name in branch.properties),
));
const globalCategories = new Set(["Formatting", "Logging", "Tasks", "Global flags"]);
const formatShortcuts = ["pretty", "json", "yaml", "csv", "markdown", "llm", "html", "pdf", "slack", "tree", "table"];

export const GLOBAL_REFERENCE_ORDER = ["id", "name", "cel", "timeout", "wait", "retry", "traces", "export", "override", "background", "continueOnError", "format", "output", "filter", "no-color", "dump-schema", "log-level", "loglevel", "json-logs", "log-to-stderr", "report-caller", "timings", "max-concurrent", "max-retries", "retry-delay", "graceful-timeout", "no-progress", "cwd", "db-url", "properties"];

export function isGlobalReferenceField(field: SchemaViewerNode, schema: Schema): boolean {
  return envelopeFields.has(field.label) || globalCategories.has(String(schema.properties?.[field.label]?.["x-cli-group"]));
}

export function buildFixtureReferenceTree(schema: Schema): SchemaViewerNode[] {
  const roots = buildSchemaTree(schema);
  const format = roots.find((node) => node.label === "format");
  if (!format) return roots;
  const shortcuts = formatShortcuts.flatMap((name) => {
    const node = roots.find((field) => field.label === name);
    return node ? [{ ...node, path: name }] : [];
  });
  if (!shortcuts.length) return roots;
  return roots.filter((node) => !formatShortcuts.includes(node.label)).map((node) => node === format
    ? { ...node, children: [...(node.children ?? []), ...shortcuts], count: shortcuts.length }
    : node);
}
