import { useSyncExternalStore } from 'react';

// useNow is the single shared 1s clock behind every relative-timestamp leaf
// ('Xs ago', uptime, sync countdown). One module-level interval drives all
// subscribers, so a tick re-renders only the components that call useNow() — the
// <RelativeTime/> leaves — instead of forcing a full-app reconcile from an
// App-level tick (the menubar's dominant main-thread cost). The interval starts
// lazily on the first subscriber and stops when the last unsubscribes.
//
// It is visibility-gated: a hidden window (the menubar webview stays resident
// when dismissed, marking its page hidden) stops ticking, and resyncs the clock
// immediately on show so drifted timestamps correct at once.

let now = Date.now();
const subscribers = new Set<() => void>();
let timer: ReturnType<typeof setInterval> | null = null;
let visibilityBound = false;

function tick() {
  now = Date.now();
  for (const notify of subscribers) notify();
}

function hidden(): boolean {
  return typeof document !== 'undefined' && document.visibilityState === 'hidden';
}

function start() {
  if (timer !== null || subscribers.size === 0 || hidden()) return;
  timer = setInterval(tick, 1000);
}

function stop() {
  if (timer === null) return;
  clearInterval(timer);
  timer = null;
}

function bindVisibility() {
  if (visibilityBound || typeof document === 'undefined') return;
  visibilityBound = true;
  document.addEventListener('visibilitychange', () => {
    if (hidden()) {
      stop();
    } else {
      // Resync immediately on show so a window hidden for a while doesn't display
      // a frozen value for up to a second before the interval fires.
      tick();
      start();
    }
  });
}

function subscribe(notify: () => void): () => void {
  bindVisibility();
  subscribers.add(notify);
  start();
  return () => {
    subscribers.delete(notify);
    if (subscribers.size === 0) stop();
  };
}

function getSnapshot(): number {
  return now;
}

export function useNow(): number {
  return useSyncExternalStore(subscribe, getSnapshot, getSnapshot);
}
