// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from "vitest";
import { envelope } from "./model";
import { connectHostBridge, saveHostSession, saveHostTarget, storedPresentation } from "./useHostBridge";

afterEach(() => {
  localStorage.clear();
  vi.restoreAllMocks();
});

describe("saving site settings through the host", () => {
  const origin = "https://example.com";
  const bridgeSession = "bridge-1";

  /** A parent document that the panel posts to; `answer` replies the way the host does. */
  function attachParent() {
    const frame = document.createElement("iframe");
    document.body.append(frame);
    const parent = frame.contentWindow!;
    const post = vi.spyOn(parent, "postMessage").mockImplementation(() => undefined);
    vi.spyOn(window, "parent", "get").mockReturnValue(parent);
    connectHostBridge({ parentOrigin: origin, session: bridgeSession });
    const requestId = () => {
      const [sent] = post.mock.calls[0] ?? [];
      if (!sent) throw new Error("The panel posted nothing to the host");
      return (sent as { message: { requestId: string } }).message.requestId;
    };
    const answer = (result: { error?: string } = {}) =>
      window.dispatchEvent(new MessageEvent("message", { origin, source: parent, data: envelope(bridgeSession, { type: "settings-result", requestId: requestId(), ...result }) }));
    return { post, answer };
  }

  it("asks the host to store the site's target alone, and settles when the host confirms", async () => {
    const { post, answer } = attachParent();
    const saved = saveHostTarget("global");
    expect(post.mock.calls[0]).toEqual([{ bridge: "gavel-extension-v1", session: bridgeSession, message: { type: "settings", requestId: expect.any(String), target: "global" } }, origin]);
    answer();
    await expect(saved).resolves.toBeUndefined();
  });

  it("asks the host to store, or with undefined clear, the page's session alone", async () => {
    const { post, answer } = attachParent();
    const saved = saveHostSession(undefined);
    expect(post.mock.calls[0]?.[0]).toMatchObject({ message: { type: "settings", session: null } });
    expect(post.mock.calls[0]?.[0]).not.toHaveProperty("message.target");
    answer();
    await saved;
  });

  it("rejects with the host's error when it cannot store the setting", async () => {
    const { answer } = attachParent();
    const saved = saveHostSession("s-1");
    answer({ error: "Grabber session must be a non-empty string" });
    await expect(saved).rejects.toThrow("Grabber session must be a non-empty string");
  });
});

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
