import { expect, it } from "vitest";
import { earlyDecision } from "./early-decision";

const enabled = JSON.stringify({ enabled: true });

it.each([
  { name: "no stored settings", read: () => null, hookDefined: false },
  { name: "settings without enabled", read: () => JSON.stringify({ project: "docs" }), hookDefined: false },
  { name: "settings without enabled while a hook exists", read: () => JSON.stringify({ project: "docs" }), hookDefined: true },
])("does nothing for $name", ({ read, hookDefined }) => {
  expect(earlyDecision({ read, hookDefined })).toEqual({ kind: "off" });
});

it("enables the site and installs the React DevTools hook when none exists", () => {
  expect(earlyDecision({ read: () => enabled, hookDefined: false })).toEqual({ kind: "on", installHook: true });
});

it("enables the site without touching a hook the page or the React DevTools extension already defined", () => {
  expect(earlyDecision({ read: () => enabled, hookDefined: true })).toEqual({ kind: "on", installHook: false });
});

it.each([
  { name: "storage that throws", read: () => { throw new DOMException("denied", "SecurityError"); }, message: /storage is not available.*denied/ },
  { name: "settings that are not valid JSON", read: () => "{nope", message: /not valid JSON/ },
  { name: "settings with an unknown field", read: () => JSON.stringify({ enabled: true, colour: "red" }), message: /unknown field "colour"/ },
])("reports $name and does not enable the page", ({ read, message }) => {
  const decision = earlyDecision({ read, hookDefined: false });
  expect(decision.kind).toBe("failed");
  expect(decision).toMatchObject({ message: expect.stringMatching(message) });
});
