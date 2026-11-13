import {
  compactAISpecRuntime,
  familiesFromRuntimeCatalog,
  type AISpecRuntimeValue,
  type RuntimeCatalogFamily,
  type SpecRuntimeFamily,
} from "@flanksource/clicky-ui/ai";
import type { ChatModel } from "@flanksource/clicky-ui/chat";

import type { AgentSessionRuntime } from "../contracts/agent-session-store";

/** What the runtime bar chooses from: captain's (or gavel's) model catalog and the agent runtimes that are available. */
export type RuntimeCatalog = { models: ChatModel[]; families: SpecRuntimeFamily[] };

/**
 * Captain's catalog narrowed to enabled, available runtimes with a default
 * model, optionally of one `mode` — agent sessions pass `"agent"`, since only
 * the agent SDK keeps one chat session alive for follow-up messages.
 */
export function availableRuntimeCatalog(
  catalog: RuntimeCatalogFamily[],
  { mode: only }: { mode?: string },
): RuntimeCatalogFamily[] {
  return catalog.flatMap((family) => {
    const modes = family.modes.filter(
      (mode) =>
        (only === undefined || mode.mode === only) &&
        !mode.disabled &&
        mode.availability?.state === "available" &&
        mode.defaultModel,
    );
    return modes.length ? [{ ...family, modes }] : [];
  });
}

/** The picker families for the available agent runtimes; empty when there are none. */
export function agentRuntimeFamilies(runtimes: RuntimeCatalogFamily[]): SpecRuntimeFamily[] {
  const agents = availableRuntimeCatalog(runtimes, { mode: "agent" });
  // familiesFromRuntimeCatalog falls back to built-in families for an
  // empty list, so an empty agent catalog must stay empty here.
  return agents.length ? familiesFromRuntimeCatalog(agents) : [];
}

/** The configured runtime's spec, else the first agent runtime's default model to start from. */
export function specValueOfRuntime(
  families: SpecRuntimeFamily[],
  current: AgentSessionRuntime | undefined,
): AISpecRuntimeValue {
  if (current) return current.spec as AISpecRuntimeValue;
  const mode = families[0]?.modes[0];
  return {
    mode: "agent",
    ...(mode?.defaultModel ? { model: mode.defaultModel } : {}),
  };
}

/** The runtime a bar value chooses, or `undefined` while it names no model. */
export function runtimeOfSpecValue(
  value: AISpecRuntimeValue,
  presets: string[] | undefined,
): AgentSessionRuntime | undefined {
  const { model, ...spec } = compactAISpecRuntime(value);
  return model ? { spec: { ...spec, model }, ...(presets ? { presets } : {}) } : undefined;
}
