import type { ButtonHTMLAttributes, ReactNode } from "react";
import { cn } from "@/lib/utils";

interface ShimmerButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode;
}

export default function ShimmerButton({ children, className, ...props }: ShimmerButtonProps) {
  return (
    <button
      {...props}
      className={cn(
        "group relative inline-flex h-11 items-center justify-center overflow-hidden rounded-md border border-white/10 bg-brand-dark px-6 font-medium text-white shadow-lg transition-transform hover:scale-[1.02]",
        className,
      )}
    >
      <span
        aria-hidden
        className="pointer-events-none absolute inset-0 -translate-x-full bg-[linear-gradient(90deg,transparent,rgba(255,255,255,0.25),transparent)] transition-transform duration-700 group-hover:translate-x-full"
      />
      <span className="relative z-10">{children}</span>
    </button>
  );
}
