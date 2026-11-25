import { describe, expect, it } from "vitest";
import { viewportCrop } from "./screenshot-viewport";

const viewport = { width: 1000, height: 800 };
const target = (rect: { x: number; y: number; width: number; height: number }, isLive = true) => ({ rect, isLive: () => isLive });

describe("viewportCrop", () => {
  it("scales the element rect by the frame-to-viewport ratio and rounds outwards", () => {
    expect(viewportCrop({ target: target({ x: 10.2, y: 20.5, width: 100.1, height: 50 }), frame: { width: 2000, height: 1600 }, viewport })).toEqual({ x: 20, y: 41, width: 201, height: 100 });
  });

  it("uses a 1:1 ratio when the frame is the viewport size", () => {
    expect(viewportCrop({ target: target({ x: 5, y: 6, width: 7, height: 8 }), frame: viewport, viewport })).toEqual({ x: 5, y: 6, width: 7, height: 8 });
  });

  it("clips an element that overhangs the frame", () => {
    expect(viewportCrop({ target: target({ x: -20, y: 790, width: 100, height: 100 }), frame: viewport, viewport })).toEqual({ x: 0, y: 790, width: 80, height: 10 });
  });

  it("reads a live element's bounds", () => {
    const element = document.createElement("div");
    document.body.append(element);
    element.getBoundingClientRect = () => ({ x: 1, y: 2, width: 30, height: 40 }) as DOMRect;
    expect(viewportCrop({ target: element, frame: viewport, viewport })).toEqual({ x: 1, y: 2, width: 30, height: 40 });
    element.remove();
  });

  it.each([
    { name: "a target that is no longer live", target: target({ x: 0, y: 0, width: 10, height: 10 }, false), frame: viewport, error: /no longer visible/ },
    { name: "an empty rect", target: target({ x: 0, y: 0, width: 0, height: 10 }), frame: viewport, error: /no longer visible/ },
    { name: "a non-finite rect", target: target({ x: Number.NaN, y: 0, width: 10, height: 10 }), frame: viewport, error: /no longer visible/ },
    { name: "an empty frame", target: target({ x: 0, y: 0, width: 10, height: 10 }), frame: { width: 0, height: 0 }, error: /no frame/ },
    { name: "a frame whose aspect ratio is not the viewport's", target: target({ x: 0, y: 0, width: 10, height: 10 }), frame: { width: 2000, height: 1000 }, error: /invalid viewport dimensions/ },
    { name: "an element outside the frame", target: target({ x: 1200, y: 0, width: 10, height: 10 }), frame: viewport, error: /outside the captured viewport/ },
  ])("rejects $name", ({ target: rejected, frame, error }) => {
    expect(() => viewportCrop({ target: rejected, frame, viewport })).toThrow(error);
  });
});
