import {
  compactAISpecRuntime,
  type AISpecRuntimeSpec,
  type AISpecRuntimeValue,
} from '@flanksource/clicky-ui/ai';

// PromptDetail is the resolved view of one registered prompt for a config layer,
// as returned by GET /api/settings/prompts/{id}. spec is the frontmatter as an
// api.Spec-shaped JSON object (camelCase keys mirroring AISpecRuntimeSpec); body
// is the .prompt template body; raw is the full .prompt text, echoed back on save
// as the merge base so unmodeled frontmatter keys survive.
export interface PromptDetail {
  id: string;
  scope: string;
  source: 'default' | 'inline' | 'file';
  path?: string;
  spec: Record<string, unknown>;
  body: string;
  raw: string;
}

// specToRuntimeValue binds a resolved prompt to the SpecRuntimeEditor value. The
// .prompt body is edited in the prompt section's "user" field — the editor has
// no separate document-body concept — so it is folded into prompt.user.
export function specToRuntimeValue(
  spec: Record<string, unknown>,
  body: string,
): AISpecRuntimeValue {
  const value = { ...(spec as AISpecRuntimeSpec) } as AISpecRuntimeValue;
  value.prompt = { ...(value.prompt ?? {}), user: body };
  return value;
}

// runtimeValueToPayload inverts specToRuntimeValue for a save: prompt.user is
// pulled back out as the document body and the remaining compacted spec keys are
// sent as frontmatter. compactAISpecRuntime prunes empty values and drops the
// client-only workflow layer.
export function runtimeValueToPayload(value: AISpecRuntimeValue): {
  spec: AISpecRuntimeSpec;
  body: string;
} {
  const compact = compactAISpecRuntime(value);
  const body = compact.prompt?.user ?? '';
  const spec: AISpecRuntimeSpec = { ...compact };
  if (compact.prompt) {
    const rest = { ...compact.prompt };
    delete rest.user;
    if (Object.keys(rest).length > 0) spec.prompt = rest;
    else delete spec.prompt;
  }
  return { spec, body };
}
