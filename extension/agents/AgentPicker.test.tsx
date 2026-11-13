import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { AgentPicker, type AgentPickerProps, type AgentPickerSession } from "./AgentPicker";
import type { RuntimeCatalog } from "./runtime-families";

afterEach(cleanup);

const FIRST: AgentPickerSession = { id: "session-a", title: "Review changes", source: "codex", model: "example-model", cwd: "/work/example", status: "running" };
const SECOND: AgentPickerSession = { id: "session-b", title: "Fix layout", source: "claude", status: "idle" };
const CATALOG: RuntimeCatalog = {
  models: [],
  families: [{ id: "claude", label: "Claude", provider: "anthropic", modes: [{ id: "agent", label: "Agent", defaultModel: "example-agent", provider: "anthropic" }] }],
};
const RUNTIME = { spec: { model: "example-agent" } };

function sessions(overrides: Partial<NonNullable<AgentPickerProps["sessions"]>> = {}): NonNullable<AgentPickerProps["sessions"]> {
  return { items: [FIRST, SECOND], state: "ready", onSelect: vi.fn(), selectLabel: (session) => `Use session ${session.id}`, ...overrides };
}

describe("AgentPicker sessions", () => {
  it("lists each session with its folder and status and reports the one clicked", () => {
    const list = sessions();
    render(<AgentPicker sessions={list} />);
    fireEvent.click(screen.getByRole("button", { name: "Use session session-a" }));
    expect(list.onSelect).toHaveBeenCalledWith(FIRST);
    expect(screen.getByText("example")).toBeTruthy();
    expect(screen.getByText("running")).toBeTruthy();
    expect(screen.getByText("Fix layout")).toBeTruthy();
  });

  it("marks the selected session and offers New session, whose choice shows the runtime bar", () => {
    const onNew = vi.fn();
    const onChange = vi.fn();
    const { rerender } = render(<AgentPicker runtime={{ value: undefined, onChange, catalog: CATALOG }} sessions={sessions({ selectedId: FIRST.id, onNew })} />);
    expect(screen.getByRole("button", { name: "Use session session-a" }).getAttribute("aria-pressed")).toBe("true");
    expect(screen.getByRole("button", { name: "Use session session-b" }).getAttribute("aria-pressed")).toBe("false");
    expect(screen.getByRole("button", { name: "New session" }).getAttribute("aria-pressed")).toBe("false");
    expect(screen.queryByLabelText("Agent runtime")).toBeNull();

    fireEvent.click(screen.getByRole("button", { name: "New session" }));
    expect(onNew).toHaveBeenCalledOnce();

    rerender(<AgentPicker runtime={{ value: undefined, onChange, catalog: CATALOG }} sessions={sessions({ selectedId: undefined, onNew })} />);
    expect(screen.getByRole("button", { name: "New session" }).getAttribute("aria-pressed")).toBe("true");
    expect(screen.getByLabelText("Agent runtime")).toBeTruthy();
    expect(screen.getByRole("button", { name: "Use session session-a" }).getAttribute("aria-pressed")).toBe("false");
  });

  it("shows no model in the inline runtime bar while none is set, rather than a default the host would not send", () => {
    render(<AgentPicker runtime={{ value: undefined, onChange: vi.fn(), catalog: CATALOG }} sessions={sessions({ selectedId: undefined, onNew: vi.fn() })} />);
    const bar = screen.getByLabelText("Agent runtime");
    expect([...bar.querySelectorAll("[title]")].map(element => element.getAttribute("title")).filter(title => title?.startsWith("Model — "))).toEqual(["Model — unspecified"]);
    expect(bar.textContent).not.toContain("example-agent");
  });

  it("says an ended managed session resumes, where a running one shows its status", () => {
    const ended: AgentPickerSession = { id: "managed-1", title: "welcome threads", source: "claude", status: "ended", resumable: true };
    render(<AgentPicker sessions={sessions({ items: [FIRST, ended] })} />);
    expect(screen.getByText("resumes")).toBeTruthy();
    expect(screen.queryByText("ended")).toBeNull();
    expect(screen.getByText("running")).toBeTruthy();
  });

  describe("assigning every open thread", () => {
    const assignAll = (overrides: Partial<NonNullable<NonNullable<AgentPickerProps["sessions"]>["assignAll"]>> = {}) => ({ count: 3, onSession: vi.fn(), onNew: vi.fn(), ...overrides });

    it("assigns all to a chosen session or to a new session, and says how many threads that is", () => {
      const all = assignAll();
      render(<AgentPicker sessions={sessions({ assignAll: all })} />);
      fireEvent.click(screen.getByRole("button", { name: "Assign all to Fix layout" }));
      expect(all.onSession).toHaveBeenCalledWith(SECOND);
      fireEvent.click(screen.getByRole("button", { name: "Assign all to a new session" }));
      expect(all.onNew).toHaveBeenCalledOnce();
      expect(screen.getByText("3 open threads")).toBeTruthy();
    });

    it("disables every assign-all button, saying why, when the page has no open thread", () => {
      render(<AgentPicker sessions={sessions({ assignAll: assignAll({ count: 0 }) })} />);
      const buttons = screen.getAllByRole("button", { name: /^Assign all to / });
      expect(buttons.map((button) => [button.getAttribute("aria-label"), button.hasAttribute("disabled")])).toEqual([
        ["Assign all to a new session", true],
        ["Assign all to Review changes", true],
        ["Assign all to Fix layout", true],
      ]);
      expect(screen.getByText("No open threads on this page")).toBeTruthy();
    });

    it("offers no assign-all buttons unless asked to", () => {
      render(<AgentPicker sessions={sessions()} />);
      expect(screen.queryByRole("button", { name: /^Assign all to / })).toBeNull();
    });
  });

  it("offers no New session choice, and marks nothing selected, unless asked to", () => {
    render(<AgentPicker sessions={sessions()} />);
    expect(screen.queryByRole("button", { name: "New session" })).toBeNull();
    expect(screen.getByRole("button", { name: "Use session session-a" }).hasAttribute("aria-pressed")).toBe(false);
  });

  it.each([
    { state: "loading" as const, items: [], text: "Loading active sessions…" },
    { state: "ready" as const, items: [], text: "No active sessions." },
  ])("says so while the sessions are $state with none to show", ({ state, items, text }) => {
    render(<AgentPicker sessions={sessions({ items, state })} />);
    expect(screen.getByText(text)).toBeTruthy();
  });

  it("shows neither a loading nor an empty message after the sessions failed to load, only the error", () => {
    render(<AgentPicker sessions={sessions({ items: [], state: "error" })} error="Sessions unavailable: down" />);
    expect(screen.queryByText("No active sessions.")).toBeNull();
    expect(screen.queryByText("Loading active sessions…")).toBeNull();
    expect(screen.getByRole("alert").textContent).toBe("Sessions unavailable: down");
  });

  it("opens the full session list from All sessions…", () => {
    const onAll = vi.fn();
    render(<AgentPicker sessions={sessions({ onAll })} />);
    fireEvent.click(screen.getByRole("button", { name: "All sessions…" }));
    expect(onAll).toHaveBeenCalledOnce();
  });

  it("has no session section at all when no sessions are given", () => {
    render(<AgentPicker heading="Agent defaults" runtime={{ value: RUNTIME, onChange: vi.fn(), catalog: CATALOG }} />);
    expect(screen.getByRole("heading", { name: "Agent defaults" })).toBeTruthy();
    expect(screen.queryByText("Active sessions")).toBeNull();
  });
});

describe("AgentPicker settings", () => {
  it("reports the chosen session grouping", () => {
    const onChange = vi.fn();
    render(<AgentPicker grouping={{ value: "shared", onChange }} />);
    expect(screen.getByRole("radio", { name: "Shared" }).getAttribute("aria-checked")).toBe("true");
    fireEvent.click(screen.getByRole("radio", { name: "Per page" }));
    expect(onChange).toHaveBeenCalledWith("page");
  });

  it("reports the auto-run switch and explains when it is unavailable", () => {
    const onChange = vi.fn();
    const { rerender } = render(<AgentPicker autoRun={{ checked: false, onChange }} />);
    fireEvent.click(screen.getByRole("switch", { name: "Auto-run new comments on this page" }));
    expect(onChange).toHaveBeenCalledWith(true);
    rerender(<AgentPicker autoRun={{ checked: false, onChange, disabled: true, title: "Open a page to set its auto-run" }} />);
    expect(screen.getByRole("switch", { name: "Auto-run new comments on this page" }).hasAttribute("disabled")).toBe(true);
  });

  it("hands the model choice to the host's chooser when it has one", () => {
    const onChoose = vi.fn();
    render(<AgentPicker runtime={{ value: RUNTIME, onChoose, ariaLabel: "Choose playground model" }} />);
    fireEvent.click(screen.getByRole("button", { name: "Choose playground model" }));
    expect(onChoose).toHaveBeenCalledOnce();
    expect(screen.queryByLabelText("Agent runtime")).toBeNull();
  });

  it("names an unset runtime and expands the runtime bar inline from the catalog", () => {
    render(<AgentPicker runtime={{ value: undefined, onChange: vi.fn(), catalog: CATALOG, emptyLabel: "Gavel's default" }} />);
    const model = screen.getByRole("button", { name: "Choose model" });
    expect(model.textContent).toBe("Gavel's default");
    expect(model.getAttribute("aria-expanded")).toBe("false");
    fireEvent.click(model);
    expect(model.getAttribute("aria-expanded")).toBe("true");
    expect(screen.getByLabelText("Agent runtime")).toBeTruthy();
    fireEvent.click(model);
    expect(screen.queryByLabelText("Agent runtime")).toBeNull();
  });

  it("says the models are loading while the catalog is, and offers to clear a chosen runtime", () => {
    const onClear = vi.fn();
    const { rerender } = render(<AgentPicker runtime={{ value: RUNTIME, onChange: vi.fn(), catalog: undefined, onClear, clearLabel: "Use global default" }} />);
    fireEvent.click(screen.getByRole("button", { name: "Choose model" }));
    expect(screen.getByRole("status").textContent).toBe("Loading models…");
    fireEvent.click(screen.getByRole("button", { name: "Use global default" }));
    expect(onClear).toHaveBeenCalledOnce();
    rerender(<AgentPicker runtime={{ value: undefined, onChange: vi.fn(), catalog: CATALOG, onClear }} />);
    expect(screen.queryByRole("button", { name: "Use global default" })).toBeNull();
  });

  it("disables every control while busy and shows what it is waiting for", () => {
    render(<AgentPicker busy notice={<p role="status">Loading playground settings…</p>} grouping={{ value: "page", onChange: vi.fn() }} runtime={{ value: RUNTIME, onChoose: vi.fn() }} />);
    expect(screen.getByRole("radio", { name: "Shared" }).hasAttribute("disabled")).toBe(true);
    expect(screen.getByRole("button", { name: "Choose model" }).hasAttribute("disabled")).toBe(true);
    expect(screen.getByRole("status").textContent).toBe("Loading playground settings…");
  });
});
