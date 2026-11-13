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

function orderFields(fields: SchemaViewerNode[], order: string[]): SchemaViewerNode[] {
  const positions = new Map(order.map((name, index) => [name, index]));
  return fields.sort((a, b) => (positions.get(a.label) ?? order.length) - (positions.get(b.label) ?? order.length));
}

export function buildFixtureKindReferenceTree(schema: Schema, kind: string): SchemaViewerNode[] {
  const fields = buildFixtureReferenceTree(schema);
  return [
    {
      key: "global",
      label: "Global fields",
      badge: "group",
      description: "Shared step fields and available Clicky/commons CLI options. Format shortcuts are nested under format.",
      children: orderFields(fields.filter(field => isGlobalReferenceField(field, schema)), GLOBAL_REFERENCE_ORDER),
    },
    {
      key: "kind",
      label: `${kind} fields`,
      badge: "group",
      description: "Configuration for this step kind.",
      children: orderFields(fields.filter(field => !isGlobalReferenceField(field, schema)), ["kind", ...(schema.required ?? [])]),
    },
  ];
}
