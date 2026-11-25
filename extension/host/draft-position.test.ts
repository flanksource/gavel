import { expect, it } from "vitest";
import { draftPosition } from "./draft-position";

it("flips above an element near the bottom and clamps horizontally", () => {
  expect(
    draftPosition(
      { left: 990, top: 700, bottom: 730 },
      { width: 1024, height: 768 },
      240,
    ),
  ).toEqual({ left: 696, top: 454, width: 320, height: 240 });
});

it("fits compact viewports and oversized composers", () => {
  expect(
    draftPosition(
      { left: -20, top: 0, bottom: 40 },
      { width: 300, height: 200 },
      500,
    ),
  ).toEqual({ left: 8, top: 8, width: 284, height: 184 });
});
