import { useEffect, useState } from "react";
import { Button, Field, InputField, Select } from "@flanksource/clicky-ui";
import { AgentPicker } from "../agents/AgentPicker";
import type { RuntimeCatalog } from "../agents/runtime-families";
import { GavelClient, fetchRuntimeCatalog, listProjects, type AgentSession } from "./gavel-client";
import { pickerSessions } from "./picker-sessions";
import { hasServerAccess, requestServerAccess } from "./server-access";
import type { Draft } from "./popup-model";
import { sameJson, serverOrigin, type GlobalDefaults } from "./settings";

export function describe(cause: unknown): string {
  return cause instanceof Error ? cause.message : String(cause);
}

/** The projects, runtimes and sessions a gavel dashboard offers for the settings being edited. */
function useGavelChoices(options: { draft: Draft; sessions: boolean }) {
  const { draft, sessions } = options;
  const [projects, setProjects] = useState<string[]>();
  const [catalog, setCatalog] = useState<RuntimeCatalog>();
  const [agentSessions, setAgentSessions] = useState<AgentSession[]>();
  const [error, setError] = useState("");
  // The server Chrome withholds access to, and a counter that re-runs the loads once granted.
  const [withheld, setWithheld] = useState<string>();
  const [granted, setGranted] = useState(0);
  useEffect(() => {
    setProjects(undefined);
    setCatalog(undefined);
    setWithheld(undefined);
    setError("");
    let cancelled = false;
    let server: string;
    try {
      server = serverOrigin(draft.server);
    } catch (cause) {
      setError(describe(cause));
      return;
    }
    void (async () => {
      if (!(await hasServerAccess(server))) {
        if (!cancelled) setWithheld(server);
        return;
      }
      fetchRuntimeCatalog({ server }).then(
        next => { if (!cancelled) setCatalog(next); },
        (cause: unknown) => { if (!cancelled) setError(describe(cause)); },
      );
      const names = await listProjects({ server });
      if (!cancelled) setProjects(names);
    })().catch((cause: unknown) => { if (!cancelled) setError(describe(cause)); });
    return () => { cancelled = true; };
  }, [draft.server, granted]);
  const grant = () => {
    if (!withheld) return;
    // permissions.request needs this click's user gesture, so it is called synchronously here.
    requestServerAccess(withheld).then(
      ok => { if (ok) setGranted(count => count + 1); else setError(`Access to ${withheld} was not granted, so the extension cannot reach it.`); },
      (cause: unknown) => setError(describe(cause)),
    );
  };
  useEffect(() => {
    setAgentSessions(undefined);
    if (!sessions || withheld || !draft.project) return;
    let cancelled = false;
    let client: GavelClient;
    try {
      client = new GavelClient({ server: serverOrigin(draft.server), project: draft.project });
    } catch {
      // The server error is already shown by the project list above.
      return;
    }
    client.agentSessions().then(
      list => { if (!cancelled) setAgentSessions(list); },
      (cause: unknown) => { if (!cancelled) setError(describe(cause)); },
    );
    return () => { cancelled = true; };
  }, [sessions, withheld, draft.server, draft.project, granted]);
  return { projects, catalog, agentSessions, error, withheld, grant };
}

function Overrides({ field, marks, text = "overrides default" }: { field: keyof Draft; marks: ReadonlySet<string>; text?: string }) {
  return marks.has(field) ? <span className="ml-1 rounded bg-primary/10 px-1 text-[10px] font-medium text-primary">{text}</span> : null;
}

export type SettingsView = "site" | "global";

/**
 * The one settings form. `site` edits this site's overrides (and its agent
 * session); `global` edits the extension-wide defaults the sites fall back to.
 */
export function SettingsForm(props: { id: string; view: SettingsView; draft: Draft; onChange: (draft: Draft) => void; marks: ReadonlySet<string>; defaults: GlobalDefaults }) {
  const { id, view, draft, onChange, marks, defaults } = props;
  const site = view === "site";
  const { projects, catalog, agentSessions, error, withheld, grant } = useGavelChoices({ draft, sessions: site });
  const set = (patch: Partial<Draft>) => onChange({ ...draft, ...patch });
  const projectNames = projects ?? [];
  return (
    <div className="space-y-3">
      <Field label={<>Gavel dashboard address<Overrides field="server" marks={marks} /></>} htmlFor={`${id}-server`}>
        <InputField id={`${id}-server`} value={draft.server} onChange={server => set({ server, project: "" })} />
      </Field>
      {withheld && (
        <div role="alert" className="space-y-2 text-xs">
          <p>Chrome is withholding this extension's access to {withheld}, so it cannot load projects or sessions.</p>
          <Button size="sm" onClick={grant}>Grant access to {withheld}</Button>
        </div>
      )}
      {!withheld && (
        <>
          <Field label={<>Project<Overrides field="project" marks={marks} /></>} htmlFor={`${id}-project`}>
            <Select
              id={`${id}-project`}
              value={draft.project}
              options={[
                { value: "", label: projects ? "Choose a project" : "Loading projects…", disabled: true },
                ...projectNames.map(name => ({ value: name, label: name })),
                ...(draft.project && projects && !projectNames.includes(draft.project) ? [{ value: draft.project, label: `${draft.project} (not on this dashboard)` }] : []),
              ]}
              onChange={event => set({ project: event.target.value })}
            />
          </Field>
          <div className="rounded border border-border">
            <AgentPicker
              heading={<>Agent<Overrides field="runtime" marks={marks} text="model overrides default" /><Overrides field="grouping" marks={marks} text="sessions override default" /></>}
              grouping={{ value: draft.grouping, onChange: grouping => set({ grouping }) }}
              runtime={{
                value: draft.runtime,
                emptyLabel: "Gavel's default",
                catalog,
                onChange: runtime => set({ runtime }),
                onClear: site && sameJson(draft.runtime, defaults.runtime) ? undefined : () => set({ runtime: site ? defaults.runtime : undefined }),
                clearLabel: site ? "Use global default" : "Use gavel's default",
              }}
              sessions={site ? {
                items: pickerSessions({ sessions: agentSessions, selectedId: draft.session || undefined }),
                state: agentSessions ? "ready" : error ? "error" : "loading",
                selectedId: draft.session || undefined,
                onSelect: session => set({ session: session.id }),
                onNew: () => set({ session: "" }),
                selectLabel: session => `Use session ${session.title}`,
              } : undefined}
            />
          </div>
        </>
      )}
      {error && <p role="alert" className="text-xs text-destructive">{error}</p>}
    </div>
  );
}
