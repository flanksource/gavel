import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { AgentPickerProps } from "../agents/AgentPicker";
import type { AgentSession } from "./gavel-client";
import { SessionTab, SessionTabLabel } from "./SessionTab";

const SERVER = "http://gavel.example.test";
const live: AgentSession = { id: "s-1", source: "claude", title: "Fix header", inCmux: false, context: { usedTokens: 50_000, windowTokens: 200_000 } };
const other: AgentSession = { id: "s-2", source: "codex", title: "Tidy footer", inCmux: false };

const picker = (onSelect = vi.fn()): AgentPickerProps => ({
  heading: "Agent",
  sessions: { items: [live, other], state: "ready", selectedId: "s-1", onSelect, onNew: vi.fn(), selectLabel: session => `Use session ${session.title}` },
});

const fetchMock = vi.fn();
beforeEach(() => {
  fetchMock.mockResolvedValue(new Response(JSON.stringify({ id: "s-1", title: "Fix header", events: [] }), { status: 200 }));
  vi.stubGlobal("fetch", fetchMock);
});
afterEach(() => { cleanup(); vi.unstubAllGlobals(); });

describe("SessionTab", () => {
  it("embeds the delivery session's transcript from gavel and links to its dashboard page", async () => {
    render(<SessionTab server={SERVER} sessionId="s-1" liveCount={2} picker={picker()} />);
    await waitFor(() => expect(fetchMock.mock.calls.map(call => new URL(String(call[0])).pathname)).toContain("/api/captain/sessions/s-1"));
    expect((await screen.findByRole("link", { name: "Open session in gavel" })).getAttribute("href")).toBe(`${SERVER}/sessions/s-1`);
  });

  it("switches the delivery session from the Agent menu and closes it", async () => {
    const onSelect = vi.fn();
    render(<SessionTab server={SERVER} sessionId="s-1" liveCount={2} picker={picker(onSelect)} />);
    const trigger = await screen.findByRole("button", { name: "Agent" });
    expect(trigger.textContent).toContain("2");
    fireEvent.click(trigger);
    fireEvent.click(await screen.findByRole("button", { name: "Use session Tidy footer" }));
    expect([onSelect.mock.calls[0]?.[0]?.id, screen.queryByRole("button", { name: "Use session Tidy footer" })]).toEqual(["s-2", null]);
  });

  it("offers the Agent menu and loads no transcript while the next send starts a new session", () => {
    render(<SessionTab server={SERVER} sessionId={undefined} liveCount={undefined} picker={picker()} />);
    expect([screen.getByRole("button", { name: "Agent" }) !== null, screen.getByText(/starts a new session/) !== null, fetchMock.mock.calls.length]).toEqual([true, true, 0]);
  });
});

describe("SessionTabLabel", () => {
  it.each([
    { sessionId: "s-1", session: live, text: "Fix header25%" },
    { sessionId: "s-1", session: undefined, text: "s-1" },
    { sessionId: undefined, session: undefined, text: "New session" },
  ])("labels the tab $text", ({ sessionId, session, text }) => {
    const { container } = render(<SessionTabLabel sessionId={sessionId} session={session} />);
    expect(container.textContent).toBe(text);
  });
});
