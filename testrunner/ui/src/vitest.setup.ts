// jsdom implements no media-query engine, so window.matchMedia is simply
// absent. Any component that sizes itself from a breakpoint therefore throws
// "window.matchMedia is not a function" on mount — clicky-ui's Modal and the
// runner's responsive split do this. Report "no match" for every query: the
// tests assert the layout a non-matching breakpoint yields.
if (typeof window !== 'undefined' && typeof window.matchMedia !== 'function') {
  window.matchMedia = (query: string): MediaQueryList =>
    ({
      matches: false,
      media: query,
      onchange: null,
      addListener: () => {},
      removeListener: () => {},
      addEventListener: () => {},
      removeEventListener: () => {},
      dispatchEvent: () => false,
    }) as MediaQueryList;
}

// jsdom has no ResizeObserver; clicky-ui measures containers with it.
if (typeof globalThis.ResizeObserver === 'undefined') {
  globalThis.ResizeObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
  };
}
