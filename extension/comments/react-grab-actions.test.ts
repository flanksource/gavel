import { afterEach, describe, expect, it, vi } from "vitest";
import type { ContextMenuAction, ContextMenuActionContext, Plugin, ReactGrabAPI } from "@flanksource/clicky-ui/react-grab";
import { SIDEBAR_PLUGIN_NAME, grabberPlugin, registerGrabberPlugin, sidebarButtonPlugin, type GavelTodoActions, type GrabberPluginOptions } from "./react-grab-actions";
import { writeClipboard } from "./clipboard";
import { copyContextMarkdown } from "./copy-context";

vi.mock("./clipboard", () => ({ writeClipboard: vi.fn(() => Promise.resolve()) }));
vi.mock("./copy-context", () => ({ copyContextMarkdown: vi.fn(() => Promise.resolve("context markdown")) }));

afterEach(() => vi.clearAllMocks());

const element = document.createElement("button");
function context(overrides: Partial<ContextMenuActionContext> = {}) {
  return { element, elements: [element], hideContextMenu: vi.fn(), cleanup: vi.fn(), ...overrides } as unknown as ContextMenuActionContext;
}
function build(overrides: Partial<GrabberPluginOptions> = {}) {
  const onComment = vi.fn<GrabberPluginOptions["onComment"]>();
  const capture = vi.fn<GrabberPluginOptions["capture"]>(() => Promise.resolve({ status: "captured", dataUrl: "data:image/png;base64,AAAA" }));
  const options: GrabberPluginOptions = { enabled: () => true, onComment, captureSupported: () => true, capture, ...overrides };
  return { onComment, capture, plugin: grabberPlugin(options) };
}
const enabledFor = (plugin: Plugin, id: string) => {
  const enabled = action(plugin, id).enabled;
  return typeof enabled === "function" ? enabled(context()) : enabled;
};
const action = (plugin: Plugin, id: string): ContextMenuAction => {
  const found = plugin.actions?.find((entry) => entry.id === id);
  if (!found) throw new Error(`no action ${id}`);
  return found;
};

describe("grabberPlugin", () => {
  it("offers Copy screenshot, Comment and Comment with screenshot, then the host's own actions", () => {
    const extra: ContextMenuAction = { id: "prompt", label: "Prompt…", onAction: vi.fn() };
    const { plugin } = build({ extraActions: [extra] });
    expect(plugin.actions?.map((entry) => [entry.id, entry.label])).toEqual([
      ["gavel-copy-screenshot", "Copy screenshot"],
      ["comment", "Comment"],
      ["gavel-comment-screenshot", "Comment with screenshot"],
      ["prompt", "Prompt…"],
    ]);
  });

  it("hands Comment and Comment with screenshot to onComment with the screenshot choice", () => {
    const { plugin, onComment } = build();
    const target = context();
    action(plugin, "comment").onAction?.(target);
    action(plugin, "gavel-comment-screenshot").onAction?.(target);
    expect(onComment.mock.calls).toEqual([[target, { screenshot: false }], [target, { screenshot: true }]]);
  });

  it("hides screenshot actions where capture is unsupported and comment actions where the element is not commentable", () => {
    const unsupported = build({ captureSupported: () => false }).plugin;
    expect(["gavel-copy-screenshot", "gavel-comment-screenshot", "comment"].map((id) => enabledFor(unsupported, id))).toEqual([false, false, true]);
    const outside = build({ enabled: () => false }).plugin;
    expect(["gavel-copy-screenshot", "gavel-comment-screenshot", "comment"].map((id) => enabledFor(outside, id))).toEqual([true, false, false]);
  });

  it("replaces a plain copy's content with the copy-context markdown", async () => {
    const { plugin } = build();
    await expect(plugin.hooks?.transformCopyContent?.("plain", [element])).resolves.toBe("context markdown");
    expect(copyContextMarkdown).toHaveBeenCalledWith([element]);
  });

  it("copies the element's PNG together with its context and closes the menu first", async () => {
    vi.stubGlobal("fetch", vi.fn(() => Promise.resolve({ blob: () => Promise.resolve(new Blob(["png"], { type: "image/png" })) })));
    const { plugin, capture } = build();
    const target = context();
    action(plugin, "gavel-copy-screenshot").onAction?.(target);
    expect(target.hideContextMenu).toHaveBeenCalledTimes(1);
    expect(target.cleanup).toHaveBeenCalledTimes(1);
    expect(capture).toHaveBeenCalledWith(element);
    const [text, image] = vi.mocked(writeClipboard).mock.calls[0]!;
    await expect(text).resolves.toBe("context markdown");
    await expect(image?.then((blob) => blob.type)).resolves.toBe("image/png");
    vi.unstubAllGlobals();
  });

  it("fails the clipboard image with the recorded reason when capture is unavailable", async () => {
    const { plugin } = build({ capture: () => Promise.resolve({ status: "unavailable", reason: "failed" }) });
    action(plugin, "gavel-copy-screenshot").onAction?.(context());
    const [, image] = vi.mocked(writeClipboard).mock.calls[0]!;
    await expect(image).rejects.toThrow("screenshot failed");
  });
});

describe("grabberPlugin with the gavel backend", () => {
  function gavelActions(): GavelTodoActions {
    return {
      addTodo: vi.fn(),
      screenshotTodo: vi.fn(),
      copyScreenshot: vi.fn(),
      copyContent: vi.fn(() => Promise.resolve("todo body")),
      onOpenFile: vi.fn(() => true),
      setup: vi.fn(),
    };
  }

  it("offers the todo actions first and a single Copy screenshot, then Screenshot only and the comment actions", () => {
    const { plugin } = build({ gavel: gavelActions() });
    expect(plugin.actions?.map((entry) => [entry.id, entry.label, entry.shortcut])).toEqual([
      ["gavel-todo", "Add to gavel todo", "T"],
      ["gavel-screenshot", "Screenshot to gavel todo", "S"],
      ["gavel-copy-screenshot", "Copy screenshot", undefined],
      ["gavel-screenshot-only", "Screenshot only", undefined],
      ["comment", "Comment", undefined],
      ["gavel-comment-screenshot", "Comment with screenshot", undefined],
    ]);
    const labels = plugin.actions?.map((entry) => entry.label) ?? [];
    expect(new Set(labels).size).toBe(labels.length);
  });

  it("runs the gavel behavior for each todo and screenshot action, and never the playground's clipboard copy", () => {
    const gavel = gavelActions();
    const { plugin, capture } = build({ gavel });
    const target = context();
    action(plugin, "gavel-todo").onAction?.(target);
    action(plugin, "gavel-screenshot").onAction?.(target);
    action(plugin, "gavel-copy-screenshot").onAction?.(target);
    action(plugin, "gavel-screenshot-only").onAction?.(target);
    expect(gavel.addTodo).toHaveBeenCalledWith(target);
    expect(gavel.screenshotTodo).toHaveBeenCalledWith(target);
    expect(vi.mocked(gavel.copyScreenshot).mock.calls).toEqual([[target, { body: true }], [target, { body: false }]]);
    expect(writeClipboard).not.toHaveBeenCalled();
    expect(capture).not.toHaveBeenCalled();
  });

  it("takes a plain copy's content, the open-file hook and the Command+G setup from the gavel behavior", async () => {
    const gavel = gavelActions();
    const { plugin } = build({ gavel });
    await expect(plugin.hooks?.transformCopyContent?.("plain", [element])).resolves.toBe("todo body");
    expect(copyContextMarkdown).not.toHaveBeenCalled();
    expect(plugin.hooks?.onOpenFile).toBe(gavel.onOpenFile);
    expect(plugin.setup).toBe(gavel.setup);
  });

  it("hides the screenshot actions where capture is unsupported but keeps Add to gavel todo", () => {
    const { plugin } = build({ gavel: gavelActions(), captureSupported: () => false });
    expect(["gavel-todo", "gavel-screenshot", "gavel-copy-screenshot", "gavel-screenshot-only"].map((id) => enabledFor(plugin, id))).toEqual([undefined, false, false, false]);
  });

  it("registers no setup or open-file hook without the gavel backend", () => {
    const { plugin } = build();
    expect([plugin.setup, plugin.hooks?.onOpenFile]).toEqual([undefined, undefined]);
  });
});

describe("sidebarButtonPlugin", () => {
  it.each([
    [false, "Open comments sidebar"],
    [true, "Close comments sidebar"],
  ])("with the sidebar open=%s offers one toolbar button labelled %s that toggles it", (open, label) => {
    const toggle = vi.fn();
    const plugin = sidebarButtonPlugin({ icon: "<svg></svg>", open, toggle });
    expect(plugin.name).toBe(SIDEBAR_PLUGIN_NAME);
    expect(plugin.toolbarButtons).toHaveLength(1);
    const [button] = plugin.toolbarButtons!;
    expect([button!.label, button!.icon, button!.pressed?.()]).toEqual([label, "<svg></svg>", open]);
    button!.onClick();
    expect(toggle).toHaveBeenCalledTimes(1);
  });
});

describe("registerGrabberPlugin", () => {
  function grabWith(options: { plugins: string[]; defaultAction: string | undefined }) {
    const calls: string[] = [];
    const grab = {
      getPlugins: () => options.plugins,
      getToolbarState: () => (options.defaultAction === undefined ? null : { defaultAction: options.defaultAction }),
      unregisterPlugin: (name: string) => calls.push(`unregister ${name}`),
      registerPlugin: (plugin: Plugin) => calls.push(`register ${plugin.name}`),
      setToolbarState: (state: { defaultAction?: string }) => calls.push(`default ${state.defaultAction}`),
    } as unknown as ReactGrabAPI;
    return { grab, calls };
  }
  const comment: Plugin = { name: "comment" };
  const plugin: Plugin = { name: "gavel-comments" };

  it("swaps out the built-in comment plugin and restores it and the default action", () => {
    const { grab, calls } = grabWith({ plugins: ["copy", "comment"], defaultAction: "copy" });
    const restore = registerGrabberPlugin({ grab, plugin, comment });
    expect(calls).toEqual(["unregister comment", "register gavel-comments", "default comment"]);
    calls.length = 0;
    restore();
    expect(calls).toEqual(["unregister gavel-comments", "register comment", "default copy"]);
  });

  it("leaves a grab without a comment plugin without one, and falls back to copy when it had no default action", () => {
    const { grab, calls } = grabWith({ plugins: ["copy"], defaultAction: undefined });
    registerGrabberPlugin({ grab, plugin, comment })();
    expect(calls).toEqual(["register gavel-comments", "default comment", "unregister gavel-comments", "default copy"]);
  });
});
