// The heading of a platform dashboard panel: the panel's name, then any figures after a "·",
// one size for every panel.
import type { ReactNode } from 'react';

export function PanelHeading({ children }: { children: ReactNode }) {
  return <h2 className="text-lg font-semibold">{children}</h2>;
}
