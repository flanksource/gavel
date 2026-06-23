import { useEffect, useState } from 'react';

// Mobile = narrower than Tailwind's `md` breakpoint (768px), matching clicky-ui's
// own mobile detection. On these widths the desktop AppShell (split panes,
// sidebar, toolbars) has no room, so the app falls back to the compact menubar
// dropdown layout instead.
const MOBILE_QUERY = '(max-width: 767px)';

// useIsMobile tracks a viewport-width media query so the App can swap between the
// desktop AppShell and the single-column menubar layout. Defaults to false
// (desktop) when matchMedia is unavailable (e.g. SSR) so the full app renders.
export function useIsMobile(): boolean {
  const [isMobile, setIsMobile] = useState(
    () => typeof window !== 'undefined' && !!window.matchMedia && window.matchMedia(MOBILE_QUERY).matches,
  );

  useEffect(() => {
    if (typeof window === 'undefined' || !window.matchMedia) return;
    const media = window.matchMedia(MOBILE_QUERY);
    const onChange = () => setIsMobile(media.matches);
    onChange();
    media.addEventListener('change', onChange);
    return () => media.removeEventListener('change', onChange);
  }, []);

  return isMobile;
}
