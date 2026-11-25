import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { EnableCard, type EnableCardProps } from "./EnableCard";
import type { AgentPickerSession } from "./AgentPicker";
import type { RuntimeCatalog } from "./runtime-families";

afterEach(cleanup);

const HOST = "app.example.test:3000";
const REVIEW: AgentPickerSession = { id: "session-review", title: "Review changes", source: "codex", model: "example-model", status: "running" };
const LAYOUT: AgentPickerSession = { id: "session-layout", title: "Fix layout", source: "claude", status: "idle" };
const ENDED: AgentPickerSession = { id: "session-ended", title: "Ended work", source: "claude", status: "ended", resumable: true };
const CATALOG: RuntimeCatalog = {
  models: [],
  families: [{ id: "claude", label: "Claude", provider: "anthropic", modes: [{ id: "agent", label: "Agent", defaultModel: "example-agent", provider: "anthropic" }] }],
};
const RUNTIME = { spec: { model: "example-agent" } };
const FOOTNOTE = "This is only the default. Each comment can still pick another session from its Runs on menu.";

function card(overrides: Partial<EnableCardProps> = {}): EnableCardProps {
  return {
    host: HOST,
    enabled: { value: true, onChange: vi.fn() },
    target: { value: "page", onChange: vi.fn() },
    runtime: { value: RUNTIME, catalog: CATALOG, onChange: vi.fn() },
    sessions: { items: [REVIEW, LAYOUT], state: "ready" },
    onStart: vi.fn().mockResolvedValue(undefined),
    onUse: vi.fn().mockResolvedValue(undefined),
    ...overrides,
  };
}

const withTarget = (value: EnableCardProps["target"]["value"]): Partial<EnableCardProps> => ({ target: { value, onChange: vi.fn() } });
const nameField = () => screen.getByLabelText<HTMLInputElement>(/New session name/);
const button = (name: string) => screen.getByRole<HTMLButtonElement>("button", { name });
const ROUTES_NAME = "Comments from this site go to";
const routes = () => screen.getByRole("group", { name: ROUTES_NAME });
const route = (name: string) => within(routes()).getByRole<HTMLInputElement>("radio", { name });

describe("EnableCard header", () => {
  it("names the host, says On, and reports the switch under its own name", () => {
    const props = card();
    render(<EnableCard {...props} />);
    expect(screen.getByText(HOST).getAttribute("title")).toBe(HOST);
    expect(screen.getByText("On")).toBeTruthy();
    const toggle = screen.getByRole("switch", { name: `Gavel on ${HOST}` });
    expect(toggle.getAttribute("aria-checked")).toBe("true");
    fireEvent.click(toggle);
    expect(props.enabled.onChange).toHaveBeenCalledWith(false);
  });

  it("says Off when disabled, and explains a switch that cannot be used in its tooltip", () => {
    const reason = "Sign in to the dashboard first";
    render(<EnableCard {...card({ enabled: { value: false, onChange: vi.fn(), disabled: true, disabledReason: reason } })} />);
    expect(screen.getByText("Off")).toBeTruthy();
    expect((screen.getByRole("switch", { name: `Gavel on ${HOST}` }) as HTMLButtonElement).disabled).toBe(true);
    expect(screen.getByTitle(reason)).toBeTruthy();
  });

  it("renders the actions beside the switch", () => {
    render(<EnableCard {...card({ actions: <button type="button">Settings</button> })} />);
    expect(button("Settings")).toBeTruthy();
  });
});

describe("EnableCard when off", () => {
  const off = (target: EnableCardProps["target"]["value"]) =>
    card({ ...withTarget(target), enabled: { value: false, onChange: vi.fn() }, context: <p>Project row</p>, footer: <p>Sync status</p> });

  it("explains how to turn it on instead of showing routes or context", () => {
    render(<EnableCard {...off("page")} />);
    expect(screen.getByText("Gavel is off on this site")).toBeTruthy();
    expect(screen.getByText("Turn it on to pick elements on the page and send comments to an agent.")).toBeTruthy();
    expect(screen.queryByRole("group", { name: ROUTES_NAME })).toBeNull();
    expect(screen.queryAllByRole("radio")).toEqual([]);
    expect(screen.queryByText("Project row")).toBeNull();
    expect(screen.queryByText(FOOTNOTE)).toBeNull();
    expect(screen.getByText("Sync status")).toBeTruthy();
  });

  it.each([
    ["page", "When on, comments go to this site's agent."],
    ["global", "When on, comments go to the global session."],
    ["auto", "When on, comments go to a session Gavel starts for you."],
  ] as const)("summarises the stored %s target", (target, summary) => {
    render(<EnableCard {...off(target)} />);
    expect(screen.getByText(summary)).toBeTruthy();
  });
});

describe("EnableCard setup", () => {
  it("replaces the context, routes and footnote while the site cannot route comments yet", () => {
    render(<EnableCard {...card({ setup: <p>Choose a dashboard</p>, context: <p>Project row</p>, footer: <p>Sync status</p>, error: "Cannot reach gavel" })} />);
    expect(screen.getByText("Choose a dashboard")).toBeTruthy();
    expect(screen.queryByRole("group", { name: ROUTES_NAME })).toBeNull();
    expect(screen.queryAllByRole("radio")).toEqual([]);
    expect(screen.queryByText("Project row")).toBeNull();
    expect(screen.queryByText(FOOTNOTE)).toBeNull();
    expect(screen.getByText("Sync status")).toBeTruthy();
    expect(screen.getByRole("alert").textContent).toBe("Cannot reach gavel");
  });
});

describe("EnableCard routes", () => {
  it("offers the three routes as native radios sharing one name, in a fieldset named by its legend, with the footnote, context and footer around it", () => {
    render(<EnableCard {...card({ context: <p>Project row</p>, footer: <p>Sync status</p> })} />);
    expect(routes().tagName).toBe("FIELDSET");
    expect(routes().querySelector("legend")?.textContent).toBe(ROUTES_NAME);
    const radios = within(routes()).getAllByRole<HTMLInputElement>("radio");
    expect(radios.map(radio => radio.closest("label")?.textContent)).toEqual(["This site's agent", "Global session", "Automatic"]);
    expect(new Set(radios.map(radio => radio.name)).size).toBe(1);
    expect(radios[0]?.name).not.toBe("");
    expect(screen.getByText("Project row")).toBeTruthy();
    expect(screen.getByText(FOOTNOTE)).toBeTruthy();
    expect(screen.getByText("Sync status")).toBeTruthy();
  });

  it.each([
    ["page", "This site's agent"],
    ["global", "Global session"],
    ["auto", "Automatic"],
  ] as const)("marks only the %s route as selected", (target, name) => {
    render(<EnableCard {...card(withTarget(target))} />);
    const checked = within(routes()).getAllByRole<HTMLInputElement>("radio").filter(radio => radio.checked);
    expect(checked.map(radio => radio.closest("label")?.textContent)).toEqual([name]);
  });

  it("reports a chosen route at once", () => {
    const props = card();
    render(<EnableCard {...props} />);
    fireEvent.click(route("Global session"));
    fireEvent.click(route("Automatic"));
    expect((props.target.onChange as ReturnType<typeof vi.fn>).mock.calls).toEqual([["global"], ["auto"]]);
  });

  it("disables the routes while a session is being started", async () => {
    const onStart = vi.fn().mockReturnValue(new Promise<void>(() => {}));
    render(<EnableCard {...card({ onStart })} />);
    fireEvent.click(button("Start session"));
    await screen.findByRole("status");
    expect(route("Global session").disabled).toBe(true);
    expect(route("Automatic").disabled).toBe(true);
  });

  it("shows each unselected session route's session and model, or that it has not started", () => {
    render(<EnableCard {...card({ ...withTarget("auto"), sessions: { page: REVIEW.id, items: [REVIEW], state: "ready" } })} />);
    const pageRow = route("This site's agent").closest("[data-route]") as HTMLElement;
    const globalRow = route("Global session").closest("[data-route]") as HTMLElement;
    expect(within(pageRow).getByText("Review changes")).toBeTruthy();
    expect(within(pageRow).getByText("example-model")).toBeTruthy();
    expect(within(pageRow).queryByRole("button")).toBeNull();
    expect(within(globalRow).getByText("Not started")).toBeTruthy();
  });

  it("names a remembered session Gavel no longer lists, without an identity", () => {
    render(<EnableCard {...card({ ...withTarget("global"), sessions: { global: "session-gone-123456", items: [REVIEW], state: "ready" } })} />);
    expect(screen.getByText("Session 123456 (not live)")).toBeTruthy();
    expect(screen.queryByText("example-model")).toBeNull();
    expect(screen.queryByLabelText(/New session name/)).toBeNull();
  });
});

describe("EnableCard selected session", () => {
  it("shows the current session, its model and status, with New session and Switch instead of the form", () => {
    render(<EnableCard {...card({ sessions: { page: REVIEW.id, items: [REVIEW, LAYOUT], state: "ready" } })} />);
    expect(screen.getByText("Review changes")).toBeTruthy();
    expect(screen.getByText("example-model")).toBeTruthy();
    expect(screen.getByText("running")).toBeTruthy();
    expect(button("New session")).toBeTruthy();
    expect(button("Switch")).toBeTruthy();
    expect(screen.queryByLabelText(/New session name/)).toBeNull();
    expect(screen.queryByRole("button", { name: "Clear" })).toBeNull();
  });

  it("says a finished session resumes on the next comment", () => {
    render(<EnableCard {...card({ sessions: { page: ENDED.id, items: [ENDED], state: "ready" } })} />);
    expect(screen.getByText("Resumes on the next comment")).toBeTruthy();
    expect(screen.queryByText("ended")).toBeNull();
  });

  it("shows the global session, not the page's, when global is selected", () => {
    render(<EnableCard {...card({ ...withTarget("global"), sessions: { page: REVIEW.id, global: LAYOUT.id, items: [REVIEW, LAYOUT], state: "ready" } })} />);
    const pageRow = route("This site's agent").closest("[data-route]") as HTMLElement;
    expect(within(pageRow).getByText("Review changes")).toBeTruthy();
    expect(screen.getAllByText("Fix layout")).toHaveLength(1);
    expect(button("Switch")).toBeTruthy();
  });

  it("switches to another running session without offering the current one", async () => {
    const props = card({ sessions: { page: REVIEW.id, items: [REVIEW, LAYOUT], state: "ready" } });
    render(<EnableCard {...props} />);
    fireEvent.click(button("Switch"));
    expect(screen.getByRole("menu", { name: "Switch session" })).toBeTruthy();
    expect(screen.queryByRole("menuitem", { name: /Review changes/ })).toBeNull();
    fireEvent.click(screen.getByRole("menuitem", { name: /Fix layout/ }));
    await waitFor(() => expect(props.onUse).toHaveBeenCalledWith({ target: "page", session: LAYOUT }));
  });

  it("says when no other session is running", () => {
    render(<EnableCard {...card({ sessions: { page: REVIEW.id, items: [REVIEW], state: "ready" } })} />);
    fireEvent.click(button("Switch"));
    expect(screen.getByText("No active sessions.")).toBeTruthy();
  });

  it("says while the active sessions are loading", () => {
    render(<EnableCard {...card({ sessions: { page: REVIEW.id, items: [], state: "loading" } })} />);
    fireEvent.click(button("Switch"));
    expect(screen.getByText("Loading active sessions…")).toBeTruthy();
  });

  it("starts a replacement from New session, keeps the old session until it starts, and cancels back", async () => {
    const props = card({ sessions: { page: REVIEW.id, items: [REVIEW], state: "ready" } });
    render(<EnableCard {...props} />);
    fireEvent.click(button("New session"));
    expect(screen.getByText("Until it starts, comments keep going to")).toBeTruthy();
    expect(screen.getByText("Review changes")).toBeTruthy();
    fireEvent.click(button("Cancel"));
    expect(screen.queryByLabelText(/New session name/)).toBeNull();
    fireEvent.click(button("New session"));
    fireEvent.change(nameField(), { target: { value: "second-run" } });
    fireEvent.click(button("Start session"));
    await waitFor(() => expect(props.onStart).toHaveBeenCalledWith({ target: "page", prefix: "second-run", runtime: RUNTIME }));
    await waitFor(() => expect(screen.queryByLabelText(/New session name/)).toBeNull());
  });
});

describe("EnableCard new session form", () => {
  it.each([
    ["page", `gavel-${HOST}`],
    ["global", "gavel-global"],
  ] as const)("prefills the %s session name and says Gavel adds a suffix", (target, expected) => {
    render(<EnableCard {...card(withTarget(target))} />);
    expect(nameField().value).toBe(expected);
    expect(screen.getByText("Gavel adds a 6-character suffix.")).toBeTruthy();
    expect(screen.getByLabelText("Agent runtime")).toBeTruthy();
    expect(screen.queryByText("Until it starts, comments keep going to")).toBeNull();
    expect(screen.queryByRole("button", { name: "Cancel" })).toBeNull();
  });

  it.each([["page", `gavel-${HOST}`], ["global", "gavel-global"]] as const)("starts the %s session with its prefilled name, and the chosen model", async (target, expected) => {
    const props = card(withTarget(target));
    render(<EnableCard {...props} />);
    fireEvent.click(button("Start session"));
    await waitFor(() => expect(props.onStart).toHaveBeenCalledWith({ target, prefix: expected, runtime: RUNTIME }));
  });

  it("starts with the trimmed edited name, and disables Start when the name is blank", async () => {
    const props = card();
    render(<EnableCard {...props} />);
    fireEvent.change(nameField(), { target: { value: "   " } });
    expect(button("Start session").disabled).toBe(true);
    fireEvent.change(nameField(), { target: { value: " checkout-review " } });
    expect(button("Start session").disabled).toBe(false);
    fireEvent.click(button("Start session"));
    await waitFor(() => expect(props.onStart).toHaveBeenCalledWith({ target: "page", prefix: "checkout-review", runtime: RUNTIME }));
  });

  it("keeps a name edited for one target when switching to the other and back", () => {
    const props = card();
    const { rerender } = render(<EnableCard {...props} />);
    fireEvent.change(nameField(), { target: { value: "page-name" } });
    rerender(<EnableCard {...props} {...withTarget("global")} />);
    expect(nameField().value).toBe("gavel-global");
    rerender(<EnableCard {...props} />);
    expect(nameField().value).toBe("page-name");
  });

  it("disables Start and says why while a session cannot be started yet", () => {
    render(<EnableCard {...card({ startBlocked: "Choose a project first." })} />);
    expect(button("Start session").disabled).toBe(true);
    expect(screen.getByText("Choose a project first.")).toBeTruthy();
  });

  it("says models are loading until the catalog arrives", () => {
    render(<EnableCard {...card({ runtime: { value: undefined, catalog: undefined, onChange: vi.fn() } })} />);
    expect(screen.getByText("Loading models…")).toBeTruthy();
  });

  it("is busy while starting and shows the failure inline, then lets the user try again", async () => {
    let fail: (cause: Error) => void = () => {};
    const onStart = vi.fn().mockReturnValue(new Promise<void>((_resolve, reject) => { fail = reject; }));
    render(<EnableCard {...card({ onStart })} />);
    fireEvent.click(button("Start session"));
    expect((await screen.findByRole("status")).textContent).toBe("Starting session…");
    expect(button("Start session").disabled).toBe(true);
    fail(new Error("Gavel could not start the session: no capacity"));
    expect((await screen.findByRole("alert")).textContent).toBe("Gavel could not start the session: no capacity");
    expect(screen.queryByRole("status")).toBeNull();
    expect(button("Start session").disabled).toBe(false);
  });

  it.each(["page", "global"] as const)("uses a running session as the %s session", async target => {
    const props = card(withTarget(target));
    render(<EnableCard {...props} />);
    fireEvent.click(button("use a running session"));
    expect(screen.getByRole("menu", { name: "Use a running session" })).toBeTruthy();
    fireEvent.click(screen.getByRole("menuitem", { name: /Review changes/ }));
    await waitFor(() => expect(props.onUse).toHaveBeenCalledWith({ target, session: REVIEW }));
  });

  it("shows a failure to use a running session inline", async () => {
    const onUse = vi.fn().mockRejectedValue(new Error("The page's settings could not be written"));
    render(<EnableCard {...card({ onUse })} />);
    fireEvent.click(button("use a running session"));
    fireEvent.click(screen.getByRole("menuitem", { name: /Review changes/ }));
    expect((await screen.findByRole("alert")).textContent).toBe("The page's settings could not be written");
  });
});

describe("EnableCard automatic route", () => {
  const grouping = (value: "page" | "shared") => ({ value, onChange: vi.fn() });

  it.each([
    [grouping("page"), "Gavel starts a session for each page"],
    [grouping("shared"), "Gavel starts one session for all pages"],
    [undefined, "Gavel starts sessions for you"],
  ])("describes the unselected route from how sessions are grouped (%#)", (value, text) => {
    render(<EnableCard {...card({ grouping: value })} />);
    expect(screen.getByText(text)).toBeTruthy();
  });

  it("explains the selected route and lets the grouping be changed", () => {
    const props = card({ ...withTarget("auto"), grouping: grouping("page") });
    render(<EnableCard {...props} />);
    expect(screen.getByText("Gavel starts sessions for you, on the default model.")).toBeTruthy();
    const choice = screen.getByRole("radiogroup", { name: "One session for" });
    expect(within(choice).getByRole("radio", { name: "Each page" }).getAttribute("aria-checked")).toBe("true");
    fireEvent.click(within(choice).getByRole("radio", { name: "All pages" }));
    expect(props.grouping?.onChange).toHaveBeenCalledWith("shared");
    expect(screen.queryByLabelText(/New session name/)).toBeNull();
  });

  it("renders the host's menu beside the grouping control in the selected route only", () => {
    const withMenu = { ...grouping("page"), menu: <button type="button">This site</button> };
    const { rerender } = render(<EnableCard {...card({ ...withTarget("auto"), grouping: withMenu })} />);
    const choice = screen.getByRole("radiogroup", { name: "One session for" });
    expect(choice.parentElement?.contains(button("This site"))).toBe(true);
    rerender(<EnableCard {...card({ ...withTarget("page"), grouping: withMenu })} />);
    expect(screen.queryByRole("button", { name: "This site" })).toBeNull();
  });

  it("offers the grouping control inside the routes group, which is not itself a radio group", () => {
    render(<EnableCard {...card({ ...withTarget("auto"), grouping: grouping("page") })} />);
    expect(within(routes()).getByRole("radiogroup", { name: "One session for" })).toBeTruthy();
    expect(screen.queryByRole("radiogroup", { name: ROUTES_NAME })).toBeNull();
  });

  it("offers no grouping control where it cannot be changed", () => {
    render(<EnableCard {...card(withTarget("auto"))} />);
    expect(screen.getByText("Gavel starts sessions for you, on the default model.")).toBeTruthy();
    expect(screen.queryByRole("radiogroup", { name: "One session for" })).toBeNull();
  });
});

describe("EnableCard messages", () => {
  it("shows the host's error and status", () => {
    render(<EnableCard {...card({ error: "Cannot reach gavel", status: "Saved." })} />);
    expect(screen.getByRole("alert").textContent).toBe("Cannot reach gavel");
    expect(screen.getByRole("status").textContent).toBe("Saved.");
  });
});
