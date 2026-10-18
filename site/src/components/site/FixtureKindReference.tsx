import bundle from "@/data/fixtures.schema.json";
import services from "../../../../fixtures/kinds/services/schema.json";
import { SchemaReference } from "@flanksource/clicky-ui/data";

export default function FixtureKindReference({ kind }: { kind: string }) {
  const branch = bundle.document.properties.steps.items.oneOf.find((step) => step.properties.kind.const === kind);
  if (!branch) throw new Error(`Unknown fixture kind: ${kind}`);
  const schema = kind === "services" ? { ...branch, definitions: services.definitions } : branch;
  return <SchemaReference schema={schema} id={`fixture-kind-${kind}`} />;
}
