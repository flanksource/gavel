import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { GavelPanel } from "./GavelPanel";
import { GavelClient } from "./gavel-client";
import { requestBackground } from "./messages";
import type { PanelTarget } from "./settings";
import { saveHostSession, saveHostTarget } from "../host/useHostBridge";

const PAGE = { url: "https://app.example.test:3000/docs", title: "Docs" };
const SESSIONS = vi.hoisted(() => [{ id: "s-1", source: "codex", title: "Review changes", model: "codex-example", inCmux: true }]);
const STARTED = { id: "s-new", source: "claude", title: "gavel-app.example.test:3000-a1b2c3", inCmux: false };
const CATALOG = {
  models: [],
  families: [{ id: "claude", label: "Claude", provider: "anthropic", modes: [{ id: "agent", label: "Agent", defaultModel: "claude-example", provider: "anthropic" }] }],
};

vi.mock("./messages");
vi.mock("../host/useHostBridge", () => ({
  useHostBridge: () => ({ state: { url: "https://app.example.test:3000/docs", title: "Docs" }, anchors: [], focus: undefined, error: "", floating: false, presentation: "rail", draftBox: undefined }),
  requestPanelMode: vi.fn(),
  sendToHost: vi.fn(),
  saveHostSession: vi.fn(),
  saveHostTarget: vi.fn(),
}));
const COMMENTS = vi.hoisted(() => ({ list: [] as unknown[], deliverAll: vi.fn() }));
vi.mock("./useGavelComments", async importOriginal => ({
  ...(await importOriginal<typeof import("./useGavelComments")>()),
  useGavelComments: () => ({
    comments: COMMENTS.list, loading: false, error: null,
    refresh: vi.fn(), reply: vi.fn(), edit: vi.fn(), editReply: vi.fn(), updateStatus: vi.fn(), acknowledge: vi.fn(),
    assign: vi.fn(), close: vi.fn(), remove: vi.fn(), applyProposal: vi.fn(), rejectProposal: vi.fn(), requestRevision: vi.fn(), loadProposalDiff: vi.fn(),
    deliver: vi.fn(), deliverAll: COMMENTS.deliverAll,
  }),
}));
// `adopt` is what the panel hands the hook: the hook calls it with the session a started run created.
const ADOPT = vi.hoisted(() => ({ watch: vi.fn(), watchNew: vi.fn(), cancel: vi.fn(), adopt: undefined as ((sessionId: string) => void) | undefined }));
vi.mock("./useGavelAgents", () => ({
  useAgentSessions: () => ({ sessions: SESSIONS, error: null, refresh: vi.fn().mockResolvedValue(undefined) }),
  useRuntimeCatalog: () => ({ catalog: CATALOG, error: null }),
  useCommentAgent: () => ({ agent: undefined, edits: "propose", ready: true, error: null, setEdits: vi.fn() }),
  useAdoptRunSession: (options: { adopt: (sessionId: string) => void }) => {
    ADOPT.adopt = options.adopt;
    return { watch: ADOPT.watch, watchNew: ADOPT.watchNew, cancel: ADOPT.cancel };
  },
}));
const PREVIEW = vi.hoisted(() => vi.fn((_options: { sessionId: string | undefined }) => ({})));
vi.mock("./DeliveryPreview", async importOriginal => ({
  ...(await importOriginal<typeof import("./DeliveryPreview")>()),
  useDeliveryPreview: PREVIEW,
  useDeliveryMentionables: () => undefined,
}));
vi.mock("./CommentQueue", () => ({
  useCommentQueue: () => ({ rows: [], retry: vi.fn(), capture: vi.fn(), error: null }),
  CommentQueue: () => null,
}));
vi.mock("./gavel-streams", () => ({ gavelEventHub: () => ({}) }));
vi.mock("./SessionTab", () => ({ SessionTab: () => null, SessionTabLabel: () => "Session" }));

const target = (overrides: Partial<PanelTarget> = {}): PanelTarget => ({
  server: "http://localhost:9092", project: "acme", grouping: "page", sessionTarget: "page", parentOrigin: "https://app.example.test:3000", bridgeSession: "bridge-1", ...overrides,
});
const openSettings = () => fireEvent.click(screen.getByRole("button", { name: "Settings" }));
const nameField = () => screen.getByLabelText<HTMLInputElement>(/New session name/);
const route = (name: string) => screen.getByRole<HTMLInputElement>("radio", { name });

beforeEach(() => {
  localStorage.clear();
  COMMENTS.list = [];
  COMMENTS.deliverAll.mockResolvedValue(undefined);
  ADOPT.adopt = undefined;
  PREVIEW.mockImplementation(() => ({}));
  vi.spyOn(GavelClient.prototype, "startSession").mockResolvedValue(STARTED);
  vi.mocked(saveHostSession).mockResolvedValue();
  vi.mocked(saveHostTarget).mockResolvedValue();
  vi.mocked(requestBackground).mockResolvedValue(undefined);
});
afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  vi.clearAllMocks();
});

describe("GavelPanel settings card", () => {
  it("starts the page's session with the panel's model and edits policy, and saves it to the page's grabber settings", async () => {
    render(<GavelPanel target={target()} />);
    openSettings();
    expect(nameField().value).toBe("gavel-app.example.test:3000");
    fireEvent.change(nameField(), { target: { value: "checkout-review" } });
    fireEvent.click(screen.getByRole("button", { name: "Start session" }));
    await waitFor(() => expect(saveHostSession).toHaveBeenCalledWith("s-new"));
    expect(GavelClient.prototype.startSession).toHaveBeenCalledWith({ page: PAGE, prefix: "checkout-review", edits: "propose" });
    expect(requestBackground).not.toHaveBeenCalled();
    expect(await screen.findByText("Session s-new (not live)")).toBeTruthy();
  });

  it("stores the choice of target in the page's grabber settings and shows the global session form", async () => {
    render(<GavelPanel target={target()} />);
    openSettings();
    fireEvent.click(route("Global session"));
    await waitFor(() => expect(saveHostTarget).toHaveBeenCalledWith("global"));
    await waitFor(() => expect(route("Global session").checked).toBe(true));
    expect(nameField().value).toBe("gavel-global");
  });

  it("stores Automatic as the target and keeps the remembered page session", async () => {
    render(<GavelPanel target={target({ pageSession: "s-1" })} />);
    openSettings();
    fireEvent.click(route("Automatic"));
    await waitFor(() => expect(saveHostTarget).toHaveBeenCalledWith("auto"));
    await waitFor(() => expect(route("Automatic").checked).toBe(true));
    expect(saveHostSession).not.toHaveBeenCalled();
    expect(screen.getByText("Review changes")).toBeTruthy();
  });

  it.each([
    { sessionTarget: "page" as const, expected: "s-page" },
    { sessionTarget: "global" as const, expected: "s-global" },
    { sessionTarget: "auto" as const, expected: undefined },
  ])("previews delivery to $expected when the target is $sessionTarget", ({ sessionTarget, expected }) => {
    render(<GavelPanel target={target({ sessionTarget, pageSession: "s-page", globalSession: "s-global" })} />);
    expect(PREVIEW.mock.lastCall?.[0].sessionId).toBe(expected);
  });

  it("starts the global session through the background worker, not the page's own storage", async () => {
    render(<GavelPanel target={target({ sessionTarget: "global", pageSession: "s-1" })} />);
    openSettings();
    fireEvent.change(nameField(), { target: { value: "shared-agent" } });
    fireEvent.click(screen.getByRole("button", { name: "Start session" }));
    await waitFor(() => expect(requestBackground).toHaveBeenCalledWith({ type: "global-session", session: "s-new" }));
    expect(saveHostSession).not.toHaveBeenCalled();
    expect(await screen.findByText("Session s-new (not live)")).toBeTruthy();
  });

  it("uses an active session as the page's session and shows a failed save inline", async () => {
    vi.mocked(saveHostSession).mockRejectedValueOnce(new Error("The page did not confirm the session change"));
    render(<GavelPanel target={target()} />);
    openSettings();
    // Inside the settings menu, a dropdown opens as a submenu, so its trigger is a menu item.
    fireEvent.click(screen.getByRole("menuitem", { name: "use a running session" }));
    fireEvent.click(await screen.findByRole("menuitem", { name: /Review changes/ }));
    expect((await screen.findByRole("alert")).textContent).toBe("The page did not confirm the session change");
    expect(GavelClient.prototype.startSession).not.toHaveBeenCalled();
  });

  it.each(["page", "global"] as const)("offers New session and Switch, and no Clear, for the %s session", sessionTarget => {
    render(<GavelPanel target={target({ sessionTarget, pageSession: "s-1", globalSession: "s-1" })} />);
    openSettings();
    fireEvent.click(screen.getByRole("button", { name: "New session" }));
    expect(nameField()).toBeTruthy();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeTruthy();
    expect(screen.queryByRole("button", { name: "Clear" })).toBeNull();
  });

  it("shows the remembered page and global sessions of the panel's URL", () => {
    render(<GavelPanel target={target({ sessionTarget: "global", pageSession: "s-1", globalSession: "s-1" })} />);
    openSettings();
    expect(screen.getAllByText("Review changes")).toHaveLength(2);
    expect(screen.queryByLabelText(/New session name/)).toBeNull();
  });
});

describe("GavelPanel adopting the session a started run created", () => {
  const RUN_SESSION = "s-run";
  const openThread = { id: "c-1", body: "Check the heading", createdAt: "2026-11-01T10:00:00Z", author: { name: "You", kind: "user" }, status: "open" };

  /** Assigns every open thread to a run Gavel would start, from the settings menu. */
  async function assignAllToANewRun(sessionTarget: PanelTarget["sessionTarget"]) {
    COMMENTS.list = [openThread];
    PREVIEW.mockImplementation(() => ({ preview: { route: { target: { kind: "run", prompt: "review", runOptions: { spec: { model: "claude-example" } } } } } }));
    render(<GavelPanel target={target({ sessionTarget, pageSession: "s-1", globalSession: "s-global" })} />);
    openSettings();
    fireEvent.click(await screen.findByRole("button", { name: /^Assign All to/ }));
    await waitFor(() => expect(COMMENTS.deliverAll).toHaveBeenCalledWith(["c-1"], expect.objectContaining({ kind: "run" })));
  }

  it("makes the run's session the page's session when the target is page", async () => {
    await assignAllToANewRun("page");
    expect(ADOPT.watch).toHaveBeenCalledWith(["c-1"]);
    ADOPT.adopt?.(RUN_SESSION);
    await waitFor(() => expect(saveHostSession).toHaveBeenCalledWith(RUN_SESSION));
  });

  it("makes the run's session the global session when the target is global", async () => {
    await assignAllToANewRun("global");
    ADOPT.adopt?.(RUN_SESSION);
    await waitFor(() => expect(requestBackground).toHaveBeenCalledWith({ type: "global-session", session: RUN_SESSION }));
    expect(saveHostSession).not.toHaveBeenCalled();
  });

  it("leaves the page's remembered session and the global one alone when the target is auto", async () => {
    await assignAllToANewRun("auto");
    expect(ADOPT.watch).not.toHaveBeenCalled();
    ADOPT.adopt?.(RUN_SESSION);
    await new Promise(resolve => setTimeout(resolve, 0));
    expect(saveHostSession).not.toHaveBeenCalled();
    expect(requestBackground).not.toHaveBeenCalled();
  });
});
