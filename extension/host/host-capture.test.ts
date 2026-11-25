import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { withOverlaysHidden } from "./host-capture";

beforeEach(() => {
  vi.stubGlobal("requestAnimationFrame", (callback: FrameRequestCallback) => {
    callback(0);
    return 1;
  });
});
afterEach(() => {
  vi.unstubAllGlobals();
  document.body.replaceChildren();
});

function overlay(attribute: string) {
  const element = document.createElement("div");
  element.setAttribute(attribute, "");
  document.body.append(element);
  return element;
}

it("hides the host and every react-grab root while the work runs, then restores their own visibility", async () => {
  const host = overlay("data-grabber-host");
  const grab = overlay("data-react-grab");
  const otherGrab = overlay("data-react-grab");
  otherGrab.style.visibility = "visible";
  const page = document.createElement("main");
  document.body.append(page);
  const during = await withOverlaysHidden({ overlays: [host], work: () => Promise.resolve([host, grab, otherGrab, page].map((element) => element.style.visibility)) });
  expect(during).toEqual(["hidden", "hidden", "hidden", ""]);
  expect([host, grab, otherGrab, page].map((element) => element.style.visibility)).toEqual(["", "", "visible", ""]);
});

it("restores the overlays and surfaces the error when the work fails", async () => {
  const host = overlay("data-grabber-host");
  await expect(withOverlaysHidden({ overlays: [host], work: () => Promise.reject(new Error("capture failed")) })).rejects.toThrow("capture failed");
  expect(host.style.visibility).toBe("");
});

it("waits for a paint with the overlays hidden before starting the work", async () => {
  const host = overlay("data-grabber-host");
  const order: string[] = [];
  vi.stubGlobal("requestAnimationFrame", (callback: FrameRequestCallback) => {
    order.push(`frame:${host.style.visibility}`);
    callback(0);
    return 1;
  });
  await withOverlaysHidden({ overlays: [host], work: () => { order.push("work"); return Promise.resolve(); } });
  expect(order).toEqual(["frame:hidden", "frame:hidden", "work"]);
});
