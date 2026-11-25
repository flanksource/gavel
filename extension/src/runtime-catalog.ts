import type { RuntimeCatalogFamily } from "@flanksource/clicky-ui/ai";
import type { ChatModel } from "@flanksource/clicky-ui/chat";
import { agentRuntimeFamilies, type RuntimeCatalog } from "../agents/runtime-families";
import { getGavelJson } from "./gavel-client";

/** The model catalog and available agent runtimes used to choose a new session's runtime. */
export async function fetchRuntimeCatalog(options: { server: string; fetch?: typeof fetch }): Promise<RuntimeCatalog> {
  const [{ body: models }, { body: runtimes }] = await Promise.all([
    getGavelJson({ ...options, path: "/api/chat/models" }),
    getGavelJson({ ...options, path: "/api/chat/runtimes" }),
  ]);
  if (!Array.isArray(models) || models.length === 0) throw new Error("Gavel's AI model catalog is empty or invalid");
  if (!Array.isArray(runtimes) || runtimes.length === 0 || runtimes.some((family: unknown) => !Array.isArray((family as { modes?: unknown } | null)?.modes) || (family as { modes: unknown[] }).modes.length === 0))
    throw new Error("Gavel's AI runtime catalog is empty or invalid");
  const families = agentRuntimeFamilies(runtimes as RuntimeCatalogFamily[]);
  if (families.length === 0) throw new Error("Gavel reports no available agent runtime");
  return { models: models as ChatModel[], families };
}
