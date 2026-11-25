// @vitest-environment jsdom
import { expect, it, vi } from "vitest";
import { mountSidebarResize, sidebarWidth } from "./host-resize";

it.each([NaN, Infinity, "400", null, 279, 641])("rejects invalid width %s", input => {
  expect(() => sidebarWidth(input)).toThrow("Sidebar width");
});

it("resizes by keyboard and restores the preferred width after a maximized detail", () => {
  Object.defineProperty(window, "innerWidth", { configurable: true, value: 1920 });
  const frame = document.createElement("iframe");
  const onWidth = vi.fn();
  const sizing = mountSidebarResize(frame, onWidth);
  sizing.layout(400, false);
  sizing.handle.dispatchEvent(new KeyboardEvent("keydown", { key: "ArrowLeft", shiftKey: true }));
  expect(frame.style.width).toBe("448px");
  expect(onWidth).toHaveBeenCalledWith(448);
  sizing.layout(448, true);
  expect(frame.style.width).toBe("100vw");
  expect(sizing.handle.hidden).toBe(true);
  sizing.layout(448, false);
  expect(frame.style.width).toBe("448px");
  expect(sizing.handle.hidden).toBe(false);
  expect(onWidth).toHaveBeenCalledTimes(1);
  Object.defineProperty(window, "innerWidth", { configurable: true, value: 390 });
  window.dispatchEvent(new Event("resize"));
  expect(frame.style.width).toBe("390px");
  expect(sizing.handle.hidden).toBe(true);
  Object.defineProperty(window, "innerWidth", { configurable: true, value: 1920 });
  window.dispatchEvent(new Event("resize"));
  expect(frame.style.width).toBe("448px");
  sizing.close();
});

it("captures the pointer so dragging continues across the iframe and stops on release", () => {
  Object.defineProperty(window, "innerWidth", { configurable: true, value: 1920 });
  const frame = document.createElement("iframe");
  const onWidth = vi.fn();
  const sizing = mountSidebarResize(frame, onWidth);
  const capture = vi.fn();
  Object.assign(sizing.handle, { setPointerCapture: capture });
  sizing.handle.dispatchEvent(new MouseEvent("pointerdown", { button: 0, clientX: 1000 }));
  sizing.handle.dispatchEvent(new MouseEvent("pointermove", { clientX: 900 }));
  expect(frame.style.width).toBe("500px");
  expect(capture).toHaveBeenCalledTimes(1);
  sizing.handle.dispatchEvent(new MouseEvent("pointerup"));
  sizing.handle.dispatchEvent(new MouseEvent("pointermove", { clientX: 800 }));
  expect(onWidth).toHaveBeenCalledTimes(1);
  sizing.close();
});
