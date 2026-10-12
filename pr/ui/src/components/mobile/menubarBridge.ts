import { useEffect } from 'react';

type WebKitExternalBridge = {
  webkit?: {
    messageHandlers?: {
      external?: {
        postMessage: (message: string) => void;
      };
    };
  };
};

export const menubarOpenExternalMessage = 'gavel:open-external';
export const menubarPointerEnterMessage = 'gavel:pointer-enter';
export const menubarPointerLeaveMessage = 'gavel:pointer-leave';

export function postMenubarMessage(type: string, payload: Record<string, unknown> = {}) {
  const bridge = (window as WebKitExternalBridge).webkit?.messageHandlers?.external;
  if (!bridge) return false;
  bridge.postMessage(JSON.stringify({ type, ...payload }));
  return true;
}

export function useMenubarExternalLinks() {
  useEffect(() => {
    const onClick = (event: MouseEvent) => {
      if (event.defaultPrevented || event.button !== 0) return;
      const target = event.target;
      if (!(target instanceof Element)) return;
      const anchor = target.closest('a[href]');
      if (!(anchor instanceof HTMLAnchorElement) || !anchor.href || anchor.hasAttribute('data-app-link')) return;

      if (!postMenubarMessage(menubarOpenExternalMessage, { url: anchor.href })) return;

      event.preventDefault();
      event.stopPropagation();
      event.stopImmediatePropagation();
    };

    document.addEventListener('click', onClick, true);
    return () => document.removeEventListener('click', onClick, true);
  }, []);
}
