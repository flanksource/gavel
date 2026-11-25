import { describe, expect, it, vi } from "vitest";
import { createAutosave } from "./autosave";
describe("autosave", () => {
  it("keeps unchanged editor mount notifications clean", () => {
    const write = vi.fn();
    const changed = vi.fn();
    const queue = createAutosave({
      source: "initial",
      revision: "r",
      write,
      changed,
    });
    queue.edit("initial");
    expect(queue.state().status).toBe("saved");
    expect(changed).not.toHaveBeenCalled();
    expect(write).not.toHaveBeenCalled();
  });
  it("serializes writes and drains edits made while a save is in flight", async () => {
    let release!: () => void;
    const gate = new Promise<void>((resolve) => {
      release = resolve;
    });
    const write = vi.fn(async (source: string, revision: string) => {
      await gate;
      return revision + source;
    });
    const queue = createAutosave({
      source: "initial",
      revision: "r",
      write,
      changed: () => {},
    });
    queue.edit("first");
    const flushing = queue.flush();
    queue.edit("second");
    release();
    await flushing;
    expect(write.mock.calls).toEqual([
      ["first", "r"],
      ["second", "rfirst"],
    ]);
    expect(queue.state()).toMatchObject({
      source: "second",
      revision: "rfirstsecond",
      status: "saved",
    });
  });
  it("retains failed drafts and stops automatic retries", async () => {
    const queue = createAutosave({
      source: "disk",
      revision: "r",
      write: async () => {
        throw new Error("conflict");
      },
      changed: () => {},
    });
    queue.edit("draft");
    await expect(queue.flush()).rejects.toThrow("conflict");
    expect(queue.state()).toMatchObject({
      source: "draft",
      status: "error",
      error: "conflict",
    });
  });
});
