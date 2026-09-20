import { afterEach } from 'vitest';

// jsdom implements no media-query engine, so window.matchMedia is simply
// absent. Any component that sizes itself from a breakpoint therefore throws
// "window.matchMedia is not a function" on mount — clicky-ui's Modal does this,
// which takes down every test that renders a dialog.
//
// Report "no match" for every query. The component tests assert the desktop
// layout, which is what a non-matching mobile breakpoint already yields, so a
// stub that always misses keeps them on the branch they mean to exercise.
// Queries seed their first render from a localStorage copy of the last good
// response (see localQueryCache), so storage one test writes would otherwise
// become the next test's placeholder data. Every test starts with it empty.
afterEach(() => {
  if (typeof localStorage !== 'undefined') localStorage.clear();
});

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
