import { useEffect, useState } from "react";
import type { RuntimeCatalog } from "../agents/runtime-families";
import { GavelClient, listProjects, type AgentSession } from "./gavel-client";
import { fetchRuntimeCatalog } from "./runtime-catalog";
import { hasServerAccess, requestServerAccess } from "./server-access";
import { serverOrigin } from "./settings";

export function describe(cause: unknown): string {
  return cause instanceof Error ? cause.message : String(cause);
}

/** A dashboard address as people say it: `localhost:9092` for `http://localhost:9092`. */
export function dashboardHost(server: string): string {
  return new URL(server).host;
}

/** The projects, runtimes and sessions a gavel dashboard offers. */
export type GavelChoices = ReturnType<typeof useGavelChoices>;

type Address = { origin: string } | { error: string };

function parseAddress(server: string): Address {
  try {
    return { origin: serverOrigin(server) };
  } catch (cause) {
    return { error: describe(cause) };
  }
}

/**
 * Loads nothing until `dashboard` is known (the popup calls this before its settings have been read). Each of the three loads owns
 * its error and clears it when it runs again: `error` is the first of them, for the card's alert, and `connectionError` is only
 * the failure that means the dashboard itself cannot be reached (an invalid address or a failed project list).
 */
export function useGavelChoices(dashboard: { server: string; project?: string | undefined } | undefined) {
  const [projects, setProjects] = useState<string[]>();
  const [catalog, setCatalog] = useState<RuntimeCatalog>();
  const [agentSessions, setAgentSessions] = useState<AgentSession[]>();
  const [connectionError, setConnectionError] = useState("");
  const [catalogError, setCatalogError] = useState("");
  const [sessionsError, setSessionsError] = useState("");
  // The server Chrome withholds access to, and a counter that re-runs the loads once granted.
  const [withheld, setWithheld] = useState<string>();
  const [granted, setGranted] = useState(0);
  useEffect(() => {
    setProjects(undefined);
    setCatalog(undefined);
    setWithheld(undefined);
    setConnectionError("");
    setCatalogError("");
    if (!dashboard) return;
    let cancelled = false;
    const address = parseAddress(dashboard.server);
    if ("error" in address) {
      setConnectionError(address.error);
      return;
    }
    const server = address.origin;
    void (async () => {
      if (!(await hasServerAccess(server))) {
        if (!cancelled) setWithheld(server);
        return;
      }
      fetchRuntimeCatalog({ server }).then(
        next => { if (!cancelled) setCatalog(next); },
        (cause: unknown) => { if (!cancelled) setCatalogError(describe(cause)); },
      );
      const names = await listProjects({ server });
      if (!cancelled) setProjects(names);
    })().catch((cause: unknown) => { if (!cancelled) setConnectionError(describe(cause)); });
    return () => { cancelled = true; };
  }, [dashboard?.server, granted]);
  const grant = () => {
    if (!withheld) return;
    // permissions.request needs this click's user gesture, so it is called synchronously here.
    requestServerAccess(withheld).then(
      ok => { if (ok) setGranted(count => count + 1); else setConnectionError(`Access to ${withheld} was not granted, so the extension cannot reach it.`); },
      (cause: unknown) => setConnectionError(describe(cause)),
    );
  };
  useEffect(() => {
    setAgentSessions(undefined);
    setSessionsError("");
    if (withheld || !dashboard?.project) return;
    const address = parseAddress(dashboard.server);
    // An invalid address is already the connection error the dashboard effect above reports; there is nothing to list from it.
    if ("error" in address) return;
    let cancelled = false;
    new GavelClient({ server: address.origin, project: dashboard.project }).agentSessions().then(
      list => { if (!cancelled) setAgentSessions(list); },
      (cause: unknown) => { if (!cancelled) setSessionsError(describe(cause)); },
    );
    return () => { cancelled = true; };
  }, [withheld, dashboard?.server, dashboard?.project, granted]);
  const error = connectionError || catalogError || sessionsError;
  return { projects, catalog, agentSessions, error, connectionError, withheld, grant };
}
