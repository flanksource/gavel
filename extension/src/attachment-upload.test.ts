import { describe, expect, it, vi } from "vitest";
import { absoluteAttachmentUrl, uploadAttachment } from "./attachment-upload";

const server = "http://localhost:9092";
// "iVBORw0KGgo=" is the 8-byte PNG signature.
const dataUrl = "data:image/png;base64,iVBORw0KGgo=";

function response(options: { status?: number; body: unknown }): Response {
  return new Response(JSON.stringify(options.body), { status: options.status ?? 200, headers: { "content-type": "application/json" } });
}

describe("absoluteAttachmentUrl", () => {
  it("resolves a server-relative path against the gavel server", () => {
    expect(absoluteAttachmentUrl("/api/todos/attachments/a.png", server)).toBe(`${server}/api/todos/attachments/a.png`);
    expect(absoluteAttachmentUrl("http://cdn.example/a.png", server)).toBe("http://cdn.example/a.png");
  });
});

describe("uploadAttachment", () => {
  it("posts the PNG as multipart field `attachment` to the todo attachments API", async () => {
    const fetcher = vi.fn((_url: RequestInfo | URL, _init?: RequestInit) => Promise.resolve(response({ body: { attachments: [{ url: "/files/a.png", filename: "shot.png" }] } })));
    await uploadAttachment({ server, dataUrl, name: "shot.png", fetcher });
    const [url, init] = fetcher.mock.calls[0]!;
    expect(url).toBe(`${server}/api/todos/attachments`);
    expect(init!.method).toBe("POST");
    const form = init!.body as FormData;
    expect([...form.keys()]).toEqual(["attachment"]);
    const file = form.get("attachment") as File;
    expect([file.name, file.type, file.size]).toEqual(["shot.png", "image/png", 8]);
    expect(init!.headers).toBeUndefined();
  });

  it("returns the first attachment with an absolute URL", async () => {
    const fetcher = () => Promise.resolve(response({ body: { attachments: [{ url: "/files/a.png", filename: "shot.png" }, { url: "/files/b.png" }] } }));
    await expect(uploadAttachment({ server, dataUrl, name: "shot.png", fetcher })).resolves.toEqual({ url: `${server}/files/a.png`, filename: "shot.png" });
  });

  it("fails with the server's error text on a rejected upload", async () => {
    const fetcher = () => Promise.resolve(response({ status: 413, body: { error: "attachment too large" } }));
    await expect(uploadAttachment({ server, dataUrl, name: "shot.png", fetcher })).rejects.toThrow("attachment too large");
  });

  it("fails with a generic message when a rejected upload names no error", async () => {
    const fetcher = () => Promise.resolve(new Response("nope", { status: 500 }));
    await expect(uploadAttachment({ server, dataUrl, name: "shot.png", fetcher })).rejects.toThrow("screenshot upload failed (HTTP 500)");
  });

  it("fails when the server answers without an attachment URL", async () => {
    const fetcher = () => Promise.resolve(response({ body: { attachments: [] } }));
    await expect(uploadAttachment({ server, dataUrl, name: "shot.png", fetcher })).rejects.toThrow("screenshot upload returned no URL");
  });
});
