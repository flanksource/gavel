import { describeElement, resolveAnchor } from "../comments/dom-anchor";
import type { AnchorState, PanelMode, Presentation } from "./model";
import { mountFabMovement, type FabPosition } from "./host-fab";
import { mountSidebarResize } from "./host-resize";
import { mountHoverResize, type HoverSize } from "./host-hover-resize";
import { draftPosition } from "./draft-position";
import { GAVEL_FILLED_SVG } from "./gavel-icon";

export type MountHostOptions = {
  frameUrl: string;
  focus: (anchor: string) => void;
  onWidth: (width: number) => void;
  onPresentation: (mode: Presentation) => void;
  onPosition?: (position: FabPosition) => void;
  onOpenChange?: (open: boolean) => void;
  /** The user resized the hover popover. */
  onHoverSize?: (size: HoverSize) => void;
};

export function mountHost({ frameUrl, focus, onWidth, onPresentation, onPosition = () => {}, onOpenChange = () => {}, onHoverSize = () => {} }: MountHostOptions) {
  const root = document.createElement("div");
  root.id = "gavel-grabber";
  root.setAttribute("data-react-grab-ignore", "");
  const shadow = root.attachShadow({ mode: "open" });
  const style = document.createElement("style");
  style.textContent = `
    :host { all: initial; position: fixed; inset: 0; pointer-events: none; z-index: 2147483646; }
    iframe { position: absolute; inset: 0 0 0 auto; width: min(400px, 100vw); height: 100%; border: 0; background: white; box-shadow: 0 0 16px #0003; pointer-events: auto; }
    .resize { position: absolute; top: 0; height: 100%; width: 5px; background: #8886; cursor: col-resize; pointer-events: auto; touch-action: none; }
    .resize:hover, .resize:focus-visible { background: #2563eb; outline: 2px solid white; }
    .resize[hidden], iframe[hidden] ~ .resize { display: none; }
    .resize-corner { position: absolute; width: 14px; height: 14px; border-radius: 4px; background: #8888; cursor: nwse-resize; pointer-events: auto; touch-action: none; }
    .resize-corner:hover, .resize-corner:focus-visible { background: #2563eb; outline: 2px solid white; }
    .resize-corner[hidden] { display: none; }
    .fab { position: absolute; right: 20px; bottom: 20px; width: 48px; height: 48px; padding: 0; border: 2px solid white; border-radius: 50%; background: #2563eb; color: white; cursor: pointer; pointer-events: auto; box-shadow: 0 2px 12px #0004; }
    .fab svg { display: block; width: 26px; height: 26px; margin: auto; pointer-events: none; }
    .fab[hidden] { display: none; }
    :host(:not([data-mode=rail])) .resize { display: none; }
    :host([data-mode=compact]) iframe { inset: auto; border-radius: 10px; }
    :host([data-mode=draft]) iframe { inset: auto; border-radius: 10px; }
    .pin { position: absolute; pointer-events: auto; min-width: 22px; height: 22px; padding: 0 4px; border: 1px solid #fff; border-radius: 12px; color: white; background: #2563eb; font: 600 12px system-ui; cursor: pointer; transform: translate(-50%, -50%); }
    .highlight { position: absolute; border: 2px solid #2563eb; background: #2563eb18; box-sizing: border-box; pointer-events: none; }
    .error { position: absolute; top: 12px; left: 12px; max-width: 480px; padding: 12px; background: #fff; color: #991b1b; border: 1px solid currentColor; font: 14px system-ui; pointer-events: auto; }
  `;
  const frame = document.createElement("iframe");
  frame.title = "Gavel comments";
  frame.src = frameUrl;
  const pins = document.createElement("div");
  const highlight = document.createElement("div");
  highlight.className = "highlight";
  highlight.hidden = true;
  const sizing = mountSidebarResize(frame, onWidth);
  const hover = mountHoverResize(size => { onHoverSize(size); schedule(); });
  hover.handle.hidden = true;
  hover.handle.addEventListener("pointermove", () => schedule());
  hover.handle.addEventListener("keydown", () => schedule());
  const fab = document.createElement("button");
  fab.className = "fab";
  fab.type = "button";
  fab.innerHTML = GAVEL_FILLED_SVG;
  fab.setAttribute("aria-label", "Gavel actions");
  fab.title = "Gavel actions; drag or use arrow keys to move";
  fab.hidden = false;
  frame.hidden = true;
  let mode: Presentation = "compact";
  root.setAttribute("data-mode", mode);
  let minimized = true;
  let previousMinimized = false;
  let draftAnchor: string | null = null;
  let draftHeight = 260;
  let draftBox = { left: 8, top: 8, bottom: 8 };
  let dialogOpen = false;
  const movement = mountFabMovement(fab, position => { onPosition(position); schedule(); });
  let previousMode: Presentation = "compact";
  const setMinimized = (value: boolean) => { if (minimized === value) return; minimized = value; onOpenChange(!value); };
  const restore = () => { setMinimized(false); frame.hidden = false; fab.hidden = true; frame.focus(); };
  const open = () => { if (mode === "rail") showRail(); else { restore(); onPresentation(mode); schedule(); } };
  fab.addEventListener("click", () => open());
  shadow.append(style, pins, highlight, frame, sizing.handle, hover.handle, fab);
  document.documentElement.append(root);
  let anchors: string[] = [];
  let selected: string | null = null;
  let lastState = "";
  let pending = 0;
  let onAnchors: (state: AnchorState[]) => void = () => {};

  function recompute() {
    pending = 0;
    hover.handle.hidden = mode !== "compact" || dialogOpen || frame.hidden;
    if (mode === "compact" && !dialogOpen) {
      const { x, y } = movement.position();
      const { width, height } = hover.size();
      const left = Math.max(8, Math.min(x, window.innerWidth - width - 8));
      const top = Math.max(8, Math.min(y - height, window.innerHeight - height - 8));
      Object.assign(frame.style, { left: `${left}px`, top: `${top}px`, width: `${width}px`, height: `${height}px` });
      Object.assign(hover.handle.style, { left: `${left - 5}px`, top: `${top - 5}px` });
    }
    if (mode === "draft" && !dialogOpen) {
      const target = draftAnchor ? resolveAnchor(document.body, draftAnchor) : null;
      if (target) draftBox = target.getBoundingClientRect();
      const position = draftPosition(draftBox, { width: window.innerWidth, height: window.innerHeight }, draftHeight);
      for (const [key, value] of Object.entries(position)) frame.style.setProperty(key, `${value}px`);
    }
    pins.replaceChildren();
    const states = anchors.map((anchor, index): AnchorState => {
      const element = resolveAnchor(document.body, anchor);
      if (!element) return { anchor, available: false, label: anchor };
      const box = element.getBoundingClientRect();
      const pin = document.createElement("button");
      pin.className = "pin";
      pin.textContent = String(index + 1);
      pin.title = describeElement(element);
      pin.style.left = `${box.right}px`;
      pin.style.top = `${box.top}px`;
      pin.addEventListener("click", () => { showRail(); focus(anchor); });
      pins.append(pin);
      return {
        anchor,
        available: true,
        top: box.top,
        label: describeElement(element),
      };
    });
    const element = selected ? resolveAnchor(document.body, selected) : null;
    highlight.hidden = !element;
    if (element) {
      const box = element.getBoundingClientRect();
      Object.assign(highlight.style, {
        left: `${box.left}px`,
        top: `${box.top}px`,
        width: `${box.width}px`,
        height: `${box.height}px`,
      });
    }
    if (JSON.stringify(states) !== lastState) {
      lastState = JSON.stringify(states);
      onAnchors(states);
    }
  }
  const schedule = () => {
    if (!pending) pending = requestAnimationFrame(recompute);
  };
  const observer = new MutationObserver(schedule);
  observer.observe(document.body, {
    subtree: true,
    childList: true,
    characterData: true,
    attributes: true,
  });
  const resize = new ResizeObserver(schedule);
  resize.observe(document.body);
  window.addEventListener("scroll", schedule, true);
  window.addEventListener("resize", schedule);
  function railLayout() {
    mode = "rail";
    root.setAttribute("data-mode", mode);
    hover.handle.hidden = true;
    frame.style.removeProperty("left");
    frame.style.removeProperty("top");
    frame.style.removeProperty("height");
    sizing.layout(preferredWidth, maximized);
    onPresentation("rail");
  }
  function showRail() {
    railLayout();
    restore();
  }
  const resume = () => { frame.hidden = minimized; fab.hidden = !minimized; schedule(); };
  function closeDraft() {
    if (previousMode === "rail") railLayout();
    else { mode = "compact"; root.setAttribute("data-mode", mode); onPresentation(mode); schedule(); }
    if (!previousMinimized) restore();
    else { setMinimized(true); resume(); }
  }
  let preferredWidth = 400;
  let maximized = false;
  return {
    root,
    frame,
    restore,
    showRail,
    /** Opens the panel in the mode it was last shown in. */
    open,
    /** Switches the panel between the side rail and the hover popover, opening it; the FAB reopens it in the same mode. */
    setPresentation(next: PanelMode) {
      if (mode === "draft") throw new Error("Finish or cancel the floating comment before switching the panel mode");
      mode = next;
      root.setAttribute("data-mode", mode);
      open();
    },
    /** Adopts a remembered mode for the next open without opening a minimized panel. */
    preferPresentation(next: PanelMode) {
      if (mode === "draft") return;
      mode = next;
      root.setAttribute("data-mode", mode);
      if (!minimized) open();
    },
    hoverLayout(size: HoverSize) { hover.layout(size); schedule(); },
    position: movement.layout,
    dialog(open: boolean) {
      dialogOpen = open;
      if (open) Object.assign(frame.style, { left: "0", top: "0", width: "100vw", height: "100vh" });
      else if (mode === "rail") railLayout();
      else schedule();
    },
    minimize() { setMinimized(true); frame.hidden = true; fab.hidden = false; fab.focus(); schedule(); },
    /** Whether the sidebar is closed to the FAB (a sidebar hidden only while picking an element is still open). */
    minimized: () => minimized,
    resume,
    /** A new URL drops any floating draft but never reopens a minimized sidebar. */
    pageChanged() {
      if (mode === "draft") closeDraft();
      else resume();
    },
    draft(active: boolean, anchor?: string) {
      if (active) {
        if (!anchor) throw new Error("Floating comment requires an anchor");
        if (mode !== "draft") { previousMinimized = minimized; previousMode = mode; }
        draftAnchor = anchor;
        mode = "draft";
        root.setAttribute("data-mode", mode);
        restore();
        onPresentation("draft");
        schedule();
      } else closeDraft();
    },
    draftSize(height: number) { draftHeight = height; schedule(); },
    layout(width: number, detail: boolean) {
      preferredWidth = width;
      maximized = detail;
      if (mode === "rail") sizing.layout(width, detail);
      else schedule();
    },
    update(next: string[], callback: typeof onAnchors) {
      anchors = next;
      onAnchors = callback;
      lastState = "";
      schedule();
    },
    highlight(anchor: string | null) {
      selected = anchor;
      schedule();
    },
    error(message: string) {
      let error = shadow.querySelector<HTMLDivElement>(".error");
      if (!error) {
        error = document.createElement("div");
        error.className = "error";
        error.setAttribute("role", "alert");
        shadow.append(error);
      }
      error.textContent = message;
    },
    close() {
      sizing.close();
      hover.close();
      movement.close();
      observer.disconnect();
      resize.disconnect();
      cancelAnimationFrame(pending);
      window.removeEventListener("scroll", schedule, true);
      window.removeEventListener("resize", schedule);
      root.remove();
    },
  };
}
