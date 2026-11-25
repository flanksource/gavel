export function sidebarWidth(input: unknown): number {
  if (typeof input !== "number" || !Number.isFinite(input) || input < 280 || input > 640) throw new Error("Sidebar width must be between 280 and 640 pixels");
  return input;
}

export function mountSidebarResize(frame: HTMLIFrameElement, onWidth: (width: number) => void) {
  const handle = document.createElement("div");
  handle.className = "resize";
  handle.tabIndex = 0;
  handle.setAttribute("role", "separator");
  handle.setAttribute("aria-label", "Resize comment sidebar");
  handle.setAttribute("aria-orientation", "vertical");
  handle.setAttribute("aria-valuemin", "280");
  let preferred = 400;
  let maximized = false;
  let drag: { x: number; width: number } | null = null;
  const max = () => Math.max(280, Math.min(640, window.innerWidth - 320));
  const width = () => window.innerWidth < 640 ? window.innerWidth : Math.min(preferred, max());
  const layout = () => {
    frame.style.width = maximized ? "100vw" : `${width()}px`;
    handle.style.right = `${width() - 2}px`;
    handle.hidden = maximized || window.innerWidth < 640;
    handle.setAttribute("aria-valuemax", String(max()));
    handle.setAttribute("aria-valuenow", String(Math.round(width())));
  };
  const change = (next: number) => {
    preferred = Math.max(280, Math.min(max(), next));
    layout();
    onWidth(preferred);
  };
  handle.addEventListener("pointerdown", event => {
    if (event.button !== 0) return;
    event.preventDefault();
    drag = { x: event.clientX, width: width() };
    handle.setPointerCapture(event.pointerId);
  });
  handle.addEventListener("pointermove", event => {
    if (drag) change(drag.width + drag.x - event.clientX);
  });
  const end = () => { drag = null; };
  handle.addEventListener("pointerup", end);
  handle.addEventListener("pointercancel", end);
  handle.addEventListener("lostpointercapture", end);
  handle.addEventListener("keydown", event => {
    const step = event.shiftKey ? 48 : 16;
    if (event.key === "Home") change(280);
    else if (event.key === "End") change(max());
    else if (event.key === "ArrowLeft") change(width() + step);
    else if (event.key === "ArrowRight") change(width() - step);
    else return;
    event.preventDefault();
  });
  window.addEventListener("resize", layout);
  layout();
  return {
    handle,
    layout(next: number, detail: boolean) {
      preferred = sidebarWidth(next);
      maximized = detail;
      layout();
    },
    close() { window.removeEventListener("resize", layout); handle.remove(); },
  };
}
