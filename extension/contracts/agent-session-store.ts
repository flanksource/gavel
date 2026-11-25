export type AgentSessionGrouping = "shared" | "page";
export type AgentSessionRuntime = {
  spec: Record<string, unknown> & { model: string };
  presets?: string[];
};
