import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { ContextMenuActionContext, PluginConfig, ReactGrabAPI } from "@flanksource/clicky-ui/react-grab";
import { writeClipboard } from "../comments/clipboard";
import { buildBody, buildTodoUrl } from "../comments/gavel-todo";
import { gavelTodoActions, type GavelTodoOptions } from "./host-gavel";

vi.mock("../comments/clipboard", () => ({ writeClipboard: vi.fn(() => Promise.resolve()) }));

const server = "http://localhost:9092";
const todoRelay = "chrome-extension://abcdefghijklmnopabcdefghijklmnop/todo-relay.html";
const png = "data:image/png;base64,iVBORw0KGgo=";

function grabStub(overrides: Partial<ReactGrabAPI> = {}) {
  return {
    getDisplayName: () => "SaveButton",
    getSource: () => Promise.resolve({ filePath: "src/Save.tsx", lineNumber: 12, columnNumber: null, componentName: "SaveButton" }),
    getStackContext: () => Promise.resolve("stack"),
    isEnabled: () => true,
    isActive: () => false,
    activate: vi.fn(),
    deactivate: vi.fn(),
    ...overrides,
  } as unknown as ReactGrabAPI;
}

function build(overrides: Partial<GavelTodoOptions> = {}) {
  const openModal = vi.fn((_options: Parameters<GavelTodoOptions["openModal"]>[0]) => ({ close: () => {} }));
  const capture = vi.fn<GavelTodoOptions["capture"]>(() => Promise.resolve({ status: "captured", dataUrl: png }));
  const upload = vi.fn<GavelTodoOptions["upload"]>(() => Promise.resolve({ url: `${server}/api/todos/attachments/a.png`, filename: "screenshot.png" }));
  const grab = grabStub();
  const actions = gavelTodoActions({ grab, server, todoRelay, capture, upload, openModal, ...overrides });
  return { actions, grab, openModal, capture, upload };
}

let button: HTMLButtonElement;
function context(overrides: Partial<ContextMenuActionContext> = {}) {
  return { element: button, elements: [button], componentName: "SaveButton", hideContextMenu: vi.fn(), cleanup: vi.fn(), ...overrides } as unknown as ContextMenuActionContext;
}
const page = location.href;
const body = (source = "stack") => buildBody({ componentName: "SaveButton", pageUrl: page, selector: "body > button:nth-child(1)", source, html: button.outerHTML });
const settle = () => new Promise((resolve) => setTimeout(resolve, 0));

beforeEach(() => {
  document.body.innerHTML = '<button aria-label="Save draft">Save</button>';
  button = document.querySelector("button")!;
});
afterEach(() => {
  document.body.replaceChildren();
  vi.clearAllMocks();
  vi.restoreAllMocks();
});

describe("Add to gavel todo", () => {
  it("opens the todo form prefilled from the element, after closing the grab overlay", async () => {
    const { actions, openModal } = build();
    const target = context();
    actions.addTodo(target);
    expect(target.hideContextMenu).toHaveBeenCalledTimes(1);
    expect(target.cleanup).toHaveBeenCalledTimes(1);
    await settle();
    expect(openModal).toHaveBeenCalledTimes(1);
    expect(openModal.mock.calls[0]![0]).toEqual({
      relay: todoRelay,
      target: buildTodoUrl({ server, title: "SaveButton: Save draft", body: body(), sourceUrl: page }),
    });
  });
});

describe("Screenshot to gavel todo", () => {
  it("captures the element after closing the overlay and opens the form with the PNG attached", async () => {
    const { actions, openModal, capture } = build();
    const target = context();
    actions.screenshotTodo(target);
    expect(target.cleanup).toHaveBeenCalledTimes(1);
    expect(capture).toHaveBeenCalledWith(button);
    await settle();
    const opened = openModal.mock.calls[0]![0];
    expect(opened.attachment?.name).toBe("screenshot.png");
    expect(opened.attachment?.blob.type).toBe("image/png");
    expect(new URL(opened.target).searchParams.get("body")).toBe(body());
  });

  it("opens the form without an image, and says why, when the capture is unavailable", async () => {
    const warn = vi.spyOn(console, "warn").mockImplementation(() => {});
    const { actions, openModal } = build({ capture: () => Promise.resolve({ status: "unavailable", reason: "failed" }) });
    actions.screenshotTodo(context());
    await settle();
    expect(openModal.mock.calls[0]![0].attachment).toBeUndefined();
    expect(warn).toHaveBeenCalledWith(expect.stringContaining("[gavel]"), expect.stringContaining("failed"));
  });
});

describe("Copy screenshot and Screenshot only", () => {
  async function copied(run: (actions: ReturnType<typeof build>["actions"], target: ContextMenuActionContext) => void, overrides: Partial<GavelTodoOptions> = {}) {
    const { actions, upload, capture } = build(overrides);
    const target = context();
    run(actions, target);
    await settle();
    const text = vi.mocked(writeClipboard).mock.calls[0]?.[0];
    return { text: text ? await text : undefined, upload, capture, target };
  }

  it("copies the todo body with the uploaded image under Attachments, starting the clipboard write inside the gesture", async () => {
    const { text, upload, target } = await copied((actions, context) => actions.copyScreenshot(context, { body: true }));
    expect(text).toBe(`${body()}\n\n## Attachments\n\n![screenshot.png](${server}/api/todos/attachments/a.png)`);
    expect(upload).toHaveBeenCalledWith({ dataUrl: png, name: "screenshot.png" });
    expect(target.cleanup).toHaveBeenCalledTimes(1);
  });

  it("copies only the image URL for Screenshot only", async () => {
    const { text } = await copied((actions, context) => actions.copyScreenshot(context, { body: false }));
    expect(text).toBe(`${server}/api/todos/attachments/a.png`);
  });

  it.each([
    ["the capture is unavailable", { capture: () => Promise.resolve({ status: "unavailable", reason: "failed" } as const) }, /screenshot failed/],
    ["the upload fails", { upload: () => Promise.reject(new Error("attachment too large")) }, /attachment too large/],
  ])("copies nothing and logs the cause when %s", async (_name, overrides, message) => {
    const error = vi.spyOn(console, "error").mockImplementation(() => {});
    vi.mocked(writeClipboard).mockImplementationOnce((text) => text.then(() => undefined));
    const { actions } = build(overrides as Partial<GavelTodoOptions>);
    actions.copyScreenshot(context(), { body: true });
    await settle();
    expect(error).toHaveBeenCalledWith(expect.stringContaining("[gavel]"), expect.objectContaining({ message: expect.stringMatching(message) }));
  });
});

describe("copy content", () => {
  it("replaces a plain copy with the todo body of each element, or keeps the content when there is none", async () => {
    const { actions } = build();
    await expect(actions.copyContent("plain", [button])).resolves.toBe(body());
    await expect(actions.copyContent("plain", [])).resolves.toBe("plain");
  });
});

describe("open file", () => {
  it("claims /@fs/ paths and leaves ordinary paths to react-grab", () => {
    vi.stubGlobal("fetch", vi.fn(() => Promise.resolve({ ok: true })));
    const { actions } = build();
    expect(actions.onOpenFile("/src/a.tsx", 2)).toBe(false);
    expect(actions.onOpenFile("/@fs/Users/me/a.tsx", 2)).toBe(true);
    expect(fetch).toHaveBeenCalledWith("/__open-in-editor?file=%2FUsers%2Fme%2Fa.tsx&column=1&line=2");
    vi.unstubAllGlobals();
  });
});

describe("Command+G one-shot", () => {
  function setup(grabOverrides: Partial<ReactGrabAPI> = {}) {
    const grab = grabStub(grabOverrides);
    const { actions, openModal } = build({ grab });
    const config = actions.setup(grab, {} as never) as PluginConfig;
    return { grab, config, openModal };
  }
  const press = (init: KeyboardEventInit, target: EventTarget = document.body) => {
    const event = new KeyboardEvent("keydown", { key: "g", metaKey: true, bubbles: true, cancelable: true, ...init });
    target.dispatchEvent(event);
    return event;
  };

  it("arms grabbing, then sends the next selection to Add to gavel todo and swallows it", async () => {
    const { grab, config, openModal } = setup();
    expect(press({}).defaultPrevented).toBe(true);
    expect(grab.activate).toHaveBeenCalledTimes(1);
    expect(config.hooks?.onElementSelect?.(button)).toBe(true);
    await settle();
    expect(grab.deactivate).toHaveBeenCalledTimes(1);
    expect(openModal).toHaveBeenCalledTimes(1);
    expect(new URL(openModal.mock.calls[0]![0].target).searchParams.get("title")).toBe("SaveButton: Save draft");
    expect(config.hooks?.onElementSelect?.(button)).toBeUndefined();
    config.cleanup?.();
  });

  it("disarms on a second press, and on deactivation", () => {
    const { grab, config } = setup();
    press({});
    press({});
    expect(grab.deactivate).toHaveBeenCalledTimes(1);
    expect(config.hooks?.onElementSelect?.(button)).toBeUndefined();
    press({});
    config.hooks?.onDeactivate?.();
    expect(config.hooks?.onElementSelect?.(button)).toBeUndefined();
    config.cleanup?.();
  });

  it.each([
    ["without Command", { metaKey: false }],
    ["with Control", { ctrlKey: true }],
    ["with Shift", { shiftKey: true }],
    ["held down", { repeat: true }],
    ["another key", { key: "k" }],
  ])("ignores the key %s", (_name, init) => {
    const { grab, config } = setup();
    press(init);
    expect(grab.activate).not.toHaveBeenCalled();
    config.cleanup?.();
  });

  it("ignores the key while typing, while the grabber is disabled, and while the todo form is open", () => {
    const { grab, config } = setup({ isEnabled: () => false } as Partial<ReactGrabAPI>);
    press({});
    expect(grab.activate).not.toHaveBeenCalled();
    config.cleanup?.();

    const enabled = setup();
    const input = document.body.appendChild(document.createElement("input"));
    press({}, input);
    expect(enabled.grab.activate).not.toHaveBeenCalled();
    const modal = document.body.appendChild(document.createElement("div"));
    modal.id = "gavel-grabber-todo";
    press({});
    expect(enabled.grab.activate).not.toHaveBeenCalled();
    enabled.config.cleanup?.();
  });

  it("stops listening and ends an armed grab when the plugin is cleaned up", () => {
    const { grab, config } = setup();
    press({});
    config.cleanup?.();
    expect(grab.deactivate).toHaveBeenCalledTimes(1);
    press({});
    expect(grab.activate).toHaveBeenCalledTimes(1);
  });
});
