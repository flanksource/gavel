import { describe, expect, it, vi } from "vitest";
import type { SpecRuntimeFamily } from "@flanksource/clicky-ui/ai";
import { commentTargets, defaultDelivery, targetDelivery } from "./comment-targets";

const CLAUDE_MODEL = "anthropic/claude-sonnet-5-5";
const CODEX_MODEL = "openai/gpt-5.5-codex";
const families: SpecRuntimeFamily[] = [
  { id: "claude", label: "Claude", provider: "anthropic", modes: [{ id: "agent", label: "Agent", defaultModel: CLAUDE_MODEL }] },
  { id: "codex", label: "Codex", provider: "openai", modes: [{ id: "agent", label: "Agent", defaultModel: CODEX_MODEL }] },
];
const sessionDelivery = { kind: "session" as const, sessionId: "s-1" };
const runTarget = vi.fn((runtime: unknown) => ({ kind: "run" as const, prompt: "editor-comments", runOptions: { spec: runtime as Record<string, unknown> } }));

describe("commentTargets", () => {
  it("offers the panel's session and every available agent runtime by name", () => {
    expect(commentTargets(families)).toEqual([
      { id: "session", name: "session", kind: "agent" },
      { id: "runtime:claude", name: "claude", kind: "agent" },
      { id: "runtime:codex", name: "codex", kind: "agent" },
    ]);
  });

  it("offers only the session while no runtime catalog has loaded", () => {
    expect(commentTargets(undefined)).toEqual([{ id: "session", name: "session", kind: "agent" }]);
  });
});

describe("targetDelivery", () => {
  it("sends an @session comment wherever Send to agent would", () => {
    expect(targetDelivery({ target: { id: "session", name: "session", kind: "agent" }, families, session: sessionDelivery, runTarget })).toBe(sessionDelivery);
  });

  it("starts a new session of the named runtime with its default agent model, keeping the panel's presets", () => {
    const delivery = targetDelivery({ target: { id: "runtime:codex", name: "codex", kind: "agent" }, families, session: sessionDelivery, runTarget, presets: ["review"] });

    expect(runTarget).toHaveBeenLastCalledWith({ spec: { mode: "agent", model: CODEX_MODEL }, presets: ["review"] });
    expect(delivery).toEqual({ kind: "run", prompt: "editor-comments", runOptions: { spec: { spec: { mode: "agent", model: CODEX_MODEL }, presets: ["review"] } }, fresh: true });
  });

  it("refuses a target that is no longer offered", () => {
    expect(() => targetDelivery({ target: { id: "runtime:gemini", name: "gemini", kind: "agent" }, families, session: sessionDelivery, runTarget }))
      .toThrow("@gemini is not an available comment target");
  });
});

describe("defaultDelivery", () => {
  const PROJECT_LAYER = "db:project:";

  it("previews an enabled page rule by its model, naming the layer that enabled it and whether it assigns", () => {
    expect(defaultDelivery({
      found: true,
      agent: { enabled: true, assign: true, prompt: "editor-comments", spec: { model: CLAUDE_MODEL } },
      provenance: { enabled: { layer: PROJECT_LAYER } },
    })).toEqual({
      label: CLAUDE_MODEL,
      description: `New comments on this page are delivered to ${CLAUDE_MODEL} by the comment agent rule in ${PROJECT_LAYER}, which also assigns the thread to it. Pick a target to send this one elsewhere.`,
    });
  });

  it("names a rule without a model by its prompt", () => {
    expect(defaultDelivery({ found: true, agent: { enabled: true, prompt: "page-agent" }, provenance: {} })?.label).toBe("page-agent");
  });

  it("says nothing is delivered when no rule applies or the rule is disabled", () => {
    expect([
      defaultDelivery({ found: false, agent: {}, provenance: {} }),
      defaultDelivery({ found: true, agent: { enabled: false, prompt: "page-agent" }, provenance: { enabled: { layer: PROJECT_LAYER } } }),
    ]).toEqual([null, null]);
  });
});
