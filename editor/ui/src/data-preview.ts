import type { RenderOptions } from "@flanksource/facet/browser-preview";
import { parse } from "yaml";

type JsonValue = RenderOptions["data"];

/** The pattern of a JSON or YAML document, which previews through a declared component. */
export const DATA_DOCUMENT = /\.(json|ya?ml)$/i;

/**
 * The data a preview component receives for a JSON or YAML document: the
 * parsed document, without the top-level `$preview` key that names the
 * component. A syntax error throws.
 */
export function previewDataOf(source: string, path: string): JsonValue {
  const value: unknown = /\.json$/i.test(path) ? JSON.parse(source) : parse(source, { uniqueKeys: true });
  if (value && typeof value === "object" && !Array.isArray(value)) {
    const { $preview: _component, ...data } = value as Record<string, unknown>;
    return data as JsonValue;
  }
  return (value ?? null) as JsonValue;
}
