import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { SyncStatus } from "./SyncStatus";
import type { SyncStatus as Status } from "./sync-status";

afterEach(cleanup);

const connected: Status = { kind: "connected" };
const syncing: Status = { kind: "syncing", current: 2, total: 9 };
const syncError: Status = { kind: "sync-error", failed: 3, detail: "Rejected: no project" };
const connectionError: Status = { kind: "connection-error", detail: "Gavel is unreachable" };
const connecting: Status = { kind: "connecting" };

describe("label", () => {
  it.each<[Status, string]>([
    [connected, "Connected"],
    [connecting, "Connecting…"],
    [syncing, "Syncing 2 of 9"],
    [syncError, "Sync error"],
    [connectionError, "Connection error"],
  ])("shows exactly one label for %o", (status, label) => {
    render(<SyncStatus status={status} />);
    expect(screen.getByRole("status").textContent).toBe(label);
  });

  it("hides the dot from assistive tech", () => {
    render(<SyncStatus status={connected} />);
    expect(screen.getByRole("status").querySelector("[aria-hidden='true']")).not.toBeNull();
  });
});

describe("title", () => {
  it("carries the failure count and first failure for a sync error", () => {
    render(<SyncStatus status={syncError} />);
    expect(screen.getByRole("status").getAttribute("title")).toBe("3 failed: Rejected: no project");
  });

  it("carries the failure count alone when the error has no text", () => {
    render(<SyncStatus status={{ kind: "sync-error", failed: 2 }} />);
    expect(screen.getByRole("status").getAttribute("title")).toBe("2 failed");
  });

  it("carries only the detail when no entry failed (a queue read failure)", () => {
    render(<SyncStatus status={{ kind: "sync-error", failed: 0, detail: "Todo storage transaction aborted" }} />);
    expect(screen.getByRole("status").getAttribute("title")).toBe("Todo storage transaction aborted");
  });

  it("carries the reason for a connection error", () => {
    render(<SyncStatus status={connectionError} />);
    expect(screen.getByRole("status").getAttribute("title")).toBe("Gavel is unreachable");
  });

  it.each([connected, syncing, connecting])("has no title for %o", status => {
    render(<SyncStatus status={status} />);
    expect(screen.getByRole("status").hasAttribute("title")).toBe(false);
  });
});

describe("colour", () => {
  const classes = (status: Status) => {
    render(<SyncStatus status={status} />);
    const root = screen.getByRole("status");
    return { text: root.className, dot: root.querySelector("[aria-hidden='true']")!.className };
  };

  it.each([syncError, connectionError])("uses the destructive colour for the dot and text of %o", status => {
    const { text, dot } = classes(status);
    expect(text).toContain("text-destructive");
    expect(dot).toContain("bg-destructive");
  });

  it("uses the primary colour for the syncing dot", () => {
    expect(classes(syncing).dot).toContain("bg-primary");
  });

  it("uses a neutral dot with muted text while connecting, not the green of a confirmed connection", () => {
    const { text, dot } = classes(connecting);
    expect(dot).toContain("bg-muted-foreground");
    expect(dot).not.toContain("bg-green-500");
    expect(text).toContain("text-muted-foreground");
  });

  it("uses a green dot with muted text when connected", () => {
    const { text, dot } = classes(connected);
    expect(dot).toContain("bg-green-500");
    expect(text).toContain("text-muted-foreground");
  });
});

describe("onOpen", () => {
  it("renders a clickable button inside the live region for a sync error", () => {
    const onOpen = vi.fn();
    render(<SyncStatus status={syncError} onOpen={onOpen} />);
    const button = screen.getByRole("button", { name: "Sync error" });
    expect(button.getAttribute("type")).toBe("button");
    expect(button.getAttribute("title")).toBe("3 failed: Rejected: no project");
    expect(screen.getByRole("status").contains(button)).toBe(true);
    fireEvent.click(button);
    expect(onOpen).toHaveBeenCalledTimes(1);
  });

  it("stays a plain status for a sync error without onOpen", () => {
    render(<SyncStatus status={syncError} />);
    expect(screen.queryByRole("button")).toBeNull();
  });

  it.each([connected, syncing, connecting, connectionError])("stays a plain status for %o even with onOpen", status => {
    render(<SyncStatus status={status} onOpen={vi.fn()} />);
    expect(screen.queryByRole("button")).toBeNull();
    expect(screen.getByRole("status")).toBeTruthy();
  });
});
