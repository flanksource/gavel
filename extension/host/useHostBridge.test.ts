// @vitest-environment jsdom
import { afterEach, expect, it } from "vitest";
import { storedPresentation } from "./useHostBridge";

afterEach(() => localStorage.clear());

it("returns nothing for a panel that never stored a layout, so the host keeps its hover default", () => {
  expect(storedPresentation()).toEqual({});
});

it("reads back the remembered mode and hover size", () => {
  localStorage.setItem("playground.bookmarklet.mode", "rail");
  localStorage.setItem("playground.bookmarklet.hover", JSON.stringify({ width: 520, height: 600 }));
  expect(storedPresentation()).toEqual({ mode: "rail", hover: { width: 520, height: 600 } });
});

it.each([
  ["playground.bookmarklet.mode", "draft", "Invalid panel mode"],
  ["playground.bookmarklet.hover", JSON.stringify({ width: 10, height: 600 }), "Hover panel size"],
])("rejects a corrupt %s record", (key, value, message) => {
  localStorage.setItem(key, value);
  expect(() => storedPresentation()).toThrow(message);
});
