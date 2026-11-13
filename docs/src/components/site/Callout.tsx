import type { ReactNode } from "react";
import { AlertTriangle, Info, Lightbulb } from "lucide-react";
import { cn } from "@/lib/utils";

type Variant = "info" | "warn" | "tip";

const variants: Record<Variant, { icon: ReactNode; className: string }> = {
  info: {
    icon: <Info size={16} />,
    className: "border-sky-400/40 bg-sky-400/5 text-sky-300",
  },
  warn: {
    icon: <AlertTriangle size={16} />,
    className: "border-amber-400/40 bg-amber-400/5 text-amber-300",
  },
  tip: {
    icon: <Lightbulb size={16} />,
    className: "border-emerald-400/40 bg-emerald-400/5 text-emerald-300",
  },
};

interface CalloutProps {
  children: ReactNode;
  variant?: Variant;
  title?: string;
}

export default function Callout({ children, variant = "info", title }: CalloutProps) {
  const v = variants[variant];
  return (
    <div className={cn("my-6 flex gap-3 rounded-lg border p-4", v.className)}>
      <span className="mt-0.5 shrink-0">{v.icon}</span>
      <div className="text-sm text-foreground/90">
        {title && <p className="mb-1 font-semibold">{title}</p>}
        {children}
      </div>
    </div>
  );
}
