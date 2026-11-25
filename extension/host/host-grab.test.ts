import { afterEach, expect, it, vi } from "vitest";
import type { ReactGrabAPI } from "@flanksource/clicky-ui/react-grab";

const created = { dispose: vi.fn() } as unknown as ReactGrabAPI;
const init = vi.fn(() => created);
vi.mock("@flanksource/clicky-ui/react-grab/core", () => ({ init }));

afterEach(() => {
  delete window.__REACT_GRAB__;
  vi.clearAllMocks();
});

it("reuses the react-grab the page already runs and does not take ownership of it", async () => {
  const pages = { dispose: vi.fn() } as unknown as ReactGrabAPI;
  window.__REACT_GRAB__ = pages;
  const { acquireGrab } = await import("./host-grab");
  expect(acquireGrab()).toEqual({ grab: pages, created: false });
  expect(init).not.toHaveBeenCalled();
});

it("starts its own full-toolbar react-grab, without telemetry, when the page has none", async () => {
  const { acquireGrab } = await import("./host-grab");
  expect(acquireGrab()).toEqual({ grab: created, created: true });
  expect(init).toHaveBeenCalledWith({ enabled: true, telemetry: false, freezeReactUpdates: false });
  expect(window.__REACT_GRAB__).toBeUndefined();
});
