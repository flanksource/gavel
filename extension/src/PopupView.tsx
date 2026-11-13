import { useEffect, useState } from "react";
import { Button, SegmentedControl, Switch } from "@flanksource/clicky-ui";
import type { SiteSettings } from "../host/site-settings";
import { requestBackground } from "./messages";
import { draftOf, hasUnsavedChanges, settingsOf, type Draft } from "./popup-model";
import { SettingsForm, describe, type SettingsView } from "./SettingsForm";
import { effectiveSettings, siteOverrides, type GlobalDefaults } from "./settings";
import { discardGlobalDefaults, readGlobalDefaults, readSiteSettings, writeGlobalDefaults, writeSiteSettings } from "./storage";

type Loaded = { tabId: number; host: string; global: GlobalDefaults; site: SiteSettings };

/** Settings stored in an outdated or corrupt shape cannot be edited, only discarded and chosen again. */
function Unreadable({ message, discard, what }: { message: string; discard: () => void; what: string }) {
  return (
    <div role="alert" className="space-y-2 text-sm">
      <p className="text-destructive">{message}</p>
      <Button size="sm" variant="outline" onClick={discard}>Discard {what}</Button>
    </div>
  );
}

export function Popup() {
  const [loaded, setLoaded] = useState<Loaded>();
  const [globalDraft, setGlobalDraft] = useState<Draft>();
  const [siteDraft, setSiteDraft] = useState<Draft>();
  const [view, setView] = useState<SettingsView>("site");
  const [loadError, setLoadError] = useState("");
  const [globalError, setGlobalError] = useState("");
  const [siteError, setSiteError] = useState("");
  const [status, setStatus] = useState("");
  const [error, setError] = useState("");
  // Bumped after unreadable settings are discarded, to read them again.
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    setLoaded(undefined);
    setSiteDraft(undefined);
    setGlobalError("");
    setSiteError("");
    void (async () => {
      const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
      if (tab?.id === undefined || !tab.url) throw new Error("No active tab to open the grabber on");
      let global: GlobalDefaults;
      try {
        global = await readGlobalDefaults();
      } catch (cause) {
        setGlobalError(describe(cause));
        return;
      }
      setGlobalDraft(draftOf(global));
      let site: SiteSettings = {};
      try {
        site = await readSiteSettings(tab.id);
        setSiteDraft(draftOf(effectiveSettings(global, site)));
      } catch (cause) {
        setSiteError(describe(cause));
      }
      setLoaded({ tabId: tab.id, host: new URL(tab.url).host || tab.url, global, site });
    })().catch((cause: unknown) => setLoadError(describe(cause)));
  }, [attempt]);

  async function run(label: string, work: () => Promise<void>) {
    setError("");
    setStatus("");
    try {
      await work();
      setStatus(label);
    } catch (cause) {
      setError(describe(cause));
    }
  }

  if (loadError) return <p role="alert" className="p-3 text-sm text-destructive">{loadError}</p>;
  if (globalError) return <div className="p-3"><Unreadable message={globalError} what="stored global defaults" discard={() => void discardGlobalDefaults().then(() => setAttempt(count => count + 1), (cause: unknown) => setLoadError(describe(cause)))} /></div>;
  if (!loaded || !globalDraft) return <p role="status" className="p-3 text-sm">Loading…</p>;
  const { tabId, host, global, site } = loaded;
  const marks = new Set(Object.keys(site));
  const dirty = !siteError && siteDraft !== undefined && hasUnsavedChanges({ draft: siteDraft, saved: draftOf(effectiveSettings(global, site)) });

  const saveSite = () => run("Saved for this site.", async () => {
    if (!siteDraft) throw new Error("This page's settings could not be read");
    const overrides = siteOverrides(global, settingsOf(siteDraft));
    await writeSiteSettings(tabId, overrides);
    setLoaded({ ...loaded, site: overrides ?? {} });
    setSiteDraft(draftOf(effectiveSettings(global, overrides ?? {})));
  });
  // Whether the grabber is enabled is not a default to reset: it keeps following the switch.
  const resetSite = () => run("This site follows the global defaults.", async () => {
    const kept: SiteSettings = site.enabled ? { enabled: true } : {};
    await writeSiteSettings(tabId, kept);
    setLoaded({ ...loaded, site: kept });
    setSiteDraft(draftOf(effectiveSettings(global, kept)));
  });
  // The switch applies at once: the page's flag and the grabber on this tab change together, and a failed change is undone.
  const setEnabled = (enabled: boolean) => run(enabled ? "Enabled on this site." : "Disabled on this site.", async () => {
    if (!siteDraft) throw new Error("This page's settings could not be read");
    const { enabled: _enabled, ...others } = site;
    const next: SiteSettings = enabled ? { ...others, enabled: true } : others;
    await writeSiteSettings(tabId, next);
    try {
      await requestBackground({ type: enabled ? "activate" : "deactivate", tabId });
    } catch (cause) {
      if (enabled) await writeSiteSettings(tabId, site);
      throw cause;
    }
    setLoaded({ ...loaded, site: next });
    setSiteDraft({ ...siteDraft, enabled });
  });
  const saveGlobal = () => run("Saved global defaults.", async () => {
    const { session: _session, enabled: _enabled, ...defaults } = settingsOf(globalDraft);
    const saved = await writeGlobalDefaults(defaults);
    setLoaded({ ...loaded, global: saved });
    setGlobalDraft(draftOf(saved));
    if (!siteError) setSiteDraft(draftOf(effectiveSettings(saved, site)));
  });
  const captureTodo = () => run("Opened todo capture.", async () => {
    const settings = effectiveSettings(global, site);
    const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
    const url = new URL(chrome.runtime.getURL("todo-relay.html"));
    url.search = new URLSearchParams({ server: settings.server, sourceUrl: tab?.url ?? "", ...(settings.project ? { project: settings.project } : {}) }).toString();
    await chrome.tabs.create({ url: url.href });
  });

  return (
    <main className="text-foreground">
      <header className="space-y-2 border-b border-border p-3">
        <h1 className="truncate text-sm font-semibold" title={host}>{host}</h1>
        <Button size="sm" variant="outline" onClick={() => void captureTodo()}>Capture todo / saved captures</Button>
        {siteDraft && (
          <>
            <Switch label="Enabled on this site" checked={siteDraft.enabled} onChange={enabled => void setEnabled(enabled)} disabled={dirty} />
            <p className="text-xs text-muted-foreground">Shows the React Grab toolbar on this site now and on every page load, until switched off or closed from the panel.</p>
            {dirty && <p role="status" className="text-xs text-muted-foreground">Unsaved changes. Save them before enabling or disabling this site.</p>}
          </>
        )}
        <SegmentedControl
          size="sm"
          aria-label="Settings to edit"
          value={view}
          options={[{ id: "site", label: "This site" }, { id: "global", label: "Global" }]}
          onChange={setView}
        />
      </header>
      <section className="space-y-3 p-3">
        {view === "site" ? (
          siteError ? (
            <Unreadable message={siteError} what="this site's settings" discard={() => void writeSiteSettings(tabId, null).then(() => setAttempt(count => count + 1), (cause: unknown) => setError(describe(cause)))} />
          ) : siteDraft && (
            <>
              <SettingsForm id="site" view="site" draft={siteDraft} onChange={setSiteDraft} marks={marks} defaults={global} />
              <div className="flex flex-wrap gap-2">
                <Button size="sm" onClick={() => void saveSite()} disabled={!dirty}>Save for this site</Button>
                <Button size="sm" variant="outline" onClick={() => void resetSite()}>Reset site to defaults</Button>
              </div>
            </>
          )
        ) : (
          <>
            <SettingsForm id="global" view="global" draft={globalDraft} onChange={setGlobalDraft} marks={new Set()} defaults={global} />
            <Button size="sm" variant="outline" onClick={() => void saveGlobal()}>Save global defaults</Button>
          </>
        )}
      </section>
      {error && <p role="alert" className="p-3 text-sm text-destructive">{error}</p>}
      {status && <p role="status" className="p-3 text-sm text-muted-foreground">{status}</p>}
    </main>
  );
}
