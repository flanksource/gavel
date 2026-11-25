export type HoverSize = { width: number; height: number };

const MIN = { width: 280, height: 240 };
const MAX_WIDTH = 720;
const MARGIN = 8;

export const DEFAULT_HOVER_SIZE: HoverSize = { width: 400, height: 480 };

export function hoverSize(input: unknown): HoverSize {
  const value = input as Partial<Record<keyof HoverSize, unknown>> | null;
  const valid = (n: unknown, min: number, max: number) => typeof n === "number" && Number.isFinite(n) && n >= min && n <= max;
  if (!value || !valid(value.width, MIN.width, MAX_WIDTH) || !valid(value.height, MIN.height, Number.MAX_SAFE_INTEGER))
    throw new Error(`Hover panel size must be ${MIN.width}–${MAX_WIDTH} pixels wide and at least ${MIN.height} high`);
  return { width: value.width as number, height: value.height as number };
}

/**
 * The hover popover's top-left grip: the popover grows up and left from the
 * FAB, so dragging the grip away from it enlarges the panel. The host positions
 * the grip; this owns the size, clamped to the viewport.
 */
export function mountHoverResize(onSize: (size: HoverSize) => void) {
  const handle = document.createElement("div");
  handle.className = "resize-corner";
  handle.tabIndex = 0;
  handle.setAttribute("role", "slider");
  handle.setAttribute("aria-label", "Resize comment panel");
  let preferred = DEFAULT_HOVER_SIZE;
  let drag: { x: number; y: number; origin: HoverSize } | null = null;
  const clamp = (next: HoverSize): HoverSize => ({
    width: Math.max(MIN.width, Math.min(MAX_WIDTH, window.innerWidth - 2 * MARGIN, next.width)),
    height: Math.max(MIN.height, Math.min(window.innerHeight - 2 * MARGIN, next.height)),
  });
  const size = () => clamp(preferred);
  const describe = () => handle.setAttribute("aria-valuetext", `${size().width} by ${size().height} pixels`);
  const change = (next: HoverSize) => { preferred = clamp(next); describe(); };
  handle.addEventListener("pointerdown", event => {
    if (event.button !== 0) return;
    event.preventDefault();
    drag = { x: event.clientX, y: event.clientY, origin: size() };
    handle.setPointerCapture(event.pointerId);
  });
  handle.addEventListener("pointermove", event => {
    if (drag) change({ width: drag.origin.width + drag.x - event.clientX, height: drag.origin.height + drag.y - event.clientY });
  });
  const end = () => { if (drag) { drag = null; onSize(size()); } };
  handle.addEventListener("pointerup", end);
  handle.addEventListener("pointercancel", end);
  handle.addEventListener("lostpointercapture", end);
  handle.addEventListener("keydown", event => {
    const step = event.shiftKey ? 48 : 16;
    const current = size();
    if (event.key === "Home") change(MIN);
    else if (event.key === "End") change({ width: MAX_WIDTH, height: window.innerHeight });
    else if (event.key === "ArrowLeft") change({ ...current, width: current.width + step });
    else if (event.key === "ArrowRight") change({ ...current, width: current.width - step });
    else if (event.key === "ArrowUp") change({ ...current, height: current.height + step });
    else if (event.key === "ArrowDown") change({ ...current, height: current.height - step });
    else return;
    event.preventDefault();
    onSize(size());
  });
  describe();
  return {
    handle,
    /** Adopts a stored size without reporting it back. */
    layout(next: HoverSize) { preferred = hoverSize(next); describe(); },
    size,
    close() { handle.remove(); },
  };
}
