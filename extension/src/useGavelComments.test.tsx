import { act, renderHook, waitFor } from "@testing-library/react";
import { expect, it, vi } from "vitest";
import type { GitDiffPayload } from "@flanksource/clicky-ui/data";
import type { StoredComment } from "../contracts/comments-store";
import type { GavelClient } from "./gavel-client";
import { useGavelComments } from "./useGavelComments";

const page = { url: "https://example.com/docs", title: "Docs" };
const root: StoredComment = { id: "c-1", version: 4, body: "Too small", createdAt: "2026-11-01T10:00:00Z", author: { name: "You", kind: "user" }, status: "open", anchor: "#hero" };
const reply: StoredComment = { id: "m-1", rootId: "c-1", parentId: "c-1", version: 4, body: "Fixed", createdAt: "2026-11-01T10:01:00Z", author: { name: "Claude", kind: "agent" }, anchor: "#hero" };

/** A dashboard client whose event stream stays open and silent, as an idle project's does. */
function fakeClient(overrides: Record<string, unknown> = {}) {
  const client = {
    scope: "gavel://project/site",
    list: vi.fn(async () => [root, reply]),
    act: vi.fn(async () => root),
    batch: vi.fn(async () => []),
    reply: vi.fn(async () => [root, reply]),
    create: vi.fn(async () => root),
    deliver: vi.fn(async () => root),
    deliverBatch: vi.fn(async (): Promise<StoredComment[]> => []),
    proposalDiff: vi.fn(async (): Promise<GitDiffPayload> => ({ diff: "" })),
    streamEvents: vi.fn(({ signal }: { signal: AbortSignal }) => new Promise<void>(resolve => signal.addEventListener("abort", () => resolve()))),
    ...overrides,
  };
  return { client: client as unknown as GavelClient, calls: client };
}

async function loaded(client: GavelClient) {
  const hook = renderHook(() => useGavelComments({ client, page }));
  await waitFor(() => expect(hook.result.current.loading).toBe(false));
  return hook;
}

it("loads the page's comments and refreshes after a lifecycle action at the version it last saw", async () => {
  const { client, calls } = fakeClient();
  const { result } = await loaded(client);
  expect(result.current.comments.map(comment => comment.id)).toEqual(["c-1", "m-1"]);
  await act(() => result.current.updateStatus("c-1", "resolved"));
  expect(calls.act).toHaveBeenCalledWith({ id: "c-1", version: 4, action: "resolve" });
  expect(calls.list).toHaveBeenCalledTimes(2);
  await act(() => result.current.updateStatus("c-1", "open"));
  expect(calls.act).toHaveBeenLastCalledWith({ id: "c-1", version: 4, action: "reopen" });
});

it("refuses a status gavel cannot set directly and lifecycle actions on replies", async () => {
  const { client } = fakeClient();
  const { result } = await loaded(client);
  expect(() => result.current.updateStatus("c-1", "in_progress")).toThrow(/cannot be set directly/);
  expect(() => result.current.close("m-1")).toThrow(/Only thread roots/);
  expect(() => result.current.close("missing")).toThrow(/not loaded/);
});

it("replies on the thread root and deletes a reply through its root", async () => {
  const { client, calls } = fakeClient();
  const { result } = await loaded(client);
  await act(() => result.current.reply({ parentId: "m-1", body: "Thanks" }));
  expect(calls.reply).toHaveBeenCalledWith({ page, root: "c-1", version: 4, body: "Thanks" });
  await act(() => result.current.remove("m-1"));
  expect(calls.batch).toHaveBeenCalledWith([{ id: "c-1", expectedVersion: 4, action: "delete-reply", messageId: "m-1" }]);
  await act(() => result.current.remove("c-1"));
  expect(calls.act).toHaveBeenLastCalledWith({ id: "c-1", version: 4, action: "delete" });
});

it("decides a proposed patch through the thread root at the version it last saw, then refreshes", async () => {
  const { client, calls } = fakeClient();
  const { result } = await loaded(client);
  await act(() => result.current.applyProposal("m-1", "stage"));
  expect(calls.act).toHaveBeenLastCalledWith({ id: "c-1", version: 4, action: "accept-proposal", messageId: "m-1", mode: "stage" });
  await act(() => result.current.rejectProposal("m-1", "Too broad"));
  expect(calls.act).toHaveBeenLastCalledWith({ id: "c-1", version: 4, action: "reject-proposal", messageId: "m-1", bodyMarkdown: "Too broad" });
  await act(() => result.current.rejectProposal("m-1"));
  expect(calls.act).toHaveBeenLastCalledWith({ id: "c-1", version: 4, action: "reject-proposal", messageId: "m-1" });
  await act(() => result.current.requestRevision("m-1", "Keep the API"));
  expect(calls.act).toHaveBeenLastCalledWith({ id: "c-1", version: 4, action: "request-revision", messageId: "m-1", bodyMarkdown: "Keep the API" });
  expect(calls.list).toHaveBeenCalledTimes(5);
});

it("loads a proposal's diff for its reply and refuses a proposal decision on a comment that is not a loaded reply", async () => {
  const payload = { diff: "diff --git a/x b/x\n", commit: "def456" };
  const { client, calls } = fakeClient({ proposalDiff: vi.fn(async () => payload) });
  const { result } = await loaded(client);
  expect(await result.current.loadProposalDiff("m-1")).toEqual(payload);
  expect(calls.proposalDiff).toHaveBeenCalledWith("c-1", "m-1");
  expect(() => result.current.applyProposal("c-1", "apply")).toThrow(/not a reply/);
  expect(() => result.current.rejectProposal("missing")).toThrow(/not loaded/);
  expect(() => result.current.loadProposalDiff("missing")).toThrow(/not loaded/);
});

it("assigns a thread with the assignee as gavel's actor, or clears it", async () => {
  const { client, calls } = fakeClient();
  const { result } = await loaded(client);
  await act(() => result.current.assign("c-1", { id: "claude", name: "Claude", kind: "agent" }));
  expect(calls.act).toHaveBeenLastCalledWith({ id: "c-1", version: 4, action: "assign", assignee: { kind: "agent", id: "claude", label: "Claude" } });
  await act(() => result.current.assign("c-1", null));
  expect(calls.act).toHaveBeenLastCalledWith({ id: "c-1", version: 4, action: "assign", assignee: null });
});

it("reloads the thread after a failed delivery so the recorded hand-off error shows", async () => {
  const { client, calls } = fakeClient({ deliver: vi.fn(async () => { throw new Error("no live agent session"); }) });
  const { result } = await loaded(client);
  await act(async () => {
    await expect(result.current.deliver("c-1", { kind: "session", sessionId: "s-1" })).rejects.toThrow(/no live agent session/);
  });
  expect(calls.deliver).toHaveBeenCalledWith({ id: "c-1", version: 4, requestId: expect.any(String), target: { kind: "session", sessionId: "s-1" } });
  expect(calls.list).toHaveBeenCalledTimes(2);
});

it("delivers several threads in one batch at the versions last seen, then reloads", async () => {
  const other: StoredComment = { ...root, id: "c-2", version: 7 };
  const { client, calls } = fakeClient({ list: vi.fn(async () => [root, other]), deliverBatch: vi.fn(async () => [root, other]) });
  const { result } = await loaded(client);
  await act(() => result.current.deliverAll(["c-1", "c-2"], { kind: "session", sessionId: "s-1" }));
  expect(calls.deliverBatch).toHaveBeenCalledWith({
    requestId: expect.any(String),
    target: { kind: "session", sessionId: "s-1" },
    threads: [{ id: "c-1", expectedVersion: 4 }, { id: "c-2", expectedVersion: 7 }],
  });
  expect(calls.list).toHaveBeenCalledTimes(2);
});

it("refuses a batch naming a thread that is not loaded and reloads after a refused batch", async () => {
  const { client, calls } = fakeClient({ deliverBatch: vi.fn(async () => { throw new Error("comment c-1 changed (409)"); }) });
  const { result } = await loaded(client);
  await act(async () => {
    await expect(result.current.deliverAll(["c-1", "missing"], { kind: "session", sessionId: "s-1" })).rejects.toThrow(/not loaded/);
    await expect(result.current.deliverAll(["c-1"], { kind: "session", sessionId: "s-1" })).rejects.toThrow(/409/);
  });
  expect(calls.list).toHaveBeenCalledTimes(2);
});

it("surfaces a failed load and a stopped event stream instead of showing no comments", async () => {
  const { client } = fakeClient({ list: vi.fn(async () => { throw new Error("Cannot reach gavel"); }), streamEvents: vi.fn(async () => { throw new Error("connection reset"); }) });
  const { result } = renderHook(() => useGavelComments({ client, page }));
  await waitFor(() => expect(result.current.error).toMatch(/Cannot reach gavel|Live updates stopped: connection reset/));
  await waitFor(() => expect(result.current.loading).toBe(false));
});
