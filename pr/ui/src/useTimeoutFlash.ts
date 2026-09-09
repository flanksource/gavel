import { useCallback, useEffect, useRef, useState } from 'react';

// useTimeoutFlash holds a transient value that auto-reverts to `idle` after
// `ms`, and guarantees the pending timer is cleared on unmount so it never
// fires setState after the component is gone. Returns the current value and a
// `flash` function that sets it and schedules the revert.
export function useTimeoutFlash<T>(idle: T, ms: number): [T, (next: T) => void] {
  const [value, setValue] = useState<T>(idle);
  const timer = useRef<number | null>(null);

  const clearTimer = useCallback(() => {
    if (timer.current !== null) {
      window.clearTimeout(timer.current);
      timer.current = null;
    }
  }, []);

  useEffect(() => clearTimer, [clearTimer]);

  const flash = useCallback((next: T) => {
    setValue(next);
    clearTimer();
    timer.current = window.setTimeout(() => {
      timer.current = null;
      setValue(idle);
    }, ms);
  }, [clearTimer, idle, ms]);

  return [value, flash];
}
