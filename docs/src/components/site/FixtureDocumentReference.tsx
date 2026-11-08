import { SchemaViewer, type SchemaViewerProps } from "@flanksource/clicky-ui/data";
import bundle from "@/data/fixtures.schema.json";

export default function FixtureDocumentReference() {
  return <div id="fixture-document-schema">
    <SchemaViewer schema={bundle.document as unknown as SchemaViewerProps["schema"]} />
  </div>;
}
