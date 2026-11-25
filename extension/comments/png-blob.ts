const PNG_PREFIX = "data:image/png;base64,";

/**
 * The PNG a data URL carries. Decoded here rather than fetched: a page's CSP can
 * refuse `fetch("data:...")`, and anything that is not a PNG is refused, not passed on.
 */
export function pngBlob(dataUrl: string): Blob {
  if (!dataUrl.startsWith(PNG_PREFIX)) throw new Error("Attachment must be a PNG data URL");
  const binary = atob(dataUrl.slice(PNG_PREFIX.length));
  return new Blob([Uint8Array.from(binary, (char) => char.charCodeAt(0))], { type: "image/png" });
}
