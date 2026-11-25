import { useState } from "react";
import { Button } from "@flanksource/clicky-ui";
import { MainView } from "./MainView";
import { effectiveSettings } from "./settings";
import { SettingsView } from "./SettingsView";
import { useGavelChoices } from "./useGavelChoices";
import { usePopupStore, type PopupStore, type Ready } from "./usePopupStore";

type Unreadable = Extract<PopupStore["load"], { kind: "unreadable" }>;

const WHAT = { global: "stored global defaults", site: "this site's settings" } as const;

/** Settings stored in an outdated or corrupt shape cannot be edited, only discarded and chosen again. */
function UnreadableSettings({ load, store }: { load: Unreadable; store: PopupStore }) {
  return (
    <div role="alert" className="space-y-2 p-3 text-sm">
      <p className="text-destructive">{load.message}</p>
      {store.error && <p className="text-destructive">{store.error}</p>}
      <Button size="sm" variant="outline" onClick={() => void store.discard(load)}>Discard {WHAT[load.scope]}</Button>
    </div>
  );
}

function ReadyPopup({ ready, store }: { ready: Ready; store: PopupStore }) {
  const [view, setView] = useState<"main" | "settings">("main");
  const choices = useGavelChoices(effectiveSettings(ready.stored.global, ready.stored.site));
  const show = (next: "main" | "settings") => {
    store.clearMessages();
    setView(next);
  };
  return (
    <main className="text-foreground">
      {view === "settings"
        ? <SettingsView ready={ready} store={store} choices={choices} onBack={() => show("main")} />
        : <MainView ready={ready} store={store} choices={choices} onSettings={() => show("settings")} />}
    </main>
  );
}

export function Popup() {
  const store = usePopupStore();
  const { load } = store;
  if (load.kind === "failed") return <p role="alert" className="p-3 text-sm text-destructive">{load.message}</p>;
  if (load.kind === "unreadable") return <UnreadableSettings load={load} store={store} />;
  if (load.kind === "loading") return <p role="status" className="p-3 text-sm">Loading…</p>;
  return <ReadyPopup ready={load} store={store} />;
}
