export type HostViewportRect = { x: number; y: number; width: number; height: number };
export type ScreenshotTarget = {
  rect: HostViewportRect;
  isLive: () => boolean;
};

/**
 * The pixels of a captured frame (a video frame or a tab image) that show
 * `target`, given the frame's size and the viewport size it was captured from.
 */
export function viewportCrop(options: {
  target: Element | ScreenshotTarget;
  frame: { width: number; height: number };
  viewport: { width: number; height: number };
}): HostViewportRect {
  const { target, frame, viewport } = options;
  const bounds = "rect" in target ? target.rect : target.getBoundingClientRect();
  const live = "rect" in target ? target.isLive() : target.isConnected;
  if (!live || !Object.values({ x: bounds.x, y: bounds.y, width: bounds.width, height: bounds.height }).every(Number.isFinite) || bounds.width <= 0 || bounds.height <= 0) {
    throw new Error("selected element is no longer visible");
  }
  if (!frame.width || !frame.height) {
    throw new Error("captured tab has no frame");
  }
  const scaleX = frame.width / viewport.width;
  const scaleY = frame.height / viewport.height;
  if (
    !Number.isFinite(scaleX) ||
    !Number.isFinite(scaleY) ||
    Math.abs(scaleX - scaleY) > Math.max(scaleX, scaleY) * 0.02
  ) {
    throw new Error("captured tab has invalid viewport dimensions");
  }
  const x = Math.max(0, Math.floor(bounds.x * scaleX));
  const y = Math.max(0, Math.floor(bounds.y * scaleY));
  const right = Math.min(frame.width, Math.ceil((bounds.x + bounds.width) * scaleX));
  const bottom = Math.min(frame.height, Math.ceil((bounds.y + bounds.height) * scaleY));
  if (right <= x || bottom <= y) {
    throw new Error("selected element is outside the captured viewport");
  }
  return { x, y, width: right - x, height: bottom - y };
}
