import { describe, expect, it, vi } from "vitest";
import type { Comment } from "@flanksource/clicky-ui/comments";
import { agentOf, sendToItems, type AgentSession } from "./agent-sessions";
import { ApiError } from "./api";
import { missingProjectOf, threadItems, type NativeComment } from "./file-comments";

const fixer: AgentSession = { id: "s-1", source: "claude", title: "rename-pass", model: "claude-opus", cwd: "/repo", inCmux: false };
const reviewer: AgentSession = { id: "s-2", source: "codex", title: "review", inCmux: true };
const agent = { kind: "agent" as const, id: "claude:s-1", label: "claude-opus" };
const root = (agentRun?: Comment["agentRun"]): Comment => ({ id: "c-1", body: "Rename x", createdAt: "2026-11-08T10:00:00Z", author: null, status: "in_progress", ...(agentRun ? { agentRun } : {}) });

describe("agentOf", () => {
  it("names the session's agent by its comments actor id and model, falling back to its source", () => {
    expect([agentOf(fixer), agentOf(reviewer)]).toEqual([
      { id: "claude:s-1", name: "claude-opus", kind: "agent" },
      { id: "codex:s-2", name: "codex", kind: "agent" },
    ]);
  });
});

describe("sendToItems", () => {
  it("offers every live session and a new one, disabling those while a session works the thread", () => {
    const items = sendToItems(root({ sessionId: "s-1", requestedAt: "2026-11-08T10:01:00Z", state: "sent" }), [fixer, reviewer], async () => {});
    expect(items.map((item) => [item.label, Boolean(item.disabled)])).toEqual([
      ["Send to rename-pass", true],
      ["Send to review", false],
      ["Run new agent session", true],
    ]);
  });

  it("re-offers the session whose last send failed, and sends to the chosen session or a new one", () => {
    const send = vi.fn(async (_sessionId?: string) => {});
    const items = sendToItems(root({ sessionId: "s-1", requestedAt: "2026-11-08T10:01:00Z", state: "error", error: "socket gone" }), [fixer], send);
    expect(items.map((item) => Boolean(item.disabled))).toEqual([false, false]);
    items[0]?.onSelect();
    items[1]?.onSelect();
    expect(send.mock.calls).toEqual([["s-1"], []]);
  });

  it("offers nothing on a reply", () => {
    expect(sendToItems({ ...root(), parentId: "c-0" }, [fixer], async () => {})).toEqual([]);
  });
});

describe("missingProjectOf", () => {
  const suggestion = { name: "notes", dir: "/home/ada/notes" };
  const noProject = JSON.stringify({ error: "notes.md is not inside a gavel project", code: "no_project", suggestion });

  it("reads the project to create from a 409 no_project answer", () => {
    expect(missingProjectOf(new ApiError(409, noProject))).toEqual({ error: "notes.md is not inside a gavel project", suggestion });
  });

  it("leaves every other failure to the error banner", () => {
    expect([
      missingProjectOf(new ApiError(409, "file changed on disk")),
      missingProjectOf(new ApiError(500, noProject)),
      missingProjectOf(new Error(noProject)),
    ]).toEqual([undefined, undefined, undefined]);
  });
});

describe("threadItems", () => {
  it("carries the root's assignee and agent hand-off, and maps replies under it", () => {
    const agentRun = { sessionId: "s-1", requestedAt: "2026-11-08T10:01:00Z", state: "sent" as const, target: { label: "rename-pass", channel: "claude-socket" } };
    const row: NativeComment = {
      id: "c-1",
      bodyMarkdown: "Rename x",
      version: 4,
      status: "in_progress",
      createdAt: "2026-11-08T10:00:00Z",
      author: { kind: "user", id: "local:501", label: "You" },
      assignee: agent,
      content: { agentRun },
      target: { selector: { kind: "document", version: 1, view: "source" }, revision: "r1" },
      messages: [{ id: "m-1", bodyMarkdown: "Renamed", createdAt: "2026-11-08T10:05:00Z", author: agent }],
    };
    const [first, reply] = threadItems(row);
    expect([first?.assignee, first?.agentRun, reply?.parentId, reply?.author]).toEqual([
      { id: "claude:s-1", name: "claude-opus", kind: "agent" },
      agentRun,
      "c-1",
      { id: "claude:s-1", name: "claude-opus", kind: "agent" },
    ]);
  });
});
