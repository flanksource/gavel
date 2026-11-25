import { afterEach, expect, it, vi } from "vitest";
import { assertServerAccess, serverAccessPattern } from "./server-access";

afterEach(() => vi.unstubAllGlobals());

it("asks Chrome about the server's whole origin", () => {
  expect(serverAccessPattern("http://localhost:9092/api/projects")).toBe("http://localhost:9092/*");
  expect(() => serverAccessPattern("file:///tmp/x")).toThrow(/HTTP\(S\)/);
});

it.each([
  { granted: true, outcome: "resolves" },
  { granted: false, outcome: "names the withheld server and how to grant it" },
])("access check $outcome when the grant is $granted", async ({ granted }) => {
  const contains = vi.fn().mockResolvedValue(granted);
  vi.stubGlobal("chrome", { permissions: { contains } });
  const check = assertServerAccess("http://localhost:9092");
  if (granted) await expect(check).resolves.toBeUndefined();
  else await expect(check).rejects.toThrow(/withholding this extension's access to http:\/\/localhost:9092.*Grant access/);
  expect(contains).toHaveBeenCalledWith({ origins: ["http://localhost:9092/*"] });
});
