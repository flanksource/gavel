import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { AgentPicker } from "../agents/AgentPicker";
import { GavelRail, type GavelRailProps } from "./GavelPanel";
import { GAVEL_COMMENT_CONFIG } from "./useGavelComments";

afterEach(() => { cleanup(); localStorage.clear(); });

function renderRail() {
  const onAutoRunChange = vi.fn();
  const onAssignAll = vi.fn();
  const props: GavelRailProps = {
    host: { send: vi.fn(), requestMode: vi.fn(), maximize: vi.fn() },
    state: { url: "https://example.test/review", title: "Review" },
    comments: {
      comments: [], loading: false, error: null,
      create: vi.fn(), reply: vi.fn(), updateStatus: vi.fn(), acknowledge: vi.fn(),
      assign: vi.fn(), close: vi.fn(), remove: vi.fn(), applyProposal: vi.fn(),
      rejectProposal: vi.fn(), requestRevision: vi.fn(), loadProposalDiff: vi.fn(),
      deliver: vi.fn(), deliverAll: vi.fn(),
    },
    config: GAVEL_COMMENT_CONFIG,
    create: vi.fn(), floating: false, presentation: "rail", anchors: [], alerts: null,
    renderRunDetails: () => null,
    session: { sessionId: undefined, session: undefined },
    sessionContent: <p>Session transcript</p>,
    settingsContent: <AgentPicker heading="Agent settings" autoRun={{ checked: false, onChange: onAutoRunChange }} />,
    extraAssignItems: () => [], assignAll: { label: "Assign all to review session", onSelect: onAssignAll },
  };
  return { ...render(<GavelRail {...props} />), onAutoRunChange, onAssignAll };
}

describe("GavelRail settings menu", () => {
  it("opens settings over the transcript without adding a settings tab", () => {
    const { onAutoRunChange } = renderRail();
    fireEvent.click(screen.getByRole("tab", { name: "New session" }));
    expect([screen.getByText("Session transcript") !== null, screen.queryByRole("switch")]).toEqual([true, null]);

    fireEvent.click(screen.getByRole("button", { name: "Settings" }));
    fireEvent.click(screen.getByRole("switch", { name: "Auto-run new comments on this page" }));
    expect([onAutoRunChange.mock.calls, screen.queryByRole("tab", { name: "Settings" }), screen.getByText("Session transcript") !== null]).toEqual([[[true]], null, true]);

    fireEvent.click(screen.getByRole("button", { name: "Settings" }));
    fireEvent.click(screen.getByRole("tab", { name: /^Comments/ }));
    expect([screen.queryByRole("switch"), screen.getByTestId("comment-compose-input") !== null]).toEqual([null, true]);
  });

  it("keeps bulk assignment in the settings menu and closes after assigning", () => {
    const { onAssignAll } = renderRail();
    fireEvent.click(screen.getByRole("button", { name: "Settings" }));
    fireEvent.click(screen.getByRole("button", { name: "Assign all to review session" }));
    expect([onAssignAll.mock.calls.length, screen.queryByRole("switch"), screen.getByRole("button", { name: "Settings" }).getAttribute("aria-expanded")]).toEqual([1, null, "false"]);
  });

  it("keeps the selected session tab when settings is opened and the sidebar is reopened", () => {
    const { unmount } = renderRail();
    fireEvent.click(screen.getByRole("tab", { name: "New session" }));
    fireEvent.click(screen.getByRole("button", { name: "Settings" }));
    unmount();
    renderRail();
    expect([screen.getByRole("tab", { name: "New session" }).getAttribute("aria-selected"), screen.queryByRole("switch")]).toEqual(["true", null]);
  });
});
