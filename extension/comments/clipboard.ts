// writeClipboard takes promises so it can be called synchronously inside the
// user gesture (Clipboard API activation) while the content is still loading.
// With an image, one item carries both: pasting where images go yields the PNG,
// elsewhere the text.
export async function writeClipboard(text: Promise<string>, image?: Promise<Blob>): Promise<void> {
  if (
    typeof navigator === "undefined" ||
    navigator.clipboard?.write === undefined
  ) {
    throw new Error("Clipboard API is unavailable");
  }
  if (typeof ClipboardItem === "undefined") {
    throw new Error("ClipboardItem is unavailable");
  }

  await navigator.clipboard.write([
    new ClipboardItem({
      "text/plain": text.then(
        (value) => new Blob([value], { type: "text/plain" }),
      ),
      ...(image ? { "image/png": image } : {}),
    }),
  ]);
}
