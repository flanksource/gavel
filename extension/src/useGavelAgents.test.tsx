import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { Comment } from "@flanksource/clicky-ui";
import type { CommentAgent, CommentAgentSettings } from "../comments/gavel-delivery";
import type { DeliverTarget, GavelClient } from "./gavel-client";
import { useAdoptRunSession, useAgentSessions, useAutoRun } from "./useGavelAgents";

const page = { url: "https://example.com/docs", title: "Docs" };
const run: Extract<DeliverTarget, { kind: "run" }> = { kind: "run", prompt: "editor-comments", groupKey: "g-1" };
const project = { kind: "project" } as const;

function resolved(agent: CommentAgent): CommentAgentSettings {
  return { settings: { positions: [], layers: [], provenance: {} }, agent: { agent, provenance: {}, found: true } };
}

/** A dashboard whose settings resolve to the agent each call is queued with. */
function fakeClient(options: { initial: CommentAgent; saves?: CommentAgent[] }) {
  const saves = [...(options.saves ?? [])];
  const client = {
    scope: "gavel://project/site",
    commentAgent: vi.fn(async () => resolved(options.initial)),
    saveCommentAgent: vi.fn(async () => resolved(saves.shift() ?? {})),
  };
  return { client: client as unknown as GavelClient, calls: client };
}

async function loaded(client: GavelClient, target: Extract<DeliverTarget, { kind: "run" }> = run) {
  const hook = renderHook((props: { run: typeof target }) => useAutoRun({ client, page, run: props.run }), { initialProps: { run: target } });
  await waitFor(() => expect(hook.result.current.ready).toBe(true));
  return hook;
}

it("reads whether the page's resolved comment agent is enabled", async () => {
  const { client, calls } = fakeClient({ initial: { enabled: true, prompt: "editor-comments" } });
  const { result } = await loaded(client);
  expect(result.current.enabled).toBe(true);
  expect(calls.commentAgent).toHaveBeenCalledWith(page);
});

it("turns auto-run on by saving a run-derived agent as the project page entry", async () => {
  const { client, calls } = fakeClient({ initial: {}, saves: [{ enabled: true, prompt: "editor-comments", group: "g-1" }] });
  const { result } = await loaded(client);
  expect(result.current.enabled).toBe(false);
  await act(() => result.current.set(true));
  expect(calls.saveCommentAgent).toHaveBeenCalledWith({ page, layer: project, agent: { enabled: true, prompt: "editor-comments", group: "g-1" } });
  expect(result.current.enabled).toBe(true);
});

it("turns auto-run off by removing the page entry when nothing inherited keeps it on", async () => {
  const { client, calls } = fakeClient({ initial: { enabled: true }, saves: [{}] });
  const { result } = await loaded(client);
  await act(() => result.current.set(false));
  expect(calls.saveCommentAgent.mock.calls).toEqual([[{ page, layer: project, agent: null }]]);
  expect(result.current.enabled).toBe(false);
});

it("saves an explicit disabled agent when an inherited default still enables the page", async () => {
  const { client, calls } = fakeClient({ initial: { enabled: true }, saves: [{ enabled: true }, { enabled: false }] });
  const { result } = await loaded(client);
  await act(() => result.current.set(false));
  expect(calls.saveCommentAgent.mock.calls).toEqual([[{ page, layer: project, agent: null }], [{ page, layer: project, agent: { enabled: false } }]]);
  expect(result.current.enabled).toBe(false);
});

it("re-saves the agent when the run target changes while auto-run is on, and not while it is off", async () => {
  const next = { ...run, runOptions: { spec: { model: "codex-example" } } };
  const on = fakeClient({ initial: { enabled: true }, saves: [{ enabled: true }] });
  const enabled = await loaded(on.client);
  enabled.rerender({ run: next });
  await waitFor(() => expect(on.calls.saveCommentAgent).toHaveBeenCalledWith({ page, layer: project, agent: { enabled: true, prompt: "editor-comments", group: "g-1", spec: { model: "codex-example" } } }));
  const off = fakeClient({ initial: {} });
  const disabled = await loaded(off.client);
  disabled.rerender({ run: next });
  expect(off.calls.saveCommentAgent).not.toHaveBeenCalled();
});

it("reports a failed save and keeps the previous state", async () => {
  const { client, calls } = fakeClient({ initial: {} });
  calls.saveCommentAgent.mockRejectedValueOnce(new Error("403 only humans can change comment agents"));
  const { result } = await loaded(client);
  await act(() => result.current.set(true));
  expect([result.current.enabled, result.current.error]).toEqual([false, "403 only humans can change comment agents"]);
});

describe("useAgentSessions", () => {
  const session = (id: string) => ({ id, source: "claude", title: id, inCmux: true });

  const renderSessions = (agentSessions: ReturnType<typeof vi.fn>) => {
    const client = { agentSessions } as unknown as GavelClient;
    return renderHook(() => useAgentSessions(client));
  };

  it("loads the sessions and reloads them on demand", async () => {
    const agentSessions = vi.fn().mockResolvedValueOnce([session("s-1")]).mockResolvedValueOnce([session("s-1"), session("s-2")]);
    const { result } = renderSessions(agentSessions);
    await waitFor(() => expect(result.current.sessions?.map(entry => entry.id)).toEqual(["s-1"]));
    await act(() => result.current.refresh());
    expect(result.current.sessions?.map(entry => entry.id)).toEqual(["s-1", "s-2"]);
  });

  it("reports a failed reload without dropping the sessions it already has", async () => {
    const agentSessions = vi.fn().mockResolvedValueOnce([session("s-1")]).mockRejectedValueOnce(new Error("Cannot reach gavel"));
    const { result } = renderSessions(agentSessions);
    await waitFor(() => expect(result.current.sessions).toHaveLength(1));
    await act(() => result.current.refresh());
    expect([result.current.sessions?.length, result.current.error]).toEqual([1, "Cannot reach gavel"]);
  });
});

describe("useAdoptRunSession", () => {
  const before = { id: "c-1", agentRun: { sessionId: "old", requestedAt: "2026-11-01T10:00:00Z", state: "done" as const } };
  const queued = { id: "c-1", agentRun: { requestedAt: "2026-11-01T10:05:00Z", state: "queued" as const } };
  const recorded = { id: "c-1", agentRun: { sessionId: "s-new", requestedAt: "2026-11-01T10:05:00Z", state: "sent" as const } };
  const render = (comments: Pick<Comment, "id" | "agentRun">[], adopt = vi.fn()) => {
    const hook = renderHook((props: { comments: Pick<Comment, "id" | "agentRun">[] }) => useAdoptRunSession({ comments: props.comments, adopt }), { initialProps: { comments } });
    return { ...hook, adopt };
  };

  it("adopts the session gavel records on the delivered thread after the hand-off, not one a previous run left", () => {
    const { result, rerender, adopt } = render([before]);
    act(() => result.current.watch(["c-1"]));
    expect(adopt).not.toHaveBeenCalled();
    rerender({ comments: [queued] });
    expect(adopt).not.toHaveBeenCalled();
    rerender({ comments: [recorded] });
    expect(adopt.mock.calls).toEqual([["s-new"]]);
  });

  it("adopts once, even when more updates follow", () => {
    const { result, rerender, adopt } = render([before]);
    act(() => result.current.watch(["c-1"]));
    rerender({ comments: [recorded] });
    rerender({ comments: [{ ...recorded, agentRun: { ...recorded.agentRun, state: "done" as const } }] });
    expect(adopt).toHaveBeenCalledTimes(1);
  });

  it("adopts from a batch's first thread that gets a session", () => {
    const other = { id: "c-2" };
    const { result, rerender, adopt } = render([before, other]);
    act(() => result.current.watch(["c-1", "c-2"]));
    rerender({ comments: [before, { id: "c-2", agentRun: { sessionId: "s-new", requestedAt: "2026-11-01T10:06:00Z", state: "sent" as const } }] });
    expect(adopt.mock.calls).toEqual([["s-new"]]);
  });

  it("stops watching a thread whose hand-off failed, so a later unrelated session is not adopted", () => {
    const { result, rerender, adopt } = render([before]);
    act(() => result.current.watch(["c-1"]));
    rerender({ comments: [{ id: "c-1", agentRun: { requestedAt: "2026-11-01T10:05:00Z", state: "error" as const, error: "no capacity" } }] });
    rerender({ comments: [recorded] });
    expect(adopt).not.toHaveBeenCalled();
  });

  it("adopts nothing after the watch is cancelled", () => {
    const { result, rerender, adopt } = render([before]);
    act(() => result.current.watch(["c-1"]));
    act(() => result.current.cancel());
    rerender({ comments: [recorded] });
    expect(adopt).not.toHaveBeenCalled();
  });
});
