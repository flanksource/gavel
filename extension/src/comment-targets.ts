import type { SpecRuntimeFamily } from "@flanksource/clicky-ui/ai";
import type { CommentDefaultDelivery, CommentMention, CommentMentionable } from "@flanksource/clicky-ui/comments";
import type { AgentSessionRuntime } from "../contracts/agent-session-store";
import { freshTarget } from "./agent-retry";
import { DEFAULT_RUN_PROMPT, type CommentAgentSettings, type DeliverTarget } from "./gavel-client";

const SESSION_TARGET = "session";
const RUNTIME_PREFIX = "runtime:";

/** What a new comment can be addressed to: `@session` (the panel's chosen session) and one `@<runtime>` per available agent runtime. */
export function commentTargets(families: SpecRuntimeFamily[] | undefined): CommentMentionable[] {
  return [
    { id: SESSION_TARGET, name: SESSION_TARGET, kind: "agent" },
    ...(families ?? []).map((family): CommentMentionable => ({ id: `${RUNTIME_PREFIX}${family.id}`, name: family.id, kind: "agent" })),
  ];
}

export type TargetDeliveryOptions = {
  target: CommentMention;
  families: SpecRuntimeFamily[] | undefined;
  /** Where Send to agent delivers: the chosen session, or a run of the panel's runtime. */
  session: DeliverTarget;
  /** The page's run target for a runtime. */
  runTarget: (runtime: AgentSessionRuntime) => DeliverTarget;
  presets?: string[] | undefined;
};

/** The composer's preview of the page's comment agent rule: what a comment without a target is delivered to, or `null` when it is only posted. */
export function defaultDelivery({ agent, provenance }: CommentAgentSettings["agent"]): CommentDefaultDelivery | null {
  if (agent.enabled !== true) return null;
  const model = typeof agent.spec?.model === "string" ? agent.spec.model : undefined;
  const label = model ?? agent.prompt ?? DEFAULT_RUN_PROMPT;
  const layer = provenance.enabled?.layer;
  return {
    label,
    description: `New comments on this page are delivered to ${label} by the comment agent rule${layer ? ` in ${layer}` : ""}${agent.assign ? ", which also assigns the thread to it" : ""}. Pick a target to send this one elsewhere.`,
  };
}

/** Where a comment addressed to `target` is delivered: `@session` follows the panel, a runtime starts a new session of it. */
export function targetDelivery({ target, families, session, runTarget, presets }: TargetDeliveryOptions): DeliverTarget {
  if (target.id === SESSION_TARGET) return session;
  const family = target.id.startsWith(RUNTIME_PREFIX) ? families?.find(entry => entry.id === target.id.slice(RUNTIME_PREFIX.length)) : undefined;
  const model = family?.modes.find(mode => mode.defaultModel)?.defaultModel;
  if (!model) throw new Error(`@${target.name} is not an available comment target`);
  return freshTarget(runTarget({ spec: { mode: "agent", model }, ...(presets ? { presets } : {}) }));
}
