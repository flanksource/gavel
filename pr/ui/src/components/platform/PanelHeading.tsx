// The headings of a platform dashboard panel and of its parts: the panel's name, then any figures
// after a "·", one size for every panel.
import type { ReactNode } from 'react';

export function PanelHeading({ children }: { children: ReactNode }) {
  return <h2 className="text-lg font-semibold">{children}</h2>;
}

export function PanelSubheading({ children }: { children: ReactNode }) {
  return <h3 className="text-base font-semibold">{children}</h3>;
}
