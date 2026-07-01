import { useEffect, useState } from 'react';

// useDocumentVisible tracks the Page Visibility API so background work — the
// per-second re-render tick, the pollers, and the SSE streams — can pause while
// the window is hidden. The menubar webview stays resident when dismissed (it is
// only ordered out, not destroyed), which marks its page hidden; in a browser
// tab it follows the tab's visibility. Defaults to visible when the API is
// unavailable (e.g. SSR), so nothing silently stops fetching.
export function useDocumentVisible(): boolean {
  const [visible, setVisible] = useState(
    () => typeof document === 'undefined' || document.visibilityState !== 'hidden',
  );

  useEffect(() => {
    const onChange = () => setVisible(document.visibilityState !== 'hidden');
    document.addEventListener('visibilitychange', onChange);
    return () => document.removeEventListener('visibilitychange', onChange);
  }, []);

  return visible;
}
