import bundle from "@/data/fixtures.schema.json";
import { SchemaViewer, type SchemaViewerProps } from "@flanksource/clicky-ui/data";
import { buildFixtureKindReferenceTree } from "@/lib/fixture-reference";

const branches = bundle.document.properties.steps.items.oneOf;

export default function FixtureKindReference({ kind }: { kind: string }) {
  const branch = branches.find((step) => step.properties.kind.const === kind);
  if (!branch) throw new Error(`Unknown fixture kind: ${kind}`);
  const schema = branch as unknown as SchemaViewerProps["schema"];
  return <div id={`fixture-kind-${kind}`}>
    <SchemaViewer schema={schema} buildTree={schema => buildFixtureKindReferenceTree(schema, kind)} defaultOpenDepth={2} />
  </div>;
}
