import { useEffect, useState, type ComponentType } from "react";
import { JsonSchemaForm } from "@flanksource/clicky-ui/components";
import { useDensityValue, useOptionalTheme } from "@flanksource/clicky-ui/hooks";
import type { DocumentSchemaFormProps } from "@flanksource/clicky-ui/monaco";
import { PageForm } from "@flanksource/page-runtime/host";
import type { JsonValue } from "@flanksource/page-runtime/schema";
import { request } from "./api";

type PreviewContext = { paths: string[]; dependencies: Record<string, string>; runtimeOrigin: string };

/**
 * The document form for file `id`. A schema that declares `x-clicky-extensions`
 * renders through the page runtime, which compiles the extensions from the
 * file's preview root (the server rewrites each `ref` into that root); any
 * other schema renders as a plain `JsonSchemaForm`.
 */
export function previewExtensionForm(id: string): ComponentType<DocumentSchemaFormProps> {
  return function PreviewExtensionForm(props: DocumentSchemaFormProps) {
    if (props.schema["x-clicky-extensions"] === undefined) return <JsonSchemaForm {...props} />;
    return <RuntimeForm id={id} {...props} />;
  };
}

function RuntimeForm({ id, schema, value, onChange, ...formProps }: DocumentSchemaFormProps & { id: string }) {
  const theme = useOptionalTheme()?.resolvedTheme ?? "light";
  const density = useDensityValue();
  const [context, setContext] = useState<PreviewContext | null>(null);
  const [error, setError] = useState<Error | null>(null);
  useEffect(() => {
    let cancelled = false;
    setContext(null);
    void request<PreviewContext>(`/api/file/${id}/preview`)
      .then((next) => !cancelled && setContext(next))
      .catch((cause: unknown) => !cancelled && setError(cause instanceof Error ? cause : new Error(String(cause))));
    return () => {
      cancelled = true;
    };
  }, [id]);
  if (error) throw new Error(`form extensions could not load: ${error.message}`);
  if (!context) return <p role="status">Loading form extensions…</p>;
  return (
    <PageForm
      runtimeOrigin={context.runtimeOrigin}
      theme={{ theme, density }}
      schema={schema as JsonValue}
      value={value as JsonValue}
      onChange={(next) => onChange(next as Record<string, unknown>)}
      extensions={[]}
      formProps={formProps}
      manifest={{ dependencies: context.dependencies }}
      sourceProvider={{ url: `${context.runtimeOrigin}/sources/${id}`, paths: context.paths }}
      onError={setError}
      onMessage={(message) => {
        if (message.type === "error") {
          const { phase, file, line, message: text } = message.error;
          setError(new Error(`${phase} ${file}${line === undefined ? "" : `:${line}`}: ${text}`));
        }
      }}
      className="w-full border-0"
    />
  );
}
