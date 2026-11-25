import { describe, expect, it, vi } from "vitest";
import { agentGlyph } from "@flanksource/clicky-ui/comments";
import { UiRobotAi, UiSliders } from "@flanksource/clicky-ui/icons";
import type { Comment } from "@flanksource/clicky-ui";
import { agentAssignItems, assignableRoots, openRoots } from "./assign-items";

const live = { id: "s-1", source: "claude", title: "Fix header", model: "claude-opus-5-5", inCmux: true, status: "active" };
const sleeping = { id: "s-2", source: "codex", title: "Docs pass", inCmux: false, status: "sleeping" };
const ended = { id: "s-3", source: "claude", title: "Old run", model: "claude-opus-5-5", inCmux: false, status: "ended", managed: true, resumable: true };
const runtime = { spec: { model: "anthropic/claude-sonnet-5-5" } };
const handlers = () => ({ onSession: vi.fn(), onNew: vi.fn(), onAdvanced: vi.fn() });
const summary = (items: { label: unknown; icon?: unknown }[]) => items.map(({ label, icon }) => ({ label, icon }));

describe("agentAssignItems", () => {
  it("lists live sessions, then ended managed ones marked as resuming, then a new session and Advanced", () => {
    const on = handlers();
    const items = agentAssignItems({ sessions: [ended, live, sleeping], currentSessionId: undefined, runtime, disabled: false, ...on });
    expect(summary(items)).toEqual([
      { label: "Fix header", icon: agentGlyph("claude-opus-5-5").icon },
      { label: "Docs pass", icon: agentGlyph("codex").icon },
      { label: "Old run · resumes", icon: agentGlyph("claude-opus-5-5").icon },
      { label: "New Sonnet 5.5", icon: agentGlyph("anthropic/claude-sonnet-5-5").icon },
      { label: "Advanced…", icon: UiSliders },
    ]);
    items.forEach(item => item.onSelect?.());
    expect(on.onSession.mock.calls).toEqual([[live], [sleeping], [ended]]);
    expect([on.onNew.mock.calls.length, on.onAdvanced.mock.calls.length]).toEqual([1, 1]);
  });

  it("skips the session the thread is already with", () => {
    const items = agentAssignItems({ sessions: [live, sleeping, ended], currentSessionId: "s-2", runtime, disabled: false, ...handlers() });
    expect(items.map(item => item.label)).toEqual(["Fix header", "Old run · resumes", "New Sonnet 5.5", "Advanced…"]);
  });

  it("explains that delivering to an ended session resumes it", () => {
    const [first] = agentAssignItems({ sessions: [ended], currentSessionId: undefined, runtime, disabled: false, ...handlers() });
    expect(first?.title).toMatch(/resumes/i);
  });

  it("names gavel's default runtime and lists no session entry while there are no sessions", () => {
    expect(summary(agentAssignItems({ sessions: [], currentSessionId: undefined, runtime: undefined, disabled: false, ...handlers() }))).toEqual([
      { label: "New session (Gavel's default)", icon: UiRobotAi },
      { label: "Advanced…", icon: UiSliders },
    ]);
  });

  it("disables every entry for a thread gavel will not hand off", () => {
    expect(agentAssignItems({ sessions: [live, ended], currentSessionId: undefined, runtime, disabled: true, ...handlers() }).map(item => item.disabled)).toEqual([true, true, true, true]);
  });
});

const thread = (id: string, overrides: Partial<Comment> = {}): Comment => ({ id, body: id, createdAt: "2026-11-01T10:00:00Z", author: { name: "You", kind: "user" }, status: "open", ...overrides }) as Comment;
const withSession = (id: string, sessionId: string) => thread(id, { agentRun: { sessionId, requestedAt: "2026-11-01T10:01:00Z", state: "sent" } });

describe("openRoots", () => {
  it("keeps open thread roots only, not replies or resolved threads", () => {
    const comments = [thread("c-1"), thread("m-1", { parentId: "c-1" }), thread("c-2", { status: "resolved" }), thread("c-3", { status: "in_progress" })];
    expect(openRoots(comments).map(comment => comment.id)).toEqual(["c-1", "c-3"]);
  });
});

it("excludes successful deliveries to the current session and keeps failed deliveries eligible", () => {
  const failed = withSession("failed", "s-1");
  const threads = [
    thread("new"),
    withSession("delivered", "s-1"),
    withSession("elsewhere", "s-2"),
    { ...failed, agentRun: { ...failed.agentRun!, state: "error" as const } },
  ];
  expect(assignableRoots(threads, "s-1").map(comment => comment.id)).toEqual(["new", "elsewhere", "failed"]);
});
