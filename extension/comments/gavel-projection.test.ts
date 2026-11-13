import { expect, it } from "vitest";
import { bodyFor, contentFor, project, type NativeComment } from "./gavel-projection";

const attachmentUrl = (commentId: string, attachmentId: string) => `https://gavel.example/api/v1/comments/${commentId}/attachments/${attachmentId}`;
const native: NativeComment = {
  id: "c-1",
  bodyMarkdown: "Make this bigger",
  author: { kind: "user", id: "u-1", label: "Ada" },
  target: { uri: "gavel://project/site/pages/https%3A%2F%2Fexample.com%2F", scopeUri: "gavel://project/site", kind: "web-page" },
  status: "open",
  version: 3,
  createdAt: "2026-11-01T10:00:00Z",
  updatedAt: "2026-11-01T10:05:00Z",
  content: { playground: { page: "https://example.com/", anchor: "#hero", element: { source: "https://example.com/", html: "<h1>Hi</h1>" } } },
  attachments: [{ id: "a-1", name: "screenshot.png" }],
  messages: [{ id: "m-1", bodyMarkdown: "Done", author: { kind: "agent", id: "claude", label: "Claude" }, createdAt: "2026-11-01T10:06:00Z", content: { playground: { page: "https://example.com/" } } }],
};

it("projects a native comment into its root and replies, pointing screenshots at the host's attachment route", () => {
  const [root, reply] = project(native, { attachmentUrl });
  expect(root).toMatchObject({
    id: "c-1",
    version: 3,
    body: "Make this bigger",
    anchor: "#hero",
    status: "open",
    author: { id: "u-1", name: "Ada", kind: "user" },
    createdAt: "2026-11-01T10:00:00.000Z",
    updatedAt: "2026-11-01T10:05:00.000Z",
    element: { html: "<h1>Hi</h1>", screenshot: { status: "captured", filename: "screenshot.png", url: attachmentUrl("c-1", "a-1") } },
  });
  expect(reply).toMatchObject({ id: "m-1", rootId: "c-1", parentId: "c-1", body: "Done", anchor: "#hero", author: { id: "claude", name: "Claude", kind: "agent" } });
});

it("rejects a native comment that carries no page metadata", () => {
  expect(() => project({ ...native, content: {} as NativeComment["content"] }, { attachmentUrl })).toThrow(/missing playground metadata/);
});

it("projects the agent hand-off gavel records beside the playground fields", () => {
  const [root] = project({ ...native, content: { ...native.content, agentRun: { sessionId: "s-9", requestedAt: "2026-11-01T11:00:00Z", state: "sent", target: { label: "claude: site", channel: "cmux" } } } }, { attachmentUrl });
  expect(root?.agentRun).toEqual({ sessionId: "s-9", requestedAt: "2026-11-01T11:00:00Z", state: "sent", target: { label: "claude: site", channel: "cmux" } });
  expect(() => project({ ...native, content: { ...native.content, agentRun: { sessionId: "s-9", requestedAt: "2026-11-01T11:00:00Z", state: "bogus", target: { label: "x" } } } }, { attachmentUrl })).toThrow(/agent run state/);
});

it.each([
  { name: "an agent reply after the hand-off", messages: [{ kind: "agent", at: "2026-11-01T11:05:00Z" }], repliedAt: "2026-11-01T11:05:00.000Z" },
  { name: "the first of several agent replies", messages: [{ kind: "user", at: "2026-11-01T11:02:00Z" }, { kind: "agent", at: "2026-11-01T11:03:00Z" }, { kind: "agent", at: "2026-11-01T11:09:00Z" }], repliedAt: "2026-11-01T11:03:00.000Z" },
  { name: "only human replies", messages: [{ kind: "user", at: "2026-11-01T11:05:00Z" }], repliedAt: undefined },
  { name: "an agent reply written before the hand-off", messages: [{ kind: "agent", at: "2026-11-01T10:59:00Z" }], repliedAt: undefined },
])("dates the agent's first reply on the hand-off from $name", ({ messages, repliedAt }) => {
  const [root] = project({
    ...native,
    content: { ...native.content, agentRun: { sessionId: "s-9", requestedAt: "2026-11-01T11:00:00Z", state: "sent" } },
    messages: messages.map(({ kind, at }, index) => ({ id: `m-${index}`, bodyMarkdown: "reply", author: { kind, id: `${kind}:${index}` }, createdAt: at })),
  }, { attachmentUrl });
  expect(root?.agentRun?.repliedAt).toBe(repliedAt);
});

it.each([
  { name: "the agent a delivery assigned", assignee: { kind: "agent", id: "claude:s-1", label: "claude-opus-5-5" }, expected: { id: "claude:s-1", name: "claude-opus-5-5", kind: "agent" } },
  { name: "a person", assignee: { kind: "user", id: "ada" }, expected: { id: "ada", name: "ada", kind: "user" } },
  { name: "nobody, even when the playground copy still names an old assignee", assignee: null, expected: undefined },
])("takes the thread's assignee from gavel: $name", ({ assignee, expected }) => {
  const stale = { id: "codex", name: "Codex", kind: "agent" as const };
  const [root] = project({ ...native, assignee, content: { ...native.content, playground: { ...native.content.playground, assignee: stale } } }, { attachmentUrl });
  expect(root?.assignee).toEqual(expected);
});

it("keeps an empty body empty and names rating-only comments", () => {
  expect(bodyFor({ id: "x", body: "", createdAt: "", author: null, rating: "positive" })).toBe("Positive feedback");
  expect(bodyFor({ id: "x", body: "", createdAt: "", author: null })).toBe("Negative feedback");
  expect(bodyFor({ id: "x", body: "", createdAt: "", author: null, reviewDecision: "approved" })).toBe("Approved");
});

it("stores a comment's own fields under playground, leaving lifecycle fields to gavel", () => {
  expect(contentFor("https://example.com/", { id: "c-1", version: 2, body: "Hi", createdAt: "t", status: "open", author: { name: "You", kind: "user" }, anchor: "#a" })).toEqual({
    playground: { body: "Hi", author: { name: "You", kind: "user" }, anchor: "#a", page: "https://example.com/" },
  });
});
