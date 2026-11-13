import { describe, expect, it } from "vitest";
import { effectiveRuntime, runCaptureError, runDefaults, type RunContext } from "./runtime";

const context: RunContext = {
  runtimes: [], models: [], efforts: ["low", "high"],
  promptDefaults: { run: { model: "acme-model", mode: "api", presets: ["careful"], spec: { effort: "high" } } },
  runtimePresets: [{ id: "careful", name: "Careful", scope: "global", spec: { model: "preset-model", mode: "agent", budget: { cost: 2 } } }],
};

describe("todo runtime inheritance", () => {
  it("keeps prompt defaults sparse on the wire", () => {
    expect(runDefaults(context, "run")).toEqual({ step: "run", spec: {}, presets: ["careful"] });
    expect(effectiveRuntime(context, { step: "run", spec: {} })).toMatchObject({ model: "acme-model", mode: "api", effort: "high" });
  });

  it("projects selected presets and explicit overrides without dropping spec fields", () => {
    const run = { step: "run", presets: ["careful"], spec: { effort: "low", setup: { cwd: "/work/acme" } } };
    expect(effectiveRuntime(context, run)).toMatchObject({ model: "preset-model", mode: "agent", effort: "low", budget: { cost: 2 }, setup: { cwd: "/work/acme" } });
    expect(run.spec).toEqual({ effort: "low", setup: { cwd: "/work/acme" } });
  });

  it("blocks an invalid lifecycle choice instead of silently dropping the requested run", () => {
    expect(runCaptureError({ step: "run", spec: {} }, context, [{ name: "plan", label: "Plan", prompt: "plan", auxiliary: false }])).toContain("Step run is unavailable");
    expect(runCaptureError(undefined, undefined, undefined)).toBe("");
  });
});
