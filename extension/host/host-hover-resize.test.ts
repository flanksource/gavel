// @vitest-environment jsdom
import { beforeEach, expect, it, vi } from "vitest";
import { hoverSize, mountHoverResize } from "./host-hover-resize";

beforeEach(() => {
  Object.defineProperty(window, "innerWidth", { configurable: true, value: 1920 });
  Object.defineProperty(window, "innerHeight", { configurable: true, value: 1080 });
});

it.each([
  [{ width: 279, height: 480 }],
  [{ width: 400, height: 239 }],
  [{ width: 721, height: 480 }],
  [{ width: "400", height: 480 }],
  [{ width: NaN, height: 480 }],
  [null],
])("rejects invalid hover size %j", input => {
  expect(() => hoverSize(input)).toThrow("Hover panel size");
});

it("grows up and left when the top-left corner is dragged, and reports the size once released", () => {
  const onSize = vi.fn();
  const sizing = mountHoverResize(onSize);
  Object.assign(sizing.handle, { setPointerCapture: vi.fn() });
  sizing.layout({ width: 400, height: 480 });
  sizing.handle.dispatchEvent(new MouseEvent("pointerdown", { button: 0, clientX: 500, clientY: 300 }));
  sizing.handle.dispatchEvent(new MouseEvent("pointermove", { clientX: 400, clientY: 250 }));
  expect(sizing.size()).toEqual({ width: 500, height: 530 });
  expect(onSize).not.toHaveBeenCalled();
  sizing.handle.dispatchEvent(new MouseEvent("pointerup"));
  expect(onSize).toHaveBeenCalledExactlyOnceWith({ width: 500, height: 530 });
  sizing.close();
});

it("clamps to the viewport and the size limits", () => {
  Object.defineProperty(window, "innerHeight", { configurable: true, value: 600 });
  const sizing = mountHoverResize(vi.fn());
  // The tallest a 600px viewport allows: its height less the 8px margin on each side.
  sizing.layout({ width: 700, height: 584 });
  sizing.handle.dispatchEvent(new KeyboardEvent("keydown", { key: "ArrowUp", shiftKey: true }));
  sizing.handle.dispatchEvent(new KeyboardEvent("keydown", { key: "ArrowLeft", shiftKey: true }));
  expect(sizing.size()).toEqual({ width: 720, height: 584 });
  sizing.handle.dispatchEvent(new KeyboardEvent("keydown", { key: "Home" }));
  expect(sizing.size()).toEqual({ width: 280, height: 240 });
  sizing.close();
});
