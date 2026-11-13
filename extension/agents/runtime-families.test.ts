import { describe, expect, it } from "vitest";
import type { RuntimeCatalogFamily } from "@flanksource/clicky-ui/ai";

import { agentRuntimeFamilies, availableRuntimeCatalog, runtimeOfSpecValue, specValueOfRuntime } from "./runtime-families";

const agent = (defaultModel: string | undefined, state: string, disabled = false) => ({
  mode: "agent",
  ...(defaultModel ? { defaultModel } : {}),
  disabled,
  availability: { state },
});

const catalog = [
  {
    family: "claude",
    provider: "anthropic",
    catalogPrefix: "anthropic",
    modes: [
      { mode: "api", defaultModel: "m-api", availability: { state: "available" } },
      agent("m-agent", "available"),
    ],
  },
  {
    family: "codex",
    provider: "openai",
    catalogPrefix: "openai",
    modes: [agent("x-agent", "missing_credentials")],
  },
  {
    family: "gemini",
    provider: "google",
    catalogPrefix: "google",
    modes: [agent("g-agent", "available", true)],
  },
  {
    family: "deepseek",
    provider: "deepseek",
    catalogPrefix: "deepseek",
    modes: [{ mode: "api", defaultModel: "d-api", availability: { state: "available" } }],
  },
] as unknown as RuntimeCatalogFamily[];

describe("availableRuntimeCatalog", () => {
  it("keeps only available, enabled agent modes with a default model, dropping emptied families", () => {
    expect(availableRuntimeCatalog(catalog, { mode: "agent" })).toEqual([
      { ...catalog[0], modes: [agent("m-agent", "available")] },
    ]);
  });

  it("keeps every available mode when no mode is required, still dropping unavailable and disabled ones", () => {
    expect(availableRuntimeCatalog(catalog, {})).toEqual([
      catalog[0],
      { ...catalog[3] },
    ]);
  });
});

describe("agentRuntimeFamilies", () => {
  it("offers the available agent runtimes as picker families", () => {
    const families = agentRuntimeFamilies(catalog);
    expect(families.map((family) => family.id)).toEqual(["claude"]);
    expect(families[0]?.modes.map((mode) => mode.id)).toEqual(["agent"]);
  });

  it("stays empty when no agent runtime is available instead of falling back to built-in families", () => {
    expect(agentRuntimeFamilies([catalog[1]!, catalog[3]!])).toEqual([]);
  });
});

describe("runtime and spec value", () => {
  const families = agentRuntimeFamilies(catalog);

  it("starts an unset runtime from the first agent runtime's default model", () => {
    expect(specValueOfRuntime(families, undefined)).toEqual({ mode: "agent", model: "m-agent" });
  });

  it("shows a configured runtime's own spec", () => {
    const spec = { model: "m-agent", effort: "high" };
    expect(specValueOfRuntime(families, { spec })).toBe(spec);
  });

  it("builds a runtime from a bar value, keeping the presets, and refuses one without a model", () => {
    expect(runtimeOfSpecValue({ mode: "agent", model: " m-agent ", id: " " }, ["fast"])).toEqual({ spec: { mode: "agent", model: "m-agent" }, presets: ["fast"] });
    expect(runtimeOfSpecValue({ mode: "agent", model: "" }, undefined)).toBeUndefined();
  });
});
