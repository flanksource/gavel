export type FabPosition = { x: number; y: number };

export function mountFabMovement(button: HTMLButtonElement, changed: (position: FabPosition) => void) {
  let position = { x: window.innerWidth - 68, y: window.innerHeight - 68 };
  let drag: { x: number; y: number; origin: FabPosition } | null = null;
  let moved = false;
  function layout(next = position) {
    position = { x: Math.max(8, Math.min(window.innerWidth - 56, next.x)), y: Math.max(8, Math.min(window.innerHeight - 56, next.y)) };
    Object.assign(button.style, { left: `${position.x}px`, top: `${position.y}px`, right: "auto", bottom: "auto" });
  }
  button.style.touchAction = "none";
  button.addEventListener("pointerdown", event => {
    if (event.button !== 0) return;
    moved = false;
    drag = { x: event.clientX, y: event.clientY, origin: position };
    button.setPointerCapture(event.pointerId);
  });
  button.addEventListener("pointermove", event => {
    if (!drag) return;
    const dx = event.clientX - drag.x, dy = event.clientY - drag.y;
    if (Math.hypot(dx, dy) > 5) moved = true;
    if (moved) layout({ x: drag.origin.x + dx, y: drag.origin.y + dy });
  });
  const end = () => { if (drag && moved) changed(position); drag = null; };
  button.addEventListener("pointerup", end);
  button.addEventListener("pointercancel", end);
  button.addEventListener("lostpointercapture", end);
  button.addEventListener("click", event => { if (moved) { event.preventDefault(); event.stopImmediatePropagation(); moved = false; } }, true);
  button.addEventListener("keydown", event => {
    const step = event.shiftKey ? 48 : 16;
    const dx = event.key === "ArrowLeft" ? -step : event.key === "ArrowRight" ? step : 0;
    const dy = event.key === "ArrowUp" ? -step : event.key === "ArrowDown" ? step : 0;
    if (!dx && !dy) return;
    event.preventDefault(); layout({ x: position.x + dx, y: position.y + dy }); changed(position);
  });
  const resize = () => layout();
  window.addEventListener("resize", resize);
  layout();
  return { layout, position: () => position, close: () => window.removeEventListener("resize", resize) };
}
