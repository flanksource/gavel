import { useCallback, useEffect, useRef, useState } from 'react';

export type CopyState = 'idle' | 'copying' | 'copied' | 'error';

interface CopyFeedbackOptions {
  copiedMs?: number;
  errorMs?: number;
}

// useCopyFeedback drives the copy-button state machine: idle → copying →
// copied/error, where the terminal copied/error state auto-reverts to idle.
// The pending reset timer is always cleared on unmount so it never fires
// setState after the component is gone.
export function useCopyFeedback({ copiedMs = 2000, errorMs = 3000 }: CopyFeedbackOptions = {}) {
  const [copyState, setCopyState] = useState<CopyState>('idle');
  const [copyError, setCopyError] = useState('');
  const timer = useRef<number | null>(null);

  const clearTimer = useCallback(() => {
    if (timer.current !== null) {
      window.clearTimeout(timer.current);
      timer.current = null;
    }
  }, []);

  useEffect(() => clearTimer, [clearTimer]);

  const beginCopy = useCallback(() => {
    setCopyState('copying');
    setCopyError('');
    clearTimer();
  }, [clearTimer]);

  const resetCopyFeedback = useCallback((next: 'copied' | 'error', error = '') => {
    setCopyState(next);
    setCopyError(error);
    clearTimer();
    timer.current = window.setTimeout(() => {
      timer.current = null;
      setCopyState('idle');
      setCopyError('');
    }, next === 'copied' ? copiedMs : errorMs);
  }, [clearTimer, copiedMs, errorMs]);

  return { copyState, copyError, beginCopy, resetCopyFeedback };
}
