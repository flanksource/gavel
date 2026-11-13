import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { SiteSettings } from "../host/site-settings";
import { Popup } from "./PopupView";
import { fetchRuntimeCatalog, listProjects } from "./gavel-client";
import { requestBackground } from "./messages";
import type { GlobalDefaults } from "./settings";
import { discardGlobalDefaults, readGlobalDefaults, readSiteSettings, writeGlobalDefaults, writeSiteSettings } from "./storage";

const agentSessions = vi.hoisted(() => vi.fn());
vi.mock("./storage");
vi.mock("./messages");
vi.mock("./server-access", () => ({ hasServerAccess: async () => true, requestServerAccess: async () => true }));
vi.mock("./gavel-client", () => ({
  listProjects: vi.fn(),
  fetchRuntimeCatalog: vi.fn(),
  GavelClient: class { agentSessions = agentSessions; },
}));

const gavel = "http://localhost:9092";
const claude = { spec: { model: "claude-example" } };
const codex = { spec: { model: "codex-example" } };
const globalDefaults: GlobalDefaults = { server: gavel, project: "acme", runtime: claude };
const CATALOG = {
  models: [],
  families: [{ id: "claude", label: "Claude", provider: "anthropic", modes: [{ id: "agent", label: "Agent", defaultModel: "claude-example", provider: "anthropic" }] }],
};
const SESSION = { id: "s-1", source: "codex", title: "Review changes", model: "codex-example", inCmux: true };

function open(options: { global?: GlobalDefaults; site?: SiteSettings } = {}) {
  vi.mocked(readGlobalDefaults).mockResolvedValue(options.global ?? globalDefaults);
  vi.mocked(readSiteSettings).mockResolvedValue(options.site ?? {});
  render(<Popup />);
}

beforeEach(() => {
  vi.stubGlobal("chrome", { tabs: { query: async () => [{ id: 7, url: "https://example.com/docs" }] } });
  vi.mocked(listProjects).mockResolvedValue(["acme", "docs"]);
  vi.mocked(fetchRuntimeCatalog).mockResolvedValue(CATALOG);
  agentSessions.mockResolvedValue([SESSION]);
  vi.mocked(writeSiteSettings).mockResolvedValue();
  vi.mocked(writeGlobalDefaults).mockImplementation(async defaults => defaults);
  vi.mocked(requestBackground).mockResolvedValue(undefined);
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  vi.clearAllMocks();
});

const radio = (name: string) => screen.getByRole("radio", { name });

describe("popup views", () => {
  it("edits this site first, with the enabled switch above the segmented control and the site-only controls", async () => {
    open({ site: { project: "docs", runtime: codex } });
    await screen.findByRole("button", { name: "Save for this site" });
    expect(radio("This site").getAttribute("aria-checked")).toBe("true");
    const enabled = screen.getByRole("switch", { name: "Enabled on this site" });
    expect(enabled.compareDocumentPosition(screen.getByRole("radiogroup", { name: "Settings to edit" })) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(screen.getByRole("button", { name: "Reset site to defaults" })).toBeTruthy();
    expect(screen.getByText("example.com")).toBeTruthy();
    expect(screen.getAllByText("overrides default")).toHaveLength(1);
    expect(screen.getByText("model overrides default")).toBeTruthy();
    expect(await screen.findByRole("button", { name: "Use session Review changes" })).toBeTruthy();
    expect(screen.queryByRole("button", { name: "Save global defaults" })).toBeNull();
  });

  it("switches to the global defaults, keeping the enabled switch and hiding the site-only session and reset", async () => {
    open();
    await screen.findByRole("button", { name: "Save for this site" });
    fireEvent.click(radio("Global"));
    expect(radio("Global").getAttribute("aria-checked")).toBe("true");
    expect(screen.getByRole("button", { name: "Save global defaults" })).toBeTruthy();
    expect(screen.getByRole("switch", { name: "Enabled on this site" })).toBeTruthy();
    expect(screen.queryByRole("button", { name: "Save for this site" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Reset site to defaults" })).toBeNull();
    expect(screen.queryByText("Active sessions")).toBeNull();
    expect(screen.queryByRole("button", { name: "New session" })).toBeNull();
    expect(screen.getByRole("radiogroup", { name: "Session grouping" })).toBeTruthy();
    expect(screen.getByRole("button", { name: "Choose model" }).textContent).not.toBe("Gavel's default");
    expect(screen.getByRole("button", { name: "Use gavel's default" })).toBeTruthy();
  });

  it("keeps the site's unsaved edits while the global view is shown, and saves each view separately", async () => {
    open();
    await screen.findByRole("button", { name: "Save for this site" });
    fireEvent.click(screen.getByRole("radio", { name: "Shared" }));
    fireEvent.click(radio("Global"));
    expect(screen.getByRole("radio", { name: "Per page" }).getAttribute("aria-checked")).toBe("true");
    fireEvent.click(radio("This site"));
    expect(screen.getByRole("radio", { name: "Shared" }).getAttribute("aria-checked")).toBe("true");
    expect(screen.getByRole("button", { name: "Save for this site" }).hasAttribute("disabled")).toBe(false);
    fireEvent.click(radio("Global"));
    fireEvent.click(screen.getByRole("radio", { name: "Shared" }));
    fireEvent.click(screen.getByRole("button", { name: "Save global defaults" }));
    await waitFor(() => expect(writeGlobalDefaults).toHaveBeenCalledWith({ server: gavel, project: "acme", runtime: claude, grouping: "shared" }));
    expect(writeSiteSettings).not.toHaveBeenCalled();
  });
});

describe("popup site settings", () => {
  it("stores only what differs from the global defaults: the chosen session and a different grouping", async () => {
    open();
    fireEvent.click(await screen.findByRole("button", { name: "Use session Review changes" }));
    fireEvent.click(screen.getByRole("radio", { name: "Shared" }));
    fireEvent.click(screen.getByRole("button", { name: "Save for this site" }));
    await waitFor(() => expect(writeSiteSettings).toHaveBeenCalledWith(7, { grouping: "shared", session: "s-1" }));
    await screen.findByText("Saved for this site.");
  });

  it("returns to a new session, which shows the runtime bar, and lets a site follow the global runtime again", async () => {
    open({ site: { runtime: codex, session: "s-1" } });
    const session = await screen.findByRole("button", { name: "Use session Review changes" });
    expect(session.getAttribute("aria-pressed")).toBe("true");
    expect(screen.queryByLabelText("Agent runtime")).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "New session" }));
    await screen.findByLabelText("Agent runtime");
    fireEvent.click(screen.getByRole("button", { name: "Use global default" }));
    fireEvent.click(screen.getByRole("button", { name: "Save for this site" }));
    await waitFor(() => expect(writeSiteSettings).toHaveBeenCalledWith(7, null));
  });

  it("resets the site to the defaults but keeps it enabled", async () => {
    open({ site: { project: "docs", enabled: true } });
    fireEvent.click(await screen.findByRole("button", { name: "Reset site to defaults" }));
    await waitFor(() => expect(writeSiteSettings).toHaveBeenCalledWith(7, { enabled: true }));
  });

  it("holds the enabled switch while edits are unsaved, and applies it at once otherwise", async () => {
    open();
    const enabled = await screen.findByRole("switch", { name: "Enabled on this site" });
    fireEvent.click(await screen.findByRole("button", { name: "Use session Review changes" }));
    await screen.findByText(/Unsaved changes/);
    expect(enabled.hasAttribute("disabled")).toBe(true);
    fireEvent.click(screen.getByRole("button", { name: "Save for this site" }));
    await waitFor(() => expect(enabled.hasAttribute("disabled")).toBe(false));
    fireEvent.click(enabled);
    await waitFor(() => expect(requestBackground).toHaveBeenCalledWith({ type: "activate", tabId: 7 }));
    expect(writeSiteSettings).toHaveBeenLastCalledWith(7, { session: "s-1", enabled: true });
  });
});

describe("popup global defaults", () => {
  it("saves the runtime and grouping without the site-only session or enabled flag, and can unset the runtime", async () => {
    open({ site: { session: "s-1", enabled: true } });
    await screen.findByRole("button", { name: "Save for this site" });
    fireEvent.click(radio("Global"));
    fireEvent.click(screen.getByRole("button", { name: "Use gavel's default" }));
    expect(screen.getByRole("button", { name: "Choose model" }).textContent).toBe("Gavel's default");
    fireEvent.click(screen.getByRole("button", { name: "Save global defaults" }));
    await waitFor(() => expect(writeGlobalDefaults).toHaveBeenCalledWith({ server: gavel, project: "acme", grouping: "page" }));
  });

  it("offers the models from the dashboard's catalog for the runtime", async () => {
    open();
    await screen.findByRole("button", { name: "Save for this site" });
    fireEvent.click(radio("Global"));
    fireEvent.click(screen.getByRole("button", { name: "Choose model" }));
    await within(await screen.findByLabelText("Agent runtime")).findByText(/Claude|claude/);
    expect(fetchRuntimeCatalog).toHaveBeenCalledWith({ server: gavel });
  });

  it("offers capture and Gavel settings before a comment project is configured, without a backend selector", async () => {
    open({ global: { server: gavel } });
    await screen.findByRole("button", { name: "Save for this site" });
    expect(screen.getByRole("button", { name: "Capture todo / saved captures" })).toBeTruthy();
    expect(screen.queryByLabelText("Backend")).toBeNull();
    expect(fetchRuntimeCatalog).toHaveBeenCalledWith({ server: gavel });
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
    await screen.findByRole("button", { name: "Save for this site" });
    expect(discardGlobalDefaults).toHaveBeenCalledOnce();
  });

  it("shows an unreadable site entry in the site view only, with a way to discard it, while the global view still works", async () => {
    vi.mocked(readGlobalDefaults).mockResolvedValue(globalDefaults);
    vi.mocked(readSiteSettings).mockRejectedValueOnce(new Error('Grabber settings in flanksource.grabber still has the removed "agent" field'));
    vi.mocked(readSiteSettings).mockResolvedValue({});
    render(<Popup />);
    expect((await screen.findByRole("alert")).textContent).toContain('removed "agent" field');
    expect(screen.queryByRole("switch", { name: "Enabled on this site" })).toBeNull();
    fireEvent.click(radio("Global"));
    expect(screen.getByRole("button", { name: "Save global defaults" })).toBeTruthy();
    fireEvent.click(radio("This site"));
    fireEvent.click(screen.getByRole("button", { name: "Discard this site's settings" }));
    await waitFor(() => expect(writeSiteSettings).toHaveBeenCalledWith(7, null));
    await screen.findByRole("button", { name: "Save for this site" });
  });
});
