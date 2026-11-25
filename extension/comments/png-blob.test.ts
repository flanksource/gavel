import { describe, expect, it } from "vitest";
import { pngBlob } from "./png-blob";

function readBytes(blob: Blob): Promise<number[]> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve([...new Uint8Array(reader.result as ArrayBuffer)]);
    reader.onerror = () => reject(reader.error);
    reader.readAsArrayBuffer(blob);
  });
}

describe("pngBlob", () => {
  it("decodes a PNG data URL", async () => {
    // "iVBORw0KGgo=" is the 8-byte PNG signature.
    const blob = pngBlob("data:image/png;base64,iVBORw0KGgo=");
    expect(blob.type).toBe("image/png");
    expect(await readBytes(blob)).toEqual([137, 80, 78, 71, 13, 10, 26, 10]);
  });

  it("refuses anything that is not a PNG data URL", () => {
    expect(() => pngBlob("data:image/jpeg;base64,AAAA")).toThrow(/PNG data URL/);
    expect(() => pngBlob("https://example.com/a.png")).toThrow(/PNG data URL/);
  });
});
