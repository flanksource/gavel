import { expect, it, vi } from "vitest";
import type { EventSourceFactory } from "@flanksource/clicky-ui/hooks";
import { gavelStreams } from "./gavel-streams";

const SERVER = "http://localhost:9092";

it("subscribes a gavel stream through the hub by its server-relative path and query", () => {
  const hub = vi.fn<EventSourceFactory>();
  gavelStreams({ server: SERVER, hub })(`${SERVER}/api/captain/sessions/s-1?follow=1&after=m-9`);
  expect(hub.mock.calls).toEqual([["/api/captain/sessions/s-1?follow=1&after=m-9"]]);
});

it.each([
  "http://elsewhere.example.test/api/captain/sessions/s-1?follow=1",
  "/api/captain/sessions/s-1?follow=1",
])("refuses %s, which is not an absolute URL on the gavel server", url => {
  const hub = vi.fn<EventSourceFactory>();
  expect(() => gavelStreams({ server: SERVER, hub })(url)).toThrow("gavel server");
  expect(hub).not.toHaveBeenCalled();
});
