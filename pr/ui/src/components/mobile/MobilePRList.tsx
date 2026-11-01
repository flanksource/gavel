import { useEffect, useRef, useState, type ComponentProps } from 'react';
import { Spinner } from '../../icons/Spinner';
import { PRList } from '../PRList';

type Props = ComponentProps<typeof PRList> & {
  onRefresh: () => void;
  refreshing: boolean;
};

const REFRESH_DISTANCE = 96;
const DIRECTION_SLOP = 8;

export function MobilePRList({ onRefresh, refreshing, ...props }: Props) {
  const ref = useRef<HTMLDivElement>(null);
  const [distance, setDistance] = useState(0);

  useEffect(() => {
    const element = ref.current;
    if (!element) return;
    return listenForPull({ element, refreshing, onRefresh, onDistance: setDistance });
  }, [onRefresh, refreshing]);

  return (
    <div ref={ref} role="region" aria-label="Pull requests" aria-busy={refreshing} className="h-full overflow-y-auto overscroll-y-contain">
      <div role="status" className="flex items-center justify-center gap-2 overflow-hidden text-xs text-muted-foreground" style={{ height: refreshing ? 48 : Math.min(distance / 2, 64) }}>
        {refreshing ? <><Spinner />Requesting refresh…</> : distance > 0 ? (distance >= REFRESH_DISTANCE ? 'Release to refresh' : 'Pull down to refresh') : null}
      </div>
      <PRList {...props} />
    </div>
  );
}

function listenForPull({ element, refreshing, onRefresh, onDistance }: {
  element: HTMLElement;
  refreshing: boolean;
  onRefresh: () => void;
  onDistance: (distance: number) => void;
}) {
  let start: Touch | undefined;
  let distance = 0;
  function reset() {
    start = undefined;
    distance = 0;
    onDistance(0);
  }
  function onStart(event: TouchEvent) {
    reset();
    if (!refreshing && element.scrollTop <= 0 && event.touches.length === 1) start = event.touches[0];
  }
  function onMove(event: TouchEvent) {
    if (!start) return;
    const touch = event.touches[0];
    if (event.touches.length !== 1 || touch.identifier !== start.identifier || element.scrollTop > 0 || !event.cancelable) return reset();
    const dy = touch.clientY - start.clientY;
    const dx = Math.abs(touch.clientX - start.clientX);
    if (dy < -DIRECTION_SLOP || (dx > DIRECTION_SLOP && dx > dy)) return reset();
    if (dy <= DIRECTION_SLOP && distance === 0) return;
    // React's touch listeners are passive; this native listener must cancel
    // overscroll before the browser takes ownership of a downward gesture.
    event.preventDefault();
    distance = Math.max(0, dy);
    onDistance(distance);
  }
  function onEnd() {
    const ready = distance >= REFRESH_DISTANCE && element.scrollTop <= 0;
    reset();
    if (ready) onRefresh();
  }

  reset();
  element.addEventListener('touchstart', onStart, { passive: true });
  element.addEventListener('touchmove', onMove, { passive: false });
  element.addEventListener('touchend', onEnd);
  element.addEventListener('touchcancel', reset);
  return () => {
    element.removeEventListener('touchstart', onStart);
    element.removeEventListener('touchmove', onMove);
    element.removeEventListener('touchend', onEnd);
    element.removeEventListener('touchcancel', reset);
  };
}
