import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { openTodoModal } from "./host-todo-modal";

const extension = "chrome-extension://abcdefghijklmnopabcdefghijklmnop";
const relay = `${extension}/todo-relay.html`;
const target = "http://localhost:9092/todos/new?embed=1&title=Card";
const open = (overrides: Partial<Parameters<typeof openTodoModal>[0]> = {}) => openTodoModal({ relay, target, parentOrigin: "https://app.example", ...overrides });
const modal = () => document.getElementById("gavel-grabber-todo");
const frame = () => modal()!.shadowRoot!.querySelector("iframe")!;
const closeButton = () => modal()!.shadowRoot!.querySelector<HTMLButtonElement>("button[aria-label=Close]")!;

/** jsdom gives no browsing context to an iframe inside a shadow tree, so the framed window is a stand-in with a recorded postMessage. */
function framedWindow() {
  const framed = { postMessage: vi.fn() } as unknown as Window;
  Object.defineProperty(frame(), "contentWindow", { get: () => framed });
  return framed;
}

function fromForm(framed: Window, data: unknown, overrides: Partial<MessageEventInit> = {}) {
  window.dispatchEvent(new MessageEvent("message", { data, origin: extension, source: framed, ...overrides }));
}

beforeEach(() => document.body.replaceChildren());
afterEach(() => {
  modal()?.remove();
  vi.restoreAllMocks();
});

describe("openTodoModal", () => {
  it("frames the extension's todo relay with the form and the page that may receive its messages", () => {
    open();
    const url = new URL(frame().src);
    expect([url.href.split("?")[0], url.searchParams.get("target"), url.searchParams.get("parentOrigin")]).toEqual([relay, target, "https://app.example"]);
    const root = modal()!;
    expect([root.hasAttribute("data-react-grab-ignore"), root.shadowRoot!.querySelector("[role=dialog]")?.getAttribute("aria-modal")]).toEqual([true, "true"]);
  });

  it.each(["todo-saved", "cancel"])("closes when the form reports %s", (type) => {
    const onClose = vi.fn();
    open({ onClose });
    fromForm(framedWindow(), { source: "gavel-react-grab", type, ref: "T-1" });
    expect(modal()).toBeNull();
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("closes from the close button and on Escape, once", () => {
    const onClose = vi.fn();
    open({ onClose });
    closeButton().click();
    expect(modal()).toBeNull();
    open({ onClose });
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    expect(modal()).toBeNull();
    expect(onClose).toHaveBeenCalledTimes(2);
  });

  it("sends the screenshot to the form when it announces it is ready, to the relay's origin only", () => {
    const blob = new Blob(["png"], { type: "image/png" });
    open({ attachment: { blob, name: "screenshot.png" } });
    const framed = framedWindow();
    fromForm(framed, { source: "gavel-react-grab", type: "embed-ready" });
    expect(framed.postMessage).toHaveBeenCalledWith({ source: "gavel-react-grab", type: "attachment", blob, name: "screenshot.png" }, extension);
  });

  it("sends nothing when there is no screenshot", () => {
    open();
    const framed = framedWindow();
    fromForm(framed, { source: "gavel-react-grab", type: "embed-ready" });
    expect(framed.postMessage).not.toHaveBeenCalled();
    expect(modal()).not.toBeNull();
  });

  it.each([
    { name: "another window", overrides: { source: window } },
    { name: "another origin", overrides: { origin: "https://evil.example" } },
  ])("ignores messages from $name, even a cancel", ({ overrides }) => {
    open();
    fromForm(framedWindow(), { source: "gavel-react-grab", type: "cancel" }, overrides);
    expect(modal()).not.toBeNull();
  });

  it("ignores messages that are not the form's and keeps one modal at a time", () => {
    open();
    const framed = framedWindow();
    fromForm(framed, { source: "something-else", type: "cancel" });
    fromForm(framed, "cancel");
    expect(modal()).not.toBeNull();
    open();
    expect(document.querySelectorAll("#gavel-grabber-todo")).toHaveLength(1);
  });
});
