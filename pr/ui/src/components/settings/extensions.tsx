// Extension factories that adapt clicky-ui's JsonSchemaForm to the settings
// redesign: section-header glyphs (pre), and value/label adornments (post) for
// prompt overrides, glob chip editors, and per-field layer provenance badges.

import type { ChatModel } from "@flanksource/clicky-ui/chat";
import type { SpecRuntimeFamily } from "@flanksource/clicky-ui/ai";
import type {
  PostExtension,
  PreExtension,
} from "@flanksource/clicky-ui/components";
import { sectionIcon } from "../../icons/settings";
import {
  PromptOverrideField,
  type PromptOverrideValue,
} from "../PromptOverrideField";
import { AIDefaultsField, type AIDefaultsValue } from "./AIDefaultsField";
import { ChipsEditor } from "./widgets/ChipsEditor";
import { LayerBadge } from "./widgets/LayerBadge";
import {
  pathOf,
  promptIdOf,
  usesPromptPicker,
  type PromptDescriptor,
} from "./schema";
import { provenanceForPath, type GavelTrace } from "./provenance";

// Config paths whose string list is edited as glob/path chips rather than the
// generic array control (large, order-insensitive, append-and-dedupe lists).
const CHIP_PATHS = new Set(["commit.gitignore", "commit.allow"]);

// sectionIconPre adds each top-level section's glyph to its form heading. It keys
// off the field name, so nested fields that reuse a section name inherit it too.
export const sectionIconPre: PreExtension = (field) => {
  const Glyph = sectionIcon[field.key];
  if (!Glyph || field.labelIcon != null) return field;
  return {
    ...field,
    labelIcon: <Glyph className="shrink-0 text-[15px] text-muted-foreground" />,
  };
};

// promptPost replaces any prompt-override field's value with the shared
// PromptOverrideField, keyed by the schema's x-prompt-id → registry default.
function promptPost(
  registry: Record<string, PromptDescriptor>,
  scopeQuery: string,
  models: ChatModel[],
  families: SpecRuntimeFamily[]
): PostExtension {
  return (field, nodes) => {
    const id = promptIdOf(field.schema);
    if (!id) return nodes;
    const desc = registry[id];
    return {
      label: nodes.label,
      value: (
        <PromptOverrideField
          value={field.value as PromptOverrideValue | undefined}
          onChange={(next) => field.onChange(next)}
          description={desc?.description}
          id={id}
          title={desc?.title ?? id}
          scopeQuery={scopeQuery}
          models={models}
          families={families}
        />
      ),
    };
  };
}

function aiDefaultsPost(
  models: ChatModel[],
  families: SpecRuntimeFamily[]
): PostExtension {
  return (field, nodes) => {
    if (!usesPromptPicker(field.schema)) return nodes;
    return {
      label: nodes.label,
      value: (
        <AIDefaultsField
          value={field.value as AIDefaultsValue | undefined}
          onChange={(next) => field.onChange(next)}
          models={models}
          families={families}
        />
      ),
    };
  };
}

// chipsPost swaps the value editor of a glob/path list field for the ChipsEditor.
const chipsPost: PostExtension = (field, nodes) => {
  const path = pathOf(field.schema);
  if (!path || !CHIP_PATHS.has(path)) return nodes;
  const items = Array.isArray(field.value) ? (field.value as string[]) : [];
  return {
    label: nodes.label,
    value: (
      <ChipsEditor
        items={items}
        onChange={(next) => field.onChange(next)}
        placeholder="Add glob…"
        unit={path === "commit.allow" ? "paths" : "globs"}
      />
    ),
  };
};

// layerBadgePost appends a provenance pill to a field's label, naming the layer
// whose value is currently in effect. Runs last so it wraps whatever label the
// other post-extensions produced.
function layerBadgePost(trace: GavelTrace | null): PostExtension {
  return (field, nodes) => {
    const path = pathOf(field.schema);
    const layer = path ? provenanceForPath(trace, path) : undefined;
    if (!layer) return nodes;
    return {
      label: (
        <span className="inline-flex min-w-0 items-center">
          {nodes.label}
          <LayerBadge layer={layer} />
        </span>
      ),
      value: nodes.value,
    };
  };
}

export function buildPre(): PreExtension[] {
  return [sectionIconPre];
}

export function buildPost(
  registry: Record<string, PromptDescriptor>,
  scopeQuery: string,
  models: ChatModel[],
  families: SpecRuntimeFamily[],
  trace: GavelTrace | null
): PostExtension[] {
  return [
    aiDefaultsPost(models, families),
    promptPost(registry, scopeQuery, models, families),
    chipsPost,
    layerBadgePost(trace),
  ];
}
