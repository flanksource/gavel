import { useEffect, useId, useState } from "react";
import { Button, Field, InputField } from "@flanksource/clicky-ui";
import { UiArrowLeft } from "@flanksource/clicky-ui/icons";
import { InlineRuntimeBar } from "../agents/AgentPicker";
import { editField, promoteField, resetField } from "./popup-model";
import { ScopeMenu } from "./ScopeMenu";
import { effectiveSettings, serverOrigin } from "./settings";
import { BlockedNotice } from "./SetupNotices";
import { dashboardHost, describe, type GavelChoices } from "./useGavelChoices";
import type { PopupStore, Ready } from "./usePopupStore";

type SectionProps = { ready: Ready; store: PopupStore; choices: GavelChoices };

const muted = "text-muted-foreground";

function ConnectionStatus({ choices }: { choices: GavelChoices }) {
  const { projects, connectionError: error } = choices;
  if (error) return <span className="text-destructive">{error}</span>;
  if (!projects) return <span>Connecting…</span>;
  return (
    <span className="flex items-center gap-1.5">
      <span aria-hidden="true" className="size-2 shrink-0 rounded-full bg-green-500" />
      Connected · {projects.length} {projects.length === 1 ? "project" : "projects"}
    </span>
  );
}

/** The dashboard address is committed on Enter or blur; an address that is not an HTTP(S) URL is refused in place and nothing is stored. */
function DashboardSection({ ready, store, choices }: SectionProps) {
  const { host, stored } = ready;
  const { server: ownServer, project: ownProject } = stored.site;
  const { server } = effectiveSettings(stored.global, stored.site);
  const id = useId();
  const [text, setText] = useState(server);
  const [invalid, setInvalid] = useState("");
  useEffect(() => {
    setText(server);
    setInvalid("");
  }, [server]);
  const commit = () => {
    let origin: string;
    try {
      origin = serverOrigin(text);
    } catch (cause) {
      setInvalid(describe(cause));
      return;
    }
    setInvalid("");
    setText(origin);
    if (origin !== server) void store.edit(before => editField(before, "server", origin));
  };
  return (
    <section aria-label="Connection" className="border-b border-border px-4 py-3">
      <Field label="Gavel dashboard" htmlFor={id} error={invalid || undefined}>
        <InputField id={id} value={text} invalid={invalid !== ""} onChange={setText} onBlur={commit} onKeyDown={event => { if (event.key === "Enter") commit(); }} />
      </Field>
      <div className={`mt-2 flex items-start justify-between gap-2 ${muted}`}>
        <div className="min-w-0 flex-1">
          {choices.withheld ? <BlockedNotice host={dashboardHost(choices.withheld)} onAllow={choices.grant} /> : <ConnectionStatus choices={choices} />}
        </div>
        {ownServer !== undefined ? (
          <ScopeMenu
            host={host}
            valueLabel={ownProject === undefined ? dashboardHost(ownServer) : `${dashboardHost(ownServer)} with project ${ownProject}`}
            defaultLabel={dashboardHost(stored.global.server)}
            onReset={() => void store.edit(before => resetField(before, "server"))}
            onPromote={() => void store.edit(before => promoteField(before, "server"))}
          />
        ) : <span className="shrink-0">Same on every site</span>}
      </div>
    </section>
  );
}

/** The model a session starts on: for sessions started from the card, and for the sessions Gavel starts when routing is automatic. */
function NewSessionsSection({ ready, store, choices }: SectionProps) {
  const { host, stored } = ready;
  const { runtime } = effectiveSettings(stored.global, stored.site);
  const ownModel = stored.site.runtime?.spec.model;
  return (
    <section aria-label="New sessions" className="px-4 py-3">
      <div className="flex items-center justify-between gap-2">
        <h2 className="font-semibold">New sessions</h2>
        {ownModel !== undefined && (
          <ScopeMenu
            host={host}
            valueLabel={ownModel}
            defaultLabel={stored.global.runtime?.spec.model}
            onReset={() => void store.edit(before => resetField(before, "runtime"))}
            onPromote={() => void store.edit(before => promoteField(before, "runtime"))}
          />
        )}
      </div>
      <p className={`mb-2 mt-0.5 ${muted}`}>The model used when you start a session, or when routing is automatic.</p>
      <InlineRuntimeBar runtime={runtime} catalog={choices.catalog} onChange={next => void store.edit(before => editField(before, "runtime", next))} />
      {!runtime && <p className={`mt-1 ${muted}`}>Gavel's default</p>}
    </section>
  );
}

/** The dashboard address and the default model; every control applies at once. */
export function SettingsView({ ready, store, choices, onBack }: SectionProps & { onBack: () => void }) {
  return (
    <div className="flex flex-col text-xs">
      <header className="flex h-[52px] items-center gap-2 border-b border-border pl-2.5 pr-4">
        <Button size="icon" variant="ghost" className="size-8 shrink-0" aria-label="Back" onClick={onBack}><UiArrowLeft className="size-4" aria-hidden /></Button>
        <h1 className="min-w-0 flex-1 truncate text-sm font-semibold">Settings</h1>
        <span className={`truncate ${muted}`} title={ready.host}>{ready.host}</span>
      </header>
      <DashboardSection ready={ready} store={store} choices={choices} />
      <NewSessionsSection ready={ready} store={store} choices={choices} />
      {store.error && <p role="alert" className="break-words px-4 pb-3 text-destructive">{store.error}</p>}
    </div>
  );
}
