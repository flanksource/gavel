import type { ReactNode } from "react";
import type { LucideIcon } from "lucide-react";
import { cn } from "@/lib/utils";

// Ring of icons orbiting a central slot. Pure CSS — outer ring rotates, inner icons counter-rotate
// so their glyphs stay upright. Reduced-motion: ring freezes at 12 o'clock.
interface IconOrbitProps {
  icons: LucideIcon[];
  center?: ReactNode;
  radius?: number;
  className?: string;
}

export default function IconOrbit({
  icons,
  center,
  radius = 72,
  className,
}: IconOrbitProps) {
  const step = 360 / Math.max(icons.length, 1);

  return (
    <div className={cn("relative flex h-full w-full items-center justify-center", className)}>
      <div className="relative" style={{ width: radius * 2, height: radius * 2 }}>
        <div className="absolute inset-0 motion-safe:[animation:orbit_30s_linear_infinite]">
          {icons.map((Icon, i) => {
            const angle = i * step;
            return (
              <span
                key={i}
                className="absolute left-1/2 top-1/2 flex h-9 w-9 -translate-x-1/2 -translate-y-1/2 items-center justify-center rounded-full border border-border bg-background text-muted-foreground shadow-sm motion-safe:[animation:orbit_30s_linear_infinite_reverse]"
                style={{
                  transform: `rotate(${angle}deg) translate(${radius}px) rotate(-${angle}deg)`,
                }}
              >
                <Icon size={16} />
              </span>
            );
          })}
        </div>
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="flex h-14 w-14 items-center justify-center rounded-full bg-brand-dark text-white shadow-md ring-4 ring-background">
            {center ?? <span className="font-mono text-xs font-semibold">gavel</span>}
          </div>
        </div>
      </div>
    </div>
  );
}
