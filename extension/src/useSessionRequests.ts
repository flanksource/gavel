import { useState } from "react";
import type { StartSessionRequest, UseSessionRequest } from "../agents/EnableCard";
import { GavelClient, type AgentSession } from "./gavel-client";
import type { EffectiveSettings } from "./settings";
import type { PopupStore, Ready } from "./usePopupStore";

/** Starting a session or choosing a running one for the page or the global target. Both reject on failure, for the card to show. */
export function useSessionRequests({ page, effective, store }: { page: Ready["page"]; effective: EffectiveSettings; store: PopupStore }) {
  // Sessions this popup started: gavel's list may not have them yet, and the card names them.
  const [started, setStarted] = useState<AgentSession[]>([]);
  const start = async ({ target, prefix, runtime }: StartSessionRequest) => {
    store.clearMessages();
    if (!effective.project) throw new Error("Choose a project to start a session");
    const session = await new GavelClient({ server: effective.server, project: effective.project }).startSession({ page, prefix, ...(runtime ? { runtime } : {}) });
    setStarted(previous => [...previous.filter(entry => entry.id !== session.id), session]);
    await store.adopt({ target, session: session.id });
    store.announce(`Started ${session.title}.`);
  };
  const choose = async ({ target, session }: UseSessionRequest) => {
    store.clearMessages();
    await store.adopt({ target, session: session.id });
    store.announce(`Using ${session.title}.`);
  };
  return { started, start, choose };
}
