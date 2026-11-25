import { useCallback, useEffect, useRef, useState } from "react";
import type { PinnedTarget } from "../agents/EnableCard";
import type { SessionTarget, SiteSettings } from "../host/site-settings";
import { requestBackground } from "./messages";
import type { Stored } from "./popup-model";
import { effectiveSettings, sameJson, type GlobalDefaults } from "./settings";
import { discardGlobalDefaults, readGlobalDefaults, readSiteSettings, writeGlobalDefaults, writeSiteSettings } from "./storage";
import { describe } from "./useGavelChoices";

export type Ready = { tabId: number; host: string; page: { url: string; title: string }; stored: Stored };
export type PopupLoad =
  | { kind: "loading" }
  | { kind: "failed"; message: string }
  | { kind: "unreadable"; scope: "global" | "site"; tabId: number; message: string }
  | ({ kind: "ready" } & Ready);

type Reading = Exclude<PopupLoad, { kind: "loading" | "failed" }>;

async function readPopup(): Promise<Reading> {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  if (tab?.id === undefined || !tab.url) throw new Error("No active tab to open the grabber on");
  const host = new URL(tab.url).host || tab.url;
  let global: GlobalDefaults;
  try {
    global = await readGlobalDefaults();
  } catch (cause) {
    return { kind: "unreadable", scope: "global", tabId: tab.id, message: describe(cause) };
  }
  let site: SiteSettings;
  try {
    site = await readSiteSettings(tab.id);
  } catch (cause) {
    return { kind: "unreadable", scope: "site", tabId: tab.id, message: describe(cause) };
  }
  return { kind: "ready", tabId: tab.id, host, page: { url: tab.url, title: tab.title || host }, stored: { global, site } };
}

const TARGET_STATUS: Record<SessionTarget, string> = {
  page: "Comments go to this site's agent.",
  global: "Comments go to the global session.",
  auto: "Comments go to a session Gavel starts.",
};

/** An empty site entry is removed from the page rather than stored as `{}`. */
const entryOf = (site: SiteSettings): SiteSettings | null => (Object.keys(site).length === 0 ? null : site);

/** What storage holds right now: the extension's defaults and this page's site entry, both read afresh. */
async function readStored(tabId: number): Promise<Stored> {
  const [global, site] = await Promise.all([readGlobalDefaults(), readSiteSettings(tabId)]);
  return { global, site };
}

/** Writes what `after` changes in the global defaults and in the site entry. */
async function persist(options: { tabId: number; before: Stored; after: Stored; preserveSession: boolean }) {
  const { tabId, before, after, preserveSession } = options;
  if (!sameJson(before.global, after.global)) await writeGlobalDefaults(after.global);
  if (!sameJson(before.site, after.site)) await writeSiteSettings(tabId, entryOf(after.site), { preserveSession });
}

/** The open panel only learns of a changed default, session or target when the grabber is activated again; activating needs a project. */
async function reactivate(tabId: number, { global, site }: Stored) {
  if (site.enabled && effectiveSettings(global, site).project !== undefined) await requestBackground({ type: "activate", tabId });
}

/** What the popup read from the tab and the extension's storage, and how to read it again after unreadable settings are discarded. */
function useReading() {
  const [load, setLoad] = useState<PopupLoad>({ kind: "loading" });
  const [attempt, setAttempt] = useState(0);
  const current = useRef<Ready | undefined>(undefined);
  const reread = useCallback(() => setAttempt(count => count + 1), []);
  useEffect(() => {
    let cancelled = false;
    setLoad({ kind: "loading" });
    readPopup().then(
      reading => {
        if (cancelled) return;
        current.current = reading.kind === "ready" ? reading : undefined;
        setLoad(reading);
      },
      (cause: unknown) => { if (!cancelled) setLoad({ kind: "failed", message: describe(cause) }); },
    );
    return () => { cancelled = true; };
  }, [attempt]);
  const ready = (): Ready => {
    if (!current.current) throw new Error("This page's settings have not been read yet");
    return current.current;
  };
  const apply = (stored: Stored) => {
    current.current = { ...ready(), stored };
    setLoad({ kind: "ready", ...current.current });
  };
  return { load, ready, apply, reread };
}

/** The one status line and the one error the popup shows; `run` reports a failed action as the error and never swallows it. */
function useMessages() {
  const [status, setStatus] = useState("");
  const [error, setError] = useState("");
  const clearMessages = () => { setError(""); setStatus(""); };
  const announce = (message: string) => { clearMessages(); setStatus(message); };
  const run = async (work: () => Promise<void>, done?: string) => {
    clearMessages();
    try {
      await work();
      if (done) setStatus(done);
    } catch (cause) {
      setError(describe(cause));
    }
  };
  return { status, error, clearMessages, announce, run };
}

/**
 * Everything the popup stores, read once and kept as one `Stored`. Every action applies at once and in the order asked: it re-reads
 * what is stored now (another popup or the open panel may have changed it meanwhile), writes, reads again, and re-activates an
 * enabled site's grabber that has a project.
 */
export function usePopupStore() {
  const { load, ready, apply, reread } = useReading();
  const { status, error, clearMessages, announce, run } = useMessages();
  const writes = useRef<Promise<unknown>>(Promise.resolve());

  // Writes run one at a time, in the order they were asked for; a failure reaches the caller and does not stop later writes.
  const serialize = <T>(work: () => Promise<T>): Promise<T> => {
    const next = writes.current.then(work);
    writes.current = next.catch(() => undefined);
    return next;
  };
  /** The global defaults are written before the site entry, so a failed write may have changed part of the edit: show what storage holds, then fail. */
  const failWithStored = async (tabId: number, cause: unknown): Promise<never> => {
    try {
      apply(await readStored(tabId));
    } catch (rereadCause) {
      throw new Error(`${describe(cause)} (and the stored settings could not be read again: ${describe(rereadCause)})`);
    }
    throw cause;
  };
  const save = (options: { change: (before: Stored) => Stored; preserveSession: boolean }) => serialize(async () => {
    const { tabId } = ready();
    const before = await readStored(tabId);
    const after = options.change(before);
    if (sameJson(before, after)) return;
    try {
      await persist({ tabId, before, after, preserveSession: options.preserveSession });
    } catch (cause) {
      return failWithStored(tabId, cause);
    }
    const saved = await readStored(tabId);
    apply(saved);
    await reactivate(tabId, saved);
  });

  /** An edit of the defaults or this site's entry; `change` is `editField`, `resetField` or `promoteField` for the row's field. */
  const edit = (change: (before: Stored) => Stored, done?: string) => run(() => save({ change, preserveSession: true }), done);
  const chooseTarget = (target: SessionTarget) => edit(({ global, site: { target: _previous, ...others } }) => ({ global, site: target === "page" ? others : { ...others, target } }), TARGET_STATUS[target]);
  /** Makes `session` the one comments go to: the global defaults' session, or this site's own. */
  const adopt = ({ target, session }: { target: PinnedTarget; session: string }) => save({
    preserveSession: false,
    change: ({ global, site }) => (target === "global" ? { global: { ...global, session }, site } : { global, site: { ...site, session } }),
  });

  /** The switch applies at once: the page's flag and the grabber on this tab change together, and a failed change is undone. */
  const setEnabled = (enabled: boolean) => run(() => serialize(async () => {
    const { tabId } = ready();
    const { site } = await readStored(tabId);
    const { enabled: _previous, ...others } = site;
    await writeSiteSettings(tabId, entryOf(enabled ? { ...others, enabled: true } : others), { preserveSession: true });
    try {
      await requestBackground({ type: enabled ? "activate" : "deactivate", tabId });
    } catch (cause) {
      await writeSiteSettings(tabId, entryOf(site), { preserveSession: true });
      throw cause;
    }
    apply(await readStored(tabId));
  }), enabled ? "Enabled on this site." : "Disabled on this site.");

  /** Settings that cannot be read are discarded and chosen again: the global defaults, or this page's entry. */
  const discard = ({ scope, tabId }: { scope: "global" | "site"; tabId: number }) => run(async () => {
    if (scope === "global") await discardGlobalDefaults();
    else await writeSiteSettings(tabId, null);
    reread();
  });

  return { load, status, error, edit, chooseTarget, adopt, setEnabled, discard, run, announce, clearMessages };
}

export type PopupStore = ReturnType<typeof usePopupStore>;
