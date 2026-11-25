import type { AgentPickerSession } from "../agents/AgentPicker";
import type { AgentSession } from "./gavel-client";
import { orderedSessions, sessionLabel } from "./assign-items";

/** Gavel's sessions as the picker lists them (live first, ended managed ones marked as resuming), plus the remembered `selectedId` when gavel no longer offers it. */
export function pickerSessions(options: { sessions: AgentSession[] | undefined; selectedId: string | undefined }): AgentPickerSession[] {
  const { sessions, selectedId } = options;
  if (!sessions) return [];
  const items: AgentPickerSession[] = orderedSessions(sessions).map(session => ({
    id: session.id,
    source: session.source,
    title: sessionLabel(session),
    ...(session.model ? { model: session.model } : {}),
    ...(session.cwd ? { cwd: session.cwd } : {}),
    ...(session.status === "ended" || session.inCmux ? (session.status ? { status: session.status } : {}) : { status: "not in cmux" }),
  }));
  if (selectedId && !sessions.some(session => session.id === selectedId)) items.push({ id: selectedId, source: "agent", title: `${selectedId} (not live)`, status: "not live" });
  return items;
}
