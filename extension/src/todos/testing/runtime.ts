import type { RunContext, RunStep } from "../runtime";

export const steps: RunStep[] = [
  { name: "plan", label: "Plan", prompt: "plan", auxiliary: false },
  { name: "run", label: "Run", prompt: "run", auxiliary: false },
];
export const context: RunContext = {
  runtimes: [{ family: "codex", provider: "openai", catalogPrefix: "openai", modes: [{ mode: "agent", schema: { type: "object", properties: { model: { type: "string" }, effort: { type: "string" }, budget: { type: "object", properties: { cost: { type: "number" }, timeout: { type: "string" } } } } } }] }],
  models: [{ id: "acme-model", provider: "openai", label: "Acme model", reasoning: true, configured: true, supportedEfforts: ["low", "high"] }],
  efforts: ["low", "high"],
  promptDefaults: { run: { model: "acme-model", mode: "agent" } },
};
