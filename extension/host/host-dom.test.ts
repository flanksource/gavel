import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { mountHost, type MountHostOptions } from "./host-dom";

const mount = (options: Partial<MountHostOptions> = {}) => mountHost({ frameUrl: "about:blank", focus: vi.fn(), onWidth: vi.fn(), onPresentation: vi.fn(), ...options });
const nextFrame = () => new Promise<void>(resolve => requestAnimationFrame(() => resolve()));

beforeEach(() =>
  vi.stubGlobal(
    "ResizeObserver",
    class {
      observe() {}
      disconnect() {}
    },
  ),
);
afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

it("starts minimized, opens compact actions, and returns to compact after posting", () => {
  const presentation = vi.fn();
  const ui = mount({ onPresentation: presentation });
  const frame = ui.frame;
  const fab = ui.root.shadowRoot!.querySelector<HTMLButtonElement>(".fab")!;
  expect([frame.hidden, fab.hidden]).toEqual([true, false]);
  ui.minimize();
  expect([frame.hidden, fab.hidden]).toEqual([true, false]);
  fab.click();
  expect([frame.hidden, fab.hidden]).toEqual([false, true]);
  ui.draft(true, "button");
  ui.minimize();
  fab.click();
  expect(ui.root.getAttribute("data-mode")).toBe("draft");
  expect(ui.frame).toBe(frame);
  expect(presentation).toHaveBeenLastCalledWith("draft");
  ui.draft(false);
  expect(ui.root.getAttribute("data-mode")).toBe("compact");
  expect(frame.hidden).toBe(false);
  ui.close();
  expect(document.getElementById("gavel-grabber")).toBeNull();
});

it.each([
  { name: "a minimized rail", setup: (ui: ReturnType<typeof mountHost>) => ui.minimize(), hidden: true },
  { name: "a draft opened while minimized", setup: (ui: ReturnType<typeof mountHost>) => { ui.minimize(); ui.draft(true, "button"); }, hidden: true },
  { name: "a rail hidden while picking", setup: (ui: ReturnType<typeof mountHost>) => { ui.showRail(); ui.frame.hidden = true; }, hidden: false },
])("keeps the sidebar's own visibility when the page changes from $name", ({ setup, hidden }) => {
  const ui = mount();
  setup(ui);
  ui.pageChanged();
  const fab = ui.root.shadowRoot!.querySelector<HTMLButtonElement>(".fab")!;
  expect({ draft: ui.root.getAttribute("data-mode") === "draft", frameHidden: ui.frame.hidden, fabHidden: fab.hidden }).toEqual({ draft: false, frameHidden: hidden, fabHidden: !hidden });
  ui.close();
});

it("shows the gavel icon on the FAB instead of a plus", () => {
  const ui = mount();
  const fab = ui.root.shadowRoot!.querySelector<HTMLButtonElement>(".fab")!;
  expect([fab.textContent, fab.querySelector("svg")?.getAttribute("viewBox")]).toEqual(["", "0 0 256 256"]);
  ui.close();
});

it("reports each change between a closed and an open sidebar, and its current state", () => {
  const changed = vi.fn();
  const ui = mount({ onOpenChange: changed });
  expect(ui.minimized()).toBe(true);
  ui.minimize();
  expect(changed).not.toHaveBeenCalled();
  ui.showRail();
  expect([ui.minimized(), changed.mock.calls]).toEqual([false, [[true]]]);
  ui.showRail();
  expect(changed).toHaveBeenCalledTimes(1);
  ui.minimize();
  expect([ui.minimized(), changed.mock.calls]).toEqual([true, [[true], [false]]]);
  ui.close();
});

it("switches an open panel between rail and hover, and the FAB reopens it in the last mode", () => {
  const presentation = vi.fn();
  const ui = mount({ onPresentation: presentation });
  const fab = ui.root.shadowRoot!.querySelector<HTMLButtonElement>(".fab")!;
  fab.click();
  ui.setPresentation("rail");
  ui.minimize();
  fab.click();
  expect([ui.root.getAttribute("data-mode"), ui.frame.hidden, presentation.mock.lastCall]).toEqual(["rail", false, ["rail"]]);
  ui.setPresentation("compact");
  ui.minimize();
  ui.open();
  expect([ui.root.getAttribute("data-mode"), ui.frame.hidden, presentation.mock.lastCall]).toEqual(["compact", false, ["compact"]]);
  ui.draft(true, "button");
  expect(() => ui.setPresentation("rail")).toThrow("floating comment");
  ui.close();
});

it("sizes the hover popover from the stored size and shows its grip only while hovering", async () => {
  Object.defineProperty(window, "innerWidth", { configurable: true, value: 1920 });
  Object.defineProperty(window, "innerHeight", { configurable: true, value: 1080 });
  const ui = mount();
  const grip = ui.root.shadowRoot!.querySelector<HTMLElement>(".resize-corner")!;
  ui.hoverLayout({ width: 520, height: 600 });
  ui.open();
  await nextFrame();
  expect([ui.frame.style.width, ui.frame.style.height, grip.hidden]).toEqual(["520px", "600px", false]);
  ui.setPresentation("rail");
  await nextFrame();
  expect(grip.hidden).toBe(true);
  ui.setPresentation("compact");
  ui.minimize();
  await nextFrame();
  expect(grip.hidden).toBe(true);
  ui.close();
});

it("leaves Command+K available to the host document", () => {
  const ui = mount();
  const event = new KeyboardEvent("keydown", {
    key: "k",
    metaKey: true,
    cancelable: true,
    bubbles: true,
  });
  document.dispatchEvent(event);
  expect(event.defaultPrevented).toBe(false);
  ui.close();
});
