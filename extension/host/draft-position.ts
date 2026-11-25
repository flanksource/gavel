export function draftPosition(
  box: { left: number; top: number; bottom: number },
  viewport: { width: number; height: number },
  requestedHeight: number,
) {
  const width = Math.max(1, Math.min(320, viewport.width - 16));
  const height = Math.max(1, Math.min(requestedHeight, viewport.height - 16));
  const top =
    box.bottom + 6 + height <= viewport.height - 8
      ? box.bottom + 6
      : box.top - height - 6;
  return {
    left: Math.max(8, Math.min(box.left, viewport.width - width - 8)),
    top: Math.max(8, Math.min(top, viewport.height - height - 8)),
    width,
    height,
  };
}
