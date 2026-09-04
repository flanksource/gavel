import { useEffect, useRef, useState, type RefObject } from 'react';

// useContainerWidth reports an element's own content-box width via a
// ResizeObserver, so a component can adapt to the space IT is given rather than
// the viewport — the PR detail panel collapses its action buttons the same way
// whether it is squeezed into the narrow menu-bar webview or a dragged-in split
// pane. Width is 0 until the first measurement (treat that as "unknown, render
// expanded"). Returns a ref to attach to the measured element.
export function useContainerWidth<T extends HTMLElement>(): [RefObject<T | null>, number] {
  const ref = useRef<T | null>(null);
  const [width, setWidth] = useState(0);

  useEffect(() => {
    const el = ref.current;
    if (!el || typeof ResizeObserver === 'undefined') return;
    const observer = new ResizeObserver(entries => {
      for (const entry of entries) setWidth(entry.contentRect.width);
    });
    observer.observe(el);
    return () => observer.disconnect();
  }, []);

  return [ref, width];
}
