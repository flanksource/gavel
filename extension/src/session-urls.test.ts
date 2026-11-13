import { expect, it } from "vitest";
import { captainSessionUrl, sessionDashboardUrl } from "./session-urls";

const SERVER = "http://localhost:9092";

it.each([
  { name: "dashboard page", url: sessionDashboardUrl, expected: `${SERVER}/sessions/a%2Fb%20c` },
  { name: "captain transcript", url: captainSessionUrl, expected: `${SERVER}/api/captain/sessions/a%2Fb%20c` },
])("builds the session's $name URL with an encoded id", ({ url, expected }) => {
  expect(url({ server: SERVER, sessionId: "a/b c" })).toBe(expected);
});
