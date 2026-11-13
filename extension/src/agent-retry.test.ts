import { describe, expect, it } from "vitest";
import type { CommentAgentRun } from "@flanksource/clicky-ui/comments";
import type { DeliverTarget } from "../comments/gavel-delivery";
import { freshTarget, retryTarget } from "./agent-retry";

const GROUP = "gavel://project/demo/pages/https%3A%2F%2Fexample.test%2Fapp";
const RUN_TARGET: DeliverTarget = { kind: "run", prompt: "editor-comments", groupKey: GROUP, runOptions: { spec: { model: "example-model" } } };
const failed = (sessionId?: string): CommentAgentRun => ({
  requestedAt: "2026-11-11T10:00:00.000Z",
  state: "error",
  error: "session is gone",
  ...(sessionId ? { sessionId } : {}),
});

describe("retryTarget", () => {
  it("re-delivers to the session the failed run recorded when the panel sends to a session", () => {
    expect(retryTarget({ run: failed("recorded"), delivery: { kind: "session", sessionId: "chosen" } })).toEqual({ kind: "session", sessionId: "recorded" });
  });

  it("falls back to the panel's session when the failed run recorded none", () => {
    expect(retryTarget({ run: failed(), delivery: { kind: "session", sessionId: "chosen" } })).toEqual({ kind: "session", sessionId: "chosen" });
  });

  it("re-delivers a failed run to the same run target and group, never marking it fresh", () => {
    expect(retryTarget({ run: failed("managed-1"), delivery: RUN_TARGET })).toEqual(RUN_TARGET);
  });
});

describe("freshTarget", () => {
  it("marks the run target fresh and keeps its group key and options", () => {
    expect(freshTarget(RUN_TARGET)).toEqual({ ...RUN_TARGET, fresh: true });
  });

  it("refuses a session target, which has no group to restart", () => {
    expect(() => freshTarget({ kind: "session", sessionId: "chosen" })).toThrow(/new session.*run target/i);
  });
});
