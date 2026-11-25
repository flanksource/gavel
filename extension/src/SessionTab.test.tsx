import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { AgentSession } from "./gavel-client";
import { SessionTab, SessionTabLabel } from "./SessionTab";

const SERVER = "http://gavel.example.test";
const live: AgentSession = { id: "s-1", source: "claude", title: "Fix header", inCmux: false, context: { usedTokens: 50_000, windowTokens: 200_000 } };

const fetchMock = vi.fn();
beforeEach(() => {
  fetchMock.mockResolvedValue(new Response(JSON.stringify({ id: "s-1", title: "Fix header", events: [] }), { status: 200 }));
  vi.stubGlobal("fetch", fetchMock);
});
afterEach(() => { cleanup(); vi.unstubAllGlobals(); });

describe("SessionTab", () => {
  it("links to the Amp conversation without requesting a transcript", () => {
    fetchMock.mockClear();
    render(<SessionTab server={SERVER} sessionId="T-review-orb" />);
    expect(screen.getByRole("link", { name: "Open in Amp" }).getAttribute("href")).toBe("https://ampcode.com/threads/T-review-orb");
    expect(fetchMock).not.toHaveBeenCalled();
  });
  it("embeds the delivery session's transcript from gavel and links to its dashboard page", async () => {
    render(<SessionTab server={SERVER} sessionId="s-1" />);
    await waitFor(() => expect(fetchMock.mock.calls.map(call => new URL(String(call[0])).pathname)).toContain("/api/captain/sessions/s-1"));
    expect((await screen.findByRole("link", { name: "Open session in gavel" })).getAttribute("href")).toBe(`${SERVER}/sessions/s-1`);
  });

  it("keeps agent settings out of the transcript view", async () => {
    render(<SessionTab server={SERVER} sessionId="s-1" />);
    await screen.findByRole("link", { name: "Open session in gavel" });
    expect([screen.queryByRole("button", { name: "Agent" }), screen.queryByRole("switch")]).toEqual([null, null]);
  });

  it("points to Settings and loads no transcript until a session is selected", () => {
    render(<SessionTab server={SERVER} sessionId={undefined} />);
    expect([screen.getByText(/Choose an existing session in Settings/) !== null, screen.queryByRole("button", { name: "Agent" }), fetchMock.mock.calls.length]).toEqual([true, null, 0]);
  });
});

describe("SessionTabLabel", () => {
  it.each(["claude", "codex"])("shows the %s provider icon with the session title", provider => {
    render(<SessionTabLabel sessionId={live.id} session={{ ...live, source: provider }} />);
    expect([screen.getByTitle(provider).querySelector("svg") !== null, screen.getByText(live.title) !== null]).toEqual([true, true]);
  });

  it.each([
    { sessionId: "s-1", session: live, text: "Fix header25%" },
    { sessionId: "s-1", session: undefined, text: "s-1" },
    { sessionId: undefined, session: undefined, text: "New session" },
  ])("labels the tab $text", ({ sessionId, session, text }) => {
    const { container } = render(<SessionTabLabel sessionId={sessionId} session={session} />);
    expect(container.textContent).toBe(text);
  });
});
