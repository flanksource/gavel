import type { ReactNode } from "react";

export default function Prerequisites({ children }: { children: ReactNode }) {
  return <aside aria-label="Prerequisites" className="my-6 rounded-lg border border-border bg-muted/40 p-4 text-sm [&_p]:my-2 [&_ul]:my-2">
    <p className="font-semibold">Prerequisites</p>
    {children}
  </aside>;
}
