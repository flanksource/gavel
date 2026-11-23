import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { Comment } from "@flanksource/clicky-ui";
import type { CommentAgent, CommentAgentSettings } from "../comments/gavel-delivery";
import type { GavelClient } from "./gavel-client";
import { useAdoptRunSession, useAgentSessions, useCommentAgent } from "./useGavelAgents";

const page = { url: "https://example.com/docs", title: "Docs" };
const project = { kind: "project" } as const;

function resolved(agent: CommentAgent): CommentAgentSettings {
  return { settings: { positions: [{ kind: "project", dir: "/work/site", files: [], db: true }], layers: [], provenance: {} }, agent: { agent, provenance: {}, found: true } };
}

/** A dashboard whose settings resolve to the agent each call is queued with. */
function fakeClient(options: { initial: CommentAgent; saves?: CommentAgent[] }) {
  const saves = [...(options.saves ?? [])];
  const client = {
    scope: "gavel://project/site",
    commentAgent: vi.fn(async () => resolved(options.initial)),
    saveCommentAgent: vi.fn(async (_request: { agent: CommentAgent | null }) => resolved(saves.shift() ?? {})),
  };
  return { client: client as unknown as GavelClient, calls: client };
}

async function loaded(client: GavelClient) {
  const hook = renderHook(() => useCommentAgent({ client, page }));
  await waitFor(() => expect(hook.result.current.ready).toBe(true));
  return hook;
}

describe("useCommentAgent", () => {
  it("reads the page's resolved comment agent without writing a rule", async () => {
    const { client, calls } = fakeClient({ initial: { enabled: true, prompt: "editor-comments" } });
    const { result } = await loaded(client);
    expect(result.current.agent?.agent).toEqual({ enabled: true, prompt: "editor-comments" });
    expect(calls.commentAgent).toHaveBeenCalledWith(page);
    expect(calls.saveCommentAgent).not.toHaveBeenCalled();
  });

  it("reads how the page's agent edits and defaults to direct edits", async () => {
    const proposing = await loaded(fakeClient({ initial: { edits: "propose" } }).client);
    expect(proposing.result.current.edits).toBe("propose");
    const unset = await loaded(fakeClient({ initial: {} }).client);
    expect(unset.result.current.edits).toBe("direct");
  });

  it("patches only the edits policy, as the project page entry", async () => {
    const { client, calls } = fakeClient({ initial: { enabled: false }, saves: [{ enabled: false, edits: "propose" }] });
    const { result } = await loaded(client);
    await act(() => result.current.setEdits("propose"));
    expect(calls.saveCommentAgent).toHaveBeenCalledWith({ page, layer: project, agent: { edits: "propose" } });
    expect(result.current.edits).toBe("propose");
  });

  it("reports a failed save and keeps the previous state", async () => {
    const { client, calls } = fakeClient({ initial: {} });
    calls.saveCommentAgent.mockRejectedValueOnce(new Error("403 only humans can change comment agents"));
    const { result } = await loaded(client);
    await act(() => result.current.setEdits("propose"));
    expect([result.current.edits, result.current.error]).toEqual(["direct", "403 only humans can change comment agents"]);
  });

  it("reports a failed settings read without initializing a rule", async () => {
    const { client, calls } = fakeClient({ initial: {} });
    calls.commentAgent.mockRejectedValueOnce(new Error("Cannot read page agent"));
    const { result } = renderHook(() => useCommentAgent({ client, page }));
    await waitFor(() => expect(result.current.error).toBe("Cannot read page agent"));
    expect(result.current.ready).toBe(false);
    expect(calls.saveCommentAgent).not.toHaveBeenCalled();
  });

  it("serializes saves so an older change cannot overwrite a later one", async () => {
    const { client, calls } = fakeClient({ initial: {}, saves: [{ edits: "direct" }] });
    let finishFirst!: (settings: CommentAgentSettings) => void;
    calls.saveCommentAgent.mockImplementationOnce(() => new Promise(resolve => { finishFirst = resolve; }));
    const { result } = await loaded(client);
    let first!: Promise<void>;
    let second!: Promise<void>;
    act(() => { first = result.current.setEdits("propose"); second = result.current.setEdits("direct"); });
    await waitFor(() => expect(calls.saveCommentAgent).toHaveBeenCalledTimes(1));
    await act(async () => { finishFirst(resolved({ edits: "propose" })); await Promise.all([first, second]); });
    expect([result.current.edits, calls.saveCommentAgent.mock.calls.map(call => call[0]?.agent)]).toEqual(["direct", [{ edits: "propose" }, { edits: "direct" }]]);
  });
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

  it("adopts a run already recorded when the created thread ID comes back", () => {
    const { result, adopt } = render([recorded]);
    act(() => result.current.watchNew(recorded.id));
    expect(adopt).toHaveBeenCalledExactlyOnceWith("s-new");
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
