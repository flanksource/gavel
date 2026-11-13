import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

// Two anchor boxes connected by a line with a traveling gradient highlight.
// Reduced-motion: the highlight freezes at midpoint. See globals.css @keyframes beam-travel.
interface AnimatedBeamProps {
  fromLabel: string;
  toLabel: string;
  fromIcon?: ReactNode;
  toIcon?: ReactNode;
  className?: string;
}

export default function AnimatedBeam({
  fromLabel,
  toLabel,
  fromIcon,
  toIcon,
  className,
}: AnimatedBeamProps) {
  return (
    <div className={cn("relative flex h-full w-full items-center justify-between px-4", className)}>
      <Anchor label={fromLabel} icon={fromIcon} />

      <svg
        viewBox="0 0 200 20"
        preserveAspectRatio="none"
        aria-hidden
        className="mx-3 h-5 flex-1"
      >
        <line x1="0" y1="10" x2="200" y2="10" strokeDasharray="3 4" className="stroke-border" strokeWidth="1.5" />
        <line x1="0" y1="10" x2="200" y2="10" stroke="url(#beam-gradient)" strokeWidth="2" />
        <defs>
          <linearGradient id="beam-gradient" x1="0" y1="0" x2="1" y2="0">
            <stop offset="0%" stopColor="transparent" />
            <stop offset="50%" stopColor="var(--color-brand-sky)" />
            <stop offset="100%" stopColor="transparent" />
            <animate
              attributeName="x1"
              values="-1;1"
              dur="2.4s"
              repeatCount="indefinite"
              className="motion-reduce:hidden"
            />
            <animate
              attributeName="x2"
              values="0;2"
              dur="2.4s"
              repeatCount="indefinite"
              className="motion-reduce:hidden"
            />
          </linearGradient>
        </defs>
      </svg>

      <Anchor label={toLabel} icon={toIcon} />
    </div>
  );
}

function Anchor({ label, icon }: { label: string; icon?: ReactNode }) {
  return (
    <div className="flex min-w-[5rem] flex-col items-center gap-1 rounded-lg border border-border bg-background px-3 py-2 text-xs font-medium shadow-sm">
      {icon && <span className="text-brand-sky">{icon}</span>}
      <span>{label}</span>
    </div>
  );
}
