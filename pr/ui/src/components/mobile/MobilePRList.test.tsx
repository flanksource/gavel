import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { MobilePRList } from './MobilePRList';

const origin = { identifier: 1, clientX: 100, clientY: 100 };
const pastThreshold = { ...origin, clientY: 220 };

function renderList(refreshing = false) {
  const onRefresh = vi.fn();
  const props = { prs: [], selected: null, onSelect: vi.fn(), onRefresh, refreshing };
  const view = render(<MobilePRList {...props} />);
  return { ...view, props, onRefresh, list: screen.getByRole('region', { name: 'Pull requests' }) };
}

function pull(list: HTMLElement, point = pastThreshold) {
  fireEvent.touchStart(list, { touches: [origin] });
  fireEvent.touchMove(list, { touches: [point] });
}

describe('mobile PR pull to refresh', () => {
  it('refreshes an empty list once on release after showing pull and release feedback', () => {
    const { list, onRefresh } = renderList();
    expect(screen.getByText('No pull requests found')).toBeTruthy();
    pull(list, { ...origin, clientY: 140 });
    expect(screen.getByRole('status').textContent).toBe('Pull down to refresh');
    fireEvent.touchMove(list, { touches: [pastThreshold] });
    expect(screen.getByRole('status').textContent).toBe('Release to refresh');
    expect(onRefresh).not.toHaveBeenCalled();

    fireEvent.touchEnd(list, { touches: [] });
    fireEvent.touchEnd(list, { touches: [] });

    expect(onRefresh).toHaveBeenCalledTimes(1);
    expect(screen.getByRole('status').textContent).toBe('');
  });

  it.each([
    ['short pull', { ...origin, clientY: 140 }],
    ['horizontal swipe', { ...origin, clientX: 260, clientY: 120 }],
    ['upward swipe', { ...origin, clientY: 40 }],
  ])('leaves a %s alone without refreshing', (_name, point) => {
    const { list, onRefresh } = renderList();
    fireEvent.touchStart(list, { touches: [origin] });
    const move = new TouchEvent('touchmove', { touches: [point as Touch], cancelable: true, bubbles: true });
    fireEvent(list, move);
    fireEvent.touchEnd(list, { touches: [] });

    expect(onRefresh).not.toHaveBeenCalled();
    expect(move.defaultPrevented).toBe(_name === 'short pull');
  });

  it('does not arm when scrolling to the top during an existing gesture', () => {
    const { list, onRefresh } = renderList();
    list.scrollTop = 100;
    fireEvent.touchStart(list, { touches: [origin] });
    list.scrollTop = 0;
    fireEvent.touchMove(list, { touches: [pastThreshold] });
    fireEvent.touchEnd(list, { touches: [] });
    expect(onRefresh).not.toHaveBeenCalled();
  });

  it.each(['touchcancel', 'multitouch', 'horizontal', 'upward', 'native scroll'])('cancels an armed pull on %s', reason => {
    const { list, onRefresh } = renderList();
    pull(list);
    expect(screen.getByRole('status').textContent).toBe('Release to refresh');

    if (reason === 'touchcancel') fireEvent.touchCancel(list);
    else if (reason === 'multitouch') fireEvent.touchStart(list, { touches: [origin, { ...origin, identifier: 2 }] });
    else if (reason === 'horizontal') fireEvent.touchMove(list, { touches: [{ ...pastThreshold, clientX: 300 }] });
    else if (reason === 'upward') fireEvent.touchMove(list, { touches: [{ ...origin, clientY: 40 }] });
    else fireEvent.touchMove(list, { touches: [pastThreshold], cancelable: false });
    fireEvent.touchMove(list, { touches: [pastThreshold] });
    fireEvent.touchEnd(list, { touches: [] });

    expect(onRefresh).not.toHaveBeenCalled();
    expect(screen.getByRole('status').textContent).toBe('');
  });

  it('disarms when the finger returns below the threshold', () => {
    const { list, onRefresh } = renderList();
    pull(list);
    fireEvent.touchMove(list, { touches: [{ ...origin, clientY: 140 }] });
    fireEvent.touchEnd(list, { touches: [] });
    expect(onRefresh).not.toHaveBeenCalled();
  });

  it.each([
    { distance: 95, requests: 0 },
    { distance: 96, requests: 1 },
  ])('requires the full pull distance at $distance pixels', ({ distance, requests }) => {
    const { list, onRefresh } = renderList();
    pull(list, { ...origin, clientY: origin.clientY + distance });
    fireEvent.touchEnd(list, { touches: [] });
    expect(onRefresh).toHaveBeenCalledTimes(requests);
  });

  it('does not refresh if the list scrolls away from the top before release', () => {
    const { list, onRefresh } = renderList();
    pull(list);
    list.scrollTop = 100;
    fireEvent.touchEnd(list, { touches: [] });
    expect(onRefresh).not.toHaveBeenCalled();
  });

  it('cancels a pull if another refresh starts before release', () => {
    const { list, onRefresh, rerender, props } = renderList();
    pull(list);
    rerender(<MobilePRList {...props} refreshing />);
    fireEvent.touchEnd(list, { touches: [] });
    expect(onRefresh).not.toHaveBeenCalled();
  });

  it('shows request progress, blocks more pulls while pending, and allows retry after settling', () => {
    const { list, onRefresh, rerender, props } = renderList(true);
    expect(screen.getByRole('status').textContent).toContain('Requesting refresh');
    pull(list);
    fireEvent.touchEnd(list, { touches: [] });
    expect(onRefresh).not.toHaveBeenCalled();

    rerender(<MobilePRList {...props} refreshing={false} />);
    pull(list);
    fireEvent.touchEnd(list, { touches: [] });
    expect(onRefresh).toHaveBeenCalledTimes(1);
  });

  it('prevents native overscroll only for a downward pull and removes listeners on unmount', () => {
    const { list, onRefresh, unmount } = renderList();
    fireEvent.touchStart(list, { touches: [origin] });
    const move = new TouchEvent('touchmove', { touches: [pastThreshold as Touch], cancelable: true, bubbles: true });
    fireEvent(list, move);
    expect(move.defaultPrevented).toBe(true);
    unmount();
    fireEvent.touchEnd(list, { touches: [] });
    expect(onRefresh).not.toHaveBeenCalled();
  });
});
