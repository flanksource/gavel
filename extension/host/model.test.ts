import { expect, it } from "vitest";
import { bridgeMessage, envelope, externalTitle } from "./model";

it("rejects layout commands from a different bridge or session", () => {
  const message = { type: "sidebar-layout", width: 400, maximized: true } as const;
  expect(bridgeMessage(envelope("session", message), "session")).toEqual(message);
  expect(bridgeMessage(envelope("other", message), "session")).toBeNull();
  expect(bridgeMessage({ bridge: "other", session: "session", message }, "session")).toBeNull();
});

it("names untitled file documents by filename and untitled websites by host", () => {
  expect(externalTitle({ url: "file:///documents/report%20draft.html", title: "" })).toBe("report draft.html");
  expect(externalTitle({ url: "https://example.com/report", title: "" })).toBe("example.com");
  expect(externalTitle({ url: "file:///documents/report.html", title: "Quarterly report" })).toBe("Quarterly report");
});
