import type { AgentSessionGrouping, AgentSessionRuntime } from "../contracts/agent-session-store";
import { DEFAULT_GROUPING, sameJson, type EffectiveSettings } from "./settings";

/** The popup form: text fields are strings, and an empty optional field means "not set". An unset runtime is gavel's own default. */
export type Draft = { server: string; project: string; runtime: AgentSessionRuntime | undefined; grouping: AgentSessionGrouping; session: string; enabled: boolean };

export function draftOf(settings: EffectiveSettings): Draft {
  return { server: settings.server, project: settings.project ?? "", runtime: settings.runtime, grouping: settings.grouping ?? DEFAULT_GROUPING, session: settings.session ?? "", enabled: settings.enabled === true };
}

export function settingsOf(draft: Draft): EffectiveSettings {
  return {
    server: draft.server,
    ...(draft.project ? { project: draft.project } : {}),
    ...(draft.runtime ? { runtime: draft.runtime } : {}),
    grouping: draft.grouping,
    ...(draft.session ? { session: draft.session } : {}),
    ...(draft.enabled ? { enabled: true as const } : {}),
  };
}

/** Whether the form differs from what is saved, ignoring the enabled switch, which applies the moment it is flipped. */
export function hasUnsavedChanges(options: { draft: Draft; saved: Draft }): boolean {
  return !sameJson({ ...options.draft, enabled: false }, { ...options.saved, enabled: false });
}
