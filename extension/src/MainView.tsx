import { useState } from "react";
import { Button } from "@flanksource/clicky-ui";
import { UiSliders } from "@flanksource/clicky-ui/icons";
import { EnableCard } from "../agents/EnableCard";
import type { AgentSessionGrouping, AgentSessionRuntime } from "../contracts/agent-session-store";
import { pickerSessions } from "./picker-sessions";
import { editField, promoteField, resetField } from "./popup-model";
import { ProjectSelect } from "./ProjectSelect";
import { ScopeMenu } from "./ScopeMenu";
import { DEFAULT_GROUPING, effectiveSettings, sameJson } from "./settings";
import { BlockedNotice, FirstRunSteps } from "./SetupNotices";
import { SyncStatus } from "./SyncStatus";
import { dashboardHost, type GavelChoices } from "./useGavelChoices";
import type { PopupStore, Ready } from "./usePopupStore";
import { useSessionRequests } from "./useSessionRequests";
import { useSyncStatus } from "./useSyncStatus";

type ViewProps = { ready: Ready; store: PopupStore; choices: GavelChoices };

const GROUPING_LABEL: Record<AgentSessionGrouping, string> = { page: "Each page", shared: "All pages" };

/** The project's "This site" menu: a project of the default dashboard chosen for this site only. On a site's own dashboard the project goes with the dashboard, which has the menu. */
function ProjectRow({ ready, store, choices }: ViewProps) {
  const { host, stored } = ready;
  const { project: ownProject, server: ownServer } = stored.site;
  return (
    <div className="flex items-center gap-2 border-b border-border px-4 py-2.5">
      <span className="w-[52px] shrink-0 text-muted-foreground">Project</span>
      <ProjectSelect className="min-w-0 flex-1" projects={choices.projects} value={effectiveSettings(stored.global, stored.site).project} onChange={project => void store.edit(before => editField(before, "project", project))} />
      {ownProject !== undefined && ownServer === undefined && (
        <ScopeMenu
          host={host}
          valueLabel={ownProject}
          defaultLabel={stored.global.project}
          onReset={() => void store.edit(before => resetField(before, "project"))}
          onPromote={() => void store.edit(before => promoteField(before, "project"))}
        />
      )}
    </div>
  );
}

function GroupingMenu({ ready, store, own }: { ready: Ready; store: PopupStore; own: AgentSessionGrouping }) {
  const { host, stored } = ready;
  return (
    <ScopeMenu
      host={host}
      valueLabel={GROUPING_LABEL[own]}
      defaultLabel={GROUPING_LABEL[stored.global.grouping ?? DEFAULT_GROUPING]}
      onReset={() => void store.edit(before => resetField(before, "grouping"))}
      onPromote={() => void store.edit(before => promoteField(before, "grouping"))}
    />
  );
}

/** The model a session started from the card runs on: the saved default until the card's bar picks another, and the new default again once that changes. */
function useStartRuntime(saved: AgentSessionRuntime | undefined) {
  const [picked, setPicked] = useState<{ base: AgentSessionRuntime | undefined; value: AgentSessionRuntime }>();
  const value = picked && sameJson(picked.base, saved) ? picked.value : saved;
  return { value, onChange: (next: AgentSessionRuntime) => setPicked({ base: saved, value: next }) };
}

/** The bundled todo form, which also lists the captures and comments still waiting to sync. */
function savedCapturesUrl({ effective, sourceUrl }: { effective: ReturnType<typeof effectiveSettings>; sourceUrl: string }): string {
  const url = new URL(chrome.runtime.getURL("todo-relay.html"));
  url.search = new URLSearchParams({ server: effective.server, sourceUrl, ...(effective.project ? { project: effective.project } : {}) }).toString();
  return url.href;
}

/** The popup's card: switch Gavel on for this host, then say where its comments go. */
export function MainView({ ready, store, choices, onSettings }: ViewProps & { onSettings: () => void }) {
  const { host, page, stored } = ready;
  const { global, site } = stored;
  const effective = effectiveSettings(global, site);
  const blocked = choices.withheld ? `Chrome is blocking Gavel from reaching ${dashboardHost(choices.withheld)}` : undefined;
  const connectionError = blocked ?? (choices.connectionError || undefined);
  const sync = useSyncStatus({ connectionError, connecting: choices.projects === undefined && connectionError === undefined });
  const sessions = useSessionRequests({ page, effective, store });
  const startRuntime = useStartRuntime(effective.runtime);
  const openSavedCaptures = () => void store.run(async () => { await chrome.tabs.create({ url: savedCapturesUrl({ effective, sourceUrl: page.url }) }); });
  const items = choices.agentSessions && [...choices.agentSessions, ...sessions.started.filter(entry => !choices.agentSessions?.some(listed => listed.id === entry.id))];
  const hasProject = effective.project !== undefined;
  const chooseFirstProject = (project: string) => void store.edit(before => editField(before, "project", project));

  return (
    <EnableCard
      host={host}
      enabled={{ value: site.enabled === true, onChange: enabled => void store.setEnabled(enabled), disabled: !hasProject && site.enabled !== true, disabledReason: "Choose a project to turn Gavel on" }}
      actions={<Button size="icon" variant="ghost" className="size-8 shrink-0" aria-label="Settings" onClick={onSettings}><UiSliders className="size-4" aria-hidden /></Button>}
      context={<ProjectRow ready={ready} store={store} choices={choices} />}
      setup={choices.withheld
        ? <BlockedNotice className="mx-4 mt-3" host={dashboardHost(choices.withheld)} onAllow={choices.grant} onUseAnother={onSettings} />
        : !hasProject && <FirstRunSteps dashboard={{ host: dashboardHost(effective.server), projects: choices.projects, error: choices.connectionError }} onChangeDashboard={onSettings} onProject={chooseFirstProject} />}
      target={{ value: site.target ?? "page", onChange: target => void store.chooseTarget(target) }}
      runtime={{ ...startRuntime, catalog: choices.catalog }}
      grouping={{ value: effective.grouping ?? DEFAULT_GROUPING, onChange: grouping => void store.edit(before => editField(before, "grouping", grouping)), menu: site.grouping === undefined ? undefined : <GroupingMenu ready={ready} store={store} own={site.grouping} /> }}
      sessions={{
        ...(site.session ? { page: site.session } : {}),
        ...(global.session ? { global: global.session } : {}),
        items: pickerSessions({ sessions: items, selectedId: undefined }),
        state: items ? "ready" : choices.error ? "error" : "loading",
      }}
      onStart={sessions.start}
      onUse={sessions.choose}
      startBlocked={blocked ?? (hasProject ? undefined : "Choose a project to start a session.")}
      status={store.status || undefined}
      error={store.error || choices.error}
      footer={<SyncStatus status={sync} onOpen={openSavedCaptures} />}
    />
  );
}
