import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { AgentRunActions } from "./AgentRunActions";

afterEach(cleanup);
const handlers = () => ({ onRetry: vi.fn(), onRetryFresh: vi.fn(), onSendFresh: vi.fn() });
const names = () => screen.queryAllByRole("button").map(button => button.textContent);

describe("AgentRunActions", () => {
  it("offers Retry and Retry in a new session for a failed hand-off and wires each to its callback", () => {
    const { onRetry, onRetryFresh, onSendFresh } = handlers();
    render(<AgentRunActions state="error" busy={false} onRetry={onRetry} onRetryFresh={onRetryFresh} onSendFresh={onSendFresh} />);
    expect(names()).toEqual(["Retry", "Retry in a new session"]);
    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    fireEvent.click(screen.getByRole("button", { name: "Retry in a new session" }));
    expect([onRetry.mock.calls.length, onRetryFresh.mock.calls.length, onSendFresh.mock.calls.length]).toEqual([1, 1, 0]);
  });

  it.each(["sent", "done"] as const)("offers only Send to a new session once the hand-off is %s", state => {
    const { onRetry, onRetryFresh, onSendFresh } = handlers();
    render(<AgentRunActions state={state} busy={false} onRetry={onRetry} onRetryFresh={onRetryFresh} onSendFresh={onSendFresh} />);
    expect(names()).toEqual(["Send to a new session"]);
    fireEvent.click(screen.getByRole("button", { name: "Send to a new session" }));
    expect([onRetry.mock.calls.length, onRetryFresh.mock.calls.length, onSendFresh.mock.calls.length]).toEqual([0, 0, 1]);
  });

  it("offers nothing while the hand-off is queued", () => {
    render(<AgentRunActions state="queued" busy={false} {...handlers()} />);
    expect(names()).toEqual([]);
  });

  it.each(["resolved", "closed"])("offers nothing on a %s thread, which gavel will not hand off again", status => {
    render(<AgentRunActions state="done" status={status} busy={false} {...handlers()} />);
    expect(names()).toEqual([]);
  });

  it("still offers actions on an in-progress thread", () => {
    render(<AgentRunActions state="error" status="in_progress" busy={false} {...handlers()} />);
    expect(names()).toEqual(["Retry", "Retry in a new session"]);
  });

  it("disables every action while a hand-off is in flight", () => {
    render(<AgentRunActions state="error" busy {...handlers()} />);
    expect(screen.getAllByRole("button").map(button => (button as HTMLButtonElement).disabled)).toEqual([true, true]);
  });
});
