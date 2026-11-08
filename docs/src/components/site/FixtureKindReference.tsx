import bundle from "@/data/fixtures.schema.json";
import { SchemaViewer, type SchemaViewerProps } from "@flanksource/clicky-ui/data";
import { buildFixtureReferenceTree, GLOBAL_REFERENCE_ORDER, isGlobalReferenceField } from "@/lib/fixture-reference";

const branches = bundle.document.properties.steps.items.oneOf;

export default function FixtureKindReference({ kind }: { kind: string }) {
  const branch = branches.find((step) => step.properties.kind.const === kind);
  if (!branch) throw new Error(`Unknown fixture kind: ${kind}`);
  const schema = branch as unknown as SchemaViewerProps["schema"];
  return <div id={`fixture-kind-${kind}`}>
    <SchemaViewer schema={schema} buildTree={buildFixtureReferenceTree} sections={[
      {
        id: "global",
        title: "Global fields",
        description: "Shared step fields and available Clicky/commons CLI options. Format shortcuts are nested under format.",
        select: (field) => isGlobalReferenceField(field, schema),
        order: GLOBAL_REFERENCE_ORDER,
      },
      {
        id: "kind",
        title: `${kind} fields`,
        description: "Configuration for this step kind.",
        select: "*",
        order: ["kind", ...branch.required.filter((name) => name !== "kind")],
      },
    ]} />
  </div>;
}
