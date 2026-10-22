import type { ReactNode } from 'react';

/** A labelled block of a detail pane, in the label style clicky-ui's own detail bodies use. */
export function Field({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="space-y-1">
      <div className="text-[10px] uppercase tracking-wide text-muted-foreground">{title}</div>
      {children}
    </div>
  );
}
