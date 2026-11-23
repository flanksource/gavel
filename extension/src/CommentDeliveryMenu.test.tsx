import { cleanup, fireEvent, render, screen, within } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { agentRuntimeFamilies } from "../agents/runtime-families";
import { CommentDeliveryMenu, type DraftTarget } from "./CommentDeliveryMenu";
import type { AgentSession, GavelClient } from "./gavel-client";
import { context } from "./todos/testing/runtime";

afterEach(cleanup);

const preview = (defaultState: Parameters<typeof CommentDeliveryMenu>[0]["preview"]["defaultState"] = { preview: {} }) =>
  ({ client: {} as GavelClient, page: undefined, sessionId: undefined, revision: undefined, defaultState, body: "Check layout" });
const openMenu = async () => {
  fireEvent.click(screen.getByRole("button", { name: /^Runs on/ }));
  return within(await screen.findByRole("dialog", { name: "Run this comment on" }));
};

it("edits the timeout in the Runs on dropdown and uses it for a new session", async () => {
  const onChange = vi.fn();
  const model = context.models[0]!;
  const catalog = {
    models: context.models,
    families: agentRuntimeFamilies(context.runtimes.map(family => ({ ...family, modes: family.modes.map(mode => ({
      ...mode, defaultModel: model.id, availability: { state: "available" as const },
    })) }))),
  };
  render(<CommentDeliveryMenu targets={[]} onChange={onChange} catalog={catalog} preview={preview()} />);
  const dropdown = await openMenu();
  expect(dropdown.getByRole("group", { name: "New session runtime" })).toBeTruthy();
  fireEvent.click(dropdown.getByRole("menuitem", { name: "Runtime options" }));
  fireEvent.click(await screen.findByRole("menuitem", { name: "Timeout" }));
  fireEvent.click(await screen.findByText("30m"));
  fireEvent.click(dropdown.getByRole("button", { name: "Use new session" }));
  expect(onChange).toHaveBeenCalledWith({ kind: "run", runtime: { spec: { mode: "agent", model: model.id, budget: { timeout: "30m" } } } });
});

it("offers Amp picker results and normalizes a pasted URL to an explicit session target", async () => {
  const onChange = vi.fn();
  const id = "T-review-orb";
  render(<CommentDeliveryMenu targets={[]} onChange={onChange}
    sessions={[{ id, source: "amp", title: "Review the header", inCmux: false }]} preview={preview()} />);
  fireEvent.click((await openMenu()).getByRole("button", { name: /Review the header \(Amp\)/ }));
  expect(onChange).toHaveBeenLastCalledWith({ kind: "session", sessionId: id });

  const menu = await openMenu();
  const field = menu.getByLabelText(/Amp thread ID or URL/);
  const button = menu.getByRole<HTMLButtonElement>("button", { name: "Use Amp thread" });
  fireEvent.change(field, { target: { value: "https://example.test/threads/T-invalid" } });
  expect(button.disabled).toBe(true);
  fireEvent.change(field, { target: { value: "https://ampcode.com/threads/T-pasted" } });
  fireEvent.click(button);
  expect(onChange).toHaveBeenLastCalledWith({ kind: "session", sessionId: "T-pasted" });
});

describe("Runs on session choices", () => {
  const GLOBAL: AgentSession = { id: "global-1", source: "claude", title: "gavel-global-a1b2c3", inCmux: true, status: "active" };
  const PAGE: AgentSession = { id: "page-1", source: "codex", title: "gavel-app.example.test-d4e5f6", inCmux: true, status: "active" };
  const OTHER: AgentSession = { id: "other-1", source: "claude", title: "Fix layout", inCmux: true, status: "idle" };
  const sessionIds = { global: GLOBAL.id, page: PAGE.id };
  const defaultState = { preview: { session: PAGE } };
  const menu = (props: Partial<Parameters<typeof CommentDeliveryMenu>[0]> = {}) => {
    const onChange = vi.fn();
    render(<CommentDeliveryMenu targets={[]} onChange={onChange} sessions={[GLOBAL, PAGE, OTHER]} sessionIds={sessionIds} preview={preview(defaultState)} {...props} />);
    return onChange;
  };

  it("lists the default named for the session it resolves to, the global and page sessions, and the other active sessions", async () => {
    menu();
    const dropdown = await openMenu();
    expect(dropdown.getByRole("button", { name: /^Default gavel-app\.example\.test-d4e5f6/ })).toBeTruthy();
    expect(dropdown.getByRole("button", { name: /^Global session gavel-global-a1b2c3/ })).toBeTruthy();
    expect(dropdown.getByRole("button", { name: /^This page's session gavel-app\.example\.test-d4e5f6/ })).toBeTruthy();
    const others = within(dropdown.getByLabelText("Active sessions"));
    expect(others.getAllByRole("button").map(button => button.textContent)).toEqual([expect.stringContaining("Fix layout")]);
  });

  it.each([
    { name: "the default", button: /^Default/, expected: undefined },
    { name: "the global session", button: /^Global session/, expected: { kind: "session", sessionId: GLOBAL.id } },
    { name: "this page's session", button: /^This page's session/, expected: { kind: "session", sessionId: PAGE.id } },
    { name: "another active session", button: /^Fix layout/, expected: { kind: "session", sessionId: OTHER.id } },
  ] satisfies { name: string; button: RegExp; expected: DraftTarget | undefined }[])("applies $name to this draft only", async ({ button, expected }) => {
    const onChange = menu();
    fireEvent.click((await openMenu()).getByRole("button", { name: button }));
    expect(onChange).toHaveBeenCalledOnce();
    expect(onChange).toHaveBeenCalledWith(expected);
    expect(screen.queryByRole("dialog", { name: "Run this comment on" })).toBeNull();
  });

  it("names the chosen session on the trigger and marks it in the menu", async () => {
    menu({ target: { kind: "session", sessionId: GLOBAL.id } });
    expect(screen.getByRole("button", { name: "Runs on gavel-global-a1b2c3" })).toBeTruthy();
    const dropdown = await openMenu();
    expect(dropdown.getByRole("button", { name: /^Global session/ }).getAttribute("aria-pressed")).toBe("true");
    expect(dropdown.getByRole("button", { name: /^Default/ }).getAttribute("aria-pressed")).toBe("false");
  });

  it("marks the default while nothing is chosen", async () => {
    menu();
    expect(screen.getByRole("button", { name: "Runs on" })).toBeTruthy();
    expect((await openMenu()).getByRole("button", { name: /^Default/ }).getAttribute("aria-pressed")).toBe("true");
  });

  it("disables the global and page entries that have no session yet", async () => {
    menu({ sessionIds: {} });
    const dropdown = await openMenu();
    for (const name of [/^Global session none started/, /^This page's session none started/]) expect(dropdown.getByRole<HTMLButtonElement>("button", { name }).disabled).toBe(true);
  });

  it("names a default session gavel no longer lists by its last characters", async () => {
    menu({ sessions: [OTHER], sessionIds: { page: "page-gone-123456" } });
    expect((await openMenu()).getByRole("button", { name: /^This page's session Session 123456/ })).toBeTruthy();
  });

  it("keeps the @mentions, but not those that are one of the sessions already listed", async () => {
    const onChange = menu({ targets: [{ id: "agent:reviewer", name: "reviewer" }, { id: `session:${OTHER.id}`, name: "Fix layout" }] as never });
    const dropdown = await openMenu();
    expect(within(dropdown.getByLabelText("Gavel targets")).getAllByRole("button").map(button => button.textContent)).toEqual(["@reviewer"]);
    fireEvent.click(dropdown.getByRole("button", { name: "@reviewer" }));
    expect(onChange).toHaveBeenCalledWith({ kind: "mention", mention: { id: "agent:reviewer", name: "reviewer" } });
  });
});
