/** Every react-grab shadow host on the page: its toolbar, selection overlays and menus. */
const REACT_GRAB_ROOTS = "[data-react-grab]";

/**
 * A screenshot of the tab shows whatever is painted, so the grabber's own UI is
 * hidden (and a paint awaited) before `work` captures, and put back afterwards,
 * whether or not the capture succeeds.
 */
export async function withOverlaysHidden<T>(options: { overlays: HTMLElement[]; work: () => Promise<T> }): Promise<T> {
  const hidden = [...new Set([...options.overlays, ...document.querySelectorAll<HTMLElement>(REACT_GRAB_ROOTS)])].map((element) => ({
    element,
    value: element.style.getPropertyValue("visibility"),
    priority: element.style.getPropertyPriority("visibility"),
  }));
  for (const { element } of hidden) element.style.setProperty("visibility", "hidden", "important");
  try {
    await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())));
    return await options.work();
  } finally {
    for (const { element, value, priority } of hidden) {
      if (value) element.style.setProperty("visibility", value, priority);
      else element.style.removeProperty("visibility");
    }
  }
}
