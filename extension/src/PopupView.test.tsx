import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { SiteSettings } from "../host/site-settings";
import { Popup } from "./PopupView";
import { listProjects } from "./gavel-client";
import { fetchRuntimeCatalog } from "./runtime-catalog";
import { requestBackground } from "./messages";
import { hasServerAccess, requestServerAccess } from "./server-access";
import type { GlobalDefaults } from "./settings";
import type { SyncStatus } from "./sync-status";
import { discardGlobalDefaults, readGlobalDefaults, readSiteSettings, writeGlobalDefaults, writeSiteSettings } from "./storage";
import { useSyncStatus } from "./useSyncStatus";

const agentSessions = vi.hoisted(() => vi.fn());
const startSession = vi.hoisted(() => vi.fn());
const clientOptions = vi.hoisted(() => [] as unknown[]);
vi.mock("./storage");
vi.mock("./messages");
vi.mock("./useSyncStatus", () => ({ useSyncStatus: vi.fn() }));
vi.mock("./runtime-catalog", () => ({ fetchRuntimeCatalog: vi.fn() }));
vi.mock("./server-access", () => ({ hasServerAccess: vi.fn(), requestServerAccess: vi.fn() }));
vi.mock("./gavel-client", () => ({
  listProjects: vi.fn(),
  GavelClient: class {
    agentSessions = agentSessions;
    startSession = startSession;
    constructor(options: unknown) { clientOptions.push(options); }
  },
}));

const gavel = "http://localhost:9092";
const otherGavel = "https://other.example.com";
const claude = { spec: { model: "claude-example" } };
const codex = { spec: { model: "codex-example" } };
const globalDefaults: GlobalDefaults = { server: gavel, project: "acme", runtime: claude };
const CATALOG = {
  models: [],
  families: [{ id: "claude", label: "Claude", provider: "anthropic", modes: [{ id: "agent", label: "Agent", defaultModel: "claude-example", provider: "anthropic" }] }],
};
const MODELS = {
  models: [
    { id: "claude-example", provider: "anthropic", label: "Claude example", reasoning: false, configured: true },
    { id: "codex-example", provider: "openai", label: "Codex example", reasoning: false, configured: true },
  ],
  families: [
    { id: "claude", label: "Claude", provider: "anthropic", modes: [{ id: "agent", label: "Agent", defaultModel: "claude-example", provider: "anthropic" }] },
    { id: "codex", label: "Codex", provider: "openai", modes: [{ id: "agent", label: "Agent", defaultModel: "codex-example", provider: "openai" }] },
  ],
};
const codexPicked = { spec: { model: "codex-example", mode: "agent" } };
const SESSION = { id: "s-1", source: "codex", title: "Review changes", model: "codex-example", inCmux: true };
const STARTED = { id: "s-new", source: "claude", title: "gavel-example.com-a1b2c3", inCmux: false };
const PAGE = { url: "https://example.com/docs", title: "Docs" };
const preserved = { preserveSession: true };
const connected: SyncStatus = { kind: "connected" };

/** Renders the popup over fake storage that keeps what is written, and returns what it holds. `setGlobal` is another popup writing the defaults. */
function open(options: { global?: GlobalDefaults; site?: SiteSettings } = {}) {
  let savedGlobal = options.global ?? globalDefaults;
  let saved = options.site ?? {};
  vi.mocked(readGlobalDefaults).mockImplementation(async () => savedGlobal);
  vi.mocked(writeGlobalDefaults).mockImplementation(async defaults => { savedGlobal = defaults; return defaults; });
  vi.mocked(readSiteSettings).mockImplementation(async () => saved);
  vi.mocked(writeSiteSettings).mockImplementation(async (_tab, next, writeOptions) => {
    saved = { ...next, ...(writeOptions?.preserveSession && saved.session ? { session: saved.session } : {}) };
  });
  render(<Popup />);
  return { site: () => saved, global: () => savedGlobal, setGlobal: (next: GlobalDefaults) => { savedGlobal = next; } };
}

beforeEach(() => {
  vi.stubGlobal("chrome", {
    tabs: { query: async () => [{ id: 7, url: PAGE.url, title: PAGE.title }], create: vi.fn() },
    runtime: { getURL: (path: string) => `chrome-extension://example-id/${path}` },
  });
  vi.mocked(listProjects).mockResolvedValue(["acme", "docs"]);
  vi.mocked(fetchRuntimeCatalog).mockResolvedValue(CATALOG);
  vi.mocked(hasServerAccess).mockResolvedValue(true);
  vi.mocked(requestServerAccess).mockResolvedValue(true);
  vi.mocked(useSyncStatus).mockReturnValue(connected);
  agentSessions.mockResolvedValue([SESSION]);
  startSession.mockResolvedValue(STARTED);
  vi.mocked(writeSiteSettings).mockResolvedValue();
  vi.mocked(writeGlobalDefaults).mockImplementation(async defaults => defaults);
  vi.mocked(requestBackground).mockResolvedValue(undefined);
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  vi.clearAllMocks();
  clientOptions.length = 0;
});

const nameField = () => screen.getByLabelText<HTMLInputElement>(/New session name/);
const enabledSwitch = () => screen.findByRole("switch", { name: "Gavel on example.com" });
/** The project select once the dashboard's projects are in it. */
const projectSelect = () => waitFor(() => {
  const select = screen.getByRole<HTMLSelectElement>("combobox", { name: "Project" });
  expect(select.options.length).toBeGreaterThan(1);
  return select;
});
const ROUTES_NAME = "Comments from this site go to";
const routes = () => screen.getByRole("group", { name: ROUTES_NAME });
const route = (name: string) => within(routes()).getByRole<HTMLInputElement>("radio", { name });
const activations = () => vi.mocked(requestBackground).mock.calls.filter(([request]) => request.type === "activate");
const openSettings = async () => fireEvent.click(await screen.findByRole("button", { name: "Settings" }));
const pickModel = (label: string) => {
  fireEvent.click(screen.getByTitle(/^Model — /));
  fireEvent.mouseDown(screen.getByRole("option", { name: new RegExp(`^${label}`) }));
};
const dashboardField = () => screen.getByRole<HTMLInputElement>("textbox", { name: "Gavel dashboard" });

describe("main view", () => {
  it("leads with the host, the switch and the three routes, with no Advanced, no save buttons and no capture button", async () => {
    open({ site: { enabled: true } });
    expect((await enabledSwitch()).getAttribute("aria-checked")).toBe("true");
    expect(screen.getByText("example.com")).toBeTruthy();
    expect(["This site's agent", "Global session", "Automatic"].map(name => [name, route(name).checked])).toEqual([["This site's agent", true], ["Global session", false], ["Automatic", false]]);
    expect(screen.getByRole("button", { name: "Settings" })).toBeTruthy();
    expect(screen.queryByText(/Advanced/)).toBeNull();
    expect(screen.queryByRole("button", { name: /^Save/ })).toBeNull();
    expect(screen.queryByRole("button", { name: /Reset site/ })).toBeNull();
    expect(screen.queryByRole("button", { name: /Capture todo/ })).toBeNull();
    expect(screen.queryByText(/Unsaved changes/)).toBeNull();
  });

  it("shows the project row with the dashboard's projects, and the model bar for a session not started yet", async () => {
    open({ site: { enabled: true } });
    const select = await projectSelect();
    expect([...select.options].map(option => option.textContent)).toEqual(["Choose a project", "acme", "docs"]);
    expect(select.value).toBe("acme");
    expect(nameField().value).toBe("gavel-example.com");
    expect(await screen.findByLabelText("Agent runtime")).toBeTruthy();
    expect(fetchRuntimeCatalog).toHaveBeenCalledWith({ server: gavel });
  });

  it("names a stored project the dashboard does not list", async () => {
    open({ global: { ...globalDefaults, project: "gone" }, site: { enabled: true } });
    expect((await projectSelect()).textContent).toContain("gone (not on this dashboard)");
  });

  it("explains an off site instead of showing the routes", async () => {
    open();
    expect((await enabledSwitch()).getAttribute("aria-checked")).toBe("false");
    expect(screen.getByText("Gavel is off on this site")).toBeTruthy();
    expect(screen.queryByRole("group", { name: ROUTES_NAME })).toBeNull();
  });

  it("shows the sync status as the only footer label, and a sync error opens the saved captures", async () => {
    vi.mocked(useSyncStatus).mockReturnValue({ kind: "sync-error", failed: 2, detail: "HTTP 500" });
    open({ site: { enabled: true } });
    fireEvent.click(await screen.findByRole("button", { name: "Sync error" }));
    expect(document.querySelector("footer")?.textContent).toBe("Sync error");
    await waitFor(() => expect(chrome.tabs.create).toHaveBeenCalledOnce());
    const [{ url }] = vi.mocked(chrome.tabs.create).mock.calls[0] as [{ url: string }];
    expect(url).toBe(`chrome-extension://example-id/todo-relay.html?server=${encodeURIComponent(gavel)}&sourceUrl=${encodeURIComponent(PAGE.url)}&project=acme`);
    await waitFor(() => expect(useSyncStatus).toHaveBeenLastCalledWith({ connectionError: undefined, connecting: false }));
  });

  it("tells the footer the dashboard is still connecting until its projects arrive", async () => {
    vi.mocked(listProjects).mockReturnValue(new Promise(() => {}));
    open({ site: { enabled: true } });
    await enabledSwitch();
    expect(useSyncStatus).toHaveBeenLastCalledWith({ connectionError: undefined, connecting: true });
  });

  it("keeps a failed session list out of the footer's connection error, and shows it in the card's alert", async () => {
    agentSessions.mockRejectedValue(new Error("sessions unavailable"));
    open({ site: { enabled: true } });
    expect((await screen.findByRole("alert")).textContent).toContain("sessions unavailable");
    expect(useSyncStatus).toHaveBeenLastCalledWith({ connectionError: undefined, connecting: false });
  });

  it("keeps a failed runtime catalog out of the footer's connection error, and shows it in the card's alert", async () => {
    vi.mocked(fetchRuntimeCatalog).mockRejectedValue(new Error("catalog unavailable"));
    open({ site: { enabled: true } });
    expect((await screen.findByRole("alert")).textContent).toContain("catalog unavailable");
    await projectSelect();
    expect(useSyncStatus).toHaveBeenLastCalledWith({ connectionError: undefined, connecting: false });
  });

  it("reads Connected in the footer when the dashboard answers", async () => {
    open({ site: { enabled: true } });
    await enabledSwitch();
    expect(document.querySelector("footer")?.textContent).toBe("Connected");
  });
});

describe("enabling", () => {
  it("applies at once: writes the flag, activates the grabber and shows the routes", async () => {
    open();
    fireEvent.click(await enabledSwitch());
    await waitFor(() => expect(activations()).toHaveLength(1));
    expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { enabled: true }, preserved);
    expect(await screen.findByRole("group", { name: ROUTES_NAME })).toBeTruthy();
    expect(await screen.findByText("Enabled on this site.")).toBeTruthy();
  });

  it("deactivates the grabber and drops the flag when switched off", async () => {
    open({ site: { enabled: true, session: "s-1" } });
    fireEvent.click(await enabledSwitch());
    await waitFor(() => expect(requestBackground).toHaveBeenCalledWith({ type: "deactivate", tabId: 7 }));
    expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { session: "s-1" }, preserved);
    expect(await screen.findByText("Gavel is off on this site")).toBeTruthy();
  });

  it("undoes the flag and shows the failure when the grabber cannot be activated", async () => {
    vi.mocked(requestBackground).mockRejectedValue(new Error("Cannot reach the tab"));
    open();
    fireEvent.click(await enabledSwitch());
    expect((await screen.findByRole("alert")).textContent).toContain("Cannot reach the tab");
    expect(writeSiteSettings).toHaveBeenNthCalledWith(1, 7, { enabled: true }, preserved);
    expect(writeSiteSettings).toHaveBeenLastCalledWith(7, null, preserved);
    expect((await enabledSwitch()).getAttribute("aria-checked")).toBe("false");
  });

  it("is disabled with its reason, and first-run setup shows, while no project is chosen", async () => {
    open({ global: { server: gavel } });
    const toggle = await enabledSwitch();
    expect((toggle as HTMLButtonElement).disabled).toBe(true);
    expect(screen.getByTitle("Choose a project to turn Gavel on")).toBeTruthy();
    expect(screen.getByText("Set up Gavel")).toBeTruthy();
    expect(screen.getByText("Choose a project, then switch Gavel on above.")).toBeTruthy();
    expect(screen.getByText("Dashboard")).toBeTruthy();
    expect(await screen.findByText("localhost:9092 · connected")).toBeTruthy();
    expect(screen.queryByRole("group", { name: ROUTES_NAME })).toBeNull();
    expect(screen.queryAllByRole("radio")).toEqual([]);
  });

  it("stores the first project chosen as the default for every site, then leaves setup and unlocks the switch", async () => {
    open({ global: { server: gavel } });
    await enabledSwitch();
    fireEvent.change(await projectSelect(), { target: { value: "acme" } });
    await waitFor(() => expect(writeGlobalDefaults).toHaveBeenCalledWith({ server: gavel, project: "acme" }));
    expect(writeSiteSettings).not.toHaveBeenCalled();
    await waitFor(() => expect(screen.queryByText("Set up Gavel")).toBeNull());
    expect((await enabledSwitch() as HTMLButtonElement).disabled).toBe(false);
    expect(screen.getByText("Gavel is off on this site")).toBeTruthy();
  });

  it("lets Change in the setup steps open the settings view", async () => {
    open({ global: { server: gavel } });
    fireEvent.click(await screen.findByRole("button", { name: "Change" }));
    expect(screen.getByRole("heading", { name: "Settings" })).toBeTruthy();
  });

  it("says a failed dashboard load in the setup step instead of connected", async () => {
    vi.mocked(listProjects).mockRejectedValue(new Error("connection refused"));
    open({ global: { server: gavel } });
    expect((await screen.findAllByText(/connection refused/)).length).toBeGreaterThan(0);
    expect(screen.queryByText("localhost:9092 · connected")).toBeNull();
    expect(useSyncStatus).toHaveBeenLastCalledWith({ connectionError: "connection refused", connecting: false });
  });
});

describe("choosing a project", () => {
  it("stores a project that differs from the default as this site's own, and offers the This site menu for it", async () => {
    open({ site: { enabled: true } });
    expect(screen.queryByRole("button", { name: "This site" })).toBeNull();
    fireEvent.change(await projectSelect(), { target: { value: "docs" } });
    await waitFor(() => expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { enabled: true, project: "docs" }, preserved));
    expect(writeGlobalDefaults).not.toHaveBeenCalled();
    await waitFor(() => expect(activations()).toHaveLength(1));
    expect(await screen.findByRole("button", { name: "This site" })).toBeTruthy();
  });

  it("writes what resetField produces for Use the default here", async () => {
    open({ site: { enabled: true, project: "docs" } });
    fireEvent.click(await screen.findByRole("button", { name: "This site" }));
    expect(screen.getByText("docs applies only to example.com.")).toBeTruthy();
    fireEvent.click(screen.getByRole("menuitem", { name: "Use the default here acme" }));
    await waitFor(() => expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { enabled: true }, preserved));
    expect(writeGlobalDefaults).not.toHaveBeenCalled();
    await waitFor(() => expect(screen.queryByRole("button", { name: "This site" })).toBeNull());
    expect((await projectSelect()).value).toBe("acme");
  });

  it("writes what promoteField produces for Make this the default for all sites", async () => {
    open({ site: { enabled: true, project: "docs" } });
    fireEvent.click(await screen.findByRole("button", { name: "This site" }));
    fireEvent.click(screen.getByRole("menuitem", { name: "Make this the default for all sites" }));
    await waitFor(() => expect(writeGlobalDefaults).toHaveBeenCalledWith({ ...globalDefaults, project: "docs" }));
    expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { enabled: true }, preserved);
    await waitFor(() => expect(screen.queryByRole("button", { name: "This site" })).toBeNull());
    expect((await projectSelect()).value).toBe("docs");
  });

  it("shows a failed write as an error and keeps what was stored", async () => {
    const storage = open({ site: { enabled: true } });
    vi.mocked(writeSiteSettings).mockRejectedValue(new Error("This page's storage is not available to the grabber: blocked"));
    fireEvent.change(await projectSelect(), { target: { value: "docs" } });
    expect((await screen.findByRole("alert")).textContent).toContain("storage is not available");
    expect(writeSiteSettings).toHaveBeenCalledWith(7, { enabled: true, project: "docs" }, preserved);
    expect(storage.site()).toEqual({ enabled: true });
    expect(storage.global()).toEqual(globalDefaults);
    expect(activations()).toEqual([]);
    expect((await projectSelect()).value).toBe("acme");
  });
});

describe("a site on its own dashboard", () => {
  it("has no project from the default dashboard: editing the dashboard shows the setup steps and does not activate the grabber", async () => {
    open({ site: { enabled: true } });
    await openSettings();
    fireEvent.change(dashboardField(), { target: { value: otherGavel } });
    fireEvent.keyDown(dashboardField(), { key: "Enter" });
    await waitFor(() => expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { enabled: true, server: otherGavel }, preserved));
    fireEvent.click(await screen.findByRole("button", { name: "This site" }));
    expect(screen.getByText("other.example.com applies only to example.com.")).toBeTruthy();
    expect(activations()).toEqual([]);
    fireEvent.click(screen.getByRole("button", { name: "Back" }));
    expect(await screen.findByText("Set up Gavel")).toBeTruthy();
    expect(activations()).toEqual([]);
  });

  it("activates the grabber once a project is chosen for it", async () => {
    open({ global: { server: gavel }, site: { enabled: true, server: otherGavel } });
    await screen.findByText("Set up Gavel");
    fireEvent.change(await projectSelect(), { target: { value: "docs" } });
    await waitFor(() => expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { enabled: true, server: otherGavel, project: "docs" }, preserved));
    await waitFor(() => expect(activations()).toHaveLength(1));
    await waitFor(() => expect(screen.queryByText("Set up Gavel")).toBeNull());
  });

  it("does not activate the grabber when the dashboard is reset on an enabled site with no default project", async () => {
    open({ global: { server: gavel }, site: { enabled: true, server: otherGavel } });
    fireEvent.click(await screen.findByRole("button", { name: "Change" }));
    fireEvent.click(await screen.findByRole("button", { name: "This site" }));
    fireEvent.click(screen.getByRole("menuitem", { name: /^Use the default here/ }));
    await waitFor(() => expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { enabled: true }, preserved));
    await waitFor(() => expect(screen.getByText("Same on every site")).toBeTruthy());
    expect(activations()).toEqual([]);
  });

  it("offers no Project menu: the project goes with the dashboard", async () => {
    open({ site: { enabled: true, server: otherGavel, project: "docs" } });
    expect((await projectSelect()).value).toBe("docs");
    expect(screen.queryByRole("button", { name: "This site" })).toBeNull();
  });

  it("uses no project of the default dashboard when its own has none", async () => {
    open({ site: { enabled: true, server: otherGavel } });
    expect(await screen.findByText("Set up Gavel")).toBeTruthy();
    expect((await projectSelect()).value).toBe("");
  });
});

describe("choosing where comments go", () => {
  it("writes the global target at once, keeps the site's session, and re-activates an enabled site", async () => {
    open({ site: { session: "s-1", enabled: true } });
    await enabledSwitch();
    fireEvent.click(route("Global session"));
    await waitFor(() => expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { session: "s-1", enabled: true, target: "global" }, preserved));
    await waitFor(() => expect(activations()).toHaveLength(1));
    expect(route("Global session").checked).toBe(true);
    fireEvent.click(route("This site's agent"));
    await waitFor(() => expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { session: "s-1", enabled: true }, preserved));
    await waitFor(() => expect(activations()).toHaveLength(2));
  });

  it("writes Automatic as the target and keeps the site's session", async () => {
    open({ site: { session: "s-1", enabled: true } });
    await enabledSwitch();
    fireEvent.click(route("Automatic"));
    await waitFor(() => expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { session: "s-1", enabled: true, target: "auto" }, preserved));
    await waitFor(() => expect(activations()).toHaveLength(1));
    expect(route("Automatic").checked).toBe(true);
  });

  it("stores a changed grouping as a site override, with the target kept", async () => {
    open({ site: { enabled: true, target: "auto" } });
    await enabledSwitch();
    const choice = screen.getByRole("radiogroup", { name: "One session for" });
    expect(within(choice).getByRole<HTMLInputElement>("radio", { name: "Each page" }).getAttribute("aria-checked")).toBe("true");
    fireEvent.click(within(choice).getByRole("radio", { name: "All pages" }));
    await waitFor(() => expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { enabled: true, target: "auto", grouping: "shared" }, preserved));
    await waitFor(() => expect(activations()).toHaveLength(1));
    await waitFor(() => expect(within(screen.getByRole("radiogroup", { name: "One session for" })).getByRole("radio", { name: "All pages" }).getAttribute("aria-checked")).toBe("true"));
    expect(await screen.findByRole("button", { name: "This site" })).toBeTruthy();
  });

  it("offers no This site menu beside a grouping that follows the default", async () => {
    open({ site: { enabled: true, target: "auto" } });
    await enabledSwitch();
    expect(screen.getByRole("radiogroup", { name: "One session for" })).toBeTruthy();
    expect(screen.queryByRole("button", { name: "This site" })).toBeNull();
  });

  it("writes what resetField produces for a site's own grouping", async () => {
    open({ site: { enabled: true, target: "auto", grouping: "shared" } });
    fireEvent.click(await screen.findByRole("button", { name: "This site" }));
    expect(screen.getByText("All pages applies only to example.com.")).toBeTruthy();
    fireEvent.click(screen.getByRole("menuitem", { name: "Use the default here Each page" }));
    await waitFor(() => expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { enabled: true, target: "auto" }, preserved));
    expect(writeGlobalDefaults).not.toHaveBeenCalled();
    await waitFor(() => expect(screen.queryByRole("button", { name: "This site" })).toBeNull());
  });

  it("writes what promoteField produces for a site's own grouping", async () => {
    open({ site: { enabled: true, target: "auto", grouping: "shared" } });
    fireEvent.click(await screen.findByRole("button", { name: "This site" }));
    fireEvent.click(screen.getByRole("menuitem", { name: "Make this the default for all sites" }));
    await waitFor(() => expect(writeGlobalDefaults).toHaveBeenCalledWith({ ...globalDefaults, grouping: "shared" }));
    expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { enabled: true, target: "auto" }, preserved);
    await waitFor(() => expect(screen.queryByRole("button", { name: "This site" })).toBeNull());
  });
});

describe("writes", () => {
  const slowFirstWrite = () => {
    const write = vi.mocked(writeSiteSettings).getMockImplementation();
    if (!write) throw new Error("open() must come first");
    let first = true;
    vi.mocked(writeSiteSettings).mockImplementation(async (...args) => {
      if (first) {
        first = false;
        await new Promise(resolve => setTimeout(resolve, 30));
      }
      return write(...args);
    });
  };

  it("run one at a time in the order asked: two quick route choices end with storage and the screen on the second", async () => {
    const storage = open({ site: { enabled: true } });
    slowFirstWrite();
    await enabledSwitch();
    fireEvent.click(route("Global session"));
    fireEvent.click(route("Automatic"));
    await waitFor(() => expect(activations()).toHaveLength(2));
    expect(storage.site()).toEqual({ enabled: true, target: "auto" });
    expect(route("Automatic").checked).toBe(true);
    expect(route("Global session").checked).toBe(false);
    expect(vi.mocked(writeSiteSettings).mock.calls.map(([, site]) => site)).toEqual([{ enabled: true, target: "global" }, { enabled: true, target: "auto" }]);
  });

  it("do not overwrite a default another popup changed between two edits", async () => {
    vi.mocked(fetchRuntimeCatalog).mockResolvedValue(MODELS);
    const storage = open({ global: { server: gavel, project: "acme" }, site: { enabled: true } });
    fireEvent.change(await projectSelect(), { target: { value: "docs" } });
    await waitFor(() => expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { enabled: true, project: "docs" }, preserved));
    storage.setGlobal({ server: gavel, project: "acme", session: "global-9" });
    await openSettings();
    pickModel("Codex example");
    await waitFor(() => expect(writeGlobalDefaults).toHaveBeenCalledWith({ server: gavel, project: "acme", session: "global-9", runtime: codexPicked }));
    expect(storage.global()).toEqual({ server: gavel, project: "acme", session: "global-9", runtime: codexPicked });
  });

  it("show the stored defaults when the site entry cannot be written after the defaults were", async () => {
    const storage = open({ site: { enabled: true, runtime: codex } });
    await openSettings();
    fireEvent.click(await screen.findByRole("button", { name: "This site" }));
    vi.mocked(writeSiteSettings).mockRejectedValue(new Error("This page's storage is not available to the grabber: blocked"));
    fireEvent.click(screen.getByRole("menuitem", { name: "Make this the default for all sites" }));
    expect((await screen.findByRole("alert")).textContent).toContain("storage is not available");
    expect(storage.global()).toEqual({ ...globalDefaults, runtime: codex });
    expect(storage.site()).toEqual({ enabled: true, runtime: codex });
    fireEvent.click(await screen.findByRole("button", { name: "This site" }));
    expect(screen.getByRole("menuitem", { name: "Use the default here codex-example" })).toBeTruthy();
  });
});

describe("starting and choosing sessions", () => {
  it("starts the page's session with the default model, writes it to the site entry and re-activates the open panel", async () => {
    open({ site: { enabled: true } });
    await enabledSwitch();
    fireEvent.change(nameField(), { target: { value: "checkout-review" } });
    fireEvent.click(screen.getByRole("button", { name: "Start session" }));
    await waitFor(() => expect(startSession).toHaveBeenCalledWith({ page: PAGE, prefix: "checkout-review", runtime: claude }));
    await waitFor(() => expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { enabled: true, session: "s-new" }, { preserveSession: false }));
    expect(writeGlobalDefaults).not.toHaveBeenCalled();
    expect(clientOptions).toEqual([{ server: gavel, project: "acme" }, { server: gavel, project: "acme" }]);
    await waitFor(() => expect(activations()).toHaveLength(1));
    expect(await screen.findByText("gavel-example.com-a1b2c3")).toBeTruthy();
    expect(screen.queryByLabelText(/New session name/)).toBeNull();
    expect(screen.getByRole("button", { name: "New session" })).toBeTruthy();
  });

  it("starts the session on the site's own default model", async () => {
    open({ site: { enabled: true, runtime: codex } });
    await enabledSwitch();
    fireEvent.change(nameField(), { target: { value: "checkout-review" } });
    fireEvent.click(screen.getByRole("button", { name: "Start session" }));
    await waitFor(() => expect(startSession).toHaveBeenCalledWith({ page: PAGE, prefix: "checkout-review", runtime: codex }));
  });

  it("shows gavel's failure inline and writes nothing when the session cannot be started", async () => {
    startSession.mockRejectedValue(new Error("prefix is required (POST /api/v1/comments/agent-sessions, HTTP 400)"));
    open({ site: { enabled: true } });
    await enabledSwitch();
    fireEvent.click(screen.getByRole("button", { name: "Start session" }));
    expect((await screen.findByRole("alert")).textContent).toContain("HTTP 400");
    expect(writeSiteSettings).not.toHaveBeenCalled();
    expect((screen.getByRole("button", { name: "Start session" }) as HTMLButtonElement).disabled).toBe(false);
  });

  it("uses a running session as the page's session without starting one", async () => {
    open({ site: { enabled: true } });
    await enabledSwitch();
    fireEvent.click(await screen.findByRole("button", { name: "use a running session" }));
    fireEvent.click(await screen.findByRole("menuitem", { name: /Review changes/ }));
    await waitFor(() => expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { enabled: true, session: "s-1" }, { preserveSession: false }));
    expect(startSession).not.toHaveBeenCalled();
    await waitFor(() => expect(activations()).toHaveLength(1));
  });

  it("starts a new session from New session, replacing the page's remembered one", async () => {
    open({ site: { session: "s-1", enabled: true } });
    await enabledSwitch();
    expect(await screen.findByText("Review changes")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "New session" }));
    fireEvent.change(nameField(), { target: { value: "second-run" } });
    fireEvent.click(screen.getByRole("button", { name: "Start session" }));
    await waitFor(() => expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { enabled: true, session: "s-new" }, { preserveSession: false }));
  });

  it("switches the page's session to another running one", async () => {
    agentSessions.mockResolvedValue([SESSION, { id: "s-2", source: "claude", title: "Fix layout", inCmux: true }]);
    open({ site: { session: "s-1", enabled: true } });
    await enabledSwitch();
    fireEvent.click(await screen.findByRole("button", { name: "Switch" }));
    fireEvent.click(await screen.findByRole("menuitem", { name: /Fix layout/ }));
    await waitFor(() => expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { enabled: true, session: "s-2" }, { preserveSession: false }));
  });

  it("starts the global session into the global defaults, leaving the site's own session alone", async () => {
    open({ site: { session: "s-1", target: "global", enabled: true } });
    await enabledSwitch();
    expect(nameField().value).toBe("gavel-global");
    fireEvent.change(nameField(), { target: { value: "shared-agent" } });
    fireEvent.click(screen.getByRole("button", { name: "Start session" }));
    await waitFor(() => expect(startSession).toHaveBeenCalledWith({ page: PAGE, prefix: "shared-agent", runtime: claude }));
    await waitFor(() => expect(writeGlobalDefaults).toHaveBeenCalledWith({ ...globalDefaults, session: "s-new" }));
    expect(writeSiteSettings).not.toHaveBeenCalled();
    await waitFor(() => expect(activations()).toHaveLength(1));
    expect(await screen.findByText("gavel-example.com-a1b2c3")).toBeTruthy();
  });

  it("explains that a session needs a project, and does not offer to start one, before a project is chosen", async () => {
    open({ global: { server: gavel }, site: { enabled: true } });
    await screen.findByText("Set up Gavel");
    expect(screen.queryByRole("button", { name: "Start session" })).toBeNull();
  });
});

describe("settings view", () => {
  it("opens from Settings with a Back button, the title and the host, and returns to the main view", async () => {
    open({ site: { enabled: true } });
    await openSettings();
    expect(screen.getByRole("heading", { name: "Settings" })).toBeTruthy();
    expect(screen.getByText("example.com")).toBeTruthy();
    expect(screen.queryByRole("switch")).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Back" }));
    expect((await enabledSwitch()).getAttribute("aria-checked")).toBe("true");
    expect(screen.queryByRole("heading", { name: "Settings" })).toBeNull();
  });

  it("shows the dashboard with its project count, and Same on every site while the site has no dashboard of its own", async () => {
    open({ site: { enabled: true } });
    await openSettings();
    expect(dashboardField().value).toBe(gavel);
    expect(await screen.findByText("Connected · 2 projects")).toBeTruthy();
    expect(screen.getByText("Same on every site")).toBeTruthy();
    expect(screen.queryByRole("button", { name: "This site" })).toBeNull();
  });

  it("commits a new dashboard on Enter as this site's own, and offers the This site menu for it", async () => {
    open({ site: { enabled: true } });
    await openSettings();
    fireEvent.change(dashboardField(), { target: { value: `${otherGavel}/path` } });
    expect(writeSiteSettings).not.toHaveBeenCalled();
    fireEvent.keyDown(dashboardField(), { key: "Enter" });
    await waitFor(() => expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { enabled: true, server: otherGavel }, preserved));
    expect(writeGlobalDefaults).not.toHaveBeenCalled();
    expect(await screen.findByRole("button", { name: "This site" })).toBeTruthy();
    expect(screen.queryByText("Same on every site")).toBeNull();
    expect(dashboardField().value).toBe(otherGavel);
  });

  it("drops the project of the old dashboard when the site's own dashboard changes, and so does not activate the grabber", async () => {
    open({ site: { enabled: true, server: otherGavel, project: "docs" } });
    await openSettings();
    fireEvent.change(dashboardField(), { target: { value: "https://third.example.com" } });
    fireEvent.keyDown(dashboardField(), { key: "Enter" });
    await waitFor(() => expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { enabled: true, server: "https://third.example.com" }, preserved));
    fireEvent.click(await screen.findByRole("button", { name: "This site" }));
    expect(await screen.findByText("third.example.com applies only to example.com.")).toBeTruthy();
    expect(activations()).toEqual([]);
  });

  it("commits a new dashboard when the field loses focus", async () => {
    open({ site: { enabled: true } });
    await openSettings();
    fireEvent.change(dashboardField(), { target: { value: otherGavel } });
    fireEvent.blur(dashboardField());
    await waitFor(() => expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { enabled: true, server: otherGavel }, preserved));
  });

  it("writes nothing when the dashboard is left as it was", async () => {
    open({ site: { enabled: true } });
    await openSettings();
    fireEvent.blur(dashboardField());
    fireEvent.change(dashboardField(), { target: { value: `${gavel}/` } });
    expect(dashboardField().value).toBe(`${gavel}/`);
    fireEvent.blur(dashboardField());
    await waitFor(() => expect(dashboardField().value).toBe(gavel));
    expect(readSiteSettings).toHaveBeenCalledTimes(1);
    expect(writeSiteSettings).not.toHaveBeenCalled();
    expect(writeGlobalDefaults).not.toHaveBeenCalled();
  });

  it("shows an invalid address inline and writes nothing", async () => {
    open({ site: { enabled: true } });
    await openSettings();
    fireEvent.change(dashboardField(), { target: { value: "localhost:9092" } });
    fireEvent.keyDown(dashboardField(), { key: "Enter" });
    expect((await screen.findByText(/must be an HTTP\(S\) URL/)).textContent).toContain("localhost:9092");
    expect(writeSiteSettings).not.toHaveBeenCalled();
    expect(writeGlobalDefaults).not.toHaveBeenCalled();
    fireEvent.change(dashboardField(), { target: { value: gavel } });
    fireEvent.blur(dashboardField());
    await waitFor(() => expect(screen.queryByText(/must be an HTTP\(S\) URL/)).toBeNull());
  });

  it("writes what resetField and promoteField produce for a site's own dashboard", async () => {
    open({ site: { enabled: true, server: otherGavel, project: "docs" } });
    await openSettings();
    fireEvent.click(await screen.findByRole("button", { name: "This site" }));
    expect(screen.getByText("other.example.com with project docs applies only to example.com.")).toBeTruthy();
    fireEvent.click(screen.getByRole("menuitem", { name: "Make this the default for all sites" }));
    await waitFor(() => expect(writeGlobalDefaults).toHaveBeenCalledWith({ server: otherGavel, project: "docs", runtime: claude }));
    expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { enabled: true }, preserved);
    await waitFor(() => expect(screen.getByText("Same on every site")).toBeTruthy());
  });

  it("names only the dashboard in its menu while the site has no project there yet", async () => {
    open({ site: { enabled: true, server: otherGavel } });
    await openSettings();
    fireEvent.click(await screen.findByRole("button", { name: "This site" }));
    expect(screen.getByText("other.example.com applies only to example.com.")).toBeTruthy();
  });

  it("shows the dashboard's failure in the status line", async () => {
    vi.mocked(listProjects).mockRejectedValue(new Error("connection refused"));
    open({ site: { enabled: true } });
    await openSettings();
    expect(await screen.findByText("connection refused")).toBeTruthy();
    expect(screen.queryByText(/Connected ·/)).toBeNull();
  });

  it("says Gavel's default while no model is set, with no This site menu", async () => {
    open({ global: { server: gavel, project: "acme" }, site: { enabled: true } });
    await openSettings();
    expect(screen.getByText("New sessions")).toBeTruthy();
    expect(screen.getByText("The model used when you start a session, or when routing is automatic.")).toBeTruthy();
    expect(screen.queryByText("Model")).toBeNull();
    expect(screen.getByText("Gavel's default")).toBeTruthy();
    expect(await screen.findByLabelText("Agent runtime")).toBeTruthy();
    expect(screen.queryByRole("button", { name: "This site" })).toBeNull();
  });

  it("stores the first model picked as the default for every site", async () => {
    vi.mocked(fetchRuntimeCatalog).mockResolvedValue(MODELS);
    open({ global: { server: gavel, project: "acme" }, site: { enabled: true } });
    await openSettings();
    pickModel("Codex example");
    await waitFor(() => expect(writeGlobalDefaults).toHaveBeenCalledWith({ server: gavel, project: "acme", runtime: codexPicked }));
    expect(writeSiteSettings).not.toHaveBeenCalled();
    await waitFor(() => expect(screen.queryByText("Gavel's default")).toBeNull());
    expect(screen.queryByRole("button", { name: "This site" })).toBeNull();
  });

  it("stores a model that differs from the default as this site's own, and offers the This site menu for it", async () => {
    vi.mocked(fetchRuntimeCatalog).mockResolvedValue(MODELS);
    open({ site: { enabled: true } });
    await openSettings();
    pickModel("Codex example");
    await waitFor(() => expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { enabled: true, runtime: codexPicked }, preserved));
    expect(writeGlobalDefaults).not.toHaveBeenCalled();
    await waitFor(() => expect(activations()).toHaveLength(1));
    expect(await screen.findByRole("button", { name: "This site" })).toBeTruthy();
  });

  it("makes a site's own model the default for every site", async () => {
    open({ site: { enabled: true, runtime: codex } });
    await openSettings();
    fireEvent.click(await screen.findByRole("button", { name: "This site" }));
    fireEvent.click(screen.getByRole("menuitem", { name: "Make this the default for all sites" }));
    await waitFor(() => expect(writeGlobalDefaults).toHaveBeenCalledWith({ ...globalDefaults, runtime: codex }));
    expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { enabled: true }, preserved);
    await waitFor(() => expect(screen.queryByRole("button", { name: "This site" })).toBeNull());
  });

  it("offers the This site menu for a site's own model, and resets it to the default", async () => {
    open({ site: { enabled: true, runtime: codex } });
    await openSettings();
    fireEvent.click(await screen.findByRole("button", { name: "This site" }));
    expect(screen.getByText("codex-example applies only to example.com.")).toBeTruthy();
    fireEvent.click(screen.getByRole("menuitem", { name: "Use the default here claude-example" }));
    await waitFor(() => expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { enabled: true }, preserved));
    await waitFor(() => expect(screen.queryByRole("button", { name: "This site" })).toBeNull());
  });

  it("starts sessions on the new default model after it changes in Settings", async () => {
    vi.mocked(fetchRuntimeCatalog).mockResolvedValue(MODELS);
    open({ site: { enabled: true } });
    await openSettings();
    pickModel("Codex example");
    await waitFor(() => expect(writeSiteSettings).toHaveBeenCalled());
    fireEvent.click(await screen.findByRole("button", { name: "Back" }));
    fireEvent.change(await screen.findByLabelText<HTMLInputElement>(/New session name/), { target: { value: "checkout-review" } });
    fireEvent.click(screen.getByRole("button", { name: "Start session" }));
    await waitFor(() => expect(startSession).toHaveBeenCalledWith({ page: PAGE, prefix: "checkout-review", runtime: codexPicked }));
  });
});

describe("Chrome withholding access to the dashboard", () => {
  beforeEach(() => {
    vi.mocked(hasServerAccess).mockResolvedValueOnce(false);
  });

  it("replaces the routes with a notice, and the footer reads Connection error", async () => {
    vi.mocked(useSyncStatus).mockImplementation(({ connectionError }) => (connectionError ? { kind: "connection-error", detail: connectionError } : connected));
    open({ site: { enabled: true } });
    const notice = await screen.findByRole("alert");
    expect(notice.textContent).toContain("Chrome is blocking Gavel from reaching localhost:9092");
    expect(notice.textContent).toContain("Projects and sessions can't load until you allow it.");
    expect(within(notice).getByRole("button", { name: "Allow access" })).toBeTruthy();
    expect(screen.queryByRole("group", { name: ROUTES_NAME })).toBeNull();
    expect(screen.queryAllByRole("radio")).toEqual([]);
    expect(screen.queryByText("Set up Gavel")).toBeNull();
    await waitFor(() => expect(document.querySelector("footer")?.textContent).toBe("Connection error"));
    expect(useSyncStatus).toHaveBeenLastCalledWith({ connectionError: "Chrome is blocking Gavel from reaching localhost:9092", connecting: false });
  });

  it("requests access synchronously in the click, then loads the routes once it is granted", async () => {
    open({ site: { enabled: true } });
    fireEvent.click(await screen.findByRole("button", { name: "Allow access" }));
    expect(requestServerAccess).toHaveBeenCalledWith(gavel);
    expect(await screen.findByRole("group", { name: ROUTES_NAME })).toBeTruthy();
    expect(screen.queryByText(/Chrome is blocking Gavel/)).toBeNull();
  });

  it("opens the settings view from Use another dashboard", async () => {
    open({ site: { enabled: true } });
    fireEvent.click(await screen.findByRole("button", { name: "Use another dashboard" }));
    expect(screen.getByRole("heading", { name: "Settings" })).toBeTruthy();
    expect(await screen.findByRole("button", { name: "Allow access" })).toBeTruthy();
    expect(screen.queryByRole("button", { name: "Use another dashboard" })).toBeNull();
  });
});

describe("popup storage errors", () => {
  it("offers to discard global defaults that still hold the removed agent field, then reads the seeded defaults again", async () => {
    vi.mocked(readGlobalDefaults).mockRejectedValueOnce(new Error('Stored global defaults still has the removed "agent" field: re-pick the model'));
    vi.mocked(readGlobalDefaults).mockResolvedValue(globalDefaults);
    vi.mocked(readSiteSettings).mockResolvedValue({});
    vi.mocked(discardGlobalDefaults).mockResolvedValue();
    render(<Popup />);
    expect((await screen.findByRole("alert")).textContent).toContain('removed "agent" field');
    fireEvent.click(screen.getByRole("button", { name: "Discard stored global defaults" }));
    await enabledSwitch();
    expect(discardGlobalDefaults).toHaveBeenCalledOnce();
  });

  it("shows an unreadable site entry with only a way to discard it, then reads the page again", async () => {
    vi.mocked(readGlobalDefaults).mockResolvedValue(globalDefaults);
    vi.mocked(readSiteSettings).mockRejectedValueOnce(new Error('Grabber settings in flanksource.grabber still has the removed "agent" field'));
    vi.mocked(readSiteSettings).mockResolvedValue({});
    render(<Popup />);
    expect((await screen.findByRole("alert")).textContent).toContain('removed "agent" field');
    expect(screen.queryByRole("switch")).toBeNull();
    expect(screen.queryByRole("button", { name: "Settings" })).toBeNull();
    expect(screen.getAllByRole("button").map(button => button.textContent)).toEqual(["Discard this site's settings"]);
    fireEvent.click(screen.getByRole("button", { name: "Discard this site's settings" }));
    await waitFor(() => expect(writeSiteSettings).toHaveBeenCalledWith(7, null));
    await enabledSwitch();
  });

  it("shows a failed discard of the site entry and keeps offering it", async () => {
    vi.mocked(readGlobalDefaults).mockResolvedValue(globalDefaults);
    vi.mocked(readSiteSettings).mockRejectedValue(new Error("Grabber settings are corrupt"));
    vi.mocked(writeSiteSettings).mockRejectedValue(new Error("This page's storage is not available to the grabber: blocked"));
    render(<Popup />);
    fireEvent.click(await screen.findByRole("button", { name: "Discard this site's settings" }));
    await waitFor(() => expect(screen.getAllByRole("alert").map(alert => alert.textContent).join(" ")).toContain("storage is not available"));
    expect(screen.getByRole("button", { name: "Discard this site's settings" })).toBeTruthy();
  });
});
